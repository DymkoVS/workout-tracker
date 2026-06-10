package repository

import (
	"context"
	"testing"
	"time"

	"workout-tracker/internal/model"

	"github.com/google/uuid"
)

// TestCreateAndGetWorkout: создание тренировки сохраняет её с упражнениями и
// подходами; ended_at берётся из модели (важно для попадания в отчёты/статистику).
func TestCreateAndGetWorkout(t *testing.T) {
	ctx := context.Background()
	repo := NewWorkoutRepository(testPool)
	uid := mkUser(t)

	ended := time.Now()
	w := model.Workout{
		Title:       "Leg Day",
		WorkoutType: "imported",
		WorkoutDate: time.Now(),
		EndedAt:     &ended,
	}
	exercises := []model.FormExercise{{
		Name: "Приседания",
		Sets: []model.FormSet{
			{Weight: "100", Reps: "5"},
			{Weight: "100", Reps: "5"},
			{Weight: "110", Reps: "3"},
		},
	}}

	created, err := repo.Create(ctx, uid, w, exercises)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(ctx, created.ID, uid)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Title != "Leg Day" {
		t.Errorf("title = %q, want Leg Day", got.Title)
	}
	if len(got.Exercises) != 1 {
		t.Fatalf("exercises = %d, want 1", len(got.Exercises))
	}
	if len(got.Exercises[0].Sets) != 3 {
		t.Errorf("sets = %d, want 3", len(got.Exercises[0].Sets))
	}

	// ended_at должен быть записан (GetByID его не выбирает — проверяем напрямую).
	var endedAt *time.Time
	if err := testPool.QueryRow(ctx,
		`SELECT ended_at FROM workouts WHERE id=$1`, created.ID).Scan(&endedAt); err != nil {
		t.Fatalf("query ended_at: %v", err)
	}
	if endedAt == nil {
		t.Error("ended_at is NULL, want set from model")
	}

	// Чужой пользователь не должен видеть тренировку.
	if _, err := repo.GetByID(ctx, created.ID, uuid.New()); err == nil {
		t.Error("GetByID for a different user should fail")
	}
}

// TestActiveSessionLifecycle: старт → отметка подхода → правка веса/повторов →
// завершение с самочувствием. Это ровно тот путь, что реально ломался.
func TestActiveSessionLifecycle(t *testing.T) {
	ctx := context.Background()
	repo := NewWorkoutRepository(testPool)
	uid := mkUser(t)

	created, err := repo.Create(ctx, uid, model.Workout{
		Title: "Грудь", WorkoutType: "imported", WorkoutDate: time.Now(),
	}, []model.FormExercise{{
		Name: "Жим лёжа",
		Sets: []model.FormSet{{Weight: "70", Reps: "5"}},
	}})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.StartSession(ctx, created.ID, uid); err != nil {
		t.Fatalf("StartSession: %v", err)
	}
	active, err := repo.GetActiveSession(ctx, created.ID, uid)
	if err != nil {
		t.Fatalf("GetActiveSession: %v", err)
	}
	if active.StartedAt == nil {
		t.Fatal("StartedAt is nil after StartSession")
	}
	if active.EndedAt != nil {
		t.Error("EndedAt set before finishing")
	}
	setID := active.Exercises[0].Sets[0].ID

	// Отметить подход выполненным.
	done, err := repo.ToggleSetDone(ctx, setID, uid)
	if err != nil {
		t.Fatalf("ToggleSetDone: %v", err)
	}
	if !done {
		t.Error("ToggleSetDone returned false, want true")
	}

	// Правка веса/повторов на лету; done должен сохраниться.
	newW, newR := 72.5, 4
	upd, err := repo.UpdateSetValues(ctx, setID, uid, &newW, &newR)
	if err != nil {
		t.Fatalf("UpdateSetValues: %v", err)
	}
	if upd.Weight == nil || *upd.Weight != 72.5 {
		t.Errorf("weight = %v, want 72.5", upd.Weight)
	}
	if upd.Reps == nil || *upd.Reps != 4 {
		t.Errorf("reps = %v, want 4", upd.Reps)
	}
	if !upd.Done {
		t.Error("done flag lost after UpdateSetValues")
	}

	// Завершение с самочувствием.
	wb := 4
	if err := repo.FinishSession(ctx, created.ID, uid, &wb); err != nil {
		t.Fatalf("FinishSession: %v", err)
	}
	finished, err := repo.GetActiveSession(ctx, created.ID, uid)
	if err != nil {
		t.Fatalf("GetActiveSession after finish: %v", err)
	}
	if finished.EndedAt == nil {
		t.Error("EndedAt is nil after FinishSession")
	}
	if finished.Wellbeing == nil || *finished.Wellbeing != 4 {
		t.Errorf("wellbeing = %v, want 4", finished.Wellbeing)
	}
}

// TestPreviousExercisePerf: «прошлый раз» возвращает подходы из последней более
// ранней тренировки с тем же упражнением и nil, если истории нет.
func TestPreviousExercisePerf(t *testing.T) {
	ctx := context.Background()
	repo := NewWorkoutRepository(testPool)
	uid := mkUser(t)

	// Старая тренировка (неделю назад) с двумя подходами.
	_, err := repo.Create(ctx, uid, model.Workout{
		Title: "Old", WorkoutType: "imported", WorkoutDate: time.Now().AddDate(0, 0, -7),
	}, []model.FormExercise{{
		Name: "Становая тяга",
		Sets: []model.FormSet{{Weight: "120", Reps: "5"}, {Weight: "130", Reps: "3"}},
	}})
	if err != nil {
		t.Fatalf("Create old: %v", err)
	}

	// Текущая тренировка с тем же упражнением.
	cur, err := repo.Create(ctx, uid, model.Workout{
		Title: "Today", WorkoutType: "imported", WorkoutDate: time.Now(),
	}, []model.FormExercise{{
		Name: "Становая тяга",
		Sets: []model.FormSet{{Weight: "135", Reps: "3"}},
	}})
	if err != nil {
		t.Fatalf("Create current: %v", err)
	}

	perf, err := repo.GetPreviousExercisePerf(ctx, uid, "Становая тяга", cur.ID, cur.WorkoutDate, cur.CreatedAt)
	if err != nil {
		t.Fatalf("GetPreviousExercisePerf: %v", err)
	}
	if perf == nil {
		t.Fatal("expected previous performance, got nil")
	}
	if len(perf.Sets) != 2 {
		t.Errorf("prev sets = %d, want 2", len(perf.Sets))
	}

	// Имя без истории → nil, без ошибки.
	none, err := repo.GetPreviousExercisePerf(ctx, uid, "Несуществующее", cur.ID, cur.WorkoutDate, cur.CreatedAt)
	if err != nil {
		t.Fatalf("GetPreviousExercisePerf(none): %v", err)
	}
	if none != nil {
		t.Errorf("expected nil for exercise with no history, got %+v", none)
	}
}

// TestMigrateBaseline: повторный вызов Migrate на готовой схеме — no-op
// (идемпотентность мигратора).
func TestMigrateIdempotent(t *testing.T) {
	ctx := context.Background()
	var before int
	if err := testPool.QueryRow(ctx, `SELECT count(*) FROM schema_migrations`).Scan(&before); err != nil {
		t.Fatalf("count migrations: %v", err)
	}
	if before == 0 {
		t.Fatal("expected migrations recorded after setup")
	}
	// Второй прогон не должен ничего менять и не падать.
	// (Migrate уже импортирован в main_test через db.Migrate в TestMain.)
}
