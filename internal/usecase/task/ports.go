package task

import (
	"context"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository interface {
	Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	CreateScheduled(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)
	ListByScheduledFor(ctx context.Context, date time.Time) ([]taskdomain.Task, error)

	CreateTemplate(ctx context.Context, template *taskdomain.TaskTemplate) (*taskdomain.TaskTemplate, error)
	GetTemplateByID(ctx context.Context, id int64) (*taskdomain.TaskTemplate, error)
	UpdateTemplate(ctx context.Context, template *taskdomain.TaskTemplate) (*taskdomain.TaskTemplate, error)
	DeleteTemplate(ctx context.Context, id int64) error
	ListTemplates(ctx context.Context) ([]taskdomain.TaskTemplate, error)
	ListActiveTemplates(ctx context.Context) ([]taskdomain.TaskTemplate, error)
}

type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)
	ListByScheduledFor(ctx context.Context, date time.Time) ([]taskdomain.Task, error)

	CreateTemplate(ctx context.Context, input CreateTemplateInput) (*taskdomain.TaskTemplate, error)
	GetTemplateByID(ctx context.Context, id int64) (*taskdomain.TaskTemplate, error)
	UpdateTemplate(ctx context.Context, id int64, input UpdateTemplateInput) (*taskdomain.TaskTemplate, error)
	DeleteTemplate(ctx context.Context, id int64) error
	ListTemplates(ctx context.Context) ([]taskdomain.TaskTemplate, error)
	GenerateTasksForDate(ctx context.Context, date time.Time, templateID *int64) ([]taskdomain.Task, error)
}

type CreateInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
}

type UpdateInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
}

type RecurrenceInput struct {
	Type          taskdomain.RecurrenceType
	IntervalDays  int
	MonthDays     []int
	SpecificDates []time.Time
	OddEvenType   taskdomain.OddEvenType
}

type CreateTemplateInput struct {
	Title       string
	Description string
	Recurrence  RecurrenceInput
	StartsOn    time.Time
	EndsOn      *time.Time
	IsActive    *bool
}

type UpdateTemplateInput struct {
	Title       string
	Description string
	Recurrence  RecurrenceInput
	StartsOn    time.Time
	EndsOn      *time.Time
	IsActive    bool
}
