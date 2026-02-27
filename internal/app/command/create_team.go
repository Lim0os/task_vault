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
	teamCmd    ports.TeamCommandRepo
	transactor ports.Transactor
}

func NewCreateTeamHandler(cmd ports.TeamCommandRepo, transactor ports.Transactor) *CreateTeamHandler {
	return &CreateTeamHandler{teamCmd: cmd, transactor: transactor}
}

func (h *CreateTeamHandler) Handle(ctx context.Context, input CreateTeamInput) (*domain.Team, error) {
	team := &domain.Team{
		Name:      input.Name,
		CreatedBy: input.CreatedBy,
	}

	err := h.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := h.teamCmd.Create(txCtx, team); err != nil {
			return fmt.Errorf("создание команды [name=%s, created_by=%s]: %w", input.Name, input.CreatedBy, err)
		}

		member := &domain.TeamMember{
			UserID: input.CreatedBy,
			TeamID: team.ID,
			Role:   domain.RoleOwner,
		}
		if err := h.teamCmd.AddMember(txCtx, member); err != nil {
			return fmt.Errorf("добавление владельца [team_id=%s]: %w", team.ID, err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return team, nil
}
