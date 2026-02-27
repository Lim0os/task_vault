package command

import (
	"context"
	"fmt"
	"task_vault/internal/domain"
	"task_vault/internal/ports"
)

type CreateTeamInput struct {
	Name      string
	CreatedBy string
}

type CreateTeamHandler struct {
	teamCmd ports.TeamCommandRepo
}

func NewCreateTeamHandler(cmd ports.TeamCommandRepo) *CreateTeamHandler {
	return &CreateTeamHandler{teamCmd: cmd}
}

func (h *CreateTeamHandler) Handle(ctx context.Context, input CreateTeamInput) (*domain.Team, error) {
	team := &domain.Team{
		Name:      input.Name,
		CreatedBy: input.CreatedBy,
	}

	if err := h.teamCmd.Create(ctx, team); err != nil {
		return nil, fmt.Errorf("создание команды [name=%s, created_by=%s]: %w", input.Name, input.CreatedBy, err)
	}

	member := &domain.TeamMember{
		UserID: input.CreatedBy,
		TeamID: team.ID,
		Role:   domain.RoleOwner,
	}
	if err := h.teamCmd.AddMember(ctx, member); err != nil {
		return nil, fmt.Errorf("создание команды, добавление владельца [team_id=%s]: %w", team.ID, err)
	}

	return team, nil
}
