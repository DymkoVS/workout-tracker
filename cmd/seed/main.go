package main

import (
	"context"
	"log"
	"workout-tracker/internal/config"
	"workout-tracker/internal/db"
	"workout-tracker/internal/model"
	"workout-tracker/internal/repository"
)

func main() {
	cfg := config.Load()
	pool, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	users := repository.NewUserRepository(pool)
	ctx := context.Background()

	// Удаляем старого admin и создаём заново с правильным паролем
	pool.Exec(ctx, `DELETE FROM users WHERE login = 'admin'`)

	u, err := users.Create(ctx, model.CreateUserInput{
		Login:    "admin",
		Email:    "admin@localhost",
		Password: "admin123",
		FullName: "Администратор",
		Role:     model.RoleTrainer,
		IsAdmin:  true,
	})
	if err != nil {
		log.Fatalf("create admin: %v", err)
	}
	log.Printf("Создан пользователь: %s (id: %s)", u.Login, u.ID)
}
