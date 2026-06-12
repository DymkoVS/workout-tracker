package main

import (
	"context"
	"log"
	"os"
	"workout-tracker/internal/config"
	"workout-tracker/internal/db"
	"workout-tracker/internal/model"
	"workout-tracker/internal/repository"
)

func main() {
	cfg := config.Load()

	// Предохранитель: seed пересоздаёт admin со слабым паролем — это только
	// для локальной разработки. Против прода без явного SEED_ALLOW=1 не работаем.
	if os.Getenv("SEED_ALLOW") != "1" {
		log.Fatal("seed: задайте SEED_ALLOW=1 (инструмент пересоздаёт admin/admin123 — только для локальной разработки)")
	}

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
