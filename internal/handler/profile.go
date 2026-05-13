package handler

import (
	"net/http"
	"workout-tracker/internal/middleware"
	"workout-tracker/internal/repository"
)

type ProfileHandler struct {
	workouts *repository.WorkoutRepository
}

func NewProfileHandler(workouts *repository.WorkoutRepository) *ProfileHandler {
	return &ProfileHandler{workouts: workouts}
}

func (h *ProfileHandler) Show(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	stats, _ := h.workouts.GetDashboardStats(r.Context(), user.ID)
	cards, _ := h.workouts.ListCards(r.Context(), user.ID)

	var totalTonnage float64
	for _, c := range cards {
		totalTonnage += c.Tonnage
	}

	renderTemplate(w, r, "profile.html", map[string]any{
		"Stats":        stats,
		"TotalCount":   len(cards),
		"TotalTonnage": totalTonnage,
	})
}
