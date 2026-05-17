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
	users     *repository.UserRepository
}

func NewProfileHandler(
	workouts *repository.WorkoutRepository,
	gyms *repository.GymRepository,
	tc *repository.TrainerClientRepository,
	templates *repository.TemplateRepository,
	users *repository.UserRepository,
) *ProfileHandler {
	return &ProfileHandler{workouts: workouts, gyms: gyms, tc: tc, templates: templates, users: users}
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
		"pwd":          r.URL.Query().Get("pwd"),
	}

	if user.IsTrainer() {
		clients, _ := h.tc.GetClients(r.Context(), user.ID)
		data["ClientCount"] = len(clients)

		templateList, _ := h.templates.List(r.Context(), user.ID)
		data["TemplateCount"] = len(templateList)
	}

	renderTemplate(w, r, "profile.html", data)
}

func (h *ProfileHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	current := r.FormValue("current_password")
	newPwd := r.FormValue("new_password")
	confirm := r.FormValue("confirm_password")

	full, _ := h.users.GetByID(r.Context(), user.ID)

	var errMsg string
	switch {
	case !h.users.CheckPassword(full, current):
		errMsg = "Неверный текущий пароль"
	case len(newPwd) < 6:
		errMsg = "Новый пароль должен быть не короче 6 символов"
	case newPwd != confirm:
		errMsg = "Пароли не совпадают"
	}

	if errMsg != "" {
		stats, _ := h.workouts.GetDashboardStats(r.Context(), user.ID)
		cards, _ := h.workouts.ListCards(r.Context(), user.ID)
		var totalTonnage float64
		for _, c := range cards {
			totalTonnage += c.Tonnage
		}
		gymList, _ := h.gyms.List(r.Context())
		data := map[string]any{
			"Stats":           stats,
			"TotalCount":      len(cards),
			"TotalTonnage":    totalTonnage,
			"GymCount":        len(gymList),
			"PasswordError":   errMsg,
			"PasswordSection": true,
		}
		if user.IsTrainer() {
			clients, _ := h.tc.GetClients(r.Context(), user.ID)
			data["ClientCount"] = len(clients)
			templateList, _ := h.templates.List(r.Context(), user.ID)
			data["TemplateCount"] = len(templateList)
		}
		renderTemplate(w, r, "profile.html", data)
		return
	}

	if err := h.users.SetPassword(r.Context(), user.ID, newPwd); err != nil {
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/profile?pwd=ok", http.StatusSeeOther)
}
