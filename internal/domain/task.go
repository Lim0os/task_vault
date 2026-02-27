package domain

import "time"

type Status string

const (
	StatusTodo       Status = "todo"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type Task struct {
	ID          string
	Title       string
	Description string
	Status      Status
	AssigneeID  *string
	TeamID      string
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
