package ports

import (
	"context"
	"task_vault/internal/domain"
)

type UserCommandRepo interface {
	Create(ctx context.Context, user *domain.User) error
}

type TeamCommandRepo interface {
	Create(ctx context.Context, team *domain.Team) error
	AddMember(ctx context.Context, member *domain.TeamMember) error
}

type TaskCommandRepo interface {
	Create(ctx context.Context, task *domain.Task) error
	Update(ctx context.Context, task *domain.Task) error
}

type CommentCommandRepo interface {
	Create(ctx context.Context, comment *domain.Comment) error
}

type HistoryCommandRepo interface {
	CreateHistoryEntry(ctx context.Context, entry *domain.TaskHistory) error
}
