package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"workout-tracker/internal/middleware"
	"workout-tracker/internal/model"
	"workout-tracker/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type WorkoutHandler struct {
	workouts  *repository.WorkoutRepository
	gyms      *repository.GymRepository
	tc        *repository.TrainerClientRepository
	users     *repository.UserRepository
	media     *repository.MediaRepository
	uploadDir string
}

func NewWorkoutHandler(
	workouts *repository.WorkoutRepository,
	gyms *repository.GymRepository,
	tc *repository.TrainerClientRepository,
	users *repository.UserRepository,
	media *repository.MediaRepository,
	uploadDir string,
) *WorkoutHandler {
	return &WorkoutHandler{workouts: workouts, gyms: gyms, tc: tc, users: users, media: media, uploadDir: uploadDir}
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

	q := r.URL.Query()
	filter := repository.WorkoutFilter{}
	filterFrom := q.Get("from")
	filterTo := q.Get("to")
	filterGymID := q.Get("gym_id")
	filterExercise := q.Get("exercise")
	filterType := q.Get("type")

	if filterFrom != "" {
		if t, err := time.Parse("2006-01-02", filterFrom); err == nil {
			filter.DateFrom = &t
		}
	}
	if filterTo != "" {
		if t, err := time.Parse("2006-01-02", filterTo); err == nil {
			filter.DateTo = &t
		}
	}
	if filterGymID != "" {
		if id, err := uuid.Parse(filterGymID); err == nil {
			filter.GymID = &id
		}
	}
	filter.ExerciseName = filterExercise
	filter.WorkoutType = filterType

	cards, err := h.workouts.ListCardsFiltered(r.Context(), user.ID, filter)
	if err != nil {
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Sparkline: up to 6 most-recent workouts in chronological order (oldest→newest).
	n := len(cards)
	end := n
	if end > 6 {
		end = 6
	}
	sparkline := make([]float64, end)
	for i := 0; i < end; i++ {
		sparkline[i] = cards[end-1-i].Tonnage
	}

	// Tonnage delta: last 30 days vs. prior 30 days (no extra DB query).
	var tonnageDelta string
	now := time.Now()
	cut30 := now.AddDate(0, 0, -30)
	cut60 := now.AddDate(0, 0, -60)
	var last30, prev30 float64
	for _, c := range cards {
		switch {
		case c.WorkoutDate.After(cut30):
			last30 += c.Tonnage
		case c.WorkoutDate.After(cut60):
			prev30 += c.Tonnage
		}
	}
	if prev30 > 0 {
		pct := (last30 - prev30) / prev30 * 100
		if pct >= 0 {
			tonnageDelta = fmt.Sprintf("▲ +%.0f%%", pct)
		} else {
			tonnageDelta = fmt.Sprintf("▼ %.0f%%", pct)
		}
	}

	gyms, _ := h.gyms.List(r.Context())

	renderTemplate(w, r, "workouts/list.html", map[string]any{
		"WorkoutGroups":     groupByMonth(cards),
		"TotalCount":        len(cards),
		"SparklineTonnages": sparkline,
		"TonnageDelta":      tonnageDelta,
		"Gyms":              gyms,
		"FilterFrom":        filterFrom,
		"FilterTo":          filterTo,
		"FilterGymID":       filterGymID,
		"FilterExercise":    filterExercise,
		"FilterType":        filterType,
		"FilterActive":      filter.IsActive(),
	})
}

func (h *WorkoutHandler) NewForm(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	gyms, _ := h.gyms.List(r.Context())

	recentForID := user.ID
	data := map[string]any{
		"Gyms":  gyms,
		"Today": time.Now().Format("02.01.2006"),
	}
	if forClientStr := r.URL.Query().Get("for_client"); forClientStr != "" && user.IsTrainer() {
		if clientID, err := uuid.Parse(forClientStr); err == nil {
			if ok, _ := h.tc.IsAssigned(r.Context(), user.ID, clientID); ok {
				if client, err := h.users.GetByID(r.Context(), clientID); err == nil {
					data["ForClient"] = client
					recentForID = clientID
				}
			}
		}
	}
	recent, _ := h.workouts.GetRecentUnique(r.Context(), recentForID, 6)
	data["RecentWorkouts"] = recent
	renderTemplate(w, r, "workouts/form.html", data)
}

func (h *WorkoutHandler) CopyFromWorkout(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	workout, err := h.workouts.GetByID(r.Context(), id, user.ID)
	if err != nil && user.IsTrainer() {
		workout, err = h.workouts.GetByIDForTrainer(r.Context(), id, user.ID)
	}
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	renderPartialWith(w, r, "workouts/partials/copy_exercises.html", []string{
		"web/templates/workouts/partials/exercise_block.html",
	}, map[string]any{
		"Exercises": workout.Exercises,
	})
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
		WorkoutType: r.FormValue("workout_type"),
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

	media, _ := h.media.ListForWorkout(r.Context(), id)

	data := map[string]any{"Workout": workout, "Media": media}
	if workout.UserID != user.ID && user.IsTrainer() {
		data["BackURL"] = fmt.Sprintf("/trainer/clients/%s/workouts", workout.UserID)
		data["CanEdit"] = workout.TrainerID != nil && *workout.TrainerID == user.ID
		data["CanStart"] = false
	} else {
		data["CanEdit"] = true
		data["CanStart"] = true
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
		WorkoutType: r.FormValue("workout_type"),
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

	_ = deleteErr
	http.Redirect(w, r, "/workouts", http.StatusSeeOther)
}

func (h *WorkoutHandler) ActiveSession(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	workout, err := h.workouts.GetActiveSession(r.Context(), id, user.ID)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if workout.StartedAt == nil {
		_ = h.workouts.StartSession(r.Context(), id, user.ID)
		now := time.Now()
		workout.StartedAt = &now
	}
	startedUnix := workout.StartedAt.Unix()
	renderTemplate(w, r, "workouts/active.html", map[string]any{
		"Workout":       workout,
		"StartedAtUnix": startedUnix,
	})
}

func (h *WorkoutHandler) ToggleSetDone(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	setID, err := uuid.Parse(chi.URLParam(r, "setID"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	done, err := h.workouts.ToggleSetDone(r.Context(), setID, user.ID)
	if err != nil {
		http.Error(w, "Ошибка", http.StatusInternalServerError)
		return
	}
	if done {
		w.Header().Set("HX-Trigger", "startRestTimer")
	}
	s, err := h.workouts.GetSetByID(r.Context(), setID, user.ID)
	if err != nil {
		http.Error(w, "Ошибка", http.StatusInternalServerError)
		return
	}
	renderPartial(w, r, "workouts/partials/active_set_row.html", map[string]any{
		"Set": s,
	})
}

func (h *WorkoutHandler) FinishSession(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	_ = h.workouts.FinishSession(r.Context(), id, user.ID)
	http.Redirect(w, r, fmt.Sprintf("/workouts/%s", id), http.StatusSeeOther)
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

func (h *WorkoutHandler) ExerciseSuggest(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if len(q) < 2 {
		w.Write([]byte(""))
		return
	}
	names, err := h.workouts.SuggestExercises(r.Context(), user.ID, q)
	if err != nil || len(names) == 0 {
		w.Write([]byte(""))
		return
	}
	renderPartial(w, r, "workouts/partials/exercise_suggest.html", map[string]any{"Names": names})
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

var allowedMIME = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"video/mp4":  ".mp4",
}

const maxUploadSize = 10 << 20 // 10 MB

func (h *WorkoutHandler) UploadMedia(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	workoutID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	// verify ownership
	workout, err := h.workouts.GetByID(r.Context(), workoutID, user.ID)
	if err != nil || workout == nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "Файл слишком большой (максимум 10 МБ)", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Файл не найден", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// detect MIME from first 512 bytes
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	mimeType := http.DetectContentType(buf[:n])
	// strip params (e.g. "image/jpeg; charset=...")
	if idx := strings.Index(mimeType, ";"); idx != -1 {
		mimeType = strings.TrimSpace(mimeType[:idx])
	}
	ext, ok := allowedMIME[mimeType]
	if !ok {
		http.Error(w, "Тип файла не поддерживается. Разрешены: JPG, PNG, WEBP, MP4", http.StatusBadRequest)
		return
	}

	// reset reader
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		http.Error(w, "Ошибка чтения файла", http.StatusInternalServerError)
		return
	}

	filename := uuid.New().String() + ext
	dir := filepath.Join(h.uploadDir, workoutID.String())
	if err := os.MkdirAll(dir, 0755); err != nil {
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	dst, err := os.Create(filepath.Join(dir, filename))
	if err != nil {
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	size, err := io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Ошибка сохранения файла", http.StatusInternalServerError)
		return
	}

	if _, err := h.media.Create(r.Context(), workoutID, filename, header.Filename, mimeType, int(size)); err != nil {
		os.Remove(filepath.Join(dir, filename))
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/workouts/"+workoutID.String(), http.StatusSeeOther)
}

func (h *WorkoutHandler) DeleteMedia(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	workoutID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	mediaID, err := uuid.Parse(chi.URLParam(r, "mediaID"))
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	// verify ownership
	workout, err := h.workouts.GetByID(r.Context(), workoutID, user.ID)
	if err != nil || workout == nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	m, err := h.media.GetByID(r.Context(), mediaID, workoutID)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	os.Remove(filepath.Join(h.uploadDir, workoutID.String(), m.Filename))
	h.media.Delete(r.Context(), mediaID)

	http.Redirect(w, r, "/workouts/"+workoutID.String(), http.StatusSeeOther)
}

func (h *WorkoutHandler) ServeMedia(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	workoutID, err := uuid.Parse(chi.URLParam(r, "workoutID"))
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	filename := chi.URLParam(r, "filename")

	// verify ownership
	workout, err := h.workouts.GetByID(r.Context(), workoutID, user.ID)
	if err != nil || workout == nil {
		// trainers can also view their clients' media
		if _, terr := h.workouts.GetByIDForTrainer(r.Context(), workoutID, user.ID); terr != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
	}

	http.ServeFile(w, r, filepath.Join(h.uploadDir, workoutID.String(), filepath.Base(filename)))
}
