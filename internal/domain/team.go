package domain

import "time"

type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)

type Team struct {
	ID        string
	Name      string
	CreatedBy string
	CreatedAt time.Time
}

type TeamMember struct {
	ID       string
	UserID   string
	TeamID   string
	Role     Role
	JoinedAt time.Time
}
