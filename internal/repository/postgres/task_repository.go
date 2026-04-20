package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		INSERT INTO tasks (title, description, status, template_id, scheduled_for, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, title, description, status, template_id, scheduled_for, created_at, updated_at
	`

	row := r.pool.QueryRow(
		ctx,
		query,
		task.Title,
		task.Description,
		task.Status,
		task.TemplateID,
		task.ScheduledFor,
		task.CreatedAt,
		task.UpdatedAt,
	)
	created, err := scanTask(row)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *Repository) CreateScheduled(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		INSERT INTO tasks (title, description, status, template_id, scheduled_for, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (template_id, scheduled_for)
		DO UPDATE SET
			title = EXCLUDED.title,
			description = EXCLUDED.description
		RETURNING id, title, description, status, template_id, scheduled_for, created_at, updated_at
	`

	row := r.pool.QueryRow(
		ctx,
		query,
		task.Title,
		task.Description,
		task.Status,
		task.TemplateID,
		task.ScheduledFor,
		task.CreatedAt,
		task.UpdatedAt,
	)

	return scanTask(row)
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, template_id, scheduled_for, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	found, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return found, nil
}

func (r *Repository) Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		UPDATE tasks
		SET title = $1,
			description = $2,
			status = $3,
			updated_at = $4
		WHERE id = $5
		RETURNING id, title, description, status, template_id, scheduled_for, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.UpdatedAt, task.ID)
	updated, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return updated, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM tasks WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return taskdomain.ErrNotFound
	}

	return nil
}

