package repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
	"workout-tracker/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WorkoutFilter struct {
	DateFrom     *time.Time
	DateTo       *time.Time
	GymID        *uuid.UUID
	ExerciseName string
}

func (f WorkoutFilter) IsActive() bool {
	return f.DateFrom != nil || f.DateTo != nil || f.GymID != nil || f.ExerciseName != ""
}

type WorkoutRepository struct {
	db *pgxpool.Pool
}

func NewWorkoutRepository(db *pgxpool.Pool) *WorkoutRepository {
	return &WorkoutRepository{db: db}
}

// List возвращает тренировки пользователя без упражнений (для списка)
func (r *WorkoutRepository) List(ctx context.Context, userID uuid.UUID) ([]model.Workout, error) {
	rows, err := r.db.Query(ctx, `
		SELECT w.id, w.user_id, w.trainer_id, w.gym_id, COALESCE(g.name,'') as gym_name,
		       w.title, w.workout_date, w.notes, w.wellbeing, w.created_at, w.updated_at
		FROM workouts w
		LEFT JOIN gyms g ON g.id = w.gym_id
		WHERE w.user_id = $1
		ORDER BY w.workout_date DESC, w.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []model.Workout
	for rows.Next() {
		w, err := scanWorkout(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, w)
	}
	return list, nil
}

// GetByID возвращает тренировку вместе с упражнениями и подходами
func (r *WorkoutRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*model.Workout, error) {
	rows, err := r.db.Query(ctx, `
		SELECT w.id, w.user_id, w.trainer_id, w.gym_id, COALESCE(g.name,'') as gym_name,
		       w.title, w.workout_date, w.notes, w.wellbeing, w.created_at, w.updated_at
		FROM workouts w
		LEFT JOIN gyms g ON g.id = w.gym_id
		WHERE w.id = $1 AND w.user_id = $2`, id, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	w, err := pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (model.Workout, error) {
		return scanWorkout(row)
	})
	if err != nil {
		return nil, err
	}

	w.Exercises, err = r.loadExercises(ctx, w.ID)
	return &w, err
}

func (r *WorkoutRepository) loadExercises(ctx context.Context, workoutID uuid.UUID) ([]model.WorkoutExercise, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, workout_id, name, order_num, notes FROM workout_exercises
		 WHERE workout_id=$1 ORDER BY order_num`, workoutID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exercises []model.WorkoutExercise
	for rows.Next() {
		var e model.WorkoutExercise
		if err := rows.Scan(&e.ID, &e.WorkoutID, &e.Name, &e.OrderNum, &e.Notes); err != nil {
			return nil, err
		}
		exercises = append(exercises, e)
	}
	rows.Close()

	for i := range exercises {
		exercises[i].Sets, err = r.loadSets(ctx, exercises[i].ID)
		if err != nil {
			return nil, err
		}
	}
	return exercises, nil
}

func (r *WorkoutRepository) loadSets(ctx context.Context, exerciseID uuid.UUID) ([]model.Set, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, workout_exercise_id, set_num, weight, reps, rpe, rest_seconds, notes
		 FROM sets WHERE workout_exercise_id=$1 ORDER BY set_num`, exerciseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sets []model.Set
	for rows.Next() {
		var s model.Set
		if err := rows.Scan(&s.ID, &s.WorkoutExerciseID, &s.SetNum,
			&s.Weight, &s.Reps, &s.RPE, &s.RestSeconds, &s.Notes); err != nil {
			return nil, err
		}
		sets = append(sets, s)
	}
	return sets, nil
}

// Create сохраняет тренировку вместе с упражнениями и подходами в одной транзакции
func (r *WorkoutRepository) Create(ctx context.Context, userID uuid.UUID, w model.Workout, exercises []model.FormExercise) (*model.Workout, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var workoutID uuid.UUID
	err = tx.QueryRow(ctx, `
		INSERT INTO workouts (user_id, trainer_id, gym_id, title, workout_date, notes, wellbeing)
		VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`,
		userID, w.TrainerID, w.GymID, w.Title, w.WorkoutDate, w.Notes, w.Wellbeing,
	).Scan(&workoutID)
	if err != nil {
		return nil, fmt.Errorf("insert workout: %w", err)
	}

	for i, ex := range exercises {
		if ex.Name == "" {
			continue
		}
		var exID uuid.UUID
		err = tx.QueryRow(ctx,
			`INSERT INTO workout_exercises (workout_id, name, order_num, notes) VALUES ($1,$2,$3,$4) RETURNING id`,
			workoutID, ex.Name, i+1, ex.Notes,
		).Scan(&exID)
		if err != nil {
			return nil, fmt.Errorf("insert exercise: %w", err)
		}

		for j, s := range ex.Sets {
			err = insertSet(ctx, tx, exID, j+1, s)
			if err != nil {
				return nil, fmt.Errorf("insert set: %w", err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.GetByID(ctx, workoutID, userID)
}

// Update заменяет тренировку целиком (удаляет упражнения и пересоздаёт)
func (r *WorkoutRepository) Update(ctx context.Context, id, userID uuid.UUID, w model.Workout, exercises []model.FormExercise) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	res, err := tx.Exec(ctx, `
		UPDATE workouts SET gym_id=$1, title=$2, workout_date=$3, notes=$4, wellbeing=$5, updated_at=NOW()
		WHERE id=$6 AND user_id=$7`,
		w.GymID, w.Title, w.WorkoutDate, w.Notes, w.Wellbeing, id, userID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("workout not found")
	}

	if _, err := tx.Exec(ctx, `DELETE FROM workout_exercises WHERE workout_id=$1`, id); err != nil {
		return err
	}

	for i, ex := range exercises {
		if ex.Name == "" {
			continue
		}
		var exID uuid.UUID
		err = tx.QueryRow(ctx,
			`INSERT INTO workout_exercises (workout_id, name, order_num, notes) VALUES ($1,$2,$3,$4) RETURNING id`,
			id, ex.Name, i+1, ex.Notes,
		).Scan(&exID)
		if err != nil {
			return err
		}
		for j, s := range ex.Sets {
			if err := insertSet(ctx, tx, exID, j+1, s); err != nil {
				return err
			}
		}
	}

	return tx.Commit(ctx)
}

func (r *WorkoutRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM workouts WHERE id=$1 AND user_id=$2`, id, userID)
	return err
}

