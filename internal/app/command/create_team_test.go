package command

import (
	"context"
	"task_vault/internal/domain"
	"task_vault/internal/ports/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateTeam_Success(t *testing.T) {
	teamCmd := new(mocks.TeamCommandRepo)

	teamCmd.On("Create", mock.Anything, mock.AnythingOfType("*domain.Team")).
		Return(nil).
		Run(func(args mock.Arguments) {
			team := args.Get(1).(*domain.Team)
			team.ID = "team-uuid-10"
		})
	teamCmd.On("AddMember", mock.Anything, mock.AnythingOfType("*domain.TeamMember")).
		Return(nil)

	handler := NewCreateTeamHandler(teamCmd)
	team, err := handler.Handle(context.Background(), CreateTeamInput{
		Name:      "My Team",
		CreatedBy: "user-uuid-1",
	})

	assert.NoError(t, err)
	assert.Equal(t, "team-uuid-10", team.ID)
	assert.Equal(t, "My Team", team.Name)

	teamCmd.AssertCalled(t, "AddMember", mock.Anything, mock.MatchedBy(func(m *domain.TeamMember) bool {
		return m.UserID == "user-uuid-1" && m.TeamID == "team-uuid-10" && m.Role == domain.RoleOwner
	}))
}
