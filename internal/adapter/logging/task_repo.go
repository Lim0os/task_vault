package logging

import (
	"context"
	"log/slog"
	"task_vault/internal/domain"
	"task_vault/internal/ports"
	"time"
)

type TaskRepoLogger struct {
	cmd       ports.TaskCommandRepo
	query     ports.TaskQueryRepo
	history   ports.HistoryCommandRepo
	analytics ports.AnalyticsQueryRepo
	logger    *slog.Logger
}

func NewTaskRepoLogger(
	cmd ports.TaskCommandRepo,
	query ports.TaskQueryRepo,
	history ports.HistoryCommandRepo,
	analytics ports.AnalyticsQueryRepo,
	logger *slog.Logger,
) *TaskRepoLogger {
	return &TaskRepoLogger{
		cmd:       cmd,
		query:     query,
		history:   history,
		analytics: analytics,
		logger:    logger,
	}
}

// --- TaskCommandRepo ---

func (l *TaskRepoLogger) Create(ctx context.Context, task *domain.Task) error {
	start := time.Now()
	err := l.cmd.Create(ctx, task)
	l.log("Create", time.Since(start), err)
	return err
}

func (l *TaskRepoLogger) Update(ctx context.Context, task *domain.Task) error {
	start := time.Now()
	err := l.cmd.Update(ctx, task)
	l.log("Update", time.Since(start), err)
	return err
}

// --- TaskQueryRepo ---

func (l *TaskRepoLogger) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	start := time.Now()
	task, err := l.query.GetByID(ctx, id)
	l.log("GetByID", time.Since(start), err)
	return task, err
}

func (l *TaskRepoLogger) List(ctx context.Context, filter ports.TaskFilter) ([]domain.Task, int64, error) {
	start := time.Now()
	tasks, total, err := l.query.List(ctx, filter)
	l.log("List", time.Since(start), err)
	return tasks, total, err
}

func (l *TaskRepoLogger) GetHistory(ctx context.Context, taskID string) ([]domain.TaskHistory, error) {
	start := time.Now()
	history, err := l.query.GetHistory(ctx, taskID)
	l.log("GetHistory", time.Since(start), err)
	return history, err
}

// --- HistoryCommandRepo ---

func (l *TaskRepoLogger) CreateHistoryEntry(ctx context.Context, entry *domain.TaskHistory) error {
	start := time.Now()
	err := l.history.CreateHistoryEntry(ctx, entry)
	l.log("CreateHistoryEntry", time.Since(start), err)
	return err
}

// --- AnalyticsQueryRepo ---

func (l *TaskRepoLogger) TeamStats(ctx context.Context) ([]ports.TeamStat, error) {
	start := time.Now()
	stats, err := l.analytics.TeamStats(ctx)
	l.log("TeamStats", time.Since(start), err)
	return stats, err
}

func (l *TaskRepoLogger) TopContributors(ctx context.Context, teamID string) ([]ports.UserRank, error) {
	start := time.Now()
	ranks, err := l.analytics.TopContributors(ctx, teamID)
	l.log("TopContributors", time.Since(start), err)
	return ranks, err
}

func (l *TaskRepoLogger) OrphanedAssignees(ctx context.Context) ([]domain.Task, error) {
	start := time.Now()
	tasks, err := l.analytics.OrphanedAssignees(ctx)
	l.log("OrphanedAssignees", time.Since(start), err)
	return tasks, err
}

func (l *TaskRepoLogger) log(method string, duration time.Duration, err error) {
	if err != nil {
		l.logger.Error("TaskRepo", "method", method, "duration", duration.String(), "error", err)
		return
	}
	l.logger.Debug("TaskRepo", "method", method, "duration", duration.String())
}
