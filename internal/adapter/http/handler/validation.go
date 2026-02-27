package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"task_vault/internal/domain"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// ValidationError — структурированная ошибка валидации с указанием полей
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string {
	return "ошибка валидации"
}

func (r *RegisterRequest) Validate() *ValidationError {
	errs := make(map[string]string)

	if strings.TrimSpace(r.Name) == "" {
		errs["name"] = "имя обязательно"
	}
	if strings.TrimSpace(r.Email) == "" {
		errs["email"] = "email обязателен"
	} else if !emailRegex.MatchString(r.Email) {
		errs["email"] = "невалидный формат email"
	}
	if len(r.Password) < 8 {
		errs["password"] = fmt.Sprintf("минимум 8 символов, получено %d", len(r.Password))
	}

	if len(errs) > 0 {
		return &ValidationError{Fields: errs}
	}
	return nil
}

func (r *LoginRequest) Validate() *ValidationError {
	errs := make(map[string]string)

	if strings.TrimSpace(r.Email) == "" {
		errs["email"] = "email обязателен"
	}
	if r.Password == "" {
		errs["password"] = "пароль обязателен"
	}

	if len(errs) > 0 {
		return &ValidationError{Fields: errs}
	}
	return nil
}

func (r *CreateTeamRequest) Validate() *ValidationError {
	if strings.TrimSpace(r.Name) == "" {
		return &ValidationError{Fields: map[string]string{"name": "название команды обязательно"}}
	}
	return nil
}

func (r *InviteRequest) Validate() *ValidationError {
	errs := make(map[string]string)

	if strings.TrimSpace(r.Email) == "" {
		errs["email"] = "email обязателен"
	} else if !emailRegex.MatchString(r.Email) {
		errs["email"] = "невалидный формат email"
	}

	if len(errs) > 0 {
		return &ValidationError{Fields: errs}
	}
	return nil
}

func (r *CreateTaskRequest) Validate() *ValidationError {
	errs := make(map[string]string)

	if strings.TrimSpace(r.Title) == "" {
		errs["title"] = "название задачи обязательно"
	}
	if r.TeamID == "" {
		errs["team_id"] = "team_id обязателен"
	}

	if len(errs) > 0 {
		return &ValidationError{Fields: errs}
	}
	return nil
}

var validStatuses = map[domain.Status]bool{
	domain.StatusTodo:       true,
	domain.StatusInProgress: true,
	domain.StatusDone:       true,
}

func (r *UpdateTaskRequest) Validate() *ValidationError {
	errs := make(map[string]string)

	if r.Title != nil && strings.TrimSpace(*r.Title) == "" {
		errs["title"] = "название не может быть пустым"
	}
	if r.Status != nil && !validStatuses[*r.Status] {
		errs["status"] = fmt.Sprintf("недопустимый статус: %s", *r.Status)
	}

	if len(errs) > 0 {
		return &ValidationError{Fields: errs}
	}
	return nil
}

func writeValidationError(w http.ResponseWriter, err *ValidationError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	resp := struct {
		Error  string            `json:"error"`
		Fields map[string]string `json:"fields"`
	}{
		Error:  "ошибка валидации",
		Fields: err.Fields,
	}
	encodeJSON(w, resp)
}

func encodeJSON(w http.ResponseWriter, v any) {
	enc := json.NewEncoder(w)
	enc.Encode(v)
}
