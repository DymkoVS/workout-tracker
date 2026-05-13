package repository

import (
	"context"
	"strings"
	"time"
	"workout-tracker/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TrainerClientRepository struct {
	db *pgxpool.Pool
}

func NewTrainerClientRepository(db *pgxpool.Pool) *TrainerClientRepository {
	return &TrainerClientRepository{db: db}
}

// GetClientStats возвращает клиентов тренера вместе с их недельной статистикой.
func (r *TrainerClientRepository) GetClientStats(ctx context.Context, trainerID uuid.UUID) ([]*model.ClientStat, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			u.id, u.login, u.email, u.password_hash, u.full_name, u.role,
			u.is_admin, u.is_active, u.created_at, u.updated_at,
			COUNT(DISTINCT CASE WHEN date_trunc('week', w.workout_date) = date_trunc('week', CURRENT_DATE)
			                    THEN w.id END)::int AS week_done,
			COUNT(DISTINCT CASE WHEN date_trunc('week', w.workout_date) = date_trunc('week', CURRENT_DATE - INTERVAL '7 days')
			                    THEN w.id END)::int AS prev_week_done,
			COUNT(DISTINCT w.id)::int AS total_workouts,
			MAX(w.workout_date) AS last_workout
		FROM trainer_clients tc
		JOIN users u ON u.id = tc.client_id
		LEFT JOIN workouts w ON w.user_id = tc.client_id
		WHERE tc.trainer_id = $1
		GROUP BY u.id, u.login, u.email, u.password_hash, u.full_name, u.role,
		         u.is_admin, u.is_active, u.created_at, u.updated_at
		ORDER BY u.full_name, u.login`, trainerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	statsByID := map[uuid.UUID]*model.ClientStat{}
	var statsList []*model.ClientStat

	for rows.Next() {
		u := &model.User{}
		cs := &model.ClientStat{User: u, WeekPlan: 4}
		var email *string
		var lastWorkout *time.Time
		if err := rows.Scan(
			&u.ID, &u.Login, &email, &u.PasswordHash, &u.FullName,
			&u.Role, &u.IsAdmin, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
			&cs.WeekDone, &cs.PrevWeekDone, &cs.TotalWorkouts, &lastWorkout,
		); err != nil {
			return nil, err
		}
		if email != nil {
			u.Email = *email
		}
		cs.LastWorkout = lastWorkout
		if lastWorkout == nil || time.Since(*lastWorkout) > 5*24*time.Hour {
			cs.Status = "off"
		} else {
			cs.Status = "on"
		}
		if cs.WeekDone >= cs.WeekPlan {
			cs.BarColor = "#D7FF1A"
		} else if cs.WeekDone > 0 {
			cs.BarColor = "#ff9f0a"
		} else {
			cs.BarColor = "#2e2e2e"
		}
		cs.Initials = clientInitials(u.FullName, u.Login)
		if lastWorkout != nil {
			cs.LastWorkoutFmt = lastWorkout.Format("02.01")
		} else {
			cs.LastWorkoutFmt = "никогда"
		}
		statsByID[u.ID] = cs
		statsList = append(statsList, cs)
	}
	rows.Close()

	if len(statsList) == 0 {
		return statsList, nil
	}

	// Вычисляем streak: получаем даты тренировок за последние 60 дней одним запросом
	dateRows, err := r.db.Query(ctx, `
		SELECT DISTINCT user_id, workout_date::date
		FROM workouts
		WHERE user_id IN (SELECT client_id FROM trainer_clients WHERE trainer_id = $1)
		  AND workout_date >= CURRENT_DATE - INTERVAL '60 days'
		ORDER BY user_id, workout_date DESC`, trainerID)
	if err != nil {
		return statsList, nil
	}
	defer dateRows.Close()

	datesByUser := map[uuid.UUID][]time.Time{}
	for dateRows.Next() {
		var uid uuid.UUID
		var d time.Time
		if err := dateRows.Scan(&uid, &d); err != nil {
			continue
		}
		datesByUser[uid] = append(datesByUser[uid], d)
	}

	today := time.Now().Truncate(24 * time.Hour)
	for uid, dates := range datesByUser {
		cs := statsByID[uid]
		if cs == nil {
			continue
		}
		for i, d := range dates {
			if d.Truncate(24*time.Hour).Equal(today.AddDate(0, 0, -i)) {
				cs.Streak++
			} else {
				break
			}
		}
	}

	return statsList, nil
}

func clientInitials(fullName, login string) string {
	parts := strings.Fields(fullName)
	if len(parts) >= 2 {
		r0, r1 := []rune(parts[0]), []rune(parts[1])
		if len(r0) > 0 && len(r1) > 0 {
			return strings.ToUpper(string(r0[:1]) + string(r1[:1]))
		}
	}
	if len(parts) == 1 {
		r := []rune(parts[0])
		if len(r) >= 2 {
			return strings.ToUpper(string(r[:2]))
		}
	}
	r := []rune(login)
	if len(r) >= 2 {
		return strings.ToUpper(string(r[:2]))
	}
	return strings.ToUpper(login)
}

// GetClients возвращает всех клиентов тренера
func (r *TrainerClientRepository) GetClients(ctx context.Context, trainerID uuid.UUID) ([]*model.User, error) {
	rows, err := r.db.Query(ctx, `
		SELECT u.id, u.login, u.email, u.password_hash, u.full_name, u.role,
		       u.is_admin, u.is_active, u.created_at, u.updated_at
		FROM users u
		JOIN trainer_clients tc ON tc.client_id = u.id
		WHERE tc.trainer_id = $1
		ORDER BY u.full_name, u.login`, trainerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []*model.User
	for rows.Next() {
		u := &model.User{}
		var email *string
		if err := rows.Scan(&u.ID, &u.Login, &email, &u.PasswordHash,
			&u.FullName, &u.Role, &u.IsAdmin, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		if email != nil {
			u.Email = *email
		}
		users = append(users, u)
	}
	return users, nil
}

// GetTrainers возвращает всех тренеров клиента
func (r *TrainerClientRepository) GetTrainers(ctx context.Context, clientID uuid.UUID) ([]*model.User, error) {
	rows, err := r.db.Query(ctx, `
		SELECT u.id, u.login, u.email, u.password_hash, u.full_name, u.role,
		       u.is_admin, u.is_active, u.created_at, u.updated_at
		FROM users u
		JOIN trainer_clients tc ON tc.trainer_id = u.id
		WHERE tc.client_id = $1
		ORDER BY u.full_name, u.login`, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []*model.User
	for rows.Next() {
		u := &model.User{}
		var email *string
		if err := rows.Scan(&u.ID, &u.Login, &email, &u.PasswordHash,
			&u.FullName, &u.Role, &u.IsAdmin, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		if email != nil {
			u.Email = *email
		}
		users = append(users, u)
	}
	return users, nil
}

// IsAssigned проверяет, назначен ли клиент к тренеру
func (r *TrainerClientRepository) IsAssigned(ctx context.Context, trainerID, clientID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM trainer_clients WHERE trainer_id=$1 AND client_id=$2)`,
		trainerID, clientID,
	).Scan(&exists)
	return exists, err
}