// ListCards returns workouts with precomputed exercise/set counts and tonnage for list views.
func (r *WorkoutRepository) ListCards(ctx context.Context, userID uuid.UUID) ([]model.WorkoutCardData, error) {
	rows, err := r.db.Query(ctx, `
		SELECT w.id, w.user_id, w.trainer_id, w.gym_id, COALESCE(g.name,'') as gym_name,
		       w.title, w.workout_date, w.notes, w.wellbeing, w.created_at, w.updated_at,
		       w.started_at, w.ended_at,
		       COUNT(DISTINCT we.id) AS exercise_count,
		       COUNT(s.id) AS set_count,
		       COALESCE(SUM(s.weight * s.reps), 0) AS tonnage
		FROM workouts w
		LEFT JOIN gyms g ON g.id = w.gym_id
		LEFT JOIN workout_exercises we ON we.workout_id = w.id
		LEFT JOIN sets s ON s.workout_exercise_id = we.id
		WHERE w.user_id = $1
		GROUP BY w.id, g.name, w.user_id, w.trainer_id, w.gym_id, w.title, w.workout_date, w.notes, w.wellbeing, w.created_at, w.updated_at, w.started_at, w.ended_at
		ORDER BY w.workout_date DESC, w.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cards []model.WorkoutCardData
	for rows.Next() {
		var c model.WorkoutCardData
		if err := rows.Scan(
			&c.ID, &c.UserID, &c.TrainerID, &c.GymID, &c.GymName,
			&c.Title, &c.WorkoutDate, &c.Notes, &c.Wellbeing, &c.CreatedAt, &c.UpdatedAt,
			&c.StartedAt, &c.EndedAt,
			&c.ExerciseCount, &c.SetCount, &c.Tonnage,
		); err != nil {
			return nil, err
		}
		cards = append(cards, c)
	}
	return cards, nil
}

func (r *WorkoutRepository) ListCardsFiltered(ctx context.Context, userID uuid.UUID, f WorkoutFilter) ([]model.WorkoutCardData, error) {
	args := []any{userID}
	where := []string{"w.user_id = $1"}

	if f.DateFrom != nil {
		args = append(args, *f.DateFrom)
		where = append(where, fmt.Sprintf("w.workout_date >= $%d", len(args)))
	}
	if f.DateTo != nil {
		args = append(args, *f.DateTo)
		where = append(where, fmt.Sprintf("w.workout_date <= $%d", len(args)))
	}
	if f.GymID != nil {
		args = append(args, *f.GymID)
		where = append(where, fmt.Sprintf("w.gym_id = $%d", len(args)))
	}
	if f.ExerciseName != "" {
		args = append(args, "%"+strings.ToLower(f.ExerciseName)+"%")
		where = append(where, fmt.Sprintf("EXISTS (SELECT 1 FROM workout_exercises we2 WHERE we2.workout_id = w.id AND lower(we2.name) LIKE $%d)", len(args)))
	}

	query := `SELECT w.id, w.user_id, w.trainer_id, w.gym_id, COALESCE(g.name,'') as gym_name,
		w.title, w.workout_date, w.notes, w.wellbeing, w.created_at, w.updated_at,
		w.started_at, w.ended_at,
		COUNT(DISTINCT we.id) AS exercise_count,
		COUNT(s.id) AS set_count,
		COALESCE(SUM(s.weight * s.reps), 0) AS tonnage
	FROM workouts w
	LEFT JOIN gyms g ON g.id = w.gym_id
	LEFT JOIN workout_exercises we ON we.workout_id = w.id
	LEFT JOIN sets s ON s.workout_exercise_id = we.id
	WHERE ` + strings.Join(where, " AND ") + `
	GROUP BY w.id, g.name, w.user_id, w.trainer_id, w.gym_id, w.title, w.workout_date, w.notes, w.wellbeing, w.created_at, w.updated_at, w.started_at, w.ended_at
	ORDER BY w.workout_date DESC, w.created_at DESC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cards []model.WorkoutCardData
	for rows.Next() {
		var c model.WorkoutCardData
		if err := rows.Scan(
			&c.ID, &c.UserID, &c.TrainerID, &c.GymID, &c.GymName,
			&c.Title, &c.WorkoutDate, &c.Notes, &c.Wellbeing, &c.CreatedAt, &c.UpdatedAt,
			&c.StartedAt, &c.EndedAt,
			&c.ExerciseCount, &c.SetCount, &c.Tonnage,
		); err != nil {
			return nil, err
		}
		cards = append(cards, c)
	}
	return cards, nil
}

// DashboardStats holds aggregated data for the home screen.
type DashboardStats struct {
	WeekCount   int
	WeekTonnage float64 // kg
	Streak      int     // consecutive calendar days
	LastCard    *model.WorkoutCardData
}

// GetDashboardStats returns week stats, streak, and last workout card for a user.
func (r *WorkoutRepository) GetDashboardStats(ctx context.Context, userID uuid.UUID) (DashboardStats, error) {
	var stats DashboardStats

	if err := r.db.QueryRow(ctx, `
		SELECT COUNT(DISTINCT w.id), COALESCE(SUM(s.weight * s.reps), 0)
		FROM workouts w
		LEFT JOIN workout_exercises we ON we.workout_id = w.id
		LEFT JOIN sets s ON s.workout_exercise_id = we.id
		WHERE w.user_id = $1
		  AND date_trunc('week', w.workout_date) = date_trunc('week', CURRENT_DATE)`,
		userID).Scan(&stats.WeekCount, &stats.WeekTonnage); err != nil {
		return stats, err
	}

	dateRows, err := r.db.Query(ctx, `
		SELECT DISTINCT workout_date::date
		FROM workouts WHERE user_id = $1
		ORDER BY 1 DESC LIMIT 60`, userID)
	if err != nil {
		return stats, err
	}
	var dates []time.Time
	for dateRows.Next() {
		var d time.Time
		if err := dateRows.Scan(&d); err != nil {
			dateRows.Close()
			return stats, err
		}
		dates = append(dates, d)
	}
	dateRows.Close()

	today := time.Now().Truncate(24 * time.Hour)
	for i, d := range dates {
		if d.Truncate(24 * time.Hour).Equal(today.AddDate(0, 0, -i)) {
			stats.Streak++
		} else {
			break
		}
	}

	var c model.WorkoutCardData
	err = r.db.QueryRow(ctx, `
		SELECT w.id, w.user_id, w.trainer_id, w.gym_id, COALESCE(g.name,'') as gym_name,
		       w.title, w.workout_date, w.notes, w.wellbeing, w.created_at, w.updated_at,
		       w.started_at, w.ended_at,
		       COUNT(DISTINCT we.id) AS exercise_count,
		       COUNT(s.id) AS set_count,
		       COALESCE(SUM(s.weight * s.reps), 0) AS tonnage
		FROM workouts w
		LEFT JOIN gyms g ON g.id = w.gym_id
		LEFT JOIN workout_exercises we ON we.workout_id = w.id
		LEFT JOIN sets s ON s.workout_exercise_id = we.id
		WHERE w.user_id = $1
		GROUP BY w.id, g.name, w.user_id, w.trainer_id, w.gym_id, w.title, w.workout_date, w.notes, w.wellbeing, w.created_at, w.updated_at, w.started_at, w.ended_at
		ORDER BY w.workout_date DESC, w.created_at DESC
		LIMIT 1`, userID).Scan(
		&c.ID, &c.UserID, &c.TrainerID, &c.GymID, &c.GymName,
		&c.Title, &c.WorkoutDate, &c.Notes, &c.Wellbeing, &c.CreatedAt, &c.UpdatedAt,
		&c.StartedAt, &c.EndedAt,
		&c.ExerciseCount, &c.SetCount, &c.Tonnage,
	)
	if err == nil {
		stats.LastCard = &c
	}
	return stats, nil
}

// GetByIDForTrainer возвращает тренировку если тренер назначен к клиенту или сам создал тренировку
func (r *WorkoutRepository) GetByIDForTrainer(ctx context.Context, id, trainerID uuid.UUID) (*model.Workout, error) {
	rows, err := r.db.Query(ctx, `
		SELECT w.id, w.user_id, w.trainer_id, w.gym_id, COALESCE(g.name,'') as gym_name,
		       w.title, w.workout_date, w.notes, w.wellbeing, w.created_at, w.updated_at
		FROM workouts w
		LEFT JOIN gyms g ON g.id = w.gym_id
		WHERE w.id = $1 AND (
			w.trainer_id = $2 OR
			EXISTS(SELECT 1 FROM trainer_clients tc WHERE tc.trainer_id=$2 AND tc.client_id=w.user_id)
		)`, id, trainerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	w, err := pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (model.Workout, error) {
		return scanWorkout(row)
	})
	if err != nil {
		return nil, err
	}
	w.Exercises, err = r.loadExercises(ctx, w.ID)
	return &w, err
}

// UpdateByTrainer обновляет только тренировки, которые тренер сам создал
func (r *WorkoutRepository) UpdateByTrainer(ctx context.Context, id, trainerID uuid.UUID, w model.Workout, exercises []model.FormExercise) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	res, err := tx.Exec(ctx, `
		UPDATE workouts SET gym_id=$1, title=$2, workout_date=$3, notes=$4, wellbeing=$5, updated_at=NOW()
		WHERE id=$6 AND trainer_id=$7`,
		w.GymID, w.Title, w.WorkoutDate, w.Notes, w.Wellbeing, id, trainerID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("workout not found or access denied")
	}
	if _, err := tx.Exec(ctx, `DELETE FROM workout_exercises WHERE workout_id=$1`, id); err != nil {
		return err
	}
	for i, ex := range exercises {
		if ex.Name == "" {
			continue
		}
		var exID uuid.UUID
		if err = tx.QueryRow(ctx,
			`INSERT INTO workout_exercises (workout_id, name, order_num, notes) VALUES ($1,$2,$3,$4) RETURNING id`,
			id, ex.Name, i+1, ex.Notes,
		).Scan(&exID); err != nil {
			return err
		}
		for j, s := range ex.Sets {
			if err := insertSet(ctx, tx, exID, j+1, s); err != nil {
				return err
			}
		}
	}
	return tx.Commit(ctx)
}

