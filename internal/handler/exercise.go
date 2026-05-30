package handler

import (
	"fmt"
	"net/http"
	"strings"
	"workout-tracker/internal/middleware"
	"workout-tracker/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ExerciseHandler struct {
	exercises *repository.ExerciseRepository
}

func NewExerciseHandler(exercises *repository.ExerciseRepository) *ExerciseHandler {
	return &ExerciseHandler{exercises: exercises}
}

func (h *ExerciseHandler) requireTrainerOrAdmin(w http.ResponseWriter, r *http.Request) bool {
	user := middleware.UserFromContext(r.Context())
	if user.IsTrainer() || user.IsAdmin {
		return true
	}
	http.Error(w, "Нет доступа", http.StatusForbidden)
	return false
}

func (h *ExerciseHandler) List(w http.ResponseWriter, r *http.Request) {
	if !h.requireTrainerOrAdmin(w, r) {
		return
	}
	list, err := h.exercises.List(r.Context())
	if err != nil {
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}
	renderTemplate(w, r, "exercises/list.html", map[string]any{
		"Exercises": list,
	})
}

func (h *ExerciseHandler) NewForm(w http.ResponseWriter, r *http.Request) {
	if !h.requireTrainerOrAdmin(w, r) {
		return
	}
	renderTemplate(w, r, "exercises/form.html", map[string]any{
		"MuscleGroups": muscleGroups,
	})
}

func (h *ExerciseHandler) Create(w http.ResponseWriter, r *http.Request) {
	if !h.requireTrainerOrAdmin(w, r) {
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ошибка формы", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		renderTemplate(w, r, "exercises/form.html", map[string]any{
			"Error":        "Название обязательно",
			"MuscleGroups": muscleGroups,
		})
		return
	}
	_, err := h.exercises.Create(r.Context(), name, r.FormValue("muscle_group"), r.FormValue("description"))
	if err != nil {
		renderTemplate(w, r, "exercises/form.html", map[string]any{
			"Error":        "Ошибка сохранения: " + err.Error(),
			"MuscleGroups": muscleGroups,
		})
		return
	}
	http.Redirect(w, r, "/exercises", http.StatusSeeOther)
}

func (h *ExerciseHandler) EditForm(w http.ResponseWriter, r *http.Request) {
	if !h.requireTrainerOrAdmin(w, r) {
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	ex, err := h.exercises.GetByID(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	renderTemplate(w, r, "exercises/form.html", map[string]any{
		"Exercise":     ex,
		"MuscleGroups": muscleGroups,
	})
}

func (h *ExerciseHandler) Update(w http.ResponseWriter, r *http.Request) {
	if !h.requireTrainerOrAdmin(w, r) {
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ошибка формы", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		ex, _ := h.exercises.GetByID(r.Context(), id)
		renderTemplate(w, r, "exercises/form.html", map[string]any{
			"Error":        "Название обязательно",
			"Exercise":     ex,
			"MuscleGroups": muscleGroups,
		})
		return
	}
	if err := h.exercises.Update(r.Context(), id, name, r.FormValue("muscle_group"), r.FormValue("description")); err != nil {
		ex, _ := h.exercises.GetByID(r.Context(), id)
		renderTemplate(w, r, "exercises/form.html", map[string]any{
			"Error":        "Ошибка сохранения: " + err.Error(),
			"Exercise":     ex,
			"MuscleGroups": muscleGroups,
		})
		return
	}
	http.Redirect(w, r, "/exercises", http.StatusSeeOther)
}

func (h *ExerciseHandler) Progress(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	ex, err := h.exercises.GetByID(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	sessions, err := h.exercises.GetProgress(r.Context(), ex.Name, user.ID, 20)
	if err != nil {
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}
	renderTemplate(w, r, "exercises/progress.html", map[string]any{
		"Exercise": ex,
		"Sessions": sessions,
	})
}

func (h *ExerciseHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if !h.requireTrainerOrAdmin(w, r) {
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := h.exercises.Delete(r.Context(), id); err != nil {
		http.Error(w, "Ошибка удаления", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/exercises"), http.StatusSeeOther)
}

var muscleGroups = []string{
	"Грудь",
	"Спина",
	"Плечи",
	"Бицепс",
	"Трицепс",
	"Ноги",
	"Ягодицы",
	"Пресс",
	"Кардио",
	"Другое",
}
