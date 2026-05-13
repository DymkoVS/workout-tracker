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
	tc       *repository.TrainerClientRepository
	users    *repository.UserRepository
}

func NewWorkoutHandler(
	workouts *repository.WorkoutRepository,
	gyms *repository.GymRepository,
	tc *repository.TrainerClientRepository,
	users *repository.UserRepository,
) *WorkoutHandler {
	return &WorkoutHandler{workouts: workouts, gyms: gyms, tc: tc, users: users}
}

// WorkoutGroup groups workout cards under a month label for the history list.
type WorkoutGroup struct {
	MonthLabel string
	Cards      []model.WorkoutCardData
}

var ruMonthsFull = [...]string{"", "ЯНВАРЬ", "ФЕВРАЛЬ", "МАРТ", "АПРЕЛЬ", "МАЙ", "ИЮНЬ", "ИЮЛЬ", "АВГУСТ", "СЕНТЯБРЬ", "ОКТЯБРЬ", "НОЯБРЬ", "ДЕКАБРЬ"}

func groupByMonth(cards []model.WorkoutCardData) []WorkoutGroup {
	var groups []WorkoutGroup
	for _, c := range cards {
		label := fmt.Sprintf("%s %d", ruMonthsFull[c.WorkoutDate.Month()], c.WorkoutDate.Year())
		if len(groups) == 0 || groups[len(groups)-1].MonthLabel != label {
			groups = append(groups, WorkoutGroup{MonthLabel: label})
		}
		groups[len(groups)-1].Cards = append(groups[len(groups)-1].Cards, c)
	}
	return groups
}

func (h *WorkoutHandler) List(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	cards, err := h.workouts.ListCards(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}
	renderTemplate(w, r, "workouts/list.html", map[string]any{
		"WorkoutGroups": groupByMonth(cards),
		"TotalCount":    len(cards),
	})
}

func (h *WorkoutHandler) NewForm(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	gyms, _ := h.gyms.List(r.Context())
	data := map[string]any{
		"Gyms":  gyms,
		"Today": time.Now().Format("02.01.2006"),
	}
	if forClientStr := r.URL.Query().Get("for_client"); forClientStr != "" && user.IsTrainer() {
		if clientID, err := uuid.Parse(forClientStr); err == nil {
			if ok, _ := h.tc.IsAssigned(r.Context(), user.ID, clientID); ok {
				if client, err := h.users.GetByID(r.Context(), clientID); err == nil {
					data["ForClient"] = client
				}
			}
		}
	}
	renderTemplate(w, r, "workouts/form.html", data)
}

func (h *WorkoutHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ошибка формы", http.StatusBadRequest)
		return
	}

	targetUserID := user.ID
	var trainerID *uuid.UUID
	var redirectAfter string

	if forClientStr := r.FormValue("for_client_id"); forClientStr != "" && user.IsTrainer() {
		if clientID, err := uuid.Parse(forClientStr); err == nil {
			if ok, _ := h.tc.IsAssigned(r.Context(), user.ID, clientID); ok {
				targetUserID = clientID
				trainerID = &user.ID
				redirectAfter = fmt.Sprintf("/trainer/clients/%s/workouts", clientID)
			}
		}
	}

	wo := model.Workout{
		Title:       r.FormValue("title"),
		Notes:       r.FormValue("notes"),
		WorkoutDate: parseDate(r.FormValue("workout_date")),
		GymID:       parseUUIDPtr(r.FormValue("gym_id")),
		TrainerID:   trainerID,
	}
	if wb := r.FormValue("wellbeing"); wb != "" {
		if v, err := strconv.Atoi(wb); err == nil {
			wo.Wellbeing = &v
		}
	}

	exercises := parseExercisesFromForm(r)
	workout, err := h.workouts.Create(r.Context(), targetUserID, wo, exercises)
	if err != nil {
		gyms, _ := h.gyms.List(r.Context())
		renderTemplate(w, r, "workouts/form.html", map[string]any{
			"Error": "Ошибка сохранения: " + err.Error(),
			"Gyms":  gyms,
			"Today": time.Now().Format("02.01.2006"),
		})
		return
	}

	if redirectAfter != "" {
		http.Redirect(w, r, redirectAfter, http.StatusSeeOther)
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
	if err != nil && user.IsTrainer() {
		workout, err = h.workouts.GetByIDForTrainer(r.Context(), id, user.ID)
	}
	if err != nil {
		http.NotFound(w, r)
		return
	}

	data := map[string]any{"Workout": workout}
	if workout.UserID != user.ID && user.IsTrainer() {
		data["BackURL"] = fmt.Sprintf("/trainer/clients/%s/workouts", workout.UserID)
		data["CanEdit"] = workout.TrainerID != nil && *workout.TrainerID == user.ID
	} else {
		data["CanEdit"] = true
	}
	renderTemplate(w, r, "workouts/show.html", data)
}

