package query

import (
	"context"
	"task_vault/internal/domain"
	"task_vault/internal/ports/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetTeams_Success(t *testing.T) {
	teamQuery := new(mocks.TeamQueryRepo)

	teamQuery.On("ListByUser", mock.Anything, "user-uuid-1").
		Return([]domain.Team{
			{ID: "team-uuid-1", Name: "Team A"},
			{ID: "team-uuid-2", Name: "Team B"},
		}, nil)

	handler := NewGetTeamsHandler(teamQuery)
	teams, err := handler.Handle(context.Background(), "user-uuid-1")

	assert.NoError(t, err)
	assert.Len(t, teams, 2)
}

func TestGetTeams_Empty(t *testing.T) {
	teamQuery := new(mocks.TeamQueryRepo)

	teamQuery.On("ListByUser", mock.Anything, "user-uuid-1").
		Return([]domain.Team{}, nil)

	handler := NewGetTeamsHandler(teamQuery)
	teams, err := handler.Handle(context.Background(), "user-uuid-1")

	assert.NoError(t, err)
	assert.Empty(t, teams)
}