// DeleteByTrainer удаляет тренировку, созданную тренером
func (r *WorkoutRepository) DeleteByTrainer(ctx context.Context, id, trainerID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM workouts WHERE id=$1 AND trainer_id=$2`, id, trainerID)
	return err
}

// StartSession sets started_at = NOW() if not already set.
func (r *WorkoutRepository) StartSession(ctx context.Context, workoutID, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE workouts SET started_at = COALESCE(started_at, NOW()) WHERE id = $1 AND user_id = $2`,
		workoutID, userID)
	return err
}

// FinishSession sets ended_at = NOW().
func (r *WorkoutRepository) FinishSession(ctx context.Context, workoutID, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE workouts SET ended_at = NOW() WHERE id = $1 AND user_id = $2`,
		workoutID, userID)
	return err
}

// GetActiveSession loads a workout with started_at/ended_at and sets with done status.
func (r *WorkoutRepository) GetActiveSession(ctx context.Context, id, userID uuid.UUID) (*model.Workout, error) {
	rows, err := r.db.Query(ctx, `
		SELECT w.id, w.user_id, w.trainer_id, w.gym_id, COALESCE(g.name,'') as gym_name,
		       w.title, w.workout_date, w.notes, w.wellbeing, w.created_at, w.updated_at,
		       w.started_at, w.ended_at
		FROM workouts w
		LEFT JOIN gyms g ON g.id = w.gym_id
		WHERE w.id = $1 AND w.user_id = $2`, id, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	w, err := pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (model.Workout, error) {
		var wo model.Workout
		err := row.Scan(
			&wo.ID, &wo.UserID, &wo.TrainerID, &wo.GymID, &wo.GymName,
			&wo.Title, &wo.WorkoutDate, &wo.Notes, &wo.Wellbeing,
			&wo.CreatedAt, &wo.UpdatedAt,
			&wo.StartedAt, &wo.EndedAt,
		)
		return wo, err
	})
	if err != nil {
		return nil, err
	}
	w.Exercises, err = r.loadExercisesActive(ctx, w.ID)
	return &w, err
}

func (r *WorkoutRepository) loadExercisesActive(ctx context.Context, workoutID uuid.UUID) ([]model.WorkoutExercise, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, workout_id, name, order_num, notes FROM workout_exercises
		 WHERE workout_id=$1 ORDER BY order_num`, workoutID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exercises []model.WorkoutExercise
	for rows.Next() {
		var e model.WorkoutExercise
		if err := rows.Scan(&e.ID, &e.WorkoutID, &e.Name, &e.OrderNum, &e.Notes); err != nil {
			return nil, err
		}
		exercises = append(exercises, e)
	}
	rows.Close()

	for i := range exercises {
		exercises[i].Sets, err = r.loadSetsActive(ctx, exercises[i].ID)
		if err != nil {
			return nil, err
		}
	}
	return exercises, nil
}