func (h *WorkoutHandler) EditForm(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	workout, err := h.workouts.GetByID(r.Context(), id, user.ID)
	isTrainerEdit := false
	if err != nil && user.IsTrainer() {
		workout, err = h.workouts.GetByIDForTrainer(r.Context(), id, user.ID)
		if err == nil {
			if workout.TrainerID == nil || *workout.TrainerID != user.ID {
				http.Error(w, "Нет доступа", http.StatusForbidden)
				return
			}
			isTrainerEdit = true
		}
	}
	if err != nil {
		http.NotFound(w, r)
		return
	}

	gyms, _ := h.gyms.List(r.Context())
	data := map[string]any{
		"Workout": workout,
		"Gyms":    gyms,
		"Today":   time.Now().Format("02.01.2006"),
	}
	if isTrainerEdit {
		if client, err := h.users.GetByID(r.Context(), workout.UserID); err == nil {
			data["ForClient"] = client
		}
	}
	renderTemplate(w, r, "workouts/form.html", data)
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

	updateErr := h.workouts.Update(r.Context(), id, user.ID, wo, exercises)
	if updateErr != nil && user.IsTrainer() {
		updateErr = h.workouts.UpdateByTrainer(r.Context(), id, user.ID, wo, exercises)
	}
	if updateErr != nil {
		http.Error(w, "Ошибка обновления", http.StatusInternalServerError)
		return
	}

	if forClientStr := r.FormValue("for_client_id"); forClientStr != "" {
		http.Redirect(w, r, fmt.Sprintf("/trainer/clients/%s/workouts", forClientStr), http.StatusSeeOther)
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

	deleteErr := h.workouts.Delete(r.Context(), id, user.ID)
	if deleteErr != nil && user.IsTrainer() {
		deleteErr = h.workouts.DeleteByTrainer(r.Context(), id, user.ID)
	}

	if ref := r.Header.Get("Referer"); ref != "" && deleteErr == nil {
		http.Redirect(w, r, ref, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/workouts", http.StatusSeeOther)
}

func (h *WorkoutHandler) AddExerciseRow(w http.ResponseWriter, r *http.Request) {
	idx, _ := strconv.Atoi(r.URL.Query().Get("idx"))
	renderPartial(w, r, "workouts/partials/exercise_row.html", map[string]any{
		"ExIdx": idx,
		"Ex":    model.FormExercise{Sets: []model.FormSet{{}}},
	})
}

func (h *WorkoutHandler) AddSetRow(w http.ResponseWriter, r *http.Request) {
	exIdx, _ := strconv.Atoi(r.URL.Query().Get("ex_idx"))
	setIdx, _ := strconv.Atoi(r.URL.Query().Get("set_idx"))
	renderPartial(w, r, "workouts/partials/set_row.html", map[string]any{
		"ExIdx":  exIdx,
		"SetIdx": setIdx,
	})
}

func parseExercisesFromForm(r *http.Request) []model.FormExercise {
	var exercises []model.FormExercise
	for i := 0; ; i++ {
		name := r.FormValue(fmt.Sprintf("exercises[%d][name]", i))
		if name == "" && i > 0 {
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
			wt := r.FormValue(fmt.Sprintf("exercises[%d][sets][%d][weight]", i, j))
			reps := r.FormValue(fmt.Sprintf("exercises[%d][sets][%d][reps]", i, j))
			rpe := r.FormValue(fmt.Sprintf("exercises[%d][sets][%d][rpe]", i, j))
			rest := r.FormValue(fmt.Sprintf("exercises[%d][sets][%d][rest]", i, j))
			notes := r.FormValue(fmt.Sprintf("exercises[%d][sets][%d][notes]", i, j))
			if wt == "" && reps == "" && j > 0 {
				break
			}
			if wt != "" || reps != "" {
				ex.Sets = append(ex.Sets, model.FormSet{
					Weight: wt, Reps: reps, RPE: rpe, RestSeconds: rest, Notes: notes,
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
	for _, layout := range []string{"02.01.2006", "2006-01-02", "01/02/2006"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Now()
}

func formatDateRU(t time.Time) string {
	return t.Format("02.01.2006")
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
