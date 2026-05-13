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
}

func NewAnalyticsHandler(analytics *repository.AnalyticsRepository, tc *repository.TrainerClientRepository) *AnalyticsHandler {
	return &AnalyticsHandler{analytics: analytics, tc: tc}
}

func (h *AnalyticsHandler) Index(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	targetUserID := user.ID
	selectedClientID := ""

	if user.IsTrainer() {
		if rawID := r.URL.Query().Get("user_id"); rawID != "" {
			if uid, err := uuid.Parse(rawID); err == nil {
				if ok, _ := h.tc.IsAssigned(r.Context(), user.ID, uid); ok {
					targetUserID = uid
					selectedClientID = rawID
				}
			}
		}
	}

	tonnage, _ := h.analytics.TonnageByDate(r.Context(), targetUserID)
	frequency, _ := h.analytics.WorkoutFrequency(r.Context(), targetUserID)
	exercises, _ := h.analytics.ExerciseNames(r.Context(), targetUserID)
	cur90, prev90, _ := h.analytics.TonnagePeriodTotals(r.Context(), targetUserID)

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

	exerciseName := r.URL.Query().Get("name")
	w.Header().Set("Content-Type", "application/json")

	if exerciseName == "" {
		w.Write([]byte(`{"labels":[],"values":[]}`))
		return
	}

	points, _ := h.analytics.ExerciseProgress(r.Context(), targetUserID, exerciseName)

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
