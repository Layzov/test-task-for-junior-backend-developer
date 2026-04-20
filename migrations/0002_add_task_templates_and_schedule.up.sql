CREATE TABLE IF NOT EXISTS task_templates (
	id BIGSERIAL PRIMARY KEY,
	title TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	recurrence_type TEXT NOT NULL,
	interval_days INTEGER,
	month_days INTEGER[] NOT NULL DEFAULT '{}',
	specific_dates DATE[] NOT NULL DEFAULT '{}',
	odd_even_type TEXT,
	starts_on DATE NOT NULL,
	ends_on DATE,
	is_active BOOLEAN NOT NULL DEFAULT TRUE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE tasks
	ADD COLUMN IF NOT EXISTS template_id BIGINT REFERENCES task_templates(id) ON DELETE SET NULL,
	ADD COLUMN IF NOT EXISTS scheduled_for DATE;

CREATE UNIQUE INDEX IF NOT EXISTS idx_tasks_template_id_scheduled_for ON tasks (template_id, scheduled_for);
CREATE INDEX IF NOT EXISTS idx_tasks_scheduled_for ON tasks (scheduled_for);
CREATE INDEX IF NOT EXISTS idx_task_templates_is_active ON task_templates (is_active);
