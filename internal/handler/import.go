package handler

import (
	"net/http"
	"strconv"
	"workout-tracker/internal/importer"
	"workout-tracker/internal/middleware"
	"workout-tracker/internal/model"
	"workout-tracker/internal/repository"
)

type ImportHandler struct {
	workouts *repository.WorkoutRepository
	gyms     *repository.GymRepository
}

func NewImportHandler(workouts *repository.WorkoutRepository, gyms *repository.GymRepository) *ImportHandler {
	return &ImportHandler{workouts: workouts, gyms: gyms}
}

func (h *ImportHandler) Form(w http.ResponseWriter, r *http.Request) {
	gyms, _ := h.gyms.List(r.Context())
	renderTemplate(w, r, "workouts/import.html", map[string]any{"Gyms": gyms})
}

func (h *ImportHandler) Preview(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(2 << 20); err != nil {
		r.ParseForm()
	}
	text := r.FormValue("text")
	if text == "" {
		http.Redirect(w, r, "/workouts/import", http.StatusSeeOther)
		return
	}

	parsed := importer.Parse(text)
	gyms, _ := h.gyms.List(r.Context())

	renderTemplate(w, r, "workouts/import.html", map[string]any{
		"Gyms":    gyms,
		"Text":    text,
		"Parsed":  parsed,
		"Preview": true,
		"GymID":   r.FormValue("gym_id"),
	})
}

func (h *ImportHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	text := r.FormValue("text")
	gymIDStr := r.FormValue("gym_id")

	parsed := importer.Parse(text)
	if len(parsed) == 0 {
		http.Redirect(w, r, "/workouts/import", http.StatusSeeOther)
		return
	}

	gymID := parseUUIDPtr(gymIDStr)

	imported := 0
	for _, pw := range parsed {
		wo := model.Workout{
			UserID:      user.ID,
			Title:       pw.Title,
			WorkoutDate: pw.Date,
			GymID:       gymID,
		}
		var exercises []model.FormExercise
		for _, ex := range pw.Exercises {
			fe := model.FormExercise{Name: ex.Name}
			for _, s := range ex.Sets {
				fs := model.FormSet{Reps: itoa(s.Reps)}
				if s.Weight != nil {
					fs.Weight = ftoa(*s.Weight)
				}
				fe.Sets = append(fe.Sets, fs)
			}
			exercises = append(exercises, fe)
		}
		if _, err := h.workouts.Create(r.Context(), user.ID, wo, exercises); err == nil {
			imported++
		}
	}

	renderTemplate(w, r, "workouts/import.html", map[string]any{
		"Done":     true,
		"Imported": imported,
		"Total":    len(parsed),
	})
}

func itoa(n int) string {
	if n == 0 {
		return ""
	}
	return strconv.Itoa(n)
}

func ftoa(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}
