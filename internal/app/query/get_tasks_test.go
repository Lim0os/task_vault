package query

import (
	"context"
	"errors"
	"task_vault/internal/domain"
	"task_vault/internal/ports"
	"task_vault/internal/ports/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetTasks_FromDB(t *testing.T) {
	taskQuery := new(mocks.TaskQueryRepo)
	cache := new(mocks.Cache)

	teamID := "team-uuid-1"
	filter := ports.TaskFilter{TeamID: &teamID, Limit: 20, Offset: 0}

	cache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("cache miss"))
	taskQuery.On("List", mock.Anything, filter).
		Return([]domain.Task{{ID: "task-uuid-1", Title: "Task 1"}}, int64(1), nil)
	cache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	handler := NewGetTasksHandler(taskQuery, cache)
	output, err := handler.Handle(context.Background(), filter)

	assert.NoError(t, err)
	assert.Len(t, output.Tasks, 1)
	assert.Equal(t, int64(1), output.Total)
	cache.AssertCalled(t, "Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestGetTasks_FromCache(t *testing.T) {
	taskQuery := new(mocks.TaskQueryRepo)
	cache := new(mocks.Cache)

	teamID := "team-uuid-1"
	filter := ports.TaskFilter{TeamID: &teamID, Limit: 20, Offset: 0}

	cache.On("Get", mock.Anything, mock.Anything, mock.Anything).
		Return(nil).
		Run(func(args mock.Arguments) {
			dest := args.Get(2).(*GetTasksOutput)
			dest.Tasks = []domain.Task{{ID: "task-uuid-1", Title: "Cached"}}
			dest.Total = 1
		})

	handler := NewGetTasksHandler(taskQuery, cache)
	output, err := handler.Handle(context.Background(), filter)

	assert.NoError(t, err)
	assert.Equal(t, "Cached", output.Tasks[0].Title)
	taskQuery.AssertNotCalled(t, "List")
}

func TestGetTasks_NoCacheWithFilters(t *testing.T) {
	taskQuery := new(mocks.TaskQueryRepo)
	cache := new(mocks.Cache)

	teamID := "team-uuid-1"
	status := domain.StatusDone
	filter := ports.TaskFilter{TeamID: &teamID, Status: &status, Limit: 20}

	taskQuery.On("List", mock.Anything, filter).
		Return([]domain.Task{{ID: "task-uuid-1"}}, int64(1), nil)

	handler := NewGetTasksHandler(taskQuery, cache)
	output, err := handler.Handle(context.Background(), filter)

	assert.NoError(t, err)
	assert.Len(t, output.Tasks, 1)
	cache.AssertNotCalled(t, "Get")
	cache.AssertNotCalled(t, "Set")
}
