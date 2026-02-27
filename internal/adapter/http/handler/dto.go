package handler

import "task_vault/internal/domain"

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type CreateTeamRequest struct {
	Name string `json:"name"`
}

type InviteRequest struct {
	Email string `json:"email"`
}

type TeamResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedBy string `json:"created_by"`
}

type CreateTaskRequest struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	AssigneeID  *string `json:"assignee_id"`
	TeamID      string  `json:"team_id"`
}

type UpdateTaskRequest struct {
	Title       *string        `json:"title"`
	Description *string        `json:"description"`
	Status      *domain.Status `json:"status"`
	AssigneeID  *string        `json:"assignee_id"`
}

type TaskResponse struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	AssigneeID  *string `json:"assignee_id"`
	TeamID      string  `json:"team_id"`
	CreatedBy   string  `json:"created_by"`
}

type TaskListResponse struct {
	Tasks []TaskResponse `json:"tasks"`
	Total int64          `json:"total"`
}

type HistoryEntry struct {
	ID        string `json:"id"`
	FieldName string `json:"field_name"`
	OldValue  string `json:"old_value"`
	NewValue  string `json:"new_value"`
	ChangedBy string `json:"changed_by"`
}

type UserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type swagUserResponse struct {
	Data  UserResponse `json:"data"`
	Error string       `json:"error,omitempty"`
}

type swagLoginResponse struct {
	Data  LoginResponse `json:"data"`
	Error string        `json:"error,omitempty"`
}

type swagTeamResponse struct {
	Data  TeamResponse `json:"data"`
	Error string       `json:"error,omitempty"`
}

type swagTeamListResponse struct {
	Data  []TeamResponse `json:"data"`
	Error string         `json:"error,omitempty"`
}

type swagStatusResponse struct {
	Data  map[string]string `json:"data"`
	Error string            `json:"error,omitempty"`
}

type swagTaskResponse struct {
	Data  TaskResponse `json:"data"`
	Error string       `json:"error,omitempty"`
}

type swagTaskListResponse struct {
	Data  TaskListResponse `json:"data"`
	Error string           `json:"error,omitempty"`
}

type swagHistoryListResponse struct {
	Data  []HistoryEntry `json:"data"`
	Error string         `json:"error,omitempty"`
}

type swagErrorResponse struct {
	Error string `json:"error"`
}
