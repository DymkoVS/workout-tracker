CREATE TABLE workout_media (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    workout_id    UUID        NOT NULL REFERENCES workouts(id) ON DELETE CASCADE,
    filename      VARCHAR(255) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    mime_type     VARCHAR(100) NOT NULL,
    size_bytes    INTEGER     NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ON workout_media(workout_id, created_at);
