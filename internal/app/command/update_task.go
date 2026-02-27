package command

import (
	"context"
	"fmt"
	"task_vault/internal/domain"
	"task_vault/internal/ports"
)

type UpdateTaskInput struct {
	TaskID      string
	Title       *string
	Description *string
	Status      *domain.Status
	AssigneeID  *string
	UpdatedBy   string
}

type UpdateTaskHandler struct {
	taskCmd    ports.TaskCommandRepo
	taskQuery  ports.TaskQueryRepo
	teamQuery  ports.TeamQueryRepo
	historyCmd ports.HistoryCommandRepo
	cache      ports.Cache
}

func NewUpdateTaskHandler(
	taskCmd ports.TaskCommandRepo,
	taskQuery ports.TaskQueryRepo,
	teamQuery ports.TeamQueryRepo,
	historyCmd ports.HistoryCommandRepo,
	cache ports.Cache,
) *UpdateTaskHandler {
	return &UpdateTaskHandler{
		taskCmd:    taskCmd,
		taskQuery:  taskQuery,
		teamQuery:  teamQuery,
		historyCmd: historyCmd,
		cache:      cache,
	}
}

func (h *UpdateTaskHandler) Handle(ctx context.Context, input UpdateTaskInput) (*domain.Task, error) {
	task, err := h.taskQuery.GetByID(ctx, input.TaskID)
	if err != nil {
		return nil, domain.ErrTaskNotFound
	}

	if err := h.checkPermission(ctx, task, input.UpdatedBy); err != nil {
		return nil, err
	}

	if input.Title != nil && *input.Title != task.Title {
		if err := h.recordChange(ctx, task.ID, input.UpdatedBy, "title", task.Title, *input.Title); err != nil {
			return nil, err
		}
		task.Title = *input.Title
	}
	if input.Description != nil && *input.Description != task.Description {
		if err := h.recordChange(ctx, task.ID, input.UpdatedBy, "description", task.Description, *input.Description); err != nil {
			return nil, err
		}
		task.Description = *input.Description
	}
	if input.Status != nil && *input.Status != task.Status {
		if err := h.recordChange(ctx, task.ID, input.UpdatedBy, "status", string(task.Status), string(*input.Status)); err != nil {
			return nil, err
		}
		task.Status = *input.Status
	}
	if input.AssigneeID != nil {
		oldVal := "nil"
		if task.AssigneeID != nil {
			oldVal = *task.AssigneeID
		}
		if err := h.recordChange(ctx, task.ID, input.UpdatedBy, "assignee_id", oldVal, *input.AssigneeID); err != nil {
			return nil, err
		}
		task.AssigneeID = input.AssigneeID
	}

	if err := h.taskCmd.Update(ctx, task); err != nil {
		return nil, fmt.Errorf("обновление задачи [task_id=%s, updated_by=%s]: %w", input.TaskID, input.UpdatedBy, err)
	}

	_ = h.cache.Delete(ctx, tasksCacheKey(task.TeamID))

	return task, nil
}

func (h *UpdateTaskHandler) checkPermission(ctx context.Context, task *domain.Task, userID string) error {
	if task.CreatedBy == userID {
		return nil
	}
	if task.AssigneeID != nil && *task.AssigneeID == userID {
		return nil
	}

	member, err := h.teamQuery.GetMember(ctx, task.TeamID, userID)
	if err != nil {
		return domain.ErrNoPermission
	}
	if member.Role == domain.RoleOwner || member.Role == domain.RoleAdmin {
		return nil
	}
	return domain.ErrNoPermission
}

func (h *UpdateTaskHandler) recordChange(ctx context.Context, taskID, changedBy string, field, oldVal, newVal string) error {
	entry := &domain.TaskHistory{
		TaskID:    taskID,
		ChangedBy: changedBy,
		FieldName: field,
		OldValue:  oldVal,
		NewValue:  newVal,
	}
	if err := h.historyCmd.CreateHistoryEntry(ctx, entry); err != nil {
		return fmt.Errorf("запись истории [task_id=%s, field=%s]: %w", taskID, field, err)
	}
	return nil
}
