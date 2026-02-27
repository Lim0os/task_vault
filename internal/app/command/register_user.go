package command

import (
	"context"
	"errors"
	"fmt"
	"task_vault/internal/domain"
	"task_vault/internal/ports"

	"golang.org/x/crypto/bcrypt"
)

type RegisterUserInput struct {
	Email    string
	Password string
	Name     string
}

type RegisterUserHandler struct {
	userCmd   ports.UserCommandRepo
	userQuery ports.UserQueryRepo
}

func NewRegisterUserHandler(cmd ports.UserCommandRepo, query ports.UserQueryRepo) *RegisterUserHandler {
	return &RegisterUserHandler{userCmd: cmd, userQuery: query}
}

func (h *RegisterUserHandler) Handle(ctx context.Context, input RegisterUserInput) (*domain.User, error) {
	_, err := h.userQuery.GetByEmail(ctx, input.Email)
	if err == nil {
		return nil, domain.ErrEmailTaken
	}
	if !errors.Is(err, domain.ErrUserNotFound) {
		return nil, fmt.Errorf("проверка email [email=%s]: %w", input.Email, err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("регистрация пользователя, хеширование пароля: %w", err)
	}

	user := &domain.User{
		Email:        input.Email,
		PasswordHash: string(hash),
		Name:         input.Name,
	}

	if err := h.userCmd.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("регистрация пользователя [email=%s]: %w", input.Email, err)
	}
	return user, nil
}
