package query

import (
	"context"
	"task_vault/internal/domain"
	"task_vault/internal/ports"
	"task_vault/internal/ports/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTeamStats_Success(t *testing.T) {
	analytics := new(mocks.AnalyticsQueryRepo)

	analytics.On("TeamStats", mock.Anything).
		Return([]ports.TeamStat{
			{TeamID: "team-uuid-1", TeamName: "Alpha", MembersCount: 5, DoneLastWeek: 12},
		}, nil)

	handler := NewTeamAnalyticsHandler(analytics)
	stats, err := handler.TeamStats(context.Background())

	assert.NoError(t, err)
	assert.Len(t, stats, 1)
	assert.Equal(t, 12, stats[0].DoneLastWeek)
}

func TestTopContributors_Success(t *testing.T) {
	analytics := new(mocks.AnalyticsQueryRepo)

	analytics.On("TopContributors", mock.Anything, "team-uuid-1").
		Return([]ports.UserRank{
			{UserID: "user-uuid-10", UserName: "Alice", TasksCreated: 15, Rank: 1},
			{UserID: "user-uuid-20", UserName: "Bob", TasksCreated: 10, Rank: 2},
		}, nil)

	handler := NewTeamAnalyticsHandler(analytics)
	ranks, err := handler.TopContributors(context.Background(), "team-uuid-1")

	assert.NoError(t, err)
	assert.Len(t, ranks, 2)
	assert.Equal(t, 1, ranks[0].Rank)
}

func TestOrphanedAssignees_Success(t *testing.T) {
	analytics := new(mocks.AnalyticsQueryRepo)

	analytics.On("OrphanedAssignees", mock.Anything).
		Return([]domain.Task{{ID: "task-uuid-5", Title: "Orphan"}}, nil)

	handler := NewTeamAnalyticsHandler(analytics)
	tasks, err := handler.OrphanedAssignees(context.Background())

	assert.NoError(t, err)
	assert.Len(t, tasks, 1)
	assert.Equal(t, "Orphan", tasks[0].Title)
}
