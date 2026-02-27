package domain

import "time"

type TaskHistory struct {
	ID        string
	TaskID    string
	ChangedBy string
	FieldName string
	OldValue  string
	NewValue  string
	ChangedAt time.Time
}
