package command

import (
	"context"
	"errors"
	"task_vault/internal/app/auth"
	"task_vault/internal/domain"
	"task_vault/internal/ports/mocks"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestLoginUser_Success(t *testing.T) {
	userQuery := new(mocks.UserQueryRepo)
	jwtManager := auth.NewJWTManager("test-secret", time.Hour)

	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	userQuery.On("GetByEmail", mock.Anything, "test@example.com").
		Return(&domain.User{ID: "user-uuid-1", Email: "test@example.com", PasswordHash: string(hash)}, nil)

	handler := NewLoginUserHandler(userQuery, jwtManager)
	output, err := handler.Handle(context.Background(), LoginUserInput{
		Email:    "test@example.com",
		Password: "password123",
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, output.Token)

	claims, err := jwtManager.Validate(output.Token)
	assert.NoError(t, err)
	assert.Equal(t, "user-uuid-1", claims.UserID)
}

func TestLoginUser_WrongPassword(t *testing.T) {
	userQuery := new(mocks.UserQueryRepo)
	jwtManager := auth.NewJWTManager("test-secret", time.Hour)

	hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
	userQuery.On("GetByEmail", mock.Anything, "test@example.com").
		Return(&domain.User{ID: "user-uuid-1", PasswordHash: string(hash)}, nil)

	handler := NewLoginUserHandler(userQuery, jwtManager)
	output, err := handler.Handle(context.Background(), LoginUserInput{
		Email:    "test@example.com",
		Password: "wrong",
	})

	assert.True(t, errors.Is(err, domain.ErrInvalidCredentials))
	assert.Nil(t, output)
}

func TestLoginUser_UserNotFound(t *testing.T) {
	userQuery := new(mocks.UserQueryRepo)
	jwtManager := auth.NewJWTManager("test-secret", time.Hour)

	userQuery.On("GetByEmail", mock.Anything, "none@example.com").
		Return(nil, domain.ErrUserNotFound)

	handler := NewLoginUserHandler(userQuery, jwtManager)
	output, err := handler.Handle(context.Background(), LoginUserInput{
		Email:    "none@example.com",
		Password: "password",
	})

	assert.True(t, errors.Is(err, domain.ErrInvalidCredentials))
	assert.Nil(t, output)
}
