package query

import (
	"context"
	"fmt"
	"task_vault/internal/domain"
	"task_vault/internal/ports"
)

type GetTeamsHandler struct {
	teamQuery ports.TeamQueryRepo
}

func NewGetTeamsHandler(teamQuery ports.TeamQueryRepo) *GetTeamsHandler {
	return &GetTeamsHandler{teamQuery: teamQuery}
}

func (h *GetTeamsHandler) Handle(ctx context.Context, userID string) ([]domain.Team, error) {
	teams, err := h.teamQuery.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("получение команд пользователя [user_id=%s]: %w", userID, err)
	}
	return teams, nil
}
