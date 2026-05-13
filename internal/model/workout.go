package model

import (
	"time"

	"github.com/google/uuid"
)

type Gym struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
}

type Workout struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	TrainerID   *uuid.UUID
	GymID       *uuid.UUID
	GymName     string
	Title       string
	WorkoutDate time.Time
	Notes       string
	Wellbeing   *int
	Exercises   []WorkoutExercise
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type WorkoutExercise struct {
	ID        uuid.UUID
	WorkoutID uuid.UUID
	Name      string
	OrderNum  int
	Notes     string
	Sets      []Set
}

type Set struct {
	ID                 uuid.UUID
	WorkoutExerciseID  uuid.UUID
	SetNum             int
	Weight             *float64
	Reps               *int
	RPE                *float64
	RestSeconds        *int
	Notes              string
}

// WorkoutCardData — workout with precomputed display stats for list/card views
type WorkoutCardData struct {
	Workout
	ExerciseCount int
	SetCount      int
	Tonnage       float64 // kg
}

// FormSet — данные одного подхода из HTML-формы (строки, до парсинга)
type FormSet struct {
	Weight      string
	Reps        string
	RPE         string
	RestSeconds string
	Notes       string
}

// FormExercise — данные одного упражнения из HTML-формы
type FormExercise struct {
	Name  string
	Notes string
	Sets  []FormSet
}

type WorkoutTemplate struct {
	ID        uuid.UUID
	TrainerID uuid.UUID
	Title     string
	Notes     string
	Exercises []TemplateExercise
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TemplateExercise struct {
	ID         uuid.UUID
	TemplateID uuid.UUID
	Name       string
	OrderNum   int
	Notes      string
	Sets       []TemplateSet
}

type TemplateSet struct {
	ID                 uuid.UUID
	TemplateExerciseID uuid.UUID
	SetNum             int
	Weight             *float64
	Reps               *int
	RPE                *float64
	RestSeconds        *int
	Notes              string
}
