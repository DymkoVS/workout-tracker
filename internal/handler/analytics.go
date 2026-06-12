package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"workout-tracker/internal/middleware"
	"workout-tracker/internal/repository"

	"github.com/google/uuid"
)

type AnalyticsHandler struct {
	analytics *repository.AnalyticsRepository
	tc        *repository.TrainerClientRepository
	gyms      *repository.GymRepository
}

func NewAnalyticsHandler(analytics *repository.AnalyticsRepository, tc *repository.TrainerClientRepository, gyms *repository.GymRepository) *AnalyticsHandler {
	return &AnalyticsHandler{analytics: analytics, tc: tc, gyms: gyms}
}

func (h *AnalyticsHandler) Index(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	targetUserID := user.ID
	selectedClientID := ""

	q := r.URL.Query()

	if user.IsTrainer() {
		if rawID := q.Get("user_id"); rawID != "" {
			if uid, err := uuid.Parse(rawID); err == nil {
				if ok, _ := h.tc.IsAssigned(r.Context(), user.ID, uid); ok {
					targetUserID = uid
					selectedClientID = rawID
				}
			}
		}
	}

	filter := repository.AnalyticsFilter{WorkoutType: q.Get("type")}
	filterGymID := q.Get("gym_id")
	if filterGymID != "" {
		if gid, err := uuid.Parse(filterGymID); err == nil {
			filter.GymID = &gid
		}
	}

	tonnage, err := h.analytics.TonnageByDate(r.Context(), targetUserID, filter)
	logErr("analytics: tonnage", err)
	frequency, err := h.analytics.WorkoutFrequency(r.Context(), targetUserID, filter)
	logErr("analytics: frequency", err)
	exercises, err := h.analytics.ExerciseNames(r.Context(), targetUserID, filter)
	logErr("analytics: exercise names", err)
	cur90, prev90, err := h.analytics.TonnagePeriodTotals(r.Context(), targetUserID, filter)
	logErr("analytics: period totals", err)

	var tonnageDelta string
	if prev90 > 0 {
		pct := (cur90 - prev90) / prev90 * 100
		if pct >= 0 {
			tonnageDelta = fmt.Sprintf("▲ +%.0f%%", pct)
		} else {
			tonnageDelta = fmt.Sprintf("▼ %.0f%%", pct)
		}
	}

	tonnageLabels := make([]string, len(tonnage))
	tonnageValues := make([]float64, len(tonnage))
	for i, p := range tonnage {
		tonnageLabels[i] = p.Date.Format("02.01")
		tonnageValues[i] = p.Value
	}

	freqLabels := make([]string, len(frequency))
	freqValues := make([]int, len(frequency))
	for i, p := range frequency {
		freqLabels[i] = p.Week
		freqValues[i] = p.Count
	}

	tlJSON, _ := json.Marshal(tonnageLabels)
	tvJSON, _ := json.Marshal(tonnageValues)
	flJSON, _ := json.Marshal(freqLabels)
	fvJSON, _ := json.Marshal(freqValues)

	gymList, err := h.gyms.List(r.Context())
	logErr("analytics: gyms", err)

	data := map[string]any{
		"TonnageLabels": template.JS(string(tlJSON)),
		"TonnageValues": template.JS(string(tvJSON)),
		"FreqLabels":    template.JS(string(flJSON)),
		"FreqValues":    template.JS(string(fvJSON)),
		"TonnageCount":  len(tonnage),
		"FreqCount":     len(frequency),
		"Exercises":     exercises,
		"TargetUserID":  targetUserID.String(),
		"TonnageDelta":  tonnageDelta,
		"Gyms":          gymList,
		"FilterGymID":   filterGymID,
		"FilterType":    q.Get("type"),
		"FilterActive":  filterGymID != "" || q.Get("type") != "",
	}

	if user.IsTrainer() {
		clients, _ := h.tc.GetClients(r.Context(), user.ID)
		data["Clients"] = clients
		data["SelectedClientID"] = selectedClientID
	}

	renderTemplate(w, r, "analytics/index.html", data)
}

func (h *AnalyticsHandler) ExerciseData(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	targetUserID := user.ID
	if user.IsTrainer() {
		if rawID := r.URL.Query().Get("user_id"); rawID != "" {
			if uid, err := uuid.Parse(rawID); err == nil {
				if ok, _ := h.tc.IsAssigned(r.Context(), user.ID, uid); ok {
					targetUserID = uid
				}
			}
		}
	}

	q := r.URL.Query()
	exerciseName := q.Get("name")
	w.Header().Set("Content-Type", "application/json")

	if exerciseName == "" {
		w.Write([]byte(`{"labels":[],"values":[]}`))
		return
	}

	filter := repository.AnalyticsFilter{WorkoutType: q.Get("type")}
	if gid, err := uuid.Parse(q.Get("gym_id")); err == nil {
		filter.GymID = &gid
	}

	points, _ := h.analytics.ExerciseProgress(r.Context(), targetUserID, exerciseName, filter)

	labels := make([]string, len(points))
	values := make([]float64, len(points))
	for i, p := range points {
		labels[i] = p.Date.Format("02.01")
		values[i] = p.Value
	}

	json.NewEncoder(w).Encode(map[string]any{
		"labels": labels,
		"values": values,
	})
}
