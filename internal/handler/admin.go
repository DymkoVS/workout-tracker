package handler

import (
	"net/http"
	"workout-tracker/internal/middleware"
	"workout-tracker/internal/model"
	"workout-tracker/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type AdminHandler struct {
	users *repository.UserRepository
	tc    *repository.TrainerClientRepository
}

func NewAdminHandler(users *repository.UserRepository, tc *repository.TrainerClientRepository) *AdminHandler {
	return &AdminHandler{users: users, tc: tc}
}

func (h *AdminHandler) UsersList(w http.ResponseWriter, r *http.Request) {
	users, err := h.users.List(r.Context())
	if err != nil {
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}
	renderTemplate(w, r, "admin/users.html", map[string]any{
		"Users":       users,
		"CurrentUser": middleware.UserFromContext(r.Context()),
	})
}

func (h *AdminHandler) NewUserForm(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "admin/user_form.html", map[string]any{
		"CurrentUser": middleware.UserFromContext(r.Context()),
	})
}

func (h *AdminHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	in := model.CreateUserInput{
		Login:    r.FormValue("login"),
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
		FullName: r.FormValue("full_name"),
		Role:     r.FormValue("role"),
		IsAdmin:  r.FormValue("is_admin") == "on",
	}

	if in.Login == "" || in.Password == "" || (in.Role != model.RoleTrainer && in.Role != model.RoleClient) {
		renderTemplate(w, r, "admin/user_form.html", map[string]any{
			"Error":       "Заполните все обязательные поля",
			"Input":       in,
			"CurrentUser": middleware.UserFromContext(r.Context()),
		})
		return
	}

	if _, err := h.users.Create(r.Context(), in); err != nil {
		renderTemplate(w, r, "admin/user_form.html", map[string]any{
			"Error":       "Пользователь с таким логином уже существует",
			"Input":       in,
			"CurrentUser": middleware.UserFromContext(r.Context()),
		})
		return
	}

	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

func (h *AdminHandler) EditUserForm(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	user, err := h.users.GetByID(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	renderTemplate(w, r, "admin/user_form.html", map[string]any{
		"EditUser":    user,
		"CurrentUser": middleware.UserFromContext(r.Context()),
	})
}

func (h *AdminHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	err = h.users.Update(r.Context(), id,
		r.FormValue("login"),
		r.FormValue("email"),
		r.FormValue("full_name"),
		r.FormValue("role"),
		r.FormValue("is_admin") == "on",
		r.FormValue("is_active") == "on",
	)
	if err != nil {
		http.Error(w, "Ошибка обновления", http.StatusInternalServerError)
		return
	}

	if newPass := r.FormValue("password"); newPass != "" {
		_ = h.users.SetPassword(r.Context(), id, newPass)
	}
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

// AssignPage показывает страницу управления связями тренер-клиент
func (h *AdminHandler) AssignPage(w http.ResponseWriter, r *http.Request) {
	trainers, err := h.tc.GetAllTrainers(r.Context())
	if err != nil {
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}
	clients, err := h.tc.GetAllClients(r.Context())
	if err != nil {
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Для каждого тренера получаем его клиентов
	type trainerRow struct {
		Trainer *model.User
		Clients []*model.User
	}
	var rows []trainerRow
	for _, t := range trainers {
		tClients, _ := h.tc.GetClients(r.Context(), t.ID)
		rows = append(rows, trainerRow{Trainer: t, Clients: tClients})
	}

	renderTemplate(w, r, "admin/assign.html", map[string]any{
		"Rows":    rows,
		"Clients": clients,
	})
}

// Assign назначает клиента к тренеру
func (h *AdminHandler) Assign(w http.ResponseWriter, r *http.Request) {
	trainerID, err1 := uuid.Parse(r.FormValue("trainer_id"))
	clientID, err2 := uuid.Parse(r.FormValue("client_id"))
	if err1 != nil || err2 != nil {
		http.Error(w, "Неверные параметры", http.StatusBadRequest)
		return
	}
	_ = h.tc.Assign(r.Context(), trainerID, clientID)
	http.Redirect(w, r, "/admin/assign", http.StatusSeeOther)
}

// Unassign убирает клиента от тренера
func (h *AdminHandler) Unassign(w http.ResponseWriter, r *http.Request) {
	trainerID, err1 := uuid.Parse(chi.URLParam(r, "trainerID"))
	clientID, err2 := uuid.Parse(chi.URLParam(r, "clientID"))
	if err1 != nil || err2 != nil {
		http.Error(w, "Неверные параметры", http.StatusBadRequest)
		return
	}
	_ = h.tc.Unassign(r.Context(), trainerID, clientID)
	http.Redirect(w, r, "/admin/assign", http.StatusSeeOther)
}
