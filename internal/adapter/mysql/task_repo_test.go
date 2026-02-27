package mysql

import (
	"context"
	"task_vault/internal/domain"
	"task_vault/internal/ports"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTaskTest(t *testing.T) (*domain.User, *domain.Team) {
	t.Helper()
	cleanTables(t)
	ctx := context.Background()

	user := createTestUser(t, "taskuser")
	teamRepo := NewTeamRepo(testDB)
	team := &domain.Team{Name: "TaskTeam", CreatedBy: user.ID}
	require.NoError(t, teamRepo.Create(ctx, team))
	require.NoError(t, teamRepo.AddMember(ctx, &domain.TeamMember{
		UserID: user.ID, TeamID: team.ID, Role: domain.RoleOwner,
	}))
	return user, team
}

func TestTaskRepo_CreateAndGetByID(t *testing.T) {
	user, team := setupTaskTest(t)
	repo := NewTaskRepo(testDB)
	ctx := context.Background()

	task := &domain.Task{
		Title:     "Test Task",
		Status:    domain.StatusTodo,
		TeamID:    team.ID,
		CreatedBy: user.ID,
	}
	require.NoError(t, repo.Create(ctx, task))
	assert.NotEmpty(t, task.ID)

	found, err := repo.GetByID(ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, "Test Task", found.Title)
	assert.Equal(t, domain.StatusTodo, found.Status)
}

func TestTaskRepo_Update(t *testing.T) {
	user, team := setupTaskTest(t)
	repo := NewTaskRepo(testDB)
	ctx := context.Background()

	task := &domain.Task{Title: "Before", Status: domain.StatusTodo, TeamID: team.ID, CreatedBy: user.ID}
	require.NoError(t, repo.Create(ctx, task))

	task.Title = "After"
	task.Status = domain.StatusDone
	require.NoError(t, repo.Update(ctx, task))

	found, err := repo.GetByID(ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, "After", found.Title)
	assert.Equal(t, domain.StatusDone, found.Status)
}

func TestTaskRepo_ListWithFilter(t *testing.T) {
	user, team := setupTaskTest(t)
	repo := NewTaskRepo(testDB)
	ctx := context.Background()

	for _, title := range []string{"A", "B", "C"} {
		require.NoError(t, repo.Create(ctx, &domain.Task{
			Title: title, Status: domain.StatusTodo, TeamID: team.ID, CreatedBy: user.ID,
		}))
	}

	teamID := team.ID
	tasks, total, err := repo.List(ctx, ports.TaskFilter{TeamID: &teamID, Limit: 2, Offset: 0})
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, tasks, 2)
}

func TestTaskRepo_History(t *testing.T) {
	user, team := setupTaskTest(t)
	repo := NewTaskRepo(testDB)
	ctx := context.Background()

	task := &domain.Task{Title: "Hist", Status: domain.StatusTodo, TeamID: team.ID, CreatedBy: user.ID}
	require.NoError(t, repo.Create(ctx, task))

	require.NoError(t, repo.CreateHistoryEntry(ctx, &domain.TaskHistory{
		TaskID: task.ID, ChangedBy: user.ID, FieldName: "status", OldValue: "todo", NewValue: "done",
	}))

	history, err := repo.GetHistory(ctx, task.ID)
	require.NoError(t, err)
	assert.Len(t, history, 1)
	assert.Equal(t, "status", history[0].FieldName)
}

func TestTaskRepo_TeamStats(t *testing.T) {
	user, team := setupTaskTest(t)
	repo := NewTaskRepo(testDB)
	ctx := context.Background()

	require.NoError(t, repo.Create(ctx, &domain.Task{
		Title: "Done", Status: domain.StatusDone, TeamID: team.ID, CreatedBy: user.ID,
	}))

	stats, err := repo.TeamStats(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, stats)

	var found bool
	for _, s := range stats {
		if s.TeamID == team.ID {
			found = true
			assert.Equal(t, 1, s.MembersCount)
		}
	}
	assert.True(t, found)
}

func TestTaskRepo_OrphanedAssignees(t *testing.T) {
	user, team := setupTaskTest(t)
	repo := NewTaskRepo(testDB)
	ctx := context.Background()

	outsider := createTestUser(t, "outsider")

	outsiderID := outsider.ID
	require.NoError(t, repo.Create(ctx, &domain.Task{
		Title: "Orphan", Status: domain.StatusTodo, TeamID: team.ID,
		CreatedBy: user.ID, AssigneeID: &outsiderID,
	}))

	orphans, err := repo.OrphanedAssignees(ctx)
	require.NoError(t, err)
	assert.Len(t, orphans, 1)
	assert.Equal(t, "Orphan", orphans[0].Title)
}
