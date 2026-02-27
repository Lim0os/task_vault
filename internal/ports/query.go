package ports

import (
	"context"
	"task_vault/internal/domain"
)

type UserQueryRepo interface {
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
}

type TeamQueryRepo interface {
	GetByID(ctx context.Context, id string) (*domain.Team, error)
	ListByUser(ctx context.Context, userID string) ([]domain.Team, error)
	GetMember(ctx context.Context, teamID, userID string) (*domain.TeamMember, error)
}

type TaskQueryRepo interface {
	GetByID(ctx context.Context, id string) (*domain.Task, error)
	List(ctx context.Context, filter TaskFilter) ([]domain.Task, int64, error)
	GetHistory(ctx context.Context, taskID string) ([]domain.TaskHistory, error)
}

type AnalyticsQueryRepo interface {
	TeamStats(ctx context.Context) ([]TeamStat, error)
	TopContributors(ctx context.Context, teamID string) ([]UserRank, error)
	OrphanedAssignees(ctx context.Context) ([]domain.Task, error)
}

type TaskFilter struct {
	TeamID     *string
	Status     *domain.Status
	AssigneeID *string
	Limit      int
	Offset     int
}

type TeamStat struct {
	TeamID       string
	TeamName     string
	MembersCount int
	DoneLastWeek int
}

type UserRank struct {
	UserID       string
	UserName     string
	TeamID       string
	TasksCreated int
	Rank         int
}