func (r *Repository) List(ctx context.Context) ([]taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, template_id, scheduled_for, created_at, updated_at
		FROM tasks
		ORDER BY id DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]taskdomain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, *task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *Repository) ListByScheduledFor(ctx context.Context, date time.Time) ([]taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, template_id, scheduled_for, created_at, updated_at
		FROM tasks
		WHERE scheduled_for = $1::date
		ORDER BY id DESC
	`

	rows, err := r.pool.Query(ctx, query, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]taskdomain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, *task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *Repository) CreateTemplate(ctx context.Context, template *taskdomain.TaskTemplate) (*taskdomain.TaskTemplate, error) {
	const query = `
		INSERT INTO task_templates (
			title, description, recurrence_type, interval_days, month_days, specific_dates, odd_even_type,
			starts_on, ends_on, is_active, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, COALESCE($5, '{}'::int[]), COALESCE($6, '{}'::date[]), $7, $8, $9, $10, $11, $12)
		RETURNING id, title, description, recurrence_type, interval_days, month_days, specific_dates, odd_even_type,
			starts_on, ends_on, is_active, created_at, updated_at
	`

	row := r.pool.QueryRow(
		ctx,
		query,
		template.Title,
		template.Description,
		template.Recurrence.Type,
		nullableInt(template.Recurrence.IntervalDays),
		template.Recurrence.MonthDays,
		template.Recurrence.SpecificDates,
		nullableString(string(template.Recurrence.OddEvenType)),
		template.StartsOn,
		template.EndsOn,
		template.IsActive,
		template.CreatedAt,
		template.UpdatedAt,
	)

	return scanTaskTemplate(row)
}

func (r *Repository) GetTemplateByID(ctx context.Context, id int64) (*taskdomain.TaskTemplate, error) {
	const query = `
		SELECT id, title, description, recurrence_type, interval_days, month_days, specific_dates, odd_even_type,
			starts_on, ends_on, is_active, created_at, updated_at
		FROM task_templates
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	template, err := scanTaskTemplate(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return template, nil
}

func (r *Repository) UpdateTemplate(ctx context.Context, template *taskdomain.TaskTemplate) (*taskdomain.TaskTemplate, error) {
	const query = `
		UPDATE task_templates
		SET title = $1,
			description = $2,
			recurrence_type = $3,
			interval_days = $4,
			month_days = COALESCE($5, '{}'::int[]),
			specific_dates = COALESCE($6, '{}'::date[]),
			odd_even_type = $7,
			starts_on = $8,
			ends_on = $9,
			is_active = $10,
			updated_at = $11
		WHERE id = $12
		RETURNING id, title, description, recurrence_type, interval_days, month_days, specific_dates, odd_even_type,
			starts_on, ends_on, is_active, created_at, updated_at
	`

	row := r.pool.QueryRow(
		ctx,
		query,
		template.Title,
		template.Description,
		template.Recurrence.Type,
		nullableInt(template.Recurrence.IntervalDays),
		template.Recurrence.MonthDays,
		template.Recurrence.SpecificDates,
		nullableString(string(template.Recurrence.OddEvenType)),
		template.StartsOn,
		template.EndsOn,
		template.IsActive,
		template.UpdatedAt,
		template.ID,
	)
	updated, err := scanTaskTemplate(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return updated, nil
}

func (r *Repository) DeleteTemplate(ctx context.Context, id int64) error {
	const query = `DELETE FROM task_templates WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return taskdomain.ErrNotFound
	}

	return nil
}

func (r *Repository) ListTemplates(ctx context.Context) ([]taskdomain.TaskTemplate, error) {
	const query = `
		SELECT id, title, description, recurrence_type, interval_days, month_days, specific_dates, odd_even_type,
			starts_on, ends_on, is_active, created_at, updated_at
		FROM task_templates
		ORDER BY id DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	templates := make([]taskdomain.TaskTemplate, 0)
	for rows.Next() {
		template, err := scanTaskTemplate(rows)
		if err != nil {
			return nil, err
		}

		templates = append(templates, *template)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return templates, nil
}

func (r *Repository) ListActiveTemplates(ctx context.Context) ([]taskdomain.TaskTemplate, error) {
	const query = `
		SELECT id, title, description, recurrence_type, interval_days, month_days, specific_dates, odd_even_type,
			starts_on, ends_on, is_active, created_at, updated_at
		FROM task_templates
		WHERE is_active = TRUE
		ORDER BY id DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	templates := make([]taskdomain.TaskTemplate, 0)
	for rows.Next() {
		template, err := scanTaskTemplate(rows)
		if err != nil {
			return nil, err
		}

		templates = append(templates, *template)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return templates, nil
}

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (*taskdomain.Task, error) {
	var (
		task         taskdomain.Task
		status       string
		templateID   sql.NullInt64
		scheduledFor sql.NullTime
	)

	if err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&status,
		&templateID,
		&scheduledFor,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)
	if templateID.Valid {
		id := templateID.Int64
		task.TemplateID = &id
	}
	if scheduledFor.Valid {
		scheduledDate := normalizeDate(scheduledFor.Time)
		task.ScheduledFor = &scheduledDate
	}

	return &task, nil
}

func scanTaskTemplate(scanner taskScanner) (*taskdomain.TaskTemplate, error) {
	var (
		template     taskdomain.TaskTemplate
		recurrence   taskdomain.Recurrence
		intervalDays sql.NullInt64
		oddEvenType  sql.NullString
		endsOn       sql.NullTime
	)

	if err := scanner.Scan(
		&template.ID,
		&template.Title,
		&template.Description,
		&recurrence.Type,
		&intervalDays,
		&recurrence.MonthDays,
		&recurrence.SpecificDates,
		&oddEvenType,
		&template.StartsOn,
		&endsOn,
		&template.IsActive,
		&template.CreatedAt,
		&template.UpdatedAt,
	); err != nil {
		return nil, err
	}

	if intervalDays.Valid {
		recurrence.IntervalDays = int(intervalDays.Int64)
	}
	if oddEvenType.Valid {
		recurrence.OddEvenType = taskdomain.OddEvenType(oddEvenType.String)
	}
	template.StartsOn = normalizeDate(template.StartsOn)
	if endsOn.Valid {
		normalizedEndsOn := normalizeDate(endsOn.Time)
		template.EndsOn = &normalizedEndsOn
	}
	for i := range recurrence.SpecificDates {
		recurrence.SpecificDates[i] = normalizeDate(recurrence.SpecificDates[i])
	}
	template.Recurrence = recurrence

	return &template, nil
}

func nullableInt(value int) *int {
	if value == 0 {
		return nil
	}
	v := value
	return &v
}

func nullableString(value string) *string {
	if value == "" {
		return nil
	}
	v := value
	return &v
}

func normalizeDate(date time.Time) time.Time {
	utc := date.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}
