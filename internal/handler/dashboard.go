package handler

import (
	"net/http"
	"workout-tracker/internal/middleware"
)

func Dashboard(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	renderTemplate(w, r, "dashboard.html", map[string]any{
		"CurrentUser": user,
	})
}
