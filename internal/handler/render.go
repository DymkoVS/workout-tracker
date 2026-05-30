package handler

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
	"time"
	"workout-tracker/internal/middleware"
	"workout-tracker/internal/model"
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
	"add":  func(a, b int) int { return a + b },
	"list": func(values ...string) []string { return values },
	"not": func(v any) bool {
		if v == nil {
			return true
		}
		if b, ok := v.(bool); ok {
			return !b
		}
		return false
	},
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
	"dateRU": func(t interface{}) string {
		switch v := t.(type) {
		case interface{ Format(string) string }:
			return v.Format("02.01.2006")
		}
		return ""
	},
	"toUpper": strings.ToUpper,
	"printf":  fmt.Sprintf,
	"formatTonnage": func(kg float64) string {
		if kg == 0 {
			return "0кг"
		}
		if kg >= 1000 {
			t := kg / 1000
			if t == float64(int(t)) {
				return fmt.Sprintf("%dт", int(t))
			}
			return fmt.Sprintf("%.1fт", t)
		}
		return fmt.Sprintf("%.0fкг", kg)
	},
	"exerciseTonnage": func(e model.WorkoutExercise) float64 {
		var t float64
		for _, s := range e.Sets {
			if s.Weight != nil && s.Reps != nil {
				t += *s.Weight * float64(*s.Reps)
			}
		}
		return t
	},
	"workoutTonnage": func(exercises []model.WorkoutExercise) float64 {
		var t float64
		for _, e := range exercises {
			for _, s := range e.Sets {
				if s.Weight != nil && s.Reps != nil {
					t += *s.Weight * float64(*s.Reps)
				}
			}
		}
		return t
	},
	"totalSets": func(exercises []model.WorkoutExercise) int {
		n := 0
		for _, e := range exercises {
			n += len(e.Sets)
		}
		return n
	},
	"topSetIdx": func(sets []model.Set) int {
		best, bestVal := -1, 0.0
		for i, s := range sets {
			if s.Weight != nil && s.Reps != nil {
				if v := *s.Weight * float64(*s.Reps); v > bestVal {
					bestVal = v
					best = i
				}
			}
		}
		return best
	},
	"workoutDuration": func(startedAt, endedAt *time.Time) string {
		if startedAt == nil || endedAt == nil {
			return ""
		}
		dur := endedAt.Sub(*startedAt)
		if dur <= 0 {
			return ""
		}
		h := int(dur.Hours())
		m := int(dur.Minutes()) % 60
		if h > 0 {
			return fmt.Sprintf("%dч %dм", h, m)
		}
		return fmt.Sprintf("%dм", m)
	},
}

func renderTemplate(w http.ResponseWriter, r *http.Request, name string, data map[string]any) {
	if data == nil {
		data = map[string]any{}
	}
	if _, ok := data["CurrentUser"]; !ok {
		data["CurrentUser"] = middleware.UserFromContext(r.Context())
	}
	if _, ok := data["CurrentPath"]; !ok {
		data["CurrentPath"] = r.URL.Path
	}

	// Собираем все шаблоны: base + страница + партиалы упражнений
	files := []string{
		filepath.Join("web", "templates", "base.html"),
		filepath.Join("web", "templates", name),
	}
	// Подключаем партиалы тренировок если они нужны форме
	partials := []string{
		"web/templates/workouts/partials/exercise_block.html",
		"web/templates/workouts/partials/exercise_row.html",
		"web/templates/workouts/partials/set_row.html",
		"web/templates/workouts/partials/active_set_row.html",
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
	renderPartialWith(w, r, name, nil, data)
}

func renderPartialWith(w http.ResponseWriter, r *http.Request, name string, extra []string, data map[string]any) {
	if data == nil {
		data = map[string]any{}
	}
	files := []string{filepath.Join("web", "templates", name)}
	files = append(files, extra...)
	tmpl, err := template.New("").Funcs(tmplFuncs).ParseFiles(files...)
	if err != nil {
		http.Error(w, "Ошибка шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.ExecuteTemplate(w, filepath.Base(name), data)
}
