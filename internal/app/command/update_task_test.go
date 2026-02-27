package command

import (
	"context"
	"errors"
	"task_vault/internal/domain"
	"task_vault/internal/ports/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateTask_Success(t *testing.T) {
	taskCmd := new(mocks.TaskCommandRepo)
	taskQuery := new(mocks.TaskQueryRepo)
	teamQuery := new(mocks.TeamQueryRepo)
	historyCmd := new(mocks.HistoryCommandRepo)
	cache := new(mocks.Cache)

	existingTask := &domain.Task{
		ID:        "task-uuid-1",
		Title:     "Old Title",
		Status:    domain.StatusTodo,
		TeamID:    "team-uuid-10",
		CreatedBy: "user-uuid-100",
	}

	taskQuery.On("GetByID", mock.Anything, "task-uuid-1").Return(existingTask, nil)
	taskCmd.On("Update", mock.Anything, mock.AnythingOfType("*domain.Task")).Return(nil)
	historyCmd.On("CreateHistoryEntry", mock.Anything, mock.AnythingOfType("*domain.TaskHistory")).Return(nil)
	cache.On("DeleteByPrefix", mock.Anything, "tasks:team:team-uuid-10").Return(nil)

	tx := new(mocks.Transactor)
	handler := NewUpdateTaskHandler(taskCmd, taskQuery, teamQuery, historyCmd, cache, tx)

	newTitle := "New Title"
	newStatus := domain.StatusDone
	task, err := handler.Handle(context.Background(), UpdateTaskInput{
		TaskID:    "task-uuid-1",
		Title:     &newTitle,
		Status:    &newStatus,
		UpdatedBy: "user-uuid-100",
	})

	assert.NoError(t, err)
	assert.Equal(t, "New Title", task.Title)
	assert.Equal(t, domain.StatusDone, task.Status)
	historyCmd.AssertNumberOfCalls(t, "CreateHistoryEntry", 2)
}

func TestUpdateTask_ByAssignee(t *testing.T) {
	taskCmd := new(mocks.TaskCommandRepo)
	taskQuery := new(mocks.TaskQueryRepo)
	teamQuery := new(mocks.TeamQueryRepo)
	historyCmd := new(mocks.HistoryCommandRepo)
	cache := new(mocks.Cache)

	assigneeID := "user-uuid-200"
	existingTask := &domain.Task{
		ID:         "task-uuid-1",
		Title:      "Task",
		Status:     domain.StatusTodo,
		AssigneeID: &assigneeID,
		TeamID:     "team-uuid-10",
		CreatedBy:  "user-uuid-100",
	}

	taskQuery.On("GetByID", mock.Anything, "task-uuid-1").Return(existingTask, nil)
	taskCmd.On("Update", mock.Anything, mock.AnythingOfType("*domain.Task")).Return(nil)
	historyCmd.On("CreateHistoryEntry", mock.Anything, mock.AnythingOfType("*domain.TaskHistory")).Return(nil)
	cache.On("DeleteByPrefix", mock.Anything, "tasks:team:team-uuid-10").Return(nil)

	tx := new(mocks.Transactor)
	handler := NewUpdateTaskHandler(taskCmd, taskQuery, teamQuery, historyCmd, cache, tx)

	newStatus := domain.StatusInProgress
	task, err := handler.Handle(context.Background(), UpdateTaskInput{
		TaskID:    "task-uuid-1",
		Status:    &newStatus,
		UpdatedBy: "user-uuid-200",
	})

	assert.NoError(t, err)
	assert.Equal(t, domain.StatusInProgress, task.Status)
}

func TestUpdateTask_ByAdmin(t *testing.T) {
	taskCmd := new(mocks.TaskCommandRepo)
	taskQuery := new(mocks.TaskQueryRepo)
	teamQuery := new(mocks.TeamQueryRepo)
	historyCmd := new(mocks.HistoryCommandRepo)
	cache := new(mocks.Cache)

	existingTask := &domain.Task{
		ID:        "task-uuid-1",
		Title:     "Task",
		TeamID:    "team-uuid-10",
		CreatedBy: "user-uuid-100",
	}

	taskQuery.On("GetByID", mock.Anything, "task-uuid-1").Return(existingTask, nil)
	teamQuery.On("GetMember", mock.Anything, "team-uuid-10", "user-uuid-300").
		Return(&domain.TeamMember{Role: domain.RoleAdmin}, nil)
	taskCmd.On("Update", mock.Anything, mock.AnythingOfType("*domain.Task")).Return(nil)
	historyCmd.On("CreateHistoryEntry", mock.Anything, mock.AnythingOfType("*domain.TaskHistory")).Return(nil)
	cache.On("DeleteByPrefix", mock.Anything, "tasks:team:team-uuid-10").Return(nil)

	tx := new(mocks.Transactor)
	handler := NewUpdateTaskHandler(taskCmd, taskQuery, teamQuery, historyCmd, cache, tx)

	newDesc := "Updated by admin"
	task, err := handler.Handle(context.Background(), UpdateTaskInput{
		TaskID:      "task-uuid-1",
		Description: &newDesc,
		UpdatedBy:   "user-uuid-300",
	})

	assert.NoError(t, err)
	assert.Equal(t, "Updated by admin", task.Description)
}

