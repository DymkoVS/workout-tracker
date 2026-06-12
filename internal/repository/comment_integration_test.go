package repository

import (
	"context"
	"testing"
	"time"

	"workout-tracker/internal/model"
)

// TestWorkoutComments: добавление и чтение комментариев к тренировке;
// автор подтягивается с именем и ролью, порядок — по created_at.
func TestWorkoutComments(t *testing.T) {
	ctx := context.Background()
	workouts := NewWorkoutRepository(testPool)
	comments := NewCommentRepository(testPool)
	uid := mkUser(t)

	ended := time.Now()
	w, err := workouts.Create(ctx, uid, model.Workout{
		Title:       "Спина",
		WorkoutType: "regular",
		WorkoutDate: time.Now(),
		EndedAt:     &ended,
	}, nil)
	if err != nil {
		t.Fatalf("create workout: %v", err)
	}

	// Пустой список — без ошибок.
	list, err := comments.ListForWorkout(ctx, w.ID)
	if err != nil {
		t.Fatalf("ListForWorkout (empty): %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("want 0 comments, got %d", len(list))
	}

	if err := comments.Add(ctx, w.ID, uid, "Первый"); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := comments.Add(ctx, w.ID, uid, "Второй"); err != nil {
		t.Fatalf("Add: %v", err)
	}

	list, err = comments.ListForWorkout(ctx, w.ID)
	if err != nil {
		t.Fatalf("ListForWorkout: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("want 2 comments, got %d", len(list))
	}
	if list[0].Body != "Первый" || list[1].Body != "Второй" {
		t.Errorf("wrong order: %q, %q", list[0].Body, list[1].Body)
	}
	if list[0].AuthorName != "Test User" {
		t.Errorf("AuthorName = %q, want Test User", list[0].AuthorName)
	}
	if list[0].AuthorRole != "client" {
		t.Errorf("AuthorRole = %q, want client", list[0].AuthorRole)
	}
}
