package query

import (
	"context"
	"fmt"
	"task_vault/internal/domain"
	"task_vault/internal/ports"
)

type TeamAnalyticsHandler struct {
	analyticsQuery ports.AnalyticsQueryRepo
}

func NewTeamAnalyticsHandler(analyticsQuery ports.AnalyticsQueryRepo) *TeamAnalyticsHandler {
	return &TeamAnalyticsHandler{analyticsQuery: analyticsQuery}
}

func (h *TeamAnalyticsHandler) TeamStats(ctx context.Context) ([]ports.TeamStat, error) {
	stats, err := h.analyticsQuery.TeamStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("получение статистики команд: %w", err)
	}
	return stats, nil
}

func (h *TeamAnalyticsHandler) TopContributors(ctx context.Context, teamID string) ([]ports.UserRank, error) {
	ranks, err := h.analyticsQuery.TopContributors(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("получение топ-контрибьюторов [team_id=%s]: %w", teamID, err)
	}
	return ranks, nil
}

func (h *TeamAnalyticsHandler) OrphanedAssignees(ctx context.Context) ([]domain.Task, error) {
	tasks, err := h.analyticsQuery.OrphanedAssignees(ctx)
	if err != nil {
		return nil, fmt.Errorf("получение задач с orphaned assignees: %w", err)
	}
	return tasks, nil
}