func TestUpdateTask_ChangeAssignee(t *testing.T) {
	taskCmd := new(mocks.TaskCommandRepo)
	taskQuery := new(mocks.TaskQueryRepo)
	teamQuery := new(mocks.TeamQueryRepo)
	historyCmd := new(mocks.HistoryCommandRepo)
	cache := new(mocks.Cache)

	oldAssignee := "user-uuid-50"
	existingTask := &domain.Task{
		ID:         "task-uuid-1",
		Title:      "Task",
		AssigneeID: &oldAssignee,
		TeamID:     "team-uuid-10",
		CreatedBy:  "user-uuid-100",
	}

	taskQuery.On("GetByID", mock.Anything, "task-uuid-1").Return(existingTask, nil)
	taskCmd.On("Update", mock.Anything, mock.AnythingOfType("*domain.Task")).Return(nil)
	historyCmd.On("CreateHistoryEntry", mock.Anything, mock.MatchedBy(func(h *domain.TaskHistory) bool {
		return h.FieldName == "assignee_id" && h.OldValue == oldAssignee
	})).Return(nil)
	cache.On("DeleteByPrefix", mock.Anything, "tasks:team:team-uuid-10").Return(nil)

	tx := new(mocks.Transactor)
	handler := NewUpdateTaskHandler(taskCmd, taskQuery, teamQuery, historyCmd, cache, tx)

	newAssignee := "user-uuid-60"
	task, err := handler.Handle(context.Background(), UpdateTaskInput{
		TaskID:     "task-uuid-1",
		AssigneeID: &newAssignee,
		UpdatedBy:  "user-uuid-100",
	})

	assert.NoError(t, err)
	assert.Equal(t, &newAssignee, task.AssigneeID)
}

func TestUpdateTask_NotFound(t *testing.T) {
	taskCmd := new(mocks.TaskCommandRepo)
	taskQuery := new(mocks.TaskQueryRepo)
	teamQuery := new(mocks.TeamQueryRepo)
	historyCmd := new(mocks.HistoryCommandRepo)
	cache := new(mocks.Cache)

	taskQuery.On("GetByID", mock.Anything, "task-uuid-999").Return(nil, domain.ErrTaskNotFound)

	tx := new(mocks.Transactor)
	handler := NewUpdateTaskHandler(taskCmd, taskQuery, teamQuery, historyCmd, cache, tx)
	task, err := handler.Handle(context.Background(), UpdateTaskInput{
		TaskID:    "task-uuid-999",
		UpdatedBy: "user-uuid-100",
	})

	assert.True(t, errors.Is(err, domain.ErrTaskNotFound))
	assert.Nil(t, task)
}

func TestUpdateTask_NoPermission(t *testing.T) {
	taskCmd := new(mocks.TaskCommandRepo)
	taskQuery := new(mocks.TaskQueryRepo)
	teamQuery := new(mocks.TeamQueryRepo)
	historyCmd := new(mocks.HistoryCommandRepo)
	cache := new(mocks.Cache)

	existingTask := &domain.Task{
		ID:        "task-uuid-1",
		TeamID:    "team-uuid-10",
		CreatedBy: "user-uuid-100",
	}

	taskQuery.On("GetByID", mock.Anything, "task-uuid-1").Return(existingTask, nil)
	teamQuery.On("GetMember", mock.Anything, "team-uuid-10", "user-uuid-999").
		Return(&domain.TeamMember{Role: domain.RoleMember}, nil)

	tx := new(mocks.Transactor)
	handler := NewUpdateTaskHandler(taskCmd, taskQuery, teamQuery, historyCmd, cache, tx)

	newTitle := "Hack"
	task, err := handler.Handle(context.Background(), UpdateTaskInput{
		TaskID:    "task-uuid-1",
		Title:     &newTitle,
		UpdatedBy: "user-uuid-999",
	})

	assert.True(t, errors.Is(err, domain.ErrNoPermission))
	assert.Nil(t, task)
}
