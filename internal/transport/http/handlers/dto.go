package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type taskMutationDTO struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
}

type recurrenceDTO struct {
	Type          taskdomain.RecurrenceType `json:"type"`
	IntervalDays  int                       `json:"interval_days,omitempty"`
	MonthDays     []int                     `json:"month_days,omitempty"`
	SpecificDates []string                  `json:"specific_dates,omitempty"`
	OddEvenType   taskdomain.OddEvenType    `json:"odd_even_type,omitempty"`
}

type taskTemplateMutationDTO struct {
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Recurrence  recurrenceDTO `json:"recurrence"`
	StartsOn    string        `json:"starts_on"`
	EndsOn      *string       `json:"ends_on,omitempty"`
	IsActive    *bool         `json:"is_active,omitempty"`
}

type taskDTO struct {
	ID           int64             `json:"id"`
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	Status       taskdomain.Status `json:"status"`
	TemplateID   *int64            `json:"template_id,omitempty"`
	ScheduledFor *time.Time        `json:"scheduled_for,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

type taskTemplateDTO struct {
	ID          int64         `json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Recurrence  recurrenceDTO `json:"recurrence"`
	StartsOn    string        `json:"starts_on"`
	EndsOn      *string       `json:"ends_on,omitempty"`
	IsActive    bool          `json:"is_active"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:           task.ID,
		Title:        task.Title,
		Description:  task.Description,
		Status:       task.Status,
		TemplateID:   task.TemplateID,
		ScheduledFor: task.ScheduledFor,
		CreatedAt:    task.CreatedAt,
		UpdatedAt:    task.UpdatedAt,
	}
}

func newTaskTemplateDTO(template *taskdomain.TaskTemplate) taskTemplateDTO {
	recurrence := recurrenceDTO{
		Type:         template.Recurrence.Type,
		IntervalDays: template.Recurrence.IntervalDays,
		MonthDays:    template.Recurrence.MonthDays,
		OddEvenType:  template.Recurrence.OddEvenType,
	}
	if len(template.Recurrence.SpecificDates) > 0 {
		recurrence.SpecificDates = make([]string, 0, len(template.Recurrence.SpecificDates))
		for _, date := range template.Recurrence.SpecificDates {
			recurrence.SpecificDates = append(recurrence.SpecificDates, date.Format(time.DateOnly))
		}
	}

	startsOn := template.StartsOn.Format(time.DateOnly)
	var endsOn *string
	if template.EndsOn != nil {
		value := template.EndsOn.Format(time.DateOnly)
		endsOn = &value
	}

	return taskTemplateDTO{
		ID:          template.ID,
		Title:       template.Title,
		Description: template.Description,
		Recurrence:  recurrence,
		StartsOn:    startsOn,
		EndsOn:      endsOn,
		IsActive:    template.IsActive,
		CreatedAt:   template.CreatedAt,
		UpdatedAt:   template.UpdatedAt,
	}
}
