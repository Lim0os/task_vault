package command

import (
	"context"
	"errors"
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
		if errors.Is(err, domain.ErrNotTeamMember) {
			return nil, domain.ErrNotTeamMember
		}
		return nil, fmt.Errorf("проверка членства [team_id=%s, user_id=%s]: %w", input.TeamID, input.CreatedBy, err)
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

	_ = h.cache.DeleteByPrefix(ctx, tasksCacheKey(input.TeamID))

	return task, nil
}
