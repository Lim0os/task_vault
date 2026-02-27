package command

import (
	"context"
	"fmt"
	"task_vault/internal/domain"
	"task_vault/internal/ports"
)

type CreateTaskInput struct {
	Title       string
	Description string
	AssigneeID  *string
	TeamID      string
	CreatedBy   string
}

type CreateTaskHandler struct {
	taskCmd   ports.TaskCommandRepo
	teamQuery ports.TeamQueryRepo
	cache     ports.Cache
}

func NewCreateTaskHandler(
	taskCmd ports.TaskCommandRepo,
	teamQuery ports.TeamQueryRepo,
	cache ports.Cache,
) *CreateTaskHandler {
	return &CreateTaskHandler{
		taskCmd:   taskCmd,
		teamQuery: teamQuery,
		cache:     cache,
	}
}

func (h *CreateTaskHandler) Handle(ctx context.Context, input CreateTaskInput) (*domain.Task, error) {
	_, err := h.teamQuery.GetMember(ctx, input.TeamID, input.CreatedBy)
	if err != nil {
		return nil, domain.ErrNotTeamMember
	}

	task := &domain.Task{
		Title:       input.Title,
		Description: input.Description,
		Status:      domain.StatusTodo,
		AssigneeID:  input.AssigneeID,
		TeamID:      input.TeamID,
		CreatedBy:   input.CreatedBy,
	}

	if err := h.taskCmd.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("создание задачи [team_id=%s, created_by=%s]: %w", input.TeamID, input.CreatedBy, err)
	}

	_ = h.cache.Delete(ctx, tasksCacheKey(input.TeamID))

	return task, nil
}
