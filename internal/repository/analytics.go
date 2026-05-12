package repository

import (
	"context"
	"workout-tracker/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AnalyticsRepository struct {
	db *pgxpool.Pool
}

func NewAnalyticsRepository(db *pgxpool.Pool) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

func (r *AnalyticsRepository) TonnageByDate(ctx context.Context, userID uuid.UUID) ([]model.AnalyticsPoint, error) {
	rows, err := r.db.Query(ctx, `
		SELECT w.workout_date, COALESCE(SUM(s.weight * s.reps), 0)
		FROM workouts w
		JOIN workout_exercises e ON e.workout_id = w.id
		JOIN sets s ON s.workout_exercise_id = e.id
		WHERE w.user_id = $1
		  AND s.weight IS NOT NULL AND s.reps IS NOT NULL
		  AND w.workout_date >= CURRENT_DATE - INTERVAL '90 days'
		GROUP BY w.workout_date
		ORDER BY w.workout_date`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []model.AnalyticsPoint
	for rows.Next() {
		var p model.AnalyticsPoint
		if err := rows.Scan(&p.Date, &p.Value); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	return points, nil
}

func (r *AnalyticsRepository) WorkoutFrequency(ctx context.Context, userID uuid.UUID) ([]model.FrequencyPoint, error) {
	rows, err := r.db.Query(ctx, `
		SELECT TO_CHAR(DATE_TRUNC('week', workout_date), 'DD.MM') AS week_start, COUNT(*)::int AS count
		FROM workouts
		WHERE user_id = $1
		  AND workout_date >= CURRENT_DATE - INTERVAL '8 weeks'
		GROUP BY DATE_TRUNC('week', workout_date)
		ORDER BY DATE_TRUNC('week', workout_date)`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []model.FrequencyPoint
	for rows.Next() {
		var p model.FrequencyPoint
		if err := rows.Scan(&p.Week, &p.Count); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	return points, nil
}

func (r *AnalyticsRepository) ExerciseNames(ctx context.Context, userID uuid.UUID) ([]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT e.name
		FROM workout_exercises e
		JOIN workouts w ON w.id = e.workout_id
		WHERE w.user_id = $1
		ORDER BY e.name`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, nil
}

func (r *AnalyticsRepository) ExerciseProgress(ctx context.Context, userID uuid.UUID, exerciseName string) ([]model.AnalyticsPoint, error) {
	rows, err := r.db.Query(ctx, `
		SELECT w.workout_date, MAX(s.weight)
		FROM workouts w
		JOIN workout_exercises e ON e.workout_id = w.id
		JOIN sets s ON s.workout_exercise_id = e.id
		WHERE w.user_id = $1
		  AND LOWER(e.name) = LOWER($2)
		  AND s.weight IS NOT NULL
		GROUP BY w.workout_date
		ORDER BY w.workout_date`, userID, exerciseName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []model.AnalyticsPoint
	for rows.Next() {
		var p model.AnalyticsPoint
		if err := rows.Scan(&p.Date, &p.Value); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	return points, nil
}
