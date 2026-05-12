package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
	"workout-tracker/internal/middleware"
	"workout-tracker/internal/model"
	"workout-tracker/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type WorkoutHandler struct {
	workouts *repository.WorkoutRepository
	gyms     *repository.GymRepository
}

func NewWorkoutHandler(workouts *repository.WorkoutRepository, gyms *repository.GymRepository) *WorkoutHandler {
	return &WorkoutHandler{workouts: workouts, gyms: gyms}
}

func (h *WorkoutHandler) List(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	list, err := h.workouts.List(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}
	renderTemplate(w, r, "workouts/list.html", map[string]any{
		"Workouts": list,
	})
}

func (h *WorkoutHandler) NewForm(w http.ResponseWriter, r *http.Request) {
	gyms, _ := h.gyms.List(r.Context())
	renderTemplate(w, r, "workouts/form.html", map[string]any{
		"Gyms":  gyms,
		"Today": time.Now().Format("2006-01-02"),
	})
}

func (h *WorkoutHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ошибка формы", http.StatusBadRequest)
		return
	}

	wo := model.Workout{
		Title:       r.FormValue("title"),
		Notes:       r.FormValue("notes"),
		WorkoutDate: parseDate(r.FormValue("workout_date")),
		GymID:       parseUUIDPtr(r.FormValue("gym_id")),
	}
	if wb := r.FormValue("wellbeing"); wb != "" {
		if v, err := strconv.Atoi(wb); err == nil {
			wo.Wellbeing = &v
		}
	}

	exercises := parseExercisesFromForm(r)

	workout, err := h.workouts.Create(r.Context(), user.ID, wo, exercises)
	if err != nil {
		gyms, _ := h.gyms.List(r.Context())
		renderTemplate(w, r, "workouts/form.html", map[string]any{
			"Error": "Ошибка сохранения: " + err.Error(),
			"Gyms":  gyms,
			"Today": time.Now().Format("2006-01-02"),
		})
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/workouts/%s", workout.ID), http.StatusSeeOther)
}

func (h *WorkoutHandler) Show(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	workout, err := h.workouts.GetByID(r.Context(), id, user.ID)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	renderTemplate(w, r, "workouts/show.html", map[string]any{
		"Workout": workout,
	})
}

func (h *WorkoutHandler) EditForm(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	workout, err := h.workouts.GetByID(r.Context(), id, user.ID)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	gyms, _ := h.gyms.List(r.Context())
	renderTemplate(w, r, "workouts/form.html", map[string]any{
		"Workout": workout,
		"Gyms":    gyms,
		"Today":   time.Now().Format("2006-01-02"),
	})
}

func (h *WorkoutHandler) Update(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ошибка формы", http.StatusBadRequest)
		return
	}

	wo := model.Workout{
		Title:       r.FormValue("title"),
		Notes:       r.FormValue("notes"),
		WorkoutDate: parseDate(r.FormValue("workout_date")),
		GymID:       parseUUIDPtr(r.FormValue("gym_id")),
	}
	if wb := r.FormValue("wellbeing"); wb != "" {
		if v, err := strconv.Atoi(wb); err == nil {
			wo.Wellbeing = &v
		}
	}

	exercises := parseExercisesFromForm(r)

	if err := h.workouts.Update(r.Context(), id, user.ID, wo, exercises); err != nil {
		http.Error(w, "Ошибка обновления", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/workouts/%s", id), http.StatusSeeOther)
}

func (h *WorkoutHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	_ = h.workouts.Delete(r.Context(), id, user.ID)
	http.Redirect(w, r, "/workouts", http.StatusSeeOther)
}

// AddExerciseRow возвращает HTMX-партиал: пустую строку нового упражнения
func (h *WorkoutHandler) AddExerciseRow(w http.ResponseWriter, r *http.Request) {
	idx, _ := strconv.Atoi(r.URL.Query().Get("idx"))
	renderPartial(w, r, "workouts/partials/exercise_row.html", map[string]any{
		"ExIdx": idx,
		"Ex":    model.FormExercise{Sets: []model.FormSet{{}}},
	})
}

// AddSetRow возвращает HTMX-партиал: пустую строку нового подхода
func (h *WorkoutHandler) AddSetRow(w http.ResponseWriter, r *http.Request) {
	exIdx, _ := strconv.Atoi(r.URL.Query().Get("ex_idx"))
	setIdx, _ := strconv.Atoi(r.URL.Query().Get("set_idx"))
	renderPartial(w, r, "workouts/partials/set_row.html", map[string]any{
		"ExIdx":  exIdx,
		"SetIdx": setIdx,
	})
}

// parseExercisesFromForm парсит вложенную структуру exercises[N][field] из формы
func parseExercisesFromForm(r *http.Request) []model.FormExercise {
	var exercises []model.FormExercise
	for i := 0; ; i++ {
		name := r.FormValue(fmt.Sprintf("exercises[%d][name]", i))
		if name == "" && i > 0 {
			// проверяем есть ли ещё упражнения дальше
			found := false
			for j := i + 1; j < i+5; j++ {
				if r.FormValue(fmt.Sprintf("exercises[%d][name]", j)) != "" {
					found = true
					break
				}
			}
			if !found {
				break
			}
		}
		ex := model.FormExercise{
			Name:  name,
			Notes: r.FormValue(fmt.Sprintf("exercises[%d][notes]", i)),
		}
		for j := 0; ; j++ {
			w := r.FormValue(fmt.Sprintf("exercises[%d][sets][%d][weight]", i, j))
			reps := r.FormValue(fmt.Sprintf("exercises[%d][sets][%d][reps]", i, j))
			rpe := r.FormValue(fmt.Sprintf("exercises[%d][sets][%d][rpe]", i, j))
			rest := r.FormValue(fmt.Sprintf("exercises[%d][sets][%d][rest]", i, j))
			notes := r.FormValue(fmt.Sprintf("exercises[%d][sets][%d][notes]", i, j))
			if w == "" && reps == "" && j > 0 {
				break
			}
			if w != "" || reps != "" {
				ex.Sets = append(ex.Sets, model.FormSet{
					Weight: w, Reps: reps, RPE: rpe, RestSeconds: rest, Notes: notes,
				})
			}
			if j > 50 {
				break
			}
		}
		exercises = append(exercises, ex)
		if i > 50 {
			break
		}
	}
	return exercises
}

func parseDate(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Now()
	}
	return t
}

func parseUUIDPtr(s string) *uuid.UUID {
	if s == "" {
		return nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return nil
	}
	return &id
}
