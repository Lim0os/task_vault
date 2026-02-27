package command

import (
	"context"
	"errors"
	"fmt"
	"task_vault/internal/app/auth"
	"task_vault/internal/domain"
	"task_vault/internal/ports"

	"golang.org/x/crypto/bcrypt"
)

type LoginUserInput struct {
	Email    string
	Password string
}

type LoginUserOutput struct {
	Token string
}

type LoginUserHandler struct {
	userQuery ports.UserQueryRepo
	jwt       *auth.JWTManager
}

func NewLoginUserHandler(query ports.UserQueryRepo, jwt *auth.JWTManager) *LoginUserHandler {
	return &LoginUserHandler{userQuery: query, jwt: jwt}
}

func (h *LoginUserHandler) Handle(ctx context.Context, input LoginUserInput) (*LoginUserOutput, error) {
	user, err := h.userQuery.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("авторизация [email=%s]: %w", input.Email, err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	token, err := h.jwt.Generate(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("авторизация, генерация токена [user_id=%s]: %w", user.ID, err)
	}

	return &LoginUserOutput{Token: token}, nil
}
