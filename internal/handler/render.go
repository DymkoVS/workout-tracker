package handler

import (
	"html/template"
	"net/http"
	"path/filepath"
	"workout-tracker/internal/middleware"
)

var tmplFuncs = template.FuncMap{
	"roleLabel": func(role string) string {
		if role == "trainer" {
			return "Тренер"
		}
		return "Клиент"
	},
}

func renderTemplate(w http.ResponseWriter, r *http.Request, name string, data map[string]any) {
	if data == nil {
		data = map[string]any{}
	}
	if _, ok := data["CurrentUser"]; !ok {
		data["CurrentUser"] = middleware.UserFromContext(r.Context())
	}

	files := []string{
		filepath.Join("web", "templates", "base.html"),
		filepath.Join("web", "templates", name),
	}

	tmpl, err := template.New("").Funcs(tmplFuncs).ParseFiles(files...)
	if err != nil {
		http.Error(w, "Ошибка шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, "Ошибка рендеринга: "+err.Error(), http.StatusInternalServerError)
	}
}

func renderPartial(w http.ResponseWriter, r *http.Request, name string, data map[string]any) {
	if data == nil {
		data = map[string]any{}
	}
	tmpl, err := template.New("").Funcs(tmplFuncs).ParseFiles(
		filepath.Join("web", "templates", name),
	)
	if err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.ExecuteTemplate(w, filepath.Base(name), data)
}
