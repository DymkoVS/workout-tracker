package handler

import (
	"net/http"
	"workout-tracker/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type GymHandler struct {
	gyms *repository.GymRepository
}

func NewGymHandler(gyms *repository.GymRepository) *GymHandler {
	return &GymHandler{gyms: gyms}
}

func (h *GymHandler) List(w http.ResponseWriter, r *http.Request) {
	gyms, err := h.gyms.List(r.Context())
	if err != nil {
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}
	renderTemplate(w, r, "gyms/list.html", map[string]any{
		"Gyms": gyms,
	})
}

func (h *GymHandler) NewForm(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "gyms/form.html", nil)
}

func (h *GymHandler) Create(w http.ResponseWriter, r *http.Request) {
	name := clampStr(r.FormValue("name"), 120)
	if name == "" {
		renderTemplate(w, r, "gyms/form.html", map[string]any{
			"Error": "Название зала обязательно",
		})
		return
	}
	if _, err := h.gyms.Create(r.Context(), name); err != nil {
		renderTemplate(w, r, "gyms/form.html", map[string]any{
			"Error": "Ошибка сохранения: " + err.Error(),
			"Name":  name,
		})
		return
	}
	http.Redirect(w, r, "/gyms", http.StatusSeeOther)
}

func (h *GymHandler) EditForm(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	gym, err := h.gyms.GetByID(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	renderTemplate(w, r, "gyms/form.html", map[string]any{
		"Gym": gym,
	})
}

func (h *GymHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	name := clampStr(r.FormValue("name"), 120)
	if name == "" {
		gym, _ := h.gyms.GetByID(r.Context(), id)
		renderTemplate(w, r, "gyms/form.html", map[string]any{
			"Error": "Название зала обязательно",
			"Gym":   gym,
		})
		return
	}
	if err := h.gyms.Update(r.Context(), id, name); err != nil {
		renderTemplate(w, r, "gyms/form.html", map[string]any{
			"Error": "Ошибка обновления: " + err.Error(),
		})
		return
	}
	http.Redirect(w, r, "/gyms", http.StatusSeeOther)
}