func (r *WorkoutRepository) loadSetsActive(ctx context.Context, exerciseID uuid.UUID) ([]model.Set, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, workout_exercise_id, set_num, weight, reps, rpe, rest_seconds, notes, done
		 FROM sets WHERE workout_exercise_id=$1 ORDER BY set_num`, exerciseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sets []model.Set
	for rows.Next() {
		var s model.Set
		if err := rows.Scan(&s.ID, &s.WorkoutExerciseID, &s.SetNum,
			&s.Weight, &s.Reps, &s.RPE, &s.RestSeconds, &s.Notes, &s.Done); err != nil {
			return nil, err
		}
		sets = append(sets, s)
	}
	return sets, nil
}

// ToggleSetDone flips the done flag and returns the new state.
func (r *WorkoutRepository) ToggleSetDone(ctx context.Context, setID, userID uuid.UUID) (bool, error) {
	var done bool
	err := r.db.QueryRow(ctx, `
		UPDATE sets SET done = NOT done
		WHERE id = $1
		  AND workout_exercise_id IN (
		      SELECT we.id FROM workout_exercises we
		      JOIN workouts w ON w.id = we.workout_id
		      WHERE w.user_id = $2
		  )
		RETURNING done`,
		setID, userID).Scan(&done)
	return done, err
}

// GetSetByID returns a single set (including done) verifying it belongs to userID.
func (r *WorkoutRepository) GetSetByID(ctx context.Context, setID, userID uuid.UUID) (*model.Set, error) {
	var s model.Set
	err := r.db.QueryRow(ctx, `
		SELECT s.id, s.workout_exercise_id, s.set_num, s.weight, s.reps, s.rpe, s.rest_seconds, s.notes, s.done
		FROM sets s
		JOIN workout_exercises we ON we.id = s.workout_exercise_id
		JOIN workouts w ON w.id = we.workout_id
		WHERE s.id = $1 AND w.user_id = $2`,
		setID, userID).Scan(
		&s.ID, &s.WorkoutExerciseID, &s.SetNum, &s.Weight, &s.Reps, &s.RPE, &s.RestSeconds, &s.Notes, &s.Done,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func insertSet(ctx context.Context, tx pgx.Tx, exID uuid.UUID, num int, s model.FormSet) error {
	weight := parseOptFloat(s.Weight)
	reps := parseOptInt(s.Reps)
	rpe := parseOptFloat(s.RPE)
	rest := parseOptInt(s.RestSeconds)
	_, err := tx.Exec(ctx,
		`INSERT INTO sets (workout_exercise_id, set_num, weight, reps, rpe, rest_seconds, notes) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		exID, num, weight, reps, rpe, rest, s.Notes,
	)
	return err
}

