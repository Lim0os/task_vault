package logging

import (
	"context"
	"log/slog"
	"task_vault/internal/domain"
	"task_vault/internal/ports"
	"time"
)

type TeamRepoLogger struct {
	cmd    ports.TeamCommandRepo
	query  ports.TeamQueryRepo
	logger *slog.Logger
}

func NewTeamRepoLogger(cmd ports.TeamCommandRepo, query ports.TeamQueryRepo, logger *slog.Logger) *TeamRepoLogger {
	return &TeamRepoLogger{cmd: cmd, query: query, logger: logger}
}

func (l *TeamRepoLogger) Create(ctx context.Context, team *domain.Team) error {
	start := time.Now()
	err := l.cmd.Create(ctx, team)
	l.log("Create", time.Since(start), err)
	return err
}

func (l *TeamRepoLogger) AddMember(ctx context.Context, member *domain.TeamMember) error {
	start := time.Now()
	err := l.cmd.AddMember(ctx, member)
	l.log("AddMember", time.Since(start), err)
	return err
}

func (l *TeamRepoLogger) GetByID(ctx context.Context, id string) (*domain.Team, error) {
	start := time.Now()
	team, err := l.query.GetByID(ctx, id)
	l.log("GetByID", time.Since(start), err)
	return team, err
}

func (l *TeamRepoLogger) ListByUser(ctx context.Context, userID string) ([]domain.Team, error) {
	start := time.Now()
	teams, err := l.query.ListByUser(ctx, userID)
	l.log("ListByUser", time.Since(start), err)
	return teams, err
}

func (l *TeamRepoLogger) GetMember(ctx context.Context, teamID, userID string) (*domain.TeamMember, error) {
	start := time.Now()
	member, err := l.query.GetMember(ctx, teamID, userID)
	l.log("GetMember", time.Since(start), err)
	return member, err
}

func (l *TeamRepoLogger) log(method string, duration time.Duration, err error) {
	if err != nil {
		l.logger.Error("TeamRepo", "method", method, "duration", duration.String(), "error", err)
		return
	}
	l.logger.Debug("TeamRepo", "method", method, "duration", duration.String())
}
