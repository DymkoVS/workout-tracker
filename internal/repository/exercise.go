package repository

import (
	"context"
	"time"
	"workout-tracker/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ExerciseRepository struct {
	db *pgxpool.Pool
}

func NewExerciseRepository(db *pgxpool.Pool) *ExerciseRepository {
	return &ExerciseRepository{db: db}
}

func (r *ExerciseRepository) List(ctx context.Context) ([]model.Exercise, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, COALESCE(muscle_group,''), COALESCE(description,''), created_at
		FROM exercises
		ORDER BY lower(name)`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Exercise
	for rows.Next() {
		var e model.Exercise
		if err := rows.Scan(&e.ID, &e.Name, &e.MuscleGroup, &e.Description, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, nil
}

func (r *ExerciseRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Exercise, error) {
	var e model.Exercise
	err := r.db.QueryRow(ctx, `
		SELECT id, name, COALESCE(muscle_group,''), COALESCE(description,''), created_at
		FROM exercises WHERE id=$1`, id).
		Scan(&e.ID, &e.Name, &e.MuscleGroup, &e.Description, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *ExerciseRepository) Create(ctx context.Context, name, muscleGroup, description string) (*model.Exercise, error) {
	var e model.Exercise
	err := r.db.QueryRow(ctx, `
		INSERT INTO exercises (name, muscle_group, description)
		VALUES ($1, NULLIF($2,''), NULLIF($3,''))
		RETURNING id, name, COALESCE(muscle_group,''), COALESCE(description,''), created_at`,
		name, muscleGroup, description).
		Scan(&e.ID, &e.Name, &e.MuscleGroup, &e.Description, &e.CreatedAt)
	return &e, err
}

func (r *ExerciseRepository) Update(ctx context.Context, id uuid.UUID, name, muscleGroup, description string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE exercises
		SET name=$2, muscle_group=NULLIF($3,''), description=NULLIF($4,'')
		WHERE id=$1`,
		id, name, muscleGroup, description)
	return err
}

func (r *ExerciseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM exercises WHERE id=$1`, id)
	return err
}

func (r *ExerciseRepository) GetProgress(ctx context.Context, exerciseName string, userID uuid.UUID, limit int) ([]model.ProgressSession, error) {
	rows, err := r.db.Query(ctx, `
		WITH recent_workouts AS (
			SELECT DISTINCT w.id, w.workout_date, w.title
			FROM workouts w
			JOIN workout_exercises we ON we.workout_id = w.id
			WHERE w.user_id = $1 AND lower(we.name) = lower($2)
			ORDER BY w.workout_date DESC
			LIMIT $3
		)
		SELECT rw.id, rw.workout_date, rw.title,
		       s.set_num, s.weight, s.reps, s.rpe
		FROM recent_workouts rw
		JOIN workout_exercises we ON we.workout_id = rw.id AND lower(we.name) = lower($2)
		JOIN sets s ON s.workout_exercise_id = we.id
		ORDER BY rw.workout_date DESC, s.set_num ASC`,
		userID, exerciseName, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessionMap := make(map[uuid.UUID]*model.ProgressSession)
	var order []uuid.UUID

	for rows.Next() {
		var wid uuid.UUID
		var wdate time.Time
		var wtitle string
		var s model.Set
		if err := rows.Scan(&wid, &wdate, &wtitle, &s.SetNum, &s.Weight, &s.Reps, &s.RPE); err != nil {
			return nil, err
		}
		if _, ok := sessionMap[wid]; !ok {
			sessionMap[wid] = &model.ProgressSession{
				WorkoutID:    wid,
				WorkoutDate:  wdate,
				WorkoutTitle: wtitle,
			}
			order = append(order, wid)
		}
		sess := sessionMap[wid]
		sess.Sets = append(sess.Sets, s)
		if s.Weight != nil {
			if *s.Weight > sess.MaxWeight {
				sess.MaxWeight = *s.Weight
			}
			if s.Reps != nil {
				sess.TotalVolume += *s.Weight * float64(*s.Reps)
			}
		}
	}

	out := make([]model.ProgressSession, 0, len(order))
	for _, id := range order {
		out = append(out, *sessionMap[id])
	}
	return out, nil
}

func (r *ExerciseRepository) ListClientExercises(ctx context.Context, userID uuid.UUID) ([]model.ClientExerciseSummary, error) {
	rows, err := r.db.Query(ctx, `
		SELECT we.name, COUNT(DISTINCT w.id) AS session_count, MAX(w.workout_date) AS last_date
		FROM workout_exercises we
		JOIN workouts w ON w.id = we.workout_id
		WHERE w.user_id = $1
		GROUP BY we.name
		ORDER BY MAX(w.workout_date) DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.ClientExerciseSummary
	for rows.Next() {
		var s model.ClientExerciseSummary
		if err := rows.Scan(&s.Name, &s.SessionCount, &s.LastDate); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

// Search returns exercises whose name starts with q (case-insensitive), up to limit.
func (r *ExerciseRepository) Search(ctx context.Context, q string, limit int) ([]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT name FROM exercises
		WHERE name ILIKE $1
		ORDER BY lower(name)
		LIMIT $2`, q+"%", limit)
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
