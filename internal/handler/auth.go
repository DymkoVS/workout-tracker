package handler

import (
	"net/http"
	"workout-tracker/internal/middleware"
	"workout-tracker/internal/repository"
	"workout-tracker/internal/session"
)

type AuthHandler struct {
	users    *repository.UserRepository
	sessions *session.Store
}

func NewAuthHandler(users *repository.UserRepository, sessions *session.Store) *AuthHandler {
	return &AuthHandler{users: users, sessions: sessions}
}

func (h *AuthHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	if middleware.UserFromContext(r.Context()) != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	data := map[string]any{
		"Error": r.URL.Query().Get("error"),
	}
	renderTemplate(w, r, "login.html", data)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	login := r.FormValue("login")
	password := r.FormValue("password")

	user, err := h.users.GetByLogin(r.Context(), login)
	if err != nil || !h.users.CheckPassword(user, password) {
		http.Redirect(w, r, "/login?error=invalid", http.StatusSeeOther)
		return
	}
	if !user.IsActive {
		http.Redirect(w, r, "/login?error=inactive", http.StatusSeeOther)
		return
	}

	if oldID, err := session.ReadCookie(r); err == nil {
		_ = h.sessions.Delete(r.Context(), oldID)
	}

	sessionID, err := h.sessions.Create(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}
	session.SetCookie(w, sessionID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if sessionID, err := session.ReadCookie(r); err == nil {
		_ = h.sessions.Delete(r.Context(), sessionID)
	}
	session.ClearCookie(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
