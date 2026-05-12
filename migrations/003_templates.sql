CREATE TABLE workout_templates (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    trainer_id UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title      VARCHAR(255) NOT NULL DEFAULT '',
    notes      TEXT         NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX ON workout_templates(trainer_id, created_at DESC);

CREATE TABLE template_exercises (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID         NOT NULL REFERENCES workout_templates(id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL,
    order_num   INTEGER      NOT NULL DEFAULT 1,
    notes       TEXT         NOT NULL DEFAULT ''
);

CREATE INDEX ON template_exercises(template_id, order_num);

CREATE TABLE template_sets (
    id                   UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    template_exercise_id UUID        NOT NULL REFERENCES template_exercises(id) ON DELETE CASCADE,
    set_num              INTEGER     NOT NULL DEFAULT 1,
    weight               NUMERIC(6,2),
    reps                 INTEGER,
    rpe                  NUMERIC(3,1),
    rest_seconds         INTEGER,
    notes                TEXT        NOT NULL DEFAULT ''
);

CREATE INDEX ON template_sets(template_exercise_id, set_num);
