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
	"wellbeingEmoji": func(v any) string {
		var n int
		switch val := v.(type) {
		case int:
			n = val
		case *int:
			if val == nil {
				return ""
			}
			n = *val
		}
		emojis := []string{"😞", "😐", "🙂", "😊", "💪"}
		if n >= 1 && n <= 5 {
			return emojis[n-1]
		}
		return ""
	},
	"add": func(a, b int) int { return a + b },
	"iterate": func(n int) []int {
		s := make([]int, n)
		for i := range s {
			s[i] = i
		}
		return s
	},
	"dict": func(values ...any) map[string]any {
		m := make(map[string]any)
		for i := 0; i+1 < len(values); i += 2 {
			m[values[i].(string)] = values[i+1]
		}
		return m
	},
	"deref": func(p *int) int {
		if p == nil {
			return 0
		}
		return *p
	},
	"derefI": func(p *int) int {
		if p == nil {
			return 0
		}
		return *p
	},
	"derefF": func(p *float64) float64 {
		if p == nil {
			return 0
		}
		return *p
	},
}

func renderTemplate(w http.ResponseWriter, r *http.Request, name string, data map[string]any) {
	if data == nil {
		data = map[string]any{}
	}
	if _, ok := data["CurrentUser"]; !ok {
		data["CurrentUser"] = middleware.UserFromContext(r.Context())
	}

	// Собираем все шаблоны: base + страница + партиалы упражнений
	files := []string{
		filepath.Join("web", "templates", "base.html"),
		filepath.Join("web", "templates", name),
	}
	// Подключаем партиалы тренировок если они нужны форме
	partials := []string{
		"web/templates/workouts/partials/exercise_row.html",
		"web/templates/workouts/partials/set_row.html",
	}
	for _, p := range partials {
		files = append(files, p)
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
		http.Error(w, "Ошибка шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.ExecuteTemplate(w, filepath.Base(name), data)
}
