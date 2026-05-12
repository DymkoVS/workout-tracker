package handler

import (
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

// Clients показывает список клиентов тренера
func (h *TrainerHandler) Clients(w http.ResponseWriter, r *http.Request) {
	trainer := middleware.UserFromContext(r.Context())
	clients, err := h.tc.GetClients(r.Context(), trainer.ID)
	if err != nil {
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}
	renderTemplate(w, r, "trainer/clients.html", map[string]any{
		"Clients": clients,
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

	workouts, err := h.workouts.List(r.Context(), clientID)
	if err != nil {
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	renderTemplate(w, r, "trainer/client_workouts.html", map[string]any{
		"Client":   client,
		"Workouts": workouts,
	})
}
