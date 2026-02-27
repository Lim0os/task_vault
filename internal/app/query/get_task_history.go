package query

import (
	"context"
	"fmt"
	"task_vault/internal/domain"
	"task_vault/internal/ports"
)

type GetTaskHistoryHandler struct {
	taskQuery ports.TaskQueryRepo
}

func NewGetTaskHistoryHandler(taskQuery ports.TaskQueryRepo) *GetTaskHistoryHandler {
	return &GetTaskHistoryHandler{taskQuery: taskQuery}
}

func (h *GetTaskHistoryHandler) Handle(ctx context.Context, taskID string) ([]domain.TaskHistory, error) {
	history, err := h.taskQuery.GetHistory(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("получение истории задачи [task_id=%s]: %w", taskID, err)
	}
	return history, nil
}
