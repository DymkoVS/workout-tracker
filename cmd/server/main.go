package main

import (
	"log"
	"net/http"
	"workout-tracker/internal/config"
	"workout-tracker/internal/db"
	"workout-tracker/internal/handler"
	"workout-tracker/internal/middleware"
	"workout-tracker/internal/repository"
	"workout-tracker/internal/session"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg := config.Load()

	pool, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	sessionStore := session.NewStore(pool)
	userRepo := repository.NewUserRepository(pool)
	gymRepo := repository.NewGymRepository(pool)
	workoutRepo := repository.NewWorkoutRepository(pool)

	authHandler := handler.NewAuthHandler(userRepo, sessionStore)
	adminHandler := handler.NewAdminHandler(userRepo)
	workoutHandler := handler.NewWorkoutHandler(workoutRepo, gymRepo)
	authMiddleware := middleware.NewAuthMiddleware(sessionStore, userRepo)

	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.CleanPath)

	// Публичные маршруты
	r.Get("/login", authHandler.LoginPage)
	r.Post("/login", authHandler.Login)
	r.Post("/logout", authHandler.Logout)

	// Защищённые маршруты
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.RequireAuth)

		r.Get("/", handler.Dashboard)

		// Тренировки
		r.Get("/workouts", workoutHandler.List)
		r.Get("/workouts/new", workoutHandler.NewForm)
		r.Post("/workouts", workoutHandler.Create)
		r.Get("/workouts/{id}", workoutHandler.Show)
		r.Get("/workouts/{id}/edit", workoutHandler.EditForm)
		r.Post("/workouts/{id}", workoutHandler.Update)
		r.Post("/workouts/{id}/delete", workoutHandler.Delete)

		// HTMX-партиалы для формы
		r.Get("/workouts/htmx/add-exercise", workoutHandler.AddExerciseRow)
		r.Get("/workouts/htmx/add-set", workoutHandler.AddSetRow)

		// Маршруты администратора
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequireAdmin)
			r.Get("/admin/users", adminHandler.UsersList)
			r.Get("/admin/users/new", adminHandler.NewUserForm)
			r.Post("/admin/users", adminHandler.CreateUser)
			r.Get("/admin/users/{id}/edit", adminHandler.EditUserForm)
			r.Post("/admin/users/{id}", adminHandler.UpdateUser)
		})
	})

	log.Printf("Сервер запущен на %s", cfg.ListenAddr)
	if err := http.ListenAndServe(cfg.ListenAddr, r); err != nil {
		log.Fatal(err)
	}
}
