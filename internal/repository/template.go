package repository

import (
	"context"
	"fmt"
	"time"
	"workout-tracker/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TemplateRepository struct {
	db *pgxpool.Pool
}

func NewTemplateRepository(db *pgxpool.Pool) *TemplateRepository {
	return &TemplateRepository{db: db}
}

func (r *TemplateRepository) List(ctx context.Context, trainerID uuid.UUID) ([]model.WorkoutTemplate, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, trainer_id, title, notes, created_at, updated_at
		FROM workout_templates
		WHERE trainer_id = $1
		ORDER BY created_at DESC`, trainerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.WorkoutTemplate
	for rows.Next() {
		var t model.WorkoutTemplate
		if err := rows.Scan(&t.ID, &t.TrainerID, &t.Title, &t.Notes, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, t)
	}

	// Загружаем количество упражнений для каждого шаблона
	for i := range list {
		exs, err := r.loadExercises(ctx, list[i].ID)
		if err != nil {
			return nil, err
		}
		list[i].Exercises = exs
	}
	return list, nil
}

func (r *TemplateRepository) GetByID(ctx context.Context, id, trainerID uuid.UUID) (*model.WorkoutTemplate, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, trainer_id, title, notes, created_at, updated_at
		FROM workout_templates
		WHERE id = $1 AND trainer_id = $2`, id, trainerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	t, err := pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (model.WorkoutTemplate, error) {
		var tm model.WorkoutTemplate
		err := row.Scan(&tm.ID, &tm.TrainerID, &tm.Title, &tm.Notes, &tm.CreatedAt, &tm.UpdatedAt)
		return tm, err
	})
	if err != nil {
		return nil, err
	}

	t.Exercises, err = r.loadExercises(ctx, t.ID)
	return &t, err
}

func (r *TemplateRepository) loadExercises(ctx context.Context, templateID uuid.UUID) ([]model.TemplateExercise, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, template_id, name, order_num, notes
		FROM template_exercises
		WHERE template_id = $1
		ORDER BY order_num`, templateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exercises []model.TemplateExercise
	for rows.Next() {
		var e model.TemplateExercise
		if err := rows.Scan(&e.ID, &e.TemplateID, &e.Name, &e.OrderNum, &e.Notes); err != nil {
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

func (r *TemplateRepository) loadSets(ctx context.Context, exerciseID uuid.UUID) ([]model.TemplateSet, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, template_exercise_id, set_num, weight, reps, rpe, rest_seconds, notes
		FROM template_sets
		WHERE template_exercise_id = $1
		ORDER BY set_num`, exerciseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sets []model.TemplateSet
	for rows.Next() {
		var s model.TemplateSet
		if err := rows.Scan(&s.ID, &s.TemplateExerciseID, &s.SetNum,
			&s.Weight, &s.Reps, &s.RPE, &s.RestSeconds, &s.Notes); err != nil {
			return nil, err
		}
		sets = append(sets, s)
	}
	return sets, nil
}

func (r *TemplateRepository) Create(ctx context.Context, trainerID uuid.UUID, title, notes string, exercises []model.FormExercise) (*model.WorkoutTemplate, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var id uuid.UUID
	if err := tx.QueryRow(ctx,
		`INSERT INTO workout_templates (trainer_id, title, notes) VALUES ($1,$2,$3) RETURNING id`,
		trainerID, title, notes,
	).Scan(&id); err != nil {
		return nil, fmt.Errorf("insert template: %w", err)
	}

	if err := insertTemplateExercises(ctx, tx, id, exercises); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id, trainerID)
}

func (r *TemplateRepository) Update(ctx context.Context, id, trainerID uuid.UUID, title, notes string, exercises []model.FormExercise) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	res, err := tx.Exec(ctx,
		`UPDATE workout_templates SET title=$1, notes=$2, updated_at=NOW() WHERE id=$3 AND trainer_id=$4`,
		title, notes, id, trainerID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("template not found")
	}

	if _, err := tx.Exec(ctx, `DELETE FROM template_exercises WHERE template_id=$1`, id); err != nil {
		return err
	}

	if err := insertTemplateExercises(ctx, tx, id, exercises); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *TemplateRepository) Delete(ctx context.Context, id, trainerID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM workout_templates WHERE id=$1 AND trainer_id=$2`, id, trainerID)
	return err
}

// Apply создаёт тренировки для каждого клиента на основе шаблона
func (r *TemplateRepository) Apply(ctx context.Context, templateID, trainerID uuid.UUID, clientIDs []uuid.UUID, date time.Time, gymID *uuid.UUID) error {
	tmpl, err := r.GetByID(ctx, templateID, trainerID)
	if err != nil {
		return fmt.Errorf("template not found: %w", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, clientID := range clientIDs {
		var workoutID uuid.UUID
		if err := tx.QueryRow(ctx, `
			INSERT INTO workouts (user_id, trainer_id, gym_id, title, workout_date, notes)
			VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`,
			clientID, trainerID, gymID, tmpl.Title, date, tmpl.Notes,
		).Scan(&workoutID); err != nil {
			return fmt.Errorf("insert workout for client %s: %w", clientID, err)
		}

		for i, ex := range tmpl.Exercises {
			var exID uuid.UUID
			if err := tx.QueryRow(ctx,
				`INSERT INTO workout_exercises (workout_id, name, order_num, notes) VALUES ($1,$2,$3,$4) RETURNING id`,
				workoutID, ex.Name, i+1, ex.Notes,
			).Scan(&exID); err != nil {
				return fmt.Errorf("insert exercise: %w", err)
			}

			for _, s := range ex.Sets {
				if _, err := tx.Exec(ctx, `
					INSERT INTO sets (workout_exercise_id, set_num, weight, reps, rpe, rest_seconds, notes)
					VALUES ($1,$2,$3,$4,$5,$6,$7)`,
					exID, s.SetNum, s.Weight, s.Reps, s.RPE, s.RestSeconds, s.Notes,
				); err != nil {
					return fmt.Errorf("insert set: %w", err)
				}
			}
		}
	}

	return tx.Commit(ctx)
}

func insertTemplateExercises(ctx context.Context, tx pgx.Tx, templateID uuid.UUID, exercises []model.FormExercise) error {
	for i, ex := range exercises {
		if ex.Name == "" {
			continue
		}
		var exID uuid.UUID
		if err := tx.QueryRow(ctx,
			`INSERT INTO template_exercises (template_id, name, order_num, notes) VALUES ($1,$2,$3,$4) RETURNING id`,
			templateID, ex.Name, i+1, ex.Notes,
		).Scan(&exID); err != nil {
			return fmt.Errorf("insert exercise: %w", err)
		}

		for j, s := range ex.Sets {
			weight := parseOptFloat(s.Weight)
			reps := parseOptInt(s.Reps)
			rpe := parseOptFloat(s.RPE)
			rest := parseOptInt(s.RestSeconds)
			if _, err := tx.Exec(ctx,
				`INSERT INTO template_sets (template_exercise_id, set_num, weight, reps, rpe, rest_seconds, notes) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
				exID, j+1, weight, reps, rpe, rest, s.Notes,
			); err != nil {
				return fmt.Errorf("insert set: %w", err)
			}
		}
	}
	return nil
}
