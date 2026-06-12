package handler

import (
	"net"
	"net/http"
	"strings"
	"time"
	"workout-tracker/internal/middleware"
	"workout-tracker/internal/ratelimit"
	"workout-tracker/internal/repository"
	"workout-tracker/internal/session"
)

type AuthHandler struct {
	users    *repository.UserRepository
	sessions *session.Store
	limiter  *ratelimit.Limiter
}

func NewAuthHandler(users *repository.UserRepository, sessions *session.Store) *AuthHandler {
	return &AuthHandler{
		users:    users,
		sessions: sessions,
		// 10 неудачных попыток за 15 минут с одного IP.
		limiter: ratelimit.New(10, 15*time.Minute),
	}
}

// clientIP — IP клиента. Приложение слушает только localhost за Caddy,
// поэтому X-Forwarded-For (его ставит Caddy) можно доверять.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.IndexByte(xff, ','); i > 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
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

	ip := clientIP(r)
	if !h.limiter.Allow(ip) {
		http.Redirect(w, r, "/login?error=toomany", http.StatusSeeOther)
		return
	}

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
	h.limiter.Reset(ip)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if sessionID, err := session.ReadCookie(r); err == nil {
		_ = h.sessions.Delete(r.Context(), sessionID)
	}
	session.ClearCookie(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
