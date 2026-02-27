package handler

import (
	"errors"
	"net/http"
	"strconv"
	"task_vault/internal/adapter/http/middleware"
	"task_vault/internal/app/command"
	"task_vault/internal/app/query"
	"task_vault/internal/domain"
	"task_vault/internal/ports"

	"github.com/go-chi/chi/v5"
)

type TaskHandler struct {
	createTask     *command.CreateTaskHandler
	updateTask     *command.UpdateTaskHandler
	getTasks       *query.GetTasksHandler
	getTaskHistory *query.GetTaskHistoryHandler
}

func NewTaskHandler(
	createTask *command.CreateTaskHandler,
	updateTask *command.UpdateTaskHandler,
	getTasks *query.GetTasksHandler,
	getTaskHistory *query.GetTaskHistoryHandler,
) *TaskHandler {
	return &TaskHandler{
		createTask:     createTask,
		updateTask:     updateTask,
		getTasks:       getTasks,
		getTaskHistory: getTaskHistory,
	}
}

// @Summary      Создание задачи
// @Description  Создает задачу в указанной команде
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input body      CreateTaskRequest true "Данные задачи"
// @Success      201   {object}  swagTaskResponse
// @Failure      400   {object}  swagErrorResponse
// @Failure      403   {object}  swagErrorResponse
// @Failure      500   {object}  swagErrorResponse
// @Router       /tasks [post]
func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateTaskRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "невалидный JSON")
		return
	}
	if verr := req.Validate(); verr != nil {
		writeValidationError(w, verr)
		return
	}

	userID := middleware.GetUserID(r.Context())
	task, err := h.createTask.Handle(r.Context(), command.CreateTaskInput{
		Title:       req.Title,
		Description: req.Description,
		AssigneeID:  req.AssigneeID,
		TeamID:      req.TeamID,
		CreatedBy:   userID,
	})
	if err != nil {
		if errors.Is(err, domain.ErrNotTeamMember) {
			writeError(w, http.StatusForbidden, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "ошибка создания задачи")
		return
	}

	writeJSON(w, http.StatusCreated, toTaskResponse(task))
}

// @Summary      Список задач
// @Description  Возвращает задачи с фильтрацией и пагинацией
// @Tags         tasks
// @Produce      json
// @Security     BearerAuth
// @Param        team_id     query string false "ID команды"
// @Param        status      query string false "Статус (todo, in_progress, done)"
// @Param        assignee_id query string false "ID исполнителя"
// @Param        limit       query int    false "Лимит (по умолчанию 20)"
// @Param        offset      query int    false "Смещение (по умолчанию 0)"
// @Success      200         {object}  swagTaskListResponse
// @Failure      500         {object}  swagErrorResponse
// @Router       /tasks [get]
func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := ports.TaskFilter{
		Limit:  20,
		Offset: 0,
	}

	if v := r.URL.Query().Get("team_id"); v != "" {
		filter.TeamID = &v
	}
	if v := r.URL.Query().Get("status"); v != "" {
		s := domain.Status(v)
		filter.Status = &s
	}
	if v := r.URL.Query().Get("assignee_id"); v != "" {
		filter.AssigneeID = &v
	}
	if v := r.URL.Query().Get("limit"); v != "" {
		n, _ := strconv.Atoi(v)
		if n > 0 {
			filter.Limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		n, _ := strconv.Atoi(v)
		if n >= 0 {
			filter.Offset = n
		}
	}

	output, err := h.getTasks.Handle(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка получения задач")
		return
	}

	tasks := make([]TaskResponse, 0, len(output.Tasks))
	for _, t := range output.Tasks {
		tasks = append(tasks, toTaskResponse(&t))
	}
	writeJSON(w, http.StatusOK, TaskListResponse{Tasks: tasks, Total: output.Total})
}

// @Summary      Обновление задачи
// @Description  Частичное обновление полей задачи
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string            true "ID задачи"
// @Param        input body      UpdateTaskRequest  true "Обновляемые поля"
// @Success      200   {object}  swagTaskResponse
// @Failure      400   {object}  swagErrorResponse
// @Failure      403   {object}  swagErrorResponse
// @Failure      404   {object}  swagErrorResponse
// @Failure      500   {object}  swagErrorResponse
// @Router       /tasks/{id} [put]
func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	if taskID == "" {
		writeError(w, http.StatusBadRequest, "невалидный ID задачи")
		return
	}

	var req UpdateTaskRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "невалидный JSON")
		return
	}
	if verr := req.Validate(); verr != nil {
		writeValidationError(w, verr)
		return
	}

	userID := middleware.GetUserID(r.Context())
	task, err := h.updateTask.Handle(r.Context(), command.UpdateTaskInput{
		TaskID:      taskID,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		AssigneeID:  req.AssigneeID,
		UpdatedBy:   userID,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrTaskNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, domain.ErrNoPermission):
			writeError(w, http.StatusForbidden, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "ошибка обновления задачи")
		}
		return
	}

	writeJSON(w, http.StatusOK, toTaskResponse(task))
}

// @Summary      История изменений задачи
// @Description  Возвращает лог изменений полей задачи
// @Tags         tasks
// @Produce      json
// @Security     BearerAuth
// @Param        id  path      string true "ID задачи"
// @Success      200 {object}  swagHistoryListResponse
// @Failure      400 {object}  swagErrorResponse
// @Failure      500 {object}  swagErrorResponse
// @Router       /tasks/{id}/history [get]
func (h *TaskHandler) History(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	if taskID == "" {
		writeError(w, http.StatusBadRequest, "невалидный ID задачи")
		return
	}

	history, err := h.getTaskHistory.Handle(r.Context(), taskID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка получения истории")
		return
	}

	result := make([]HistoryEntry, 0, len(history))
	for _, h := range history {
		result = append(result, HistoryEntry{
			ID:        h.ID,
			FieldName: h.FieldName,
			OldValue:  h.OldValue,
			NewValue:  h.NewValue,
			ChangedBy: h.ChangedBy,
		})
	}
	writeJSON(w, http.StatusOK, result)
}

func toTaskResponse(t *domain.Task) TaskResponse {
	return TaskResponse{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Status:      string(t.Status),
		AssigneeID:  t.AssigneeID,
		TeamID:      t.TeamID,
		CreatedBy:   t.CreatedBy,
	}
}
