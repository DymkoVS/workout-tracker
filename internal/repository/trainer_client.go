package repository

import (
	"context"
	"workout-tracker/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TrainerClientRepository struct {
	db *pgxpool.Pool
}

func NewTrainerClientRepository(db *pgxpool.Pool) *TrainerClientRepository {
	return &TrainerClientRepository{db: db}
}

// GetClients возвращает всех клиентов тренера
func (r *TrainerClientRepository) GetClients(ctx context.Context, trainerID uuid.UUID) ([]*model.User, error) {
	rows, err := r.db.Query(ctx, `
		SELECT u.id, u.login, u.email, u.password_hash, u.full_name, u.role,
		       u.is_admin, u.is_active, u.created_at, u.updated_at
		FROM users u
		JOIN trainer_clients tc ON tc.client_id = u.id
		WHERE tc.trainer_id = $1
		ORDER BY u.full_name, u.login`, trainerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []*model.User
	for rows.Next() {
		u := &model.User{}
		var email *string
		if err := rows.Scan(&u.ID, &u.Login, &email, &u.PasswordHash,
			&u.FullName, &u.Role, &u.IsAdmin, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		if email != nil {
			u.Email = *email
		}
		users = append(users, u)
	}
	return users, nil
}

// GetTrainers возвращает всех тренеров клиента
func (r *TrainerClientRepository) GetTrainers(ctx context.Context, clientID uuid.UUID) ([]*model.User, error) {
	rows, err := r.db.Query(ctx, `
		SELECT u.id, u.login, u.email, u.password_hash, u.full_name, u.role,
		       u.is_admin, u.is_active, u.created_at, u.updated_at
		FROM users u
		JOIN trainer_clients tc ON tc.trainer_id = u.id
		WHERE tc.client_id = $1
		ORDER BY u.full_name, u.login`, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []*model.User
	for rows.Next() {
		u := &model.User{}
		var email *string
		if err := rows.Scan(&u.ID, &u.Login, &email, &u.PasswordHash,
			&u.FullName, &u.Role, &u.IsAdmin, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		if email != nil {
			u.Email = *email
		}
		users = append(users, u)
	}
	return users, nil
}

// IsAssigned проверяет, назначен ли клиент к тренеру
func (r *TrainerClientRepository) IsAssigned(ctx context.Context, trainerID, clientID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM trainer_clients WHERE trainer_id=$1 AND client_id=$2)`,
		trainerID, clientID,
	).Scan(&exists)
	return exists, err
}

// Assign назначает клиента к тренеру
func (r *TrainerClientRepository) Assign(ctx context.Context, trainerID, clientID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO trainer_clients (trainer_id, client_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		trainerID, clientID)
	return err
}

// Unassign убирает клиента от тренера
func (r *TrainerClientRepository) Unassign(ctx context.Context, trainerID, clientID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM trainer_clients WHERE trainer_id=$1 AND client_id=$2`,
		trainerID, clientID)
	return err
}

// GetAllTrainers возвращает всех тренеров (для админа)
func (r *TrainerClientRepository) GetAllTrainers(ctx context.Context) ([]*model.User, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, login, email, password_hash, full_name, role, is_admin, is_active, created_at, updated_at
		FROM users WHERE role = 'trainer' AND is_active = TRUE
		ORDER BY full_name, login`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []*model.User
	for rows.Next() {
		u := &model.User{}
		var email *string
		if err := rows.Scan(&u.ID, &u.Login, &email, &u.PasswordHash,
			&u.FullName, &u.Role, &u.IsAdmin, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		if email != nil {
			u.Email = *email
		}
		users = append(users, u)
	}
	return users, nil
}

// GetAllClients возвращает всех клиентов (для админа)
func (r *TrainerClientRepository) GetAllClients(ctx context.Context) ([]*model.User, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, login, email, password_hash, full_name, role, is_admin, is_active, created_at, updated_at
		FROM users WHERE role = 'client' AND is_active = TRUE
		ORDER BY full_name, login`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []*model.User
	for rows.Next() {
		u := &model.User{}
		var email *string
		if err := rows.Scan(&u.ID, &u.Login, &email, &u.PasswordHash,
			&u.FullName, &u.Role, &u.IsAdmin, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		if email != nil {
			u.Email = *email
		}
		users = append(users, u)
	}
	return users, nil
}
