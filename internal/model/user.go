package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	RoleTrainer = "trainer"
	RoleClient  = "client"
)

type User struct {
	ID           uuid.UUID
	Login        string
	Email        string
	PasswordHash string
	FullName     string
	Role         string
	IsAdmin      bool
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (u *User) IsTrainer() bool { return u.Role == RoleTrainer }
func (u *User) IsClient() bool  { return u.Role == RoleClient }

func (u *User) DisplayName() string {
	if u.FullName != "" {
		return u.FullName
	}
	return u.Login
}

type CreateUserInput struct {
	Login    string
	Email    string
	Password string
	FullName string
	Role     string
	IsAdmin  bool
}
