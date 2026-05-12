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
	tcRepo := repository.NewTrainerClientRepository(pool)
	templateRepo := repository.NewTemplateRepository(pool)
	analyticsRepo := repository.NewAnalyticsRepository(pool)

	authHandler := handler.NewAuthHandler(userRepo, sessionStore)
	adminHandler := handler.NewAdminHandler(userRepo, tcRepo)
	workoutHandler := handler.NewWorkoutHandler(workoutRepo, gymRepo, tcRepo, userRepo)
	gymHandler := handler.NewGymHandler(gymRepo)
	trainerHandler := handler.NewTrainerHandler(tcRepo, workoutRepo, userRepo)
	templateHandler := handler.NewTemplateHandler(templateRepo, tcRepo, gymRepo)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsRepo, tcRepo)
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

		// Залы
		r.Get("/gyms", gymHandler.List)
		r.Get("/gyms/new", gymHandler.NewForm)
		r.Post("/gyms", gymHandler.Create)
		r.Get("/gyms/{id}/edit", gymHandler.EditForm)
		r.Post("/gyms/{id}", gymHandler.Update)

		// Тренер: список клиентов и их тренировки
		r.Get("/trainer/clients", trainerHandler.Clients)
		r.Get("/trainer/clients/{id}/workouts", trainerHandler.ClientWorkouts)

		// Шаблоны тренировок (только для тренеров)
		r.Get("/templates", templateHandler.List)
		r.Get("/templates/new", templateHandler.NewForm)
		r.Post("/templates", templateHandler.Create)
		r.Get("/templates/{id}", templateHandler.Show)
		r.Get("/templates/{id}/edit", templateHandler.EditForm)
		r.Post("/templates/{id}", templateHandler.Update)
		r.Post("/templates/{id}/delete", templateHandler.Delete)
		r.Get("/templates/{id}/apply", templateHandler.ApplyForm)
		r.Post("/templates/{id}/apply", templateHandler.Apply)

		// Аналитика
		r.Get("/analytics", analyticsHandler.Index)
		r.Get("/analytics/exercise-data", analyticsHandler.ExerciseData)

		// Маршруты администратора
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequireAdmin)
			r.Get("/admin/users", adminHandler.UsersList)
			r.Get("/admin/users/new", adminHandler.NewUserForm)
			r.Post("/admin/users", adminHandler.CreateUser)
			r.Get("/admin/users/{id}/edit", adminHandler.EditUserForm)
			r.Post("/admin/users/{id}", adminHandler.UpdateUser)
			r.Get("/admin/assign", adminHandler.AssignPage)
			r.Post("/admin/assign", adminHandler.Assign)
			r.Post("/admin/assign/{trainerID}/{clientID}/remove", adminHandler.Unassign)
		})
	})

	log.Printf("Сервер запущен на %s", cfg.ListenAddr)
	if err := http.ListenAndServe(cfg.ListenAddr, r); err != nil {
		log.Fatal(err)
	}
}