func scanWorkout(row interface{ Scan(...any) error }) (model.Workout, error) {
	var w model.Workout
	err := row.Scan(
		&w.ID, &w.UserID, &w.TrainerID, &w.GymID, &w.GymName,
		&w.Title, &w.WorkoutDate, &w.Notes, &w.Wellbeing,
		&w.CreatedAt, &w.UpdatedAt,
	)
	return w, err
}

func parseOptFloat(s string) *float64 {
	if s == "" {
		return nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return &v
}

func parseOptInt(s string) *int {
	if s == "" {
		return nil
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &v
}

// GetRecentPRs returns up to 3 personal-weight records set in the last 30 days,
// ordered by improvement delta descending.
func (r *WorkoutRepository) GetRecentPRs(ctx context.Context, userID uuid.UUID) ([]model.RecentPR, error) {
	rows, err := r.db.Query(ctx, `
		WITH recent AS (
			SELECT DISTINCT ON (we.name)
				we.name,
				s.weight          AS new_weight,
				COALESCE(s.reps, 0) AS reps
			FROM sets s
			JOIN workout_exercises we ON we.id = s.workout_exercise_id
			JOIN workouts w ON w.id = we.workout_id
			WHERE w.user_id = $1
			  AND s.weight IS NOT NULL AND s.weight > 0
			  AND w.workout_date >= NOW() - INTERVAL '30 days'
			ORDER BY we.name, s.weight DESC
		),
		historical AS (
			SELECT we.name, MAX(s.weight) AS prev_max
			FROM sets s
			JOIN workout_exercises we ON we.id = s.workout_exercise_id
			JOIN workouts w ON w.id = we.workout_id
			WHERE w.user_id = $1
			  AND s.weight IS NOT NULL
			  AND w.workout_date < NOW() - INTERVAL '30 days'
			GROUP BY we.name
		)
		SELECT r.name, r.new_weight, r.reps,
		       r.new_weight - COALESCE(h.prev_max, 0) AS delta
		FROM recent r
		LEFT JOIN historical h ON h.name = r.name
		WHERE r.new_weight > COALESCE(h.prev_max, 0)
		ORDER BY delta DESC
		LIMIT 3
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var prs []model.RecentPR
	for rows.Next() {
		var pr model.RecentPR
		if err := rows.Scan(&pr.ExerciseName, &pr.NewWeight, &pr.Reps, &pr.Delta); err != nil {
			return nil, err
		}
		prs = append(prs, pr)
	}
	return prs, nil
}

// SuggestExercises returns exercise names from the user's history matching the prefix.
func (r *WorkoutRepository) SuggestExercises(ctx context.Context, userID uuid.UUID, q string) ([]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT we.name, COUNT(*) AS cnt
		FROM workout_exercises we
		JOIN workouts w ON w.id = we.workout_id
		WHERE w.user_id = $1 AND we.name ILIKE $2
		GROUP BY we.name
		ORDER BY cnt DESC
		LIMIT 6`,
		userID, q+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var name string
		var cnt int
		if err := rows.Scan(&name, &cnt); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, nil
}
