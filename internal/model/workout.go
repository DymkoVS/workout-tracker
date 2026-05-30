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
	WorkoutType string
	WorkoutDate time.Time
	Notes       string
	Wellbeing   *int
	Exercises   []WorkoutExercise
	CreatedAt   time.Time
	UpdatedAt   time.Time
	StartedAt   *time.Time
	EndedAt     *time.Time
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
	ID                uuid.UUID
	WorkoutExerciseID uuid.UUID
	SetNum            int
	Weight            *float64
	Reps              *int
	RPE               *float64
	RestSeconds       *int
	Notes             string
	Done              bool
}

type Exercise struct {
	ID          uuid.UUID
	Name        string
	MuscleGroup string
	Description string
	CreatedAt   time.Time
}

// WorkoutCardData — workout with precomputed display stats for list/card views
type WorkoutCardData struct {
	Workout
	ExerciseCount int
	SetCount      int
	Tonnage       float64 // kg
}

// RecentPR — personal record set within the last 30 days.
type RecentPR struct {
	ExerciseName string
	NewWeight    float64
	Reps         int
	Delta        float64 // how much heavier than the previous all-time max (0 = first time)
}

// FormSet — данные одного подхода из HTML-формы (строки, до парсинга)
type FormSet struct {
	Weight      string
	Reps        string
	RPE         string
	RestSeconds string
	Notes       string
}

type WorkoutMedia struct {
	ID           uuid.UUID
	WorkoutID    uuid.UUID
	Filename     string
	OriginalName string
	MimeType     string
	SizeBytes    int
	CreatedAt    time.Time
}

func (m WorkoutMedia) IsVideo() bool {
	return len(m.MimeType) >= 5 && m.MimeType[:5] == "video"
}

type ClientExerciseSummary struct {
	Name         string
	SessionCount int
	LastDate     time.Time
}

type ProgressSession struct {
	WorkoutID    uuid.UUID
	WorkoutDate  time.Time
	WorkoutTitle string
	Sets         []Set
	MaxWeight    float64
	TotalVolume  float64
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
	Type      string // "Сила" | "Кардио" | "Аксессуар"
	UsedCount int    // number of workouts created from this template
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
