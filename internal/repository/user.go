package repository

import (
	"context"
	"fmt"
	"workout-tracker/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return r.scanOne(ctx,
		`SELECT id, login, email, password_hash, full_name, role, is_admin, is_active, created_at, updated_at
		 FROM users WHERE id = $1`, id)
}

func (r *UserRepository) GetByLogin(ctx context.Context, login string) (*model.User, error) {
	return r.scanOne(ctx,
		`SELECT id, login, email, password_hash, full_name, role, is_admin, is_active, created_at, updated_at
		 FROM users WHERE login = $1`, login)
}

func (r *UserRepository) List(ctx context.Context) ([]*model.User, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, login, email, password_hash, full_name, role, is_admin, is_active, created_at, updated_at
		 FROM users ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (*model.User, error) {
		return scanUser(row)
	})
}

func (r *UserRepository) Create(ctx context.Context, in model.CreateUserInput) (*model.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	var email *string
	if in.Email != "" {
		email = &in.Email
	}
	return r.scanOne(ctx,
		`INSERT INTO users (login, email, password_hash, full_name, role, is_admin)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, login, email, password_hash, full_name, role, is_admin, is_active, created_at, updated_at`,
		in.Login, email, string(hash), in.FullName, in.Role, in.IsAdmin,
	)
}

func (r *UserRepository) Update(ctx context.Context, id uuid.UUID, login, email, fullName, role string, isAdmin, isActive bool) error {
	var emailPtr *string
	if email != "" {
		emailPtr = &email
	}
	_, err := r.db.Exec(ctx,
		`UPDATE users SET login=$1, email=$2, full_name=$3, role=$4, is_admin=$5, is_active=$6, updated_at=NOW()
		 WHERE id=$7`,
		login, emailPtr, fullName, role, isAdmin, isActive, id,
	)
	return err
}

func (r *UserRepository) SetPassword(ctx context.Context, id uuid.UUID, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, `UPDATE users SET password_hash=$1, updated_at=NOW() WHERE id=$2`, string(hash), id)
	return err
}

func (r *UserRepository) CheckPassword(u *model.User, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) == nil
}

func (r *UserRepository) scanOne(ctx context.Context, query string, args ...any) (*model.User, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (*model.User, error) {
		return scanUser(row)
	})
}

func scanUser(row pgx.CollectableRow) (*model.User, error) {
	u := &model.User{}
	var email *string
	err := row.Scan(
		&u.ID, &u.Login, &email, &u.PasswordHash,
		&u.FullName, &u.Role, &u.IsAdmin, &u.IsActive,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if email != nil {
		u.Email = *email
	}
	return u, err
}