// Assign назначает клиента к тренеру
func (r *TrainerClientRepository) Assign(ctx context.Context, trainerID, clientID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO trainer_clients (trainer_id, client_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		trainerID, clientID)
	return err
}

// Unassign убирает клиента от тренера
func (r *TrainerClientRepository) Unassign(ctx context.Context, trainerID, clientID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM trainer_clients WHERE trainer_id=$1 AND client_id=$2`,
		trainerID, clientID)
	return err
}

// GetAllTrainers возвращает всех тренеров (для админа)
func (r *TrainerClientRepository) GetAllTrainers(ctx context.Context) ([]*model.User, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, login, email, password_hash, full_name, role, is_admin, is_active, created_at, updated_at
		FROM users WHERE role = 'trainer' AND is_active = TRUE
		ORDER BY full_name, login`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []*model.User
	for rows.Next() {
		u := &model.User{}
		var email *string
		if err := rows.Scan(&u.ID, &u.Login, &email, &u.PasswordHash,
			&u.FullName, &u.Role, &u.IsAdmin, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		if email != nil {
			u.Email = *email
		}
		users = append(users, u)
	}
	return users, nil
}

// GetAllClients возвращает всех клиентов (для админа)
func (r *TrainerClientRepository) GetAllClients(ctx context.Context) ([]*model.User, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, login, email, password_hash, full_name, role, is_admin, is_active, created_at, updated_at
		FROM users WHERE role = 'client' AND is_active = TRUE
		ORDER BY full_name, login`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []*model.User
	for rows.Next() {
		u := &model.User{}
		var email *string
		if err := rows.Scan(&u.ID, &u.Login, &email, &u.PasswordHash,
			&u.FullName, &u.Role, &u.IsAdmin, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		if email != nil {
			u.Email = *email
		}
		users = append(users, u)
	}
	return users, nil
}
