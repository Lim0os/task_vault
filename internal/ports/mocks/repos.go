package mocks

import (
	"context"
	"task_vault/internal/domain"
	"task_vault/internal/ports"
	"time"

	"github.com/stretchr/testify/mock"
)

type UserCommandRepo struct{ mock.Mock }

func (m *UserCommandRepo) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

type UserQueryRepo struct{ mock.Mock }

func (m *UserQueryRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *UserQueryRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

type TeamCommandRepo struct{ mock.Mock }

func (m *TeamCommandRepo) Create(ctx context.Context, team *domain.Team) error {
	args := m.Called(ctx, team)
	return args.Error(0)
}

func (m *TeamCommandRepo) AddMember(ctx context.Context, member *domain.TeamMember) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}

type TeamQueryRepo struct{ mock.Mock }

func (m *TeamQueryRepo) GetByID(ctx context.Context, id string) (*domain.Team, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Team), args.Error(1)
}

func (m *TeamQueryRepo) ListByUser(ctx context.Context, userID string) ([]domain.Team, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Team), args.Error(1)
}

func (m *TeamQueryRepo) GetMember(ctx context.Context, teamID, userID string) (*domain.TeamMember, error) {
	args := m.Called(ctx, teamID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TeamMember), args.Error(1)
}

type TaskCommandRepo struct{ mock.Mock }

func (m *TaskCommandRepo) Create(ctx context.Context, task *domain.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *TaskCommandRepo) Update(ctx context.Context, task *domain.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

type TaskQueryRepo struct{ mock.Mock }

func (m *TaskQueryRepo) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Task), args.Error(1)
}

func (m *TaskQueryRepo) List(ctx context.Context, filter ports.TaskFilter) ([]domain.Task, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]domain.Task), args.Get(1).(int64), args.Error(2)
}

func (m *TaskQueryRepo) GetHistory(ctx context.Context, taskID string) ([]domain.TaskHistory, error) {
	args := m.Called(ctx, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TaskHistory), args.Error(1)
}

type HistoryCommandRepo struct{ mock.Mock }

func (m *HistoryCommandRepo) CreateHistoryEntry(ctx context.Context, entry *domain.TaskHistory) error {
	args := m.Called(ctx, entry)
	return args.Error(0)
}

type AnalyticsQueryRepo struct{ mock.Mock }

func (m *AnalyticsQueryRepo) TeamStats(ctx context.Context) ([]ports.TeamStat, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]ports.TeamStat), args.Error(1)
}

func (m *AnalyticsQueryRepo) TopContributors(ctx context.Context, teamID string) ([]ports.UserRank, error) {
	args := m.Called(ctx, teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]ports.UserRank), args.Error(1)
}

func (m *AnalyticsQueryRepo) OrphanedAssignees(ctx context.Context) ([]domain.Task, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Task), args.Error(1)
}

type Transactor struct{}

func (m *Transactor) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type Cache struct{ mock.Mock }

func (m *Cache) Get(ctx context.Context, key string, dest any) error {
	args := m.Called(ctx, key, dest)
	return args.Error(0)
}

func (m *Cache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *Cache) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *Cache) DeleteByPrefix(ctx context.Context, prefix string) error {
	args := m.Called(ctx, prefix)
	return args.Error(0)
}
