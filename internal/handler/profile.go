package handler

import (
	"net/http"
	"workout-tracker/internal/middleware"
	"workout-tracker/internal/repository"
)

type ProfileHandler struct {
	workouts  *repository.WorkoutRepository
	gyms      *repository.GymRepository
	tc        *repository.TrainerClientRepository
	templates *repository.TemplateRepository
}

func NewProfileHandler(
	workouts *repository.WorkoutRepository,
	gyms *repository.GymRepository,
	tc *repository.TrainerClientRepository,
	templates *repository.TemplateRepository,
) *ProfileHandler {
	return &ProfileHandler{workouts: workouts, gyms: gyms, tc: tc, templates: templates}
}

func (h *ProfileHandler) Show(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	stats, _ := h.workouts.GetDashboardStats(r.Context(), user.ID)
	cards, _ := h.workouts.ListCards(r.Context(), user.ID)

	var totalTonnage float64
	for _, c := range cards {
		totalTonnage += c.Tonnage
	}

	gymList, _ := h.gyms.List(r.Context())

	data := map[string]any{
		"Stats":        stats,
		"TotalCount":   len(cards),
		"TotalTonnage": totalTonnage,
		"GymCount":     len(gymList),
	}

	if user.IsTrainer() {
		clients, _ := h.tc.GetClients(r.Context(), user.ID)
		data["ClientCount"] = len(clients)

		templateList, _ := h.templates.List(r.Context(), user.ID)
		data["TemplateCount"] = len(templateList)
	}

	renderTemplate(w, r, "profile.html", data)
}
