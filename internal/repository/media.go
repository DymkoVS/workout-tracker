package repository

import (
	"context"
	"workout-tracker/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MediaRepository struct {
	db *pgxpool.Pool
}

func NewMediaRepository(db *pgxpool.Pool) *MediaRepository {
	return &MediaRepository{db: db}
}

func (r *MediaRepository) Create(ctx context.Context, workoutID uuid.UUID, filename, originalName, mimeType string, sizeBytes int) (model.WorkoutMedia, error) {
	var m model.WorkoutMedia
	err := r.db.QueryRow(ctx, `
		INSERT INTO workout_media (workout_id, filename, original_name, mime_type, size_bytes)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, workout_id, filename, original_name, mime_type, size_bytes, created_at`,
		workoutID, filename, originalName, mimeType, sizeBytes,
	).Scan(&m.ID, &m.WorkoutID, &m.Filename, &m.OriginalName, &m.MimeType, &m.SizeBytes, &m.CreatedAt)
	return m, err
}

func (r *MediaRepository) ListForWorkout(ctx context.Context, workoutID uuid.UUID) ([]model.WorkoutMedia, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, workout_id, filename, original_name, mime_type, size_bytes, created_at
		FROM workout_media WHERE workout_id = $1 ORDER BY created_at`, workoutID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var media []model.WorkoutMedia
	for rows.Next() {
		var m model.WorkoutMedia
		if err := rows.Scan(&m.ID, &m.WorkoutID, &m.Filename, &m.OriginalName, &m.MimeType, &m.SizeBytes, &m.CreatedAt); err != nil {
			return nil, err
		}
		media = append(media, m)
	}
	return media, nil
}

func (r *MediaRepository) GetByID(ctx context.Context, id, workoutID uuid.UUID) (*model.WorkoutMedia, error) {
	var m model.WorkoutMedia
	err := r.db.QueryRow(ctx, `
		SELECT id, workout_id, filename, original_name, mime_type, size_bytes, created_at
		FROM workout_media WHERE id = $1 AND workout_id = $2`, id, workoutID,
	).Scan(&m.ID, &m.WorkoutID, &m.Filename, &m.OriginalName, &m.MimeType, &m.SizeBytes, &m.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *MediaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM workout_media WHERE id = $1`, id)
	return err
}
