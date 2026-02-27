package logging

import (
	"context"
	"log/slog"
	"task_vault/internal/domain"
	"task_vault/internal/ports"
	"time"
)

type UserRepoLogger struct {
	next   ports.UserCommandRepo
	query  ports.UserQueryRepo
	logger *slog.Logger
}

func NewUserRepoLogger(cmd ports.UserCommandRepo, query ports.UserQueryRepo, logger *slog.Logger) *UserRepoLogger {
	return &UserRepoLogger{next: cmd, query: query, logger: logger}
}

func (l *UserRepoLogger) Create(ctx context.Context, user *domain.User) error {
	start := time.Now()
	err := l.next.Create(ctx, user)
	l.log("Create", time.Since(start), err)
	return err
}

func (l *UserRepoLogger) GetByID(ctx context.Context, id string) (*domain.User, error) {
	start := time.Now()
	user, err := l.query.GetByID(ctx, id)
	l.log("GetByID", time.Since(start), err)
	return user, err
}

func (l *UserRepoLogger) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	start := time.Now()
	user, err := l.query.GetByEmail(ctx, email)
	l.log("GetByEmail", time.Since(start), err)
	return user, err
}

func (l *UserRepoLogger) log(method string, duration time.Duration, err error) {
	if err != nil {
		l.logger.Error("UserRepo", "method", method, "duration", duration.String(), "error", err)
		return
	}
	l.logger.Debug("UserRepo", "method", method, "duration", duration.String())
}
