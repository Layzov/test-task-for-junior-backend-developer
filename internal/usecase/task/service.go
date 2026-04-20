package task

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error) {
	normalized, err := validateCreateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
	}
	now := s.now()
	model.CreatedAt = now
	model.UpdatedAt = now

	created, err := s.repo.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		ID:          id,
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		UpdatedAt:   s.now(),
	}

	updated, err := s.repo.Update(ctx, model)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]taskdomain.Task, error) {
	return s.repo.List(ctx)
}

func (s *Service) ListByScheduledFor(ctx context.Context, date time.Time) ([]taskdomain.Task, error) {
	return s.repo.ListByScheduledFor(ctx, normalizeDate(date))
}

func (s *Service) CreateTemplate(ctx context.Context, input CreateTemplateInput) (*taskdomain.TaskTemplate, error) {
	normalized, err := validateCreateTemplateInput(input)
	if err != nil {
		return nil, err
	}

	now := s.now()
	model := &taskdomain.TaskTemplate{
		Title:       normalized.Title,
		Description: normalized.Description,
		Recurrence:  toDomainRecurrence(normalized.Recurrence),
		StartsOn:    normalizeDate(normalized.StartsOn),
		EndsOn:      normalizeDatePtr(normalized.EndsOn),
		IsActive:    *normalized.IsActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return s.repo.CreateTemplate(ctx, model)
}

func (s *Service) GetTemplateByID(ctx context.Context, id int64) (*taskdomain.TaskTemplate, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetTemplateByID(ctx, id)
}

func (s *Service) UpdateTemplate(ctx context.Context, id int64, input UpdateTemplateInput) (*taskdomain.TaskTemplate, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateTemplateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.TaskTemplate{
		ID:          id,
		Title:       normalized.Title,
		Description: normalized.Description,
		Recurrence:  toDomainRecurrence(normalized.Recurrence),
		StartsOn:    normalizeDate(normalized.StartsOn),
		EndsOn:      normalizeDatePtr(normalized.EndsOn),
		IsActive:    normalized.IsActive,
		UpdatedAt:   s.now(),
	}

	return s.repo.UpdateTemplate(ctx, model)
}

func (s *Service) DeleteTemplate(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.DeleteTemplate(ctx, id)
}

func (s *Service) ListTemplates(ctx context.Context) ([]taskdomain.TaskTemplate, error) {
	return s.repo.ListTemplates(ctx)
}

func (s *Service) GenerateTasksForDate(ctx context.Context, date time.Time, templateID *int64) ([]taskdomain.Task, error) {
	scheduledFor := normalizeDate(date)
	templates := make([]taskdomain.TaskTemplate, 0)
	if templateID != nil {
		if *templateID <= 0 {
			return nil, fmt.Errorf("%w: template_id must be positive", ErrInvalidInput)
		}

		template, err := s.repo.GetTemplateByID(ctx, *templateID)
		if err != nil {
			return nil, err
		}
		templates = append(templates, *template)
	} else {
		activeTemplates, err := s.repo.ListActiveTemplates(ctx)
		if err != nil {
			return nil, err
		}
		templates = activeTemplates
	}

	createdTasks := make([]taskdomain.Task, 0, len(templates))
	for i := range templates {
		template := templates[i]
		if !shouldGenerateOnDate(template, scheduledFor) {
			continue
		}

		now := s.now()
		task := &taskdomain.Task{
			Title:        template.Title,
			Description:  template.Description,
			Status:       taskdomain.StatusNew,
			TemplateID:   &template.ID,
			ScheduledFor: &scheduledFor,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		created, err := s.repo.CreateScheduled(ctx, task)
		if err != nil {
			return nil, err
		}

		createdTasks = append(createdTasks, *created)
	}

	return createdTasks, nil
}

func validateCreateInput(input CreateInput) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if input.Status == "" {
		input.Status = taskdomain.StatusNew
	}

	if !input.Status.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	return input, nil
}

func validateUpdateInput(input UpdateInput) (UpdateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return UpdateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if !input.Status.Valid() {
		return UpdateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	return input, nil
}

func validateCreateTemplateInput(input CreateTemplateInput) (CreateTemplateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateTemplateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	recurrence, err := validateRecurrenceInput(input.Recurrence)
	if err != nil {
		return CreateTemplateInput{}, err
	}

	if input.StartsOn.IsZero() {
		return CreateTemplateInput{}, fmt.Errorf("%w: starts_on is required", ErrInvalidInput)
	}

	startsOn := normalizeDate(input.StartsOn)
	var endsOn *time.Time
	if input.EndsOn != nil {
		normalizedEndsOn := normalizeDate(*input.EndsOn)
		if normalizedEndsOn.Before(startsOn) {
			return CreateTemplateInput{}, fmt.Errorf("%w: ends_on must not be before starts_on", ErrInvalidInput)
		}

		endsOn = &normalizedEndsOn
	}
	input.EndsOn = endsOn

	if input.IsActive == nil {
		active := true
		input.IsActive = &active
	}

	input.Recurrence = recurrence
	input.StartsOn = startsOn
	return input, nil
}

func validateUpdateTemplateInput(input UpdateTemplateInput) (UpdateTemplateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return UpdateTemplateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	recurrence, err := validateRecurrenceInput(input.Recurrence)
	if err != nil {
		return UpdateTemplateInput{}, err
	}

	if input.StartsOn.IsZero() {
		return UpdateTemplateInput{}, fmt.Errorf("%w: starts_on is required", ErrInvalidInput)
	}

	startsOn := normalizeDate(input.StartsOn)
	var endsOn *time.Time
	if input.EndsOn != nil {
		normalizedEndsOn := normalizeDate(*input.EndsOn)
		if normalizedEndsOn.Before(startsOn) {
			return UpdateTemplateInput{}, fmt.Errorf("%w: ends_on must not be before starts_on", ErrInvalidInput)
		}

		endsOn = &normalizedEndsOn
	}

	input.Recurrence = recurrence
	input.StartsOn = startsOn
	input.EndsOn = endsOn
	return input, nil
}

func validateRecurrenceInput(input RecurrenceInput) (RecurrenceInput, error) {
	if !input.Type.Valid() {
		return RecurrenceInput{}, fmt.Errorf("%w: invalid recurrence type", ErrInvalidInput)
	}

	switch input.Type {
	case taskdomain.RecurrenceDailyInterval:
		if input.IntervalDays < 1 {
			return RecurrenceInput{}, fmt.Errorf("%w: interval_days must be >= 1", ErrInvalidInput)
		}
	case taskdomain.RecurrenceMonthlyDays:
		if len(input.MonthDays) == 0 {
			return RecurrenceInput{}, fmt.Errorf("%w: month_days must not be empty", ErrInvalidInput)
		}

		monthDays := make([]int, 0, len(input.MonthDays))
		for _, day := range input.MonthDays {
			if day < 1 || day > 30 {
				return RecurrenceInput{}, fmt.Errorf("%w: month_days values must be from 1 to 30", ErrInvalidInput)
			}
			if slices.Contains(monthDays, day) {
				continue
			}
			monthDays = append(monthDays, day)
		}
		slices.Sort(monthDays)
		input.MonthDays = monthDays
	case taskdomain.RecurrenceSpecificDates:
		if len(input.SpecificDates) == 0 {
			return RecurrenceInput{}, fmt.Errorf("%w: specific_dates must not be empty", ErrInvalidInput)
		}

		specificDates := make([]time.Time, 0, len(input.SpecificDates))
		for _, dt := range input.SpecificDates {
			if dt.IsZero() {
				return RecurrenceInput{}, fmt.Errorf("%w: specific_dates contains zero date", ErrInvalidInput)
			}
			normalizedDate := normalizeDate(dt)
			if slices.ContainsFunc(specificDates, func(candidate time.Time) bool {
				return candidate.Equal(normalizedDate)
			}) {
				continue
			}
			specificDates = append(specificDates, normalizedDate)
		}
		slices.SortFunc(specificDates, func(a, b time.Time) int {
			if a.Before(b) {
				return -1
			}
			if a.After(b) {
				return 1
			}
			return 0
		})
		input.SpecificDates = specificDates
	case taskdomain.RecurrenceOddEven:
		if !input.OddEvenType.Valid() {
			return RecurrenceInput{}, fmt.Errorf("%w: odd_even_type must be odd or even", ErrInvalidInput)
		}
	default:
		return RecurrenceInput{}, fmt.Errorf("%w: invalid recurrence type", ErrInvalidInput)
	}

	return input, nil
}

func shouldGenerateOnDate(template taskdomain.TaskTemplate, date time.Time) bool {
	if !template.IsActive {
		return false
	}
	if date.Before(template.StartsOn) {
		return false
	}
	if template.EndsOn != nil && date.After(*template.EndsOn) {
		return false
	}

	switch template.Recurrence.Type {
	case taskdomain.RecurrenceDailyInterval:
		daysFromStart := int(date.Sub(template.StartsOn).Hours() / 24)
		return daysFromStart%template.Recurrence.IntervalDays == 0
	case taskdomain.RecurrenceMonthlyDays:
		return slices.Contains(template.Recurrence.MonthDays, date.Day())
	case taskdomain.RecurrenceSpecificDates:
		return slices.ContainsFunc(template.Recurrence.SpecificDates, func(candidate time.Time) bool {
			return candidate.Equal(date)
		})
	case taskdomain.RecurrenceOddEven:
		isEven := date.Day()%2 == 0
		return (isEven && template.Recurrence.OddEvenType == taskdomain.OddEvenEven) ||
			(!isEven && template.Recurrence.OddEvenType == taskdomain.OddEvenOdd)
	default:
		return false
	}
}

func normalizeDate(date time.Time) time.Time {
	utc := date.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}

func normalizeDatePtr(date *time.Time) *time.Time {
	if date == nil {
		return nil
	}
	normalized := normalizeDate(*date)
	return &normalized
}

func toDomainRecurrence(input RecurrenceInput) taskdomain.Recurrence {
	return taskdomain.Recurrence{
		Type:          input.Type,
		IntervalDays:  input.IntervalDays,
		MonthDays:     input.MonthDays,
		SpecificDates: input.SpecificDates,
		OddEvenType:   input.OddEvenType,
	}
}
