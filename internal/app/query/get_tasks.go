package query

import (
	"context"
	"fmt"
	"task_vault/internal/domain"
	"task_vault/internal/ports"
	"time"
)

const tasksCacheTTL = 5 * time.Minute

type GetTasksOutput struct {
	Tasks []domain.Task
	Total int64
}

type GetTasksHandler struct {
	taskQuery ports.TaskQueryRepo
	cache     ports.Cache
}

func NewGetTasksHandler(taskQuery ports.TaskQueryRepo, cache ports.Cache) *GetTasksHandler {
	return &GetTasksHandler{taskQuery: taskQuery, cache: cache}
}

func (h *GetTasksHandler) Handle(ctx context.Context, filter ports.TaskFilter) (*GetTasksOutput, error) {
	// Кешируем только запросы по team_id без дополнительных фильтров
	cacheKey := ""
	canCache := filter.TeamID != nil && filter.Status == nil && filter.AssigneeID == nil
	if canCache {
		cacheKey = fmt.Sprintf("tasks:team:%s:offset:%d:limit:%d", *filter.TeamID, filter.Offset, filter.Limit)
		var cached GetTasksOutput
		if err := h.cache.Get(ctx, cacheKey, &cached); err == nil {
			return &cached, nil
		}
	}

	tasks, total, err := h.taskQuery.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("получение списка задач: %w", err)
	}

	output := &GetTasksOutput{Tasks: tasks, Total: total}

	if canCache {
		_ = h.cache.Set(ctx, cacheKey, output, tasksCacheTTL)
	}

	return output, nil
}
