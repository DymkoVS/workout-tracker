package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	"workout-tracker/internal/middleware"
	"workout-tracker/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type TemplateHandler struct {
	templates *repository.TemplateRepository
	tc        *repository.TrainerClientRepository
	gyms      *repository.GymRepository
}

func NewTemplateHandler(
	templates *repository.TemplateRepository,
	tc *repository.TrainerClientRepository,
	gyms *repository.GymRepository,
) *TemplateHandler {
	return &TemplateHandler{templates: templates, tc: tc, gyms: gyms}
}

func (h *TemplateHandler) List(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	list, err := h.templates.List(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}
	renderTemplate(w, r, "templates/list.html", map[string]any{
		"Templates": list,
	})
}

func (h *TemplateHandler) NewForm(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "templates/form.html", nil)
}

func (h *TemplateHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ошибка формы", http.StatusBadRequest)
		return
	}
	exercises := parseExercisesFromForm(r)
	tmpl, err := h.templates.Create(r.Context(), user.ID, r.FormValue("title"), r.FormValue("notes"), r.FormValue("type"), exercises)
	if err != nil {
		renderTemplate(w, r, "templates/form.html", map[string]any{
			"Error": "Ошибка сохранения: " + err.Error(),
		})
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/templates/%s", tmpl.ID), http.StatusSeeOther)
}

func (h *TemplateHandler) Show(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	tmpl, err := h.templates.GetByID(r.Context(), id, user.ID)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	renderTemplate(w, r, "templates/show.html", map[string]any{
		"Template": tmpl,
	})
}

func (h *TemplateHandler) EditForm(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	tmpl, err := h.templates.GetByID(r.Context(), id, user.ID)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	renderTemplate(w, r, "templates/form.html", map[string]any{
		"Template": tmpl,
	})
}

func (h *TemplateHandler) Update(w http.ResponseWriter, r *http.Request) {
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
	exercises := parseExercisesFromForm(r)
	if err := h.templates.Update(r.Context(), id, user.ID, r.FormValue("title"), r.FormValue("notes"), r.FormValue("type"), exercises); err != nil {
		http.Error(w, "Ошибка обновления", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/templates/%s", id), http.StatusSeeOther)
}

func (h *TemplateHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	_ = h.templates.Delete(r.Context(), id, user.ID)
	http.Redirect(w, r, "/templates", http.StatusSeeOther)
}

func (h *TemplateHandler) ApplyForm(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	tmpl, err := h.templates.GetByID(r.Context(), id, user.ID)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	clients, _ := h.tc.GetClients(r.Context(), user.ID)
	gyms, _ := h.gyms.List(r.Context())
	renderTemplate(w, r, "templates/apply.html", map[string]any{
		"Template": tmpl,
		"Clients":  clients,
		"Gyms":     gyms,
		"Today":    time.Now().Format("02.01.2006"),
	})
}

func (h *TemplateHandler) Apply(w http.ResponseWriter, r *http.Request) {
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

	clientStrs := r.Form["client_ids"]
	if len(clientStrs) == 0 {
		http.Redirect(w, r, fmt.Sprintf("/templates/%s/apply", id), http.StatusSeeOther)
		return
	}

	var clientIDs []uuid.UUID
	for _, s := range clientStrs {
		if cid, err := uuid.Parse(strings.TrimSpace(s)); err == nil {
			clientIDs = append(clientIDs, cid)
		}
	}

	date := parseDate(r.FormValue("workout_date"))
	gymID := parseUUIDPtr(r.FormValue("gym_id"))

	if err := h.templates.Apply(r.Context(), id, user.ID, clientIDs, date, gymID); err != nil {
		clients, _ := h.tc.GetClients(r.Context(), user.ID)
		gyms, _ := h.gyms.List(r.Context())
		tmpl, _ := h.templates.GetByID(r.Context(), id, user.ID)
		renderTemplate(w, r, "templates/apply.html", map[string]any{
			"Template": tmpl,
			"Clients":  clients,
			"Gyms":     gyms,
			"Today":    time.Now().Format("02.01.2006"),
			"Error":    "Ошибка применения: " + err.Error(),
		})
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/templates/%s", id), http.StatusSeeOther)
}
