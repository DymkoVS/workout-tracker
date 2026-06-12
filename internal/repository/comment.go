package repository

import (
	"context"
	"workout-tracker/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CommentRepository struct {
	db *pgxpool.Pool
}

func NewCommentRepository(db *pgxpool.Pool) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) Add(ctx context.Context, workoutID, authorID uuid.UUID, body string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO workout_comments (workout_id, author_id, body) VALUES ($1, $2, $3)`,
		workoutID, authorID, body)
	return err
}

func (r *CommentRepository) ListForWorkout(ctx context.Context, workoutID uuid.UUID) ([]*model.WorkoutComment, error) {
	rows, err := r.db.Query(ctx, `
		SELECT c.id, c.workout_id, c.author_id, c.body, c.created_at,
		       COALESCE(NULLIF(u.full_name, ''), u.login), u.role
		FROM workout_comments c
		JOIN users u ON u.id = c.author_id
		WHERE c.workout_id = $1
		ORDER BY c.created_at`,
		workoutID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*model.WorkoutComment
	for rows.Next() {
		c := &model.WorkoutComment{}
		if err := rows.Scan(&c.ID, &c.WorkoutID, &c.AuthorID, &c.Body, &c.CreatedAt,
			&c.AuthorName, &c.AuthorRole); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, rows.Err()
}
