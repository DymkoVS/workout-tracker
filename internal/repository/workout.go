package repository

import (
	"context"
	"fmt"
	"strconv"
	"workout-tracker/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

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
