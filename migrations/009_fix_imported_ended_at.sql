-- Импортированные исторические тренировки не имеют ended_at, потому что
-- создавались через importer.Confirm до того, как он начал передавать ended_at.
-- Помечаем их как завершённые: ended_at = workout_date.
-- Условие: прошедшая дата, ни started_at, ни ended_at не заполнены.
UPDATE workouts
SET ended_at = workout_date
WHERE ended_at IS NULL
  AND started_at IS NULL
  AND workout_date < CURRENT_DATE;
