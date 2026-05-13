-- Add type field to workout_templates (Сила / Кардио / Аксессуар)
ALTER TABLE workout_templates
    ADD COLUMN IF NOT EXISTS type TEXT NOT NULL DEFAULT 'Сила';

-- Link workouts back to the template they were created from
ALTER TABLE workouts
    ADD COLUMN IF NOT EXISTS template_id UUID REFERENCES workout_templates(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS workouts_template_id_idx ON workouts(template_id);
