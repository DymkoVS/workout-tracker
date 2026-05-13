package handler

import (
	"fmt"
	"net/http"
	"time"
	"workout-tracker/internal/middleware"
	"workout-tracker/internal/repository"
)

type DashboardHandler struct {
	workouts *repository.WorkoutRepository
}

func NewDashboardHandler(workouts *repository.WorkoutRepository) *DashboardHandler {
	return &DashboardHandler{workouts: workouts}
}

var ruMonthsGen   = [...]string{"", "ЯНВ", "ФЕВ", "МАР", "АПР", "МАЯ", "ИЮН", "ИЮЛ", "АВГ", "СЕН", "ОКТ", "НОЯ", "ДЕК"}
var ruWeekdaysShort = [...]string{"ВС", "ПН", "ВТ", "СР", "ЧТ", "ПТ", "СБ"}

func fmtTonnage(kg float64) string {
	if kg == 0 {
		return "0кг"
	}
	if kg >= 1000 {
		t := kg / 1000
		if t == float64(int(t)) {
			return fmt.Sprintf("%dт", int(t))
		}
		return fmt.Sprintf("%.1fт", t)
	}
	return fmt.Sprintf("%.0fкг", kg)
}

func (h *DashboardHandler) Index(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	now := time.Now()
	todayRU := fmt.Sprintf("%d %s · %s", now.Day(), ruMonthsGen[now.Month()], ruWeekdaysShort[now.Weekday()])

	stats, _ := h.workouts.GetDashboardStats(r.Context(), user.ID)
	recentPRs, _ := h.workouts.GetRecentPRs(r.Context(), user.ID)

	renderTemplate(w, r, "dashboard.html", map[string]any{
		"CurrentUser": user,
		"TodayRU":     todayRU,
		"WeekCount":   stats.WeekCount,
		"WeekTonnage": fmtTonnage(stats.WeekTonnage),
		"Streak":      stats.Streak,
		"LastCard":    stats.LastCard,
		"RecentPRs":   recentPRs,
	})
}
