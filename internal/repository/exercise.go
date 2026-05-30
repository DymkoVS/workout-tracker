package repository

import (
	"context"
	"workout-tracker/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ExerciseRepository struct {
	db *pgxpool.Pool
}

func NewExerciseRepository(db *pgxpool.Pool) *ExerciseRepository {
	return &ExerciseRepository{db: db}
}

func (r *ExerciseRepository) List(ctx context.Context) ([]model.Exercise, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, COALESCE(muscle_group,''), COALESCE(description,''), created_at
		FROM exercises
		ORDER BY lower(name)`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Exercise
	for rows.Next() {
		var e model.Exercise
		if err := rows.Scan(&e.ID, &e.Name, &e.MuscleGroup, &e.Description, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, nil
}

func (r *ExerciseRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Exercise, error) {
	var e model.Exercise
	err := r.db.QueryRow(ctx, `
		SELECT id, name, COALESCE(muscle_group,''), COALESCE(description,''), created_at
		FROM exercises WHERE id=$1`, id).
		Scan(&e.ID, &e.Name, &e.MuscleGroup, &e.Description, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *ExerciseRepository) Create(ctx context.Context, name, muscleGroup, description string) (*model.Exercise, error) {
	var e model.Exercise
	err := r.db.QueryRow(ctx, `
		INSERT INTO exercises (name, muscle_group, description)
		VALUES ($1, NULLIF($2,''), NULLIF($3,''))
		RETURNING id, name, COALESCE(muscle_group,''), COALESCE(description,''), created_at`,
		name, muscleGroup, description).
		Scan(&e.ID, &e.Name, &e.MuscleGroup, &e.Description, &e.CreatedAt)
	return &e, err
}

func (r *ExerciseRepository) Update(ctx context.Context, id uuid.UUID, name, muscleGroup, description string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE exercises
		SET name=$2, muscle_group=NULLIF($3,''), description=NULLIF($4,'')
		WHERE id=$1`,
		id, name, muscleGroup, description)
	return err
}

func (r *ExerciseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM exercises WHERE id=$1`, id)
	return err
}

// Search returns exercises whose name starts with q (case-insensitive), up to limit.
func (r *ExerciseRepository) Search(ctx context.Context, q string, limit int) ([]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT name FROM exercises
		WHERE name ILIKE $1
		ORDER BY lower(name)
		LIMIT $2`, q+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, nil
}
