package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"task_vault/internal/domain"
	"task_vault/internal/ports"

	"github.com/google/uuid"
)

type TaskRepo struct {
	db *sql.DB
}

func NewTaskRepo(db *sql.DB) *TaskRepo {
	return &TaskRepo{db: db}
}

func (r *TaskRepo) conn(ctx context.Context) DBTX {
	return ConnFromContext(ctx, r.db)
}

func (r *TaskRepo) Create(ctx context.Context, task *domain.Task) error {
	task.ID = uuid.New().String()
	_, err := r.conn(ctx).ExecContext(ctx,
		`INSERT INTO tasks (id, title, description, status, assignee_id, team_id, created_by)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		task.ID, task.Title, task.Description, task.Status, task.AssigneeID, task.TeamID, task.CreatedBy,
	)
	return err
}

func (r *TaskRepo) Update(ctx context.Context, task *domain.Task) error {
	_, err := r.conn(ctx).ExecContext(ctx,
		`UPDATE tasks SET title = ?, description = ?, status = ?, assignee_id = ?
		 WHERE id = ?`,
		task.Title, task.Description, task.Status, task.AssigneeID, task.ID,
	)
	return err
}

func (r *TaskRepo) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	t := &domain.Task{}
	err := r.conn(ctx).QueryRowContext(ctx,
		`SELECT id, title, description, status, assignee_id, team_id, created_by, created_at, updated_at
		 FROM tasks WHERE id = ?`, id,
	).Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.AssigneeID,
		&t.TeamID, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrTaskNotFound
		}
		return nil, fmt.Errorf("получение задачи [id=%s]: %w", id, err)
	}
	return t, nil
}

func (r *TaskRepo) List(ctx context.Context, filter ports.TaskFilter) ([]domain.Task, int64, error) {
	var conditions []string
	var args []any

	if filter.TeamID != nil {
		conditions = append(conditions, "team_id = ?")
		args = append(args, *filter.TeamID)
	}
	if filter.Status != nil {
		conditions = append(conditions, "status = ?")
		args = append(args, *filter.Status)
	}
	if filter.AssigneeID != nil {
		conditions = append(conditions, "assignee_id = ?")
		args = append(args, *filter.AssigneeID)
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tasks %s", where)
	if err := r.conn(ctx).QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}

	query := fmt.Sprintf(
		`SELECT id, title, description, status, assignee_id, team_id, created_by, created_at, updated_at
		 FROM tasks %s ORDER BY created_at DESC LIMIT ? OFFSET ?`, where,
	)
	args = append(args, limit, filter.Offset)

	rows, err := r.conn(ctx).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tasks []domain.Task
	for rows.Next() {
		var t domain.Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.AssigneeID,
			&t.TeamID, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, t)
	}
	return tasks, total, rows.Err()
}

func (r *TaskRepo) GetHistory(ctx context.Context, taskID string) ([]domain.TaskHistory, error) {
	rows, err := r.conn(ctx).QueryContext(ctx,
		`SELECT id, task_id, changed_by, field_name, old_value, new_value, changed_at
		 FROM task_history WHERE task_id = ? ORDER BY changed_at DESC`, taskID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []domain.TaskHistory
	for rows.Next() {
		var h domain.TaskHistory
		if err := rows.Scan(&h.ID, &h.TaskID, &h.ChangedBy, &h.FieldName,
			&h.OldValue, &h.NewValue, &h.ChangedAt); err != nil {
			return nil, err
		}
		history = append(history, h)
	}
	return history, rows.Err()
}

func (r *TaskRepo) CreateHistoryEntry(ctx context.Context, entry *domain.TaskHistory) error {
	entry.ID = uuid.New().String()
	_, err := r.conn(ctx).ExecContext(ctx,
		`INSERT INTO task_history (id, task_id, changed_by, field_name, old_value, new_value)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		entry.ID, entry.TaskID, entry.ChangedBy, entry.FieldName, entry.OldValue, entry.NewValue,
	)
	return err
}

func (r *TaskRepo) TeamStats(ctx context.Context) ([]ports.TeamStat, error) {
	rows, err := r.conn(ctx).QueryContext(ctx, `
		SELECT t.id, t.name,
		       COUNT(DISTINCT tm.user_id) AS members_count,
		       COUNT(DISTINCT CASE
		           WHEN tk.status = 'done' AND tk.updated_at >= NOW() - INTERVAL 7 DAY
		           THEN tk.id
		       END) AS done_last_week
		FROM teams t
		LEFT JOIN team_members tm ON tm.team_id = t.id
		LEFT JOIN tasks tk ON tk.team_id = t.id
		GROUP BY t.id, t.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []ports.TeamStat
	for rows.Next() {
		var s ports.TeamStat
		if err := rows.Scan(&s.TeamID, &s.TeamName, &s.MembersCount, &s.DoneLastWeek); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, rows.Err()
}

func (r *TaskRepo) TopContributors(ctx context.Context, teamID string) ([]ports.UserRank, error) {
	rows, err := r.conn(ctx).QueryContext(ctx, `
		SELECT user_id, user_name, team_id, tasks_created, rn FROM (
		    SELECT u.id AS user_id, u.name AS user_name, tm.team_id,
		           COUNT(tk.id) AS tasks_created,
		           ROW_NUMBER() OVER (PARTITION BY tm.team_id ORDER BY COUNT(tk.id) DESC) AS rn
		    FROM users u
		    JOIN team_members tm ON tm.user_id = u.id
		    LEFT JOIN tasks tk ON tk.created_by = u.id AND tk.team_id = tm.team_id
		        AND tk.created_at >= NOW() - INTERVAL 1 MONTH
		    WHERE tm.team_id = ?
		    GROUP BY u.id, u.name, tm.team_id
		) ranked WHERE rn <= 3`, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ranks []ports.UserRank
	for rows.Next() {
		var r ports.UserRank
		if err := rows.Scan(&r.UserID, &r.UserName, &r.TeamID, &r.TasksCreated, &r.Rank); err != nil {
			return nil, err
		}
		ranks = append(ranks, r)
	}
	return ranks, rows.Err()
}

func (r *TaskRepo) OrphanedAssignees(ctx context.Context) ([]domain.Task, error) {
	rows, err := r.conn(ctx).QueryContext(ctx, `
		SELECT tk.id, tk.title, tk.description, tk.status, tk.assignee_id,
		       tk.team_id, tk.created_by, tk.created_at, tk.updated_at
		FROM tasks tk
		WHERE tk.assignee_id IS NOT NULL
		  AND NOT EXISTS (
		      SELECT 1 FROM team_members tm
		      WHERE tm.team_id = tk.team_id AND tm.user_id = tk.assignee_id
		  )`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []domain.Task
	for rows.Next() {
		var t domain.Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.AssigneeID,
			&t.TeamID, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}
