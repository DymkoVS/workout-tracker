-- Комментарии к тренировке: тренер пишет фидбэк клиенту, клиент может ответить.
CREATE TABLE IF NOT EXISTS workout_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workout_id UUID NOT NULL REFERENCES workouts(id) ON DELETE CASCADE,
    author_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    body TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_workout_comments_workout
    ON workout_comments (workout_id, created_at);
