package mysql

import (
	"context"
	"task_vault/internal/domain"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestUser(t *testing.T, name string) *domain.User {
	t.Helper()
	user := &domain.User{Email: name + "@test.com", PasswordHash: "h", Name: name}
	require.NoError(t, NewUserRepo(testDB).Create(context.Background(), user))
	return user
}

func TestTeamRepo_CreateAndGetByID(t *testing.T) {
	cleanTables(t)
	ctx := context.Background()
	user := createTestUser(t, "owner")
	repo := NewTeamRepo(testDB)

	team := &domain.Team{Name: "Alpha", CreatedBy: user.ID}
	require.NoError(t, repo.Create(ctx, team))
	assert.NotEmpty(t, team.ID)

	found, err := repo.GetByID(ctx, team.ID)
	require.NoError(t, err)
	assert.Equal(t, "Alpha", found.Name)
}

func TestTeamRepo_AddMemberAndGetMember(t *testing.T) {
	cleanTables(t)
	ctx := context.Background()
	user := createTestUser(t, "member")
	repo := NewTeamRepo(testDB)

	team := &domain.Team{Name: "Beta", CreatedBy: user.ID}
	require.NoError(t, repo.Create(ctx, team))

	member := &domain.TeamMember{UserID: user.ID, TeamID: team.ID, Role: domain.RoleOwner}
	require.NoError(t, repo.AddMember(ctx, member))
	assert.NotEmpty(t, member.ID)

	found, err := repo.GetMember(ctx, team.ID, user.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.RoleOwner, found.Role)
}

func TestTeamRepo_ListByUser(t *testing.T) {
	cleanTables(t)
	ctx := context.Background()
	user := createTestUser(t, "multi")
	repo := NewTeamRepo(testDB)

	for _, name := range []string{"Team1", "Team2"} {
		team := &domain.Team{Name: name, CreatedBy: user.ID}
		require.NoError(t, repo.Create(ctx, team))
		require.NoError(t, repo.AddMember(ctx, &domain.TeamMember{
			UserID: user.ID, TeamID: team.ID, Role: domain.RoleOwner,
		}))
	}

	teams, err := repo.ListByUser(ctx, user.ID)
	require.NoError(t, err)
	assert.Len(t, teams, 2)
}
