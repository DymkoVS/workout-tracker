-- Спортивные залы
CREATE TABLE gyms (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO gyms (name) VALUES ('Основной зал');

-- Тренировки
CREATE TABLE workouts (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    trainer_id   UUID        REFERENCES users(id) ON DELETE SET NULL,
    gym_id       UUID        REFERENCES gyms(id) ON DELETE SET NULL,
    title        VARCHAR(255) NOT NULL DEFAULT '',
    workout_date DATE        NOT NULL DEFAULT CURRENT_DATE,
    notes        TEXT        NOT NULL DEFAULT '',
    wellbeing    SMALLINT    CHECK (wellbeing BETWEEN 1 AND 5),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ON workouts(user_id, workout_date DESC);

-- Упражнения в тренировке
CREATE TABLE workout_exercises (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    workout_id  UUID        NOT NULL REFERENCES workouts(id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL,
    order_num   INTEGER     NOT NULL DEFAULT 0,
    notes       TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ON workout_exercises(workout_id, order_num);

-- Подходы
CREATE TABLE sets (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    workout_exercise_id UUID        NOT NULL REFERENCES workout_exercises(id) ON DELETE CASCADE,
    set_num             INTEGER     NOT NULL DEFAULT 1,
    weight              NUMERIC(6,2),
    reps                INTEGER,
    rpe                 NUMERIC(3,1),
    rest_seconds        INTEGER,
    notes               TEXT        NOT NULL DEFAULT ''
);

CREATE INDEX ON sets(workout_exercise_id, set_num);
