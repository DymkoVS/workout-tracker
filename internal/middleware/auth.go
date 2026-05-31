package middleware

import (
	"context"
	"net/http"
	"workout-tracker/internal/model"
	"workout-tracker/internal/repository"
	"workout-tracker/internal/session"
)

type contextKey string

const UserKey contextKey = "user"
const ActiveWorkoutKey contextKey = "active_workout"

type AuthMiddleware struct {
	sessions *session.Store
	users    *repository.UserRepository
	workouts *repository.WorkoutRepository
}

func NewAuthMiddleware(sessions *session.Store, users *repository.UserRepository, workouts *repository.WorkoutRepository) *AuthMiddleware {
	return &AuthMiddleware{sessions: sessions, users: users, workouts: workouts}
}

func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := m.userFromRequest(r)
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		if !user.IsActive {
			session.ClearCookie(w)
			http.Redirect(w, r, "/login?error=inactive", http.StatusSeeOther)
			return
		}
		ctx := context.WithValue(r.Context(), UserKey, user)
		if aw := m.workouts.FindActiveWorkout(ctx, user.ID); aw != nil {
			ctx = context.WithValue(ctx, ActiveWorkoutKey, aw)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *AuthMiddleware) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := UserFromContext(r.Context())
		if user == nil || !user.IsAdmin {
			http.Error(w, "Доступ запрещён", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (m *AuthMiddleware) RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := UserFromContext(r.Context())
			if user == nil || user.Role != role {
				http.Error(w, "Доступ запрещён", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (m *AuthMiddleware) userFromRequest(r *http.Request) *model.User {
	sessionID, err := session.ReadCookie(r)
	if err != nil {
		return nil
	}
	userID, err := m.sessions.GetUserID(r.Context(), sessionID)
	if err != nil {
		return nil
	}
	user, err := m.users.GetByID(r.Context(), userID)
	if err != nil {
		return nil
	}
	return user
}

func UserFromContext(ctx context.Context) *model.User {
	u, _ := ctx.Value(UserKey).(*model.User)
	return u
}

func ActiveWorkoutFromContext(ctx context.Context) *model.Workout {
	w, _ := ctx.Value(ActiveWorkoutKey).(*model.Workout)
	return w
}
