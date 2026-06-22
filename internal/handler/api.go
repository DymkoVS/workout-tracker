package handler

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"workout-tracker/internal/model"
	"workout-tracker/internal/repository"

	"github.com/google/uuid"
)

// APIHandler serves the server-to-server import API used by the Telegram bot.
// All writes go through the same validated WorkoutRepository.Create path as the
// web UI, so the bot no longer builds raw SQL and can't drift from the schema.
// Auth is a single shared bearer token (IMPORT_API_TOKEN); if it is unset the
// API is disabled (fail closed).
type APIHandler struct {
	workouts *repository.WorkoutRepository
	gyms     *repository.GymRepository
	users    *repository.UserRepository
	token    string
}

func NewAPIHandler(w *repository.WorkoutRepository, g *repository.GymRepository, u *repository.UserRepository, token string) *APIHandler {
	return &APIHandler{workouts: w, gyms: g, users: u, token: token}
}

// authOK does a constant-time compare of the bearer token. Empty configured
// token → always false (API disabled).
func (h *APIHandler) authOK(r *http.Request) bool {
	if h.token == "" {
		return false
	}
	got := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	return subtle.ConstantTimeCompare([]byte(got), []byte(h.token)) == 1
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// Gyms returns the list of gym names so the bot can give Claude the real names
// to match against (replaces the bot reading the gyms table directly).
func (h *APIHandler) Gyms(w http.ResponseWriter, r *http.Request) {
	if !h.authOK(r) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	gyms, err := h.gyms.List(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "db"})
		return
	}
	names := make([]string, 0, len(gyms))
	for _, g := range gyms {
		names = append(names, g.Name)
	}
	writeJSON(w, http.StatusOK, map[string]any{"gyms": names})
}

// ── Активная сессия (источник для Apple Watch-пульта) ─────────────────────────

type apiActiveSet struct {
	SetNum      int      `json:"set_num"`
	Weight      *float64 `json:"weight"`
	Reps        *int     `json:"reps"`
	RPE         *float64 `json:"rpe"`
	RestSeconds *int     `json:"rest_seconds"`
	Done        bool     `json:"done"`
}

type apiActiveExercise struct {
	Name  string         `json:"name"`
	Order int            `json:"order"`
	Sets  []apiActiveSet `json:"sets"`
}

type apiActiveResp struct {
	ID        string              `json:"id"`
	Title     string              `json:"title"`
	StartedAt *time.Time          `json:"started_at"`
	Exercises []apiActiveExercise `json:"exercises"`
}

// ActiveSession возвращает текущую активную тренировку пользователя (начата, не
// завершена) с упражнениями и целевыми подходами — источник для Apple Watch-пульта.
// Нет активной сессии → {"active": null}. Только чтение; auth — общий токен.
func (h *APIHandler) ActiveSession(w http.ResponseWriter, r *http.Request) {
	if !h.authOK(r) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	login := strings.TrimSpace(r.URL.Query().Get("login"))
	if login == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "login required"})
		return
	}
	user, err := h.users.GetByLogin(r.Context(), login)
	if err != nil || user == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unknown user: " + login})
		return
	}

	active := h.workouts.FindActiveWorkout(r.Context(), user.ID)
	if active == nil {
		writeJSON(w, http.StatusOK, map[string]any{"active": nil})
		return
	}
	full, err := h.workouts.GetActiveSession(r.Context(), active.ID, user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "db"})
		return
	}

	resp := apiActiveResp{ID: full.ID.String(), Title: full.Title, StartedAt: full.StartedAt}
	for _, ex := range full.Exercises {
		ae := apiActiveExercise{Name: ex.Name, Order: ex.OrderNum}
		for _, s := range ex.Sets {
			ae.Sets = append(ae.Sets, apiActiveSet{
				SetNum: s.SetNum, Weight: s.Weight, Reps: s.Reps,
				RPE: s.RPE, RestSeconds: s.RestSeconds, Done: s.Done,
			})
		}
		resp.Exercises = append(resp.Exercises, ae)
	}
	writeJSON(w, http.StatusOK, map[string]any{"active": resp})
}

type apiImportSet struct {
	Weight      *float64 `json:"weight"`
	Reps        *int     `json:"reps"`
	RPE         *float64 `json:"rpe"`
	RestSeconds *int     `json:"rest_seconds"`
	Notes       string   `json:"notes"`
}

type apiImportExercise struct {
	Name  string         `json:"name"`
	Notes string         `json:"notes"`
	Sets  []apiImportSet `json:"sets"`
}

type apiImportReq struct {
	Login     string              `json:"login"`
	Title     string              `json:"title"`
	Date      string              `json:"date"` // YYYY-MM-DD; empty → today
	Gym       string              `json:"gym"`  // gym name; empty/unknown → none
	Notes     string              `json:"notes"`
	Exercises []apiImportExercise `json:"exercises"`
}

// Import creates a workout from the bot's parsed JSON via the shared Create path.
func (h *APIHandler) Import(w http.ResponseWriter, r *http.Request) {
	if !h.authOK(r) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req apiImportReq
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "bad json: " + err.Error()})
		return
	}
	if strings.TrimSpace(req.Login) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "login required"})
		return
	}
	if len(req.Exercises) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "no exercises"})
		return
	}

	user, err := h.users.GetByLogin(r.Context(), req.Login)
	if err != nil || user == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unknown user: " + req.Login})
		return
	}

	// Дата: YYYY-MM-DD или сегодня.
	date := time.Now()
	if s := strings.TrimSpace(req.Date); s != "" {
		if d, err := time.Parse("2006-01-02", s); err == nil {
			date = d
		} else {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "bad date, want YYYY-MM-DD"})
			return
		}
	}

	// Зал: точное совпадение по имени (без учёта регистра). Бот присылает имя,
	// уже выбранное Claude из реального списка, так что fuzzy не нужен.
	var gymID *uuid.UUID
	if name := strings.TrimSpace(req.Gym); name != "" {
		if gyms, err := h.gyms.List(r.Context()); err == nil {
			for _, g := range gyms {
				if strings.EqualFold(g.Name, name) {
					id := g.ID
					gymID = &id
					break
				}
			}
		}
	}

	ended := date // импортированная тренировка считается завершённой
	wk := model.Workout{
		Title:       strings.TrimSpace(req.Title),
		WorkoutType: "imported",
		WorkoutDate: date,
		Notes:       req.Notes,
		GymID:       gymID,
		EndedAt:     &ended,
	}

	exercises := make([]model.FormExercise, 0, len(req.Exercises))
	for _, ex := range req.Exercises {
		if strings.TrimSpace(ex.Name) == "" {
			continue
		}
		fe := model.FormExercise{Name: ex.Name, Notes: ex.Notes}
		for _, s := range ex.Sets {
			fe.Sets = append(fe.Sets, model.FormSet{
				Weight:      fmtOptFloat(s.Weight),
				Reps:        fmtOptInt(s.Reps),
				RPE:         fmtOptFloat(s.RPE),
				RestSeconds: fmtOptInt(s.RestSeconds),
				Notes:       s.Notes,
			})
		}
		exercises = append(exercises, fe)
	}
	if len(exercises) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "no valid exercises"})
		return
	}

	created, err := h.workouts.Create(r.Context(), user.ID, wk, exercises)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "create failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"id": created.ID.String()})
}

func fmtOptFloat(f *float64) string {
	if f == nil {
		return ""
	}
	return strconv.FormatFloat(*f, 'g', -1, 64)
}

func fmtOptInt(i *int) string {
	if i == nil {
		return ""
	}
	return strconv.Itoa(*i)
}
