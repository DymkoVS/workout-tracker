package repository

import (
	"context"
	"fmt"
	"strings"
	"workout-tracker/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AnalyticsFilter holds optional filters for analytics queries.
type AnalyticsFilter struct {
	GymID       *uuid.UUID
	WorkoutType string
}

func (f AnalyticsFilter) where(startIdx int) (clauses []string, args []any) {
	if f.GymID != nil {
		args = append(args, *f.GymID)
		clauses = append(clauses, fmt.Sprintf("w.gym_id = $%d", startIdx+len(args)-1))
	}
	if f.WorkoutType != "" {
		args = append(args, f.WorkoutType)
		clauses = append(clauses, fmt.Sprintf("w.workout_type = $%d", startIdx+len(args)-1))
	}
	return
}

type AnalyticsRepository struct {
	db *pgxpool.Pool
}

func NewAnalyticsRepository(db *pgxpool.Pool) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

func (r *AnalyticsRepository) TonnageByDate(ctx context.Context, userID uuid.UUID, f AnalyticsFilter) ([]model.AnalyticsPoint, error) {
	extra, fArgs := f.where(2)
	extraSQL := ""
	if len(extra) > 0 {
		extraSQL = " AND " + strings.Join(extra, " AND ")
	}
	args := append([]any{userID}, fArgs...)
	rows, err := r.db.Query(ctx, `
		SELECT w.workout_date, COALESCE(SUM(s.weight * s.reps), 0)
		FROM workouts w
		JOIN workout_exercises e ON e.workout_id = w.id
		JOIN sets s ON s.workout_exercise_id = e.id
		WHERE w.user_id = $1
		  AND s.weight IS NOT NULL AND s.reps IS NOT NULL
		  AND w.workout_date >= CURRENT_DATE - INTERVAL '90 days'`+extraSQL+`
		GROUP BY w.workout_date
		ORDER BY w.workout_date`, args...)
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

func (r *AnalyticsRepository) WorkoutFrequency(ctx context.Context, userID uuid.UUID, f AnalyticsFilter) ([]model.FrequencyPoint, error) {
	extra, fArgs := f.where(2)
	extraSQL := ""
	if len(extra) > 0 {
		extraSQL = " AND " + strings.Join(extra, " AND ")
	}
	args := append([]any{userID}, fArgs...)
	rows, err := r.db.Query(ctx, `
		SELECT TO_CHAR(DATE_TRUNC('week', w.workout_date), 'DD.MM') AS week_start, COUNT(*)::int AS count
		FROM workouts w
		WHERE w.user_id = $1
		  AND w.ended_at IS NOT NULL
		  AND w.workout_date >= CURRENT_DATE - INTERVAL '8 weeks'`+extraSQL+`
		GROUP BY DATE_TRUNC('week', w.workout_date)
		ORDER BY DATE_TRUNC('week', w.workout_date)`, args...)
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

func (r *AnalyticsRepository) ExerciseNames(ctx context.Context, userID uuid.UUID, f AnalyticsFilter) ([]string, error) {
	extra, fArgs := f.where(2)
	extraSQL := ""
	if len(extra) > 0 {
		extraSQL = " AND " + strings.Join(extra, " AND ")
	}
	args := append([]any{userID}, fArgs...)
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT e.name
		FROM workout_exercises e
		JOIN workouts w ON w.id = e.workout_id
		WHERE w.user_id = $1`+extraSQL+`
		ORDER BY e.name`, args...)
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

// TonnagePeriodTotals returns total tonnage for the current 90 days and the
// preceding 90 days (days 91–180 back). Used to compute the ▲/▼ delta.
func (r *AnalyticsRepository) TonnagePeriodTotals(ctx context.Context, userID uuid.UUID, f AnalyticsFilter) (current, prev float64, err error) {
	extra, fArgs := f.where(2)
	extraSQL := ""
	if len(extra) > 0 {
		extraSQL = " AND " + strings.Join(extra, " AND ")
	}
	args := append([]any{userID}, fArgs...)
	err = r.db.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(CASE WHEN w.workout_date >= CURRENT_DATE - INTERVAL '90 days'
			                  THEN s.weight * s.reps END), 0),
			COALESCE(SUM(CASE WHEN w.workout_date >= CURRENT_DATE - INTERVAL '180 days'
			                   AND w.workout_date < CURRENT_DATE - INTERVAL '90 days'
			                  THEN s.weight * s.reps END), 0)
		FROM workouts w
		JOIN workout_exercises e ON e.workout_id = w.id
		JOIN sets s ON s.workout_exercise_id = e.id
		WHERE w.user_id = $1
		  AND s.weight IS NOT NULL AND s.reps IS NOT NULL
		  AND w.workout_date >= CURRENT_DATE - INTERVAL '180 days'`+extraSQL,
		args...).Scan(&current, &prev)
	return
}

func (r *AnalyticsRepository) ExerciseProgress(ctx context.Context, userID uuid.UUID, exerciseName string, f AnalyticsFilter) ([]model.AnalyticsPoint, error) {
	extra, fArgs := f.where(3)
	extraSQL := ""
	if len(extra) > 0 {
		extraSQL = " AND " + strings.Join(extra, " AND ")
	}
	args := append([]any{userID, exerciseName}, fArgs...)
	rows, err := r.db.Query(ctx, `
		SELECT w.workout_date, MAX(s.weight)
		FROM workouts w
		JOIN workout_exercises e ON e.workout_id = w.id
		JOIN sets s ON s.workout_exercise_id = e.id
		WHERE w.user_id = $1
		  AND LOWER(e.name) = LOWER($2)
		  AND s.weight IS NOT NULL`+extraSQL+`
		GROUP BY w.workout_date
		ORDER BY w.workout_date`, args...)
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
