package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	taskdomain "example.com/taskservice/internal/domain/task"
	taskusecase "example.com/taskservice/internal/usecase/task"
)

type TaskHandler struct {
	usecase taskusecase.Usecase
}

func NewTaskHandler(usecase taskusecase.Usecase) *TaskHandler {
	return &TaskHandler{usecase: usecase}
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req taskMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	created, err := h.usecase.Create(r.Context(), taskusecase.CreateInput{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, newTaskDTO(created))
}

func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	task, err := h.usecase.GetByID(r.Context(), id)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newTaskDTO(task))
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var req taskMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	updated, err := h.usecase.Update(r.Context(), id, taskusecase.UpdateInput{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newTaskDTO(updated))
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.usecase.Delete(r.Context(), id); err != nil {
		writeUsecaseError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	if scheduledForRaw := r.URL.Query().Get("scheduled_for"); scheduledForRaw != "" {
		scheduledFor, err := time.Parse(time.DateOnly, scheduledForRaw)
		if err != nil {
			writeError(w, http.StatusBadRequest, errors.New("invalid scheduled_for date, expected YYYY-MM-DD"))
			return
		}

		tasks, err := h.usecase.ListByScheduledFor(r.Context(), scheduledFor)
		if err != nil {
			writeUsecaseError(w, err)
			return
		}

		response := make([]taskDTO, 0, len(tasks))
		for i := range tasks {
			response = append(response, newTaskDTO(&tasks[i]))
		}

		writeJSON(w, http.StatusOK, response)
		return
	}

	tasks, err := h.usecase.List(r.Context())
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	response := make([]taskDTO, 0, len(tasks))
	for i := range tasks {
		response = append(response, newTaskDTO(&tasks[i]))
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *TaskHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	var req taskTemplateMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	input, err := toCreateTemplateInput(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	created, err := h.usecase.CreateTemplate(r.Context(), input)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, newTaskTemplateDTO(created))
}

func (h *TaskHandler) GetTemplateByID(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	template, err := h.usecase.GetTemplateByID(r.Context(), id)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newTaskTemplateDTO(template))
}

func (h *TaskHandler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var req taskTemplateMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	input, err := toUpdateTemplateInput(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	updated, err := h.usecase.UpdateTemplate(r.Context(), id, input)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newTaskTemplateDTO(updated))
}

func (h *TaskHandler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.usecase.DeleteTemplate(r.Context(), id); err != nil {
		writeUsecaseError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	templates, err := h.usecase.ListTemplates(r.Context())
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	response := make([]taskTemplateDTO, 0, len(templates))
	for i := range templates {
		response = append(response, newTaskTemplateDTO(&templates[i]))
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *TaskHandler) GenerateTasksForDate(w http.ResponseWriter, r *http.Request) {
	dateRaw := r.URL.Query().Get("date")
	if dateRaw == "" {
		writeError(w, http.StatusBadRequest, errors.New("missing required query parameter: date"))
		return
	}

	date, err := time.Parse(time.DateOnly, dateRaw)
	if err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid date format, expected YYYY-MM-DD"))
		return
	}

	var templateID *int64
	if templateIDRaw := r.URL.Query().Get("template_id"); templateIDRaw != "" {
		parsedTemplateID, err := strconv.ParseInt(templateIDRaw, 10, 64)
		if err != nil || parsedTemplateID <= 0 {
			writeError(w, http.StatusBadRequest, errors.New("invalid template_id, expected positive integer"))
			return
		}
		templateID = &parsedTemplateID
	}

	tasks, err := h.usecase.GenerateTasksForDate(r.Context(), date, templateID)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	response := make([]taskDTO, 0, len(tasks))
	for i := range tasks {
		response = append(response, newTaskDTO(&tasks[i]))
	}

	writeJSON(w, http.StatusOK, response)
}

func getIDFromRequest(r *http.Request) (int64, error) {
	rawID := mux.Vars(r)["id"]
	if rawID == "" {
		return 0, errors.New("missing task id")
	}

	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		return 0, errors.New("invalid task id")
	}

	if id <= 0 {
		return 0, errors.New("invalid task id")
	}

	return id, nil
}

func decodeJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return err
	}

	return nil
}

func toCreateTemplateInput(req taskTemplateMutationDTO) (taskusecase.CreateTemplateInput, error) {
	startsOn, err := parseDate(req.StartsOn, "starts_on")
	if err != nil {
		return taskusecase.CreateTemplateInput{}, err
	}

	var endsOn *time.Time
	if req.EndsOn != nil {
		parsedEndsOn, err := parseDate(*req.EndsOn, "ends_on")
		if err != nil {
			return taskusecase.CreateTemplateInput{}, err
		}
		endsOn = &parsedEndsOn
	}

	recurrence, err := toRecurrenceInput(req.Recurrence)
	if err != nil {
		return taskusecase.CreateTemplateInput{}, err
	}

	return taskusecase.CreateTemplateInput{
		Title:       req.Title,
		Description: req.Description,
		Recurrence:  recurrence,
		StartsOn:    startsOn,
		EndsOn:      endsOn,
		IsActive:    req.IsActive,
	}, nil
}

func toUpdateTemplateInput(req taskTemplateMutationDTO) (taskusecase.UpdateTemplateInput, error) {
	startsOn, err := parseDate(req.StartsOn, "starts_on")
	if err != nil {
		return taskusecase.UpdateTemplateInput{}, err
	}

	var endsOn *time.Time
	if req.EndsOn != nil {
		parsedEndsOn, err := parseDate(*req.EndsOn, "ends_on")
		if err != nil {
			return taskusecase.UpdateTemplateInput{}, err
		}
		endsOn = &parsedEndsOn
	}

	recurrence, err := toRecurrenceInput(req.Recurrence)
	if err != nil {
		return taskusecase.UpdateTemplateInput{}, err
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	return taskusecase.UpdateTemplateInput{
		Title:       req.Title,
		Description: req.Description,
		Recurrence:  recurrence,
		StartsOn:    startsOn,
		EndsOn:      endsOn,
		IsActive:    isActive,
	}, nil
}

func toRecurrenceInput(req recurrenceDTO) (taskusecase.RecurrenceInput, error) {
	var specificDates []time.Time
	if len(req.SpecificDates) > 0 {
		specificDates = make([]time.Time, 0, len(req.SpecificDates))
		for _, raw := range req.SpecificDates {
			parsedDate, err := parseDate(raw, "specific_dates")
			if err != nil {
				return taskusecase.RecurrenceInput{}, err
			}
			specificDates = append(specificDates, parsedDate)
		}
	}

	return taskusecase.RecurrenceInput{
		Type:          req.Type,
		IntervalDays:  req.IntervalDays,
		MonthDays:     req.MonthDays,
		SpecificDates: specificDates,
		OddEvenType:   req.OddEvenType,
	}, nil
}

func parseDate(raw string, field string) (time.Time, error) {
	date, err := time.Parse(time.DateOnly, raw)
	if err != nil {
		return time.Time{}, errors.New("invalid " + field + " date, expected YYYY-MM-DD")
	}

	return date, nil
}

func writeUsecaseError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, taskdomain.ErrNotFound):
		writeError(w, http.StatusNotFound, err)
	case errors.Is(err, taskusecase.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err)
	default:
		writeError(w, http.StatusInternalServerError, err)
	}
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{
		"error": err.Error(),
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(payload)
}
