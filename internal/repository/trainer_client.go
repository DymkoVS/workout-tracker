package repository

import (
	"context"
	"fmt"
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
		LEFT JOIN workouts w ON w.user_id = tc.client_id AND w.ended_at IS NOT NULL
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

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(statsList) == 0 {
		return statsList, nil
	}

	// Вычисляем streak: получаем даты тренировок за последние 60 дней одним запросом
	dateRows, err := r.db.Query(ctx, `
		SELECT DISTINCT user_id, workout_date::date
		FROM workouts
		WHERE user_id IN (SELECT client_id FROM trainer_clients WHERE trainer_id = $1)
		  AND ended_at IS NOT NULL
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
			if d.Truncate(24 * time.Hour).Equal(today.AddDate(0, 0, -i)) {
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

// GetClientDetailData возвращает полные данные по одному клиенту для страницы детали.
func (r *TrainerClientRepository) GetClientDetailData(ctx context.Context, trainerID, clientID uuid.UUID) (*model.ClientDetailData, error) {
	u := &model.User{}
	cd := &model.ClientDetailData{User: u, WeekPlan: 4}
	var email *string
	var lastWorkout *time.Time

	err := r.db.QueryRow(ctx, `
		SELECT
			u.id, u.login, u.email, u.password_hash, u.full_name, u.role,
			u.is_admin, u.is_active, u.created_at, u.updated_at,
			COUNT(DISTINCT w.id)::int AS total_workouts,
			COUNT(DISTINCT CASE WHEN date_trunc('week', w.workout_date) = date_trunc('week', CURRENT_DATE)
			                    THEN w.id END)::int AS week_done,
			MAX(w.workout_date) AS last_workout
		FROM trainer_clients tc
		JOIN users u ON u.id = tc.client_id
		LEFT JOIN workouts w ON w.user_id = tc.client_id AND w.ended_at IS NOT NULL
		WHERE tc.trainer_id = $1 AND tc.client_id = $2
		GROUP BY u.id, u.login, u.email, u.password_hash, u.full_name, u.role,
		         u.is_admin, u.is_active, u.created_at, u.updated_at`,
		trainerID, clientID,
	).Scan(
		&u.ID, &u.Login, &email, &u.PasswordHash, &u.FullName,
		&u.Role, &u.IsAdmin, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
		&cd.TotalWorkouts, &cd.WeekDone, &lastWorkout,
	)
	if err != nil {
		return nil, err
	}
	if email != nil {
		u.Email = *email
	}
	if lastWorkout == nil || time.Since(*lastWorkout) > 5*24*time.Hour {
		cd.Status = "off"
	} else {
		cd.Status = "on"
	}
	cd.Initials = clientInitials(u.FullName, u.Login)

	// Compliance grid: workouts per week for last 4 weeks → 16 cells (4 per week)
	compRows, err := r.db.Query(ctx, `
		SELECT date_trunc('week', workout_date)::date AS week_start, COUNT(DISTINCT id)::int
		FROM workouts
		WHERE user_id = $1
		  AND ended_at IS NOT NULL
		  AND workout_date >= date_trunc('week', CURRENT_DATE) - INTERVAL '21 days'
		GROUP BY week_start`, clientID)
	if err == nil {
		weekCounts := map[time.Time]int{}
		for compRows.Next() {
			var ws time.Time
			var cnt int
			if scanErr := compRows.Scan(&ws, &cnt); scanErr == nil {
				weekCounts[ws.UTC().Truncate(24*time.Hour)] = cnt
			}
		}
		compRows.Close()

		// Current ISO week Monday
		today := time.Now().UTC().Truncate(24 * time.Hour)
		weekday := int(today.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		currentWeekStart := today.AddDate(0, 0, -(weekday - 1))

		cd.Compliance = make([]bool, 16)
		doneTotal := 0
		for w := 3; w >= 0; w-- {
			ws := currentWeekStart.AddDate(0, 0, -w*7)
			cnt := weekCounts[ws]
			if cnt > 4 {
				cnt = 4
			}
			base := (3 - w) * 4
			for i := 0; i < cnt; i++ {
				cd.Compliance[base+i] = true
				doneTotal++
			}
		}
		pct := float64(doneTotal) / 16.0 * 100
		cd.CompliancePct = fmt.Sprintf("%.0f%%", pct)
	}

	// Streak: consecutive calendar days with workouts going back from today
	dateRows, err := r.db.Query(ctx, `
		SELECT DISTINCT workout_date::date
		FROM workouts
		WHERE user_id = $1
		  AND ended_at IS NOT NULL
		  AND workout_date >= CURRENT_DATE - INTERVAL '60 days'
		ORDER BY workout_date DESC`, clientID)
	if err == nil {
		today := time.Now().Truncate(24 * time.Hour)
		for i := 0; dateRows.Next(); i++ {
			var d time.Time
			if scanErr := dateRows.Scan(&d); scanErr != nil {
				break
			}
			if d.Truncate(24 * time.Hour).Equal(today.AddDate(0, 0, -i)) {
				cd.Streak++
			} else {
				break
			}
		}
		dateRows.Close()
	}

	// Average tonnage per workout over last 90 days (result already in kg)
	var avgTonnage float64
	_ = r.db.QueryRow(ctx, `
		SELECT COALESCE(AVG(workout_tonnage), 0)
		FROM (
			SELECT w.id, COALESCE(SUM(s.weight * s.reps), 0) AS workout_tonnage
			FROM workouts w
			JOIN workout_exercises e ON e.workout_id = w.id
			JOIN sets s ON s.workout_exercise_id = e.id
			WHERE w.user_id = $1
			  AND s.weight IS NOT NULL AND s.reps IS NOT NULL
			  AND w.workout_date >= CURRENT_DATE - INTERVAL '90 days'
			GROUP BY w.id
		) pt`, clientID,
	).Scan(&avgTonnage)
	avgTonnage /= 1000
	switch {
	case avgTonnage >= 1:
		cd.AvgTonnageFmt = fmt.Sprintf("%.1fт", avgTonnage)
	case avgTonnage > 0:
		cd.AvgTonnageFmt = fmt.Sprintf("%.2fт", avgTonnage)
	default:
		cd.AvgTonnageFmt = "—"
	}

	// Recent workouts (last 4)
	recentRows, err := r.db.Query(ctx, `
		SELECT w.title, w.workout_date, w.wellbeing,
		       COALESCE(SUM(s.weight * s.reps), 0) AS tonnage
		FROM workouts w
		LEFT JOIN workout_exercises e ON e.workout_id = w.id
		LEFT JOIN sets s ON s.workout_exercise_id = e.id
		         AND s.weight IS NOT NULL AND s.reps IS NOT NULL
		WHERE w.user_id = $1
		  AND w.ended_at IS NOT NULL
		GROUP BY w.id, w.title, w.workout_date, w.wellbeing
		ORDER BY w.workout_date DESC
		LIMIT 4`, clientID)
	if err == nil {
		for recentRows.Next() {
			var title string
			var wdate time.Time
			var wellbeing *int
			var tonnage float64
			if scanErr := recentRows.Scan(&title, &wdate, &wellbeing, &tonnage); scanErr != nil {
				continue
			}
			rw := model.ClientRecentWorkout{
				DateFmt:   wdate.Format("02.01"),
				Title:     title,
				Wellbeing: wellbeing,
			}
			tKg := tonnage / 1000
			switch {
			case tKg >= 1:
				rw.TonnageFmt = fmt.Sprintf("%.1fт", tKg)
			case tKg > 0:
				rw.TonnageFmt = fmt.Sprintf("%.2fт", tKg)
			}
			cd.RecentWorkouts = append(cd.RecentWorkouts, rw)
		}
		recentRows.Close()
	}

	return cd, nil
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
	return users, rows.Err()
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
	return users, rows.Err()
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
	return users, rows.Err()
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
	return users, rows.Err()
}
