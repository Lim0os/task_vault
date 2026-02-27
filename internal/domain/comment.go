package domain

import "time"

type Comment struct {
	ID        string
	TaskID    string
	UserID    string
	Content   string
	CreatedAt time.Time
}
