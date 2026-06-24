package handler

import (
	"bytes"
	"html/template"
	"strings"
	"testing"

	"workout-tracker/internal/model"
)

// renderTmpl исполняет шаблон ровно как прод: те же tmplFuncs, те же файлы.
func renderTmpl(t *testing.T, exec string, data map[string]any, files ...string) string {
	t.Helper()
	tmpl, err := template.New("").Funcs(tmplFuncs).ParseFiles(files...)
	if err != nil {
		t.Fatalf("parse %v: %v", files, err)
	}
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, exec, data); err != nil {
		t.Fatalf("execute %s: %v", exec, err)
	}
	return buf.String()
}

const (
	pSet   = "../../web/templates/workouts/partials/set_row.html"
	pBlock = "../../web/templates/workouts/partials/exercise_block.html"
	pRow   = "../../web/templates/workouts/partials/exercise_row.html"
)

// 2.2: единый set_row. Пустой подход (HTMX add-set) — без значений, без паники
// на отсутствующем .Set.
func TestSetRowEmpty(t *testing.T) {
	out := renderTmpl(t, "set_row.html", map[string]any{"ExIdx": 0, "SetIdx": 0}, pSet)
	if !strings.Contains(out, `name="exercises[0][sets][0][weight]"`) {
		t.Fatalf("нет имени поля веса:\n%s", out)
	}
	if !strings.Contains(out, `onclick="removeSet(this)"`) {
		t.Fatalf("нет кнопки удаления подхода")
	}
}

// set_row с заполненным model.Set (*float64/*int) — edit-путь.
func TestSetRowFilled(t *testing.T) {
	w, reps, rpe := 82.5, 6, 8.0
	set := model.Set{SetNum: 1, Weight: &w, Reps: &reps, RPE: &rpe}
	out := renderTmpl(t, "set_row.html", map[string]any{"ExIdx": 2, "SetIdx": 3, "Set": set}, pSet)
	for _, want := range []string{`name="exercises[2][sets][3][weight]"`, `value="82.5"`, `value="6"`, `value="8"`} {
		if !strings.Contains(out, want) {
			t.Fatalf("ожидали %q в:\n%s", want, out)
		}
	}
}

// 2.2: exercise_block для HTMX-add (model.FormExercise, Weight=string "") —
// должен исполниться без ошибки типов и с пустым подходом.
func TestExerciseBlockFormExercise(t *testing.T) {
	ex := model.FormExercise{Sets: []model.FormSet{{}}}
	out := renderTmpl(t, "exercise_block", map[string]any{"ExIdx": 1, "Ex": ex}, pBlock, pSet)
	if !strings.Contains(out, `name="exercises[1][name]"`) {
		t.Fatalf("нет имени упражнения:\n%s", out)
	}
	if !strings.Contains(out, `name="exercises[1][sets][0][weight]"`) {
		t.Fatalf("нет первого подхода")
	}
}

// exercise_block для edit (model.WorkoutExercise с model.Set) — derefF/derefI.
func TestExerciseBlockWorkoutExercise(t *testing.T) {
	w := 100.0
	reps := 5
	ex := model.WorkoutExercise{
		Name: "Присед", Notes: "техника",
		Sets: []model.Set{{SetNum: 1, Weight: &w, Reps: &reps}},
	}
	out := renderTmpl(t, "exercise_block", map[string]any{"ExIdx": 0, "Ex": ex}, pBlock, pSet)
	for _, want := range []string{`value="Присед"`, `value="100"`, `value="5"`, `removeExercise(this)`} {
		if !strings.Contains(out, want) {
			t.Fatalf("ожидали %q в:\n%s", want, out)
		}
	}
}

// exercise_block с nil Ex — пустой блок новой тренировки.
func TestExerciseBlockNilEx(t *testing.T) {
	out := renderTmpl(t, "exercise_block", map[string]any{"ExIdx": 0, "Ex": nil}, pBlock, pSet)
	if !strings.Contains(out, `name="exercises[0][name]"`) {
		t.Fatalf("пустой блок не отрендерился:\n%s", out)
	}
	if strings.Contains(out, "[sets][0][weight]") {
		t.Fatalf("у пустого блока не должно быть подходов:\n%s", out)
	}
}

// 2.2: exercise_row делегирует в exercise_block (как renderPartialWith с extra).
func TestExerciseRowDelegates(t *testing.T) {
	ex := model.FormExercise{Sets: []model.FormSet{{}}}
	out := renderTmpl(t, "exercise_row.html", map[string]any{"ExIdx": 3, "Ex": ex}, pRow, pBlock, pSet)
	if !strings.Contains(out, `name="exercises[3][name]"`) {
		t.Fatalf("делегация не сработала:\n%s", out)
	}
}
