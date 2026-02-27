package command

import (
	"context"
	"database/sql"
	"errors"
	"task_vault/internal/domain"
	"task_vault/internal/ports/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateTask_Success(t *testing.T) {
	taskCmd := new(mocks.TaskCommandRepo)
	teamQuery := new(mocks.TeamQueryRepo)
	cache := new(mocks.Cache)

	teamQuery.On("GetMember", mock.Anything, "team-uuid-1", "user-uuid-100").
		Return(&domain.TeamMember{}, nil)
	taskCmd.On("Create", mock.Anything, mock.AnythingOfType("*domain.Task")).
		Return(nil).
		Run(func(args mock.Arguments) {
			task := args.Get(1).(*domain.Task)
			task.ID = "task-uuid-42"
		})
	cache.On("Delete", mock.Anything, "tasks:team:team-uuid-1").Return(nil)

	handler := NewCreateTaskHandler(taskCmd, teamQuery, cache)
	task, err := handler.Handle(context.Background(), CreateTaskInput{
		Title:     "New Task",
		TeamID:    "team-uuid-1",
		CreatedBy: "user-uuid-100",
	})

	assert.NoError(t, err)
	assert.Equal(t, "task-uuid-42", task.ID)
	assert.Equal(t, domain.StatusTodo, task.Status)
	cache.AssertCalled(t, "Delete", mock.Anything, "tasks:team:team-uuid-1")
}

func TestCreateTask_WithAssignee(t *testing.T) {
	taskCmd := new(mocks.TaskCommandRepo)
	teamQuery := new(mocks.TeamQueryRepo)
	cache := new(mocks.Cache)

	assigneeID := "user-uuid-200"
	teamQuery.On("GetMember", mock.Anything, "team-uuid-1", "user-uuid-100").
		Return(&domain.TeamMember{}, nil)
	taskCmd.On("Create", mock.Anything, mock.AnythingOfType("*domain.Task")).
		Return(nil).
		Run(func(args mock.Arguments) {
			task := args.Get(1).(*domain.Task)
			task.ID = "task-uuid-43"
		})
	cache.On("Delete", mock.Anything, "tasks:team:team-uuid-1").Return(nil)

	handler := NewCreateTaskHandler(taskCmd, teamQuery, cache)
	task, err := handler.Handle(context.Background(), CreateTaskInput{
		Title:       "Assigned Task",
		Description: "Some description",
		AssigneeID:  &assigneeID,
		TeamID:      "team-uuid-1",
		CreatedBy:   "user-uuid-100",
	})

	assert.NoError(t, err)
	assert.Equal(t, &assigneeID, task.AssigneeID)
	assert.Equal(t, "Some description", task.Description)
}

func TestCreateTask_NotTeamMember(t *testing.T) {
	taskCmd := new(mocks.TaskCommandRepo)
	teamQuery := new(mocks.TeamQueryRepo)
	cache := new(mocks.Cache)

	teamQuery.On("GetMember", mock.Anything, "team-uuid-1", "user-uuid-100").
		Return(nil, sql.ErrNoRows)

	handler := NewCreateTaskHandler(taskCmd, teamQuery, cache)
	task, err := handler.Handle(context.Background(), CreateTaskInput{
		Title:     "New Task",
		TeamID:    "team-uuid-1",
		CreatedBy: "user-uuid-100",
	})

	assert.True(t, errors.Is(err, domain.ErrNotTeamMember))
	assert.Nil(t, task)
}

func TestCreateTask_DBError(t *testing.T) {
	taskCmd := new(mocks.TaskCommandRepo)
	teamQuery := new(mocks.TeamQueryRepo)
	cache := new(mocks.Cache)

	teamQuery.On("GetMember", mock.Anything, "team-uuid-1", "user-uuid-100").
		Return(&domain.TeamMember{}, nil)
	taskCmd.On("Create", mock.Anything, mock.AnythingOfType("*domain.Task")).
		Return(assert.AnError)

	handler := NewCreateTaskHandler(taskCmd, teamQuery, cache)
	task, err := handler.Handle(context.Background(), CreateTaskInput{
		Title:     "Task",
		TeamID:    "team-uuid-1",
		CreatedBy: "user-uuid-100",
	})

	assert.Error(t, err)
	assert.Nil(t, task)
}
