package repository

import (
	"context"
	"workout-tracker/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GymRepository struct {
	db *pgxpool.Pool
}

func NewGymRepository(db *pgxpool.Pool) *GymRepository {
	return &GymRepository{db: db}
}

func (r *GymRepository) List(ctx context.Context) ([]model.Gym, error) {
	rows, err := r.db.Query(ctx, `SELECT id, name, created_at FROM gyms ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var gyms []model.Gym
	for rows.Next() {
		var g model.Gym
		if err := rows.Scan(&g.ID, &g.Name, &g.CreatedAt); err != nil {
			return nil, err
		}
		gyms = append(gyms, g)
	}
	return gyms, rows.Err()
}

func (r *GymRepository) Create(ctx context.Context, name string) (model.Gym, error) {
	var g model.Gym
	err := r.db.QueryRow(ctx,
		`INSERT INTO gyms (name) VALUES ($1) RETURNING id, name, created_at`, name,
	).Scan(&g.ID, &g.Name, &g.CreatedAt)
	return g, err
}

func (r *GymRepository) GetByID(ctx context.Context, id uuid.UUID) (model.Gym, error) {
	var g model.Gym
	err := r.db.QueryRow(ctx,
		`SELECT id, name, created_at FROM gyms WHERE id=$1`, id,
	).Scan(&g.ID, &g.Name, &g.CreatedAt)
	return g, err
}

func (r *GymRepository) Update(ctx context.Context, id uuid.UUID, name string) error {
	_, err := r.db.Exec(ctx, `UPDATE gyms SET name=$1 WHERE id=$2`, name, id)
	return err
}
