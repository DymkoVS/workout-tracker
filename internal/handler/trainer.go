package handler

import (
	"fmt"
	"net/http"
	"workout-tracker/internal/middleware"
	"workout-tracker/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type TrainerHandler struct {
	tc       *repository.TrainerClientRepository
	workouts *repository.WorkoutRepository
	users    *repository.UserRepository
}

func NewTrainerHandler(tc *repository.TrainerClientRepository, workouts *repository.WorkoutRepository, users *repository.UserRepository) *TrainerHandler {
	return &TrainerHandler{tc: tc, workouts: workouts, users: users}
}

// Clients показывает список клиентов тренера с недельной статистикой.
func (h *TrainerHandler) Clients(w http.ResponseWriter, r *http.Request) {
	trainer := middleware.UserFromContext(r.Context())
	stats, err := h.tc.GetClientStats(r.Context(), trainer.ID)
	if err != nil {
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	activeCount := 0
	weekDoneTotal, weekPlanTotal, prevWeekDoneTotal := 0, 0, 0
	for _, cs := range stats {
		if cs.IsActive {
			activeCount++
		}
		weekDoneTotal += cs.WeekDone
		weekPlanTotal += cs.WeekPlan
		prevWeekDoneTotal += cs.PrevWeekDone
	}

	weekPulsePct := ""
	weekPulseRatio := ""
	weekPulseDelta := ""
	weekPulseDeltaNeg := false

	if weekPlanTotal > 0 {
		pct := float64(weekDoneTotal) / float64(weekPlanTotal) * 100
		weekPulsePct = fmt.Sprintf("%.0f%%", pct)
		weekPulseRatio = fmt.Sprintf("%d/%d трен.", weekDoneTotal, weekPlanTotal)
		prevPct := float64(prevWeekDoneTotal) / float64(weekPlanTotal) * 100
		delta := pct - prevPct
		if delta > 0.5 {
			weekPulseDelta = fmt.Sprintf("▲ +%.0f%%", delta)
		} else if delta < -0.5 {
			weekPulseDelta = fmt.Sprintf("▼ %.0f%%", delta)
			weekPulseDeltaNeg = true
		}
	}

	renderTemplate(w, r, "trainer/clients.html", map[string]any{
		"ClientStats":       stats,
		"ActiveCount":       activeCount,
		"WeekPulsePct":      weekPulsePct,
		"WeekPulseRatio":    weekPulseRatio,
		"WeekPulseDelta":    weekPulseDelta,
		"WeekPulseDeltaNeg": weekPulseDeltaNeg,
	})
}

// ClientWorkouts показывает тренировки клиента (только для назначенного тренера)
func (h *TrainerHandler) ClientWorkouts(w http.ResponseWriter, r *http.Request) {
	trainer := middleware.UserFromContext(r.Context())
	clientID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	ok, err := h.tc.IsAssigned(r.Context(), trainer.ID, clientID)
	if err != nil || !ok {
		http.Error(w, "Нет доступа", http.StatusForbidden)
		return
	}

	client, err := h.users.GetByID(r.Context(), clientID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	cards, err := h.workouts.ListCards(r.Context(), clientID)
	if err != nil {
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	renderTemplate(w, r, "trainer/client_workouts.html", map[string]any{
		"Client":  client,
		"Cards":   cards,
		"BackURL": "/trainer/clients",
	})
}
