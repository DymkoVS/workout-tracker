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
}

func NewAdminHandler(users *repository.UserRepository) *AdminHandler {
	return &AdminHandler{users: users}
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

	if r.Header.Get("HX-Request") == "true" {
		users, _ := h.users.List(r.Context())
		renderTemplate(w, r, "admin/users_table.html", map[string]any{"Users": users})
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
