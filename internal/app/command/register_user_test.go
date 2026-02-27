package command

import (
	"context"
	"database/sql"
	"errors"
	"task_vault/internal/domain"
	"task_vault/internal/ports/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegisterUser_Success(t *testing.T) {
	userCmd := new(mocks.UserCommandRepo)
	userQuery := new(mocks.UserQueryRepo)

	userQuery.On("GetByEmail", mock.Anything, "test@example.com").
		Return(nil, sql.ErrNoRows)
	userCmd.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).
		Return(nil).
		Run(func(args mock.Arguments) {
			user := args.Get(1).(*domain.User)
			user.ID = "user-uuid-1"
		})

	handler := NewRegisterUserHandler(userCmd, userQuery)
	user, err := handler.Handle(context.Background(), RegisterUserInput{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	})

	assert.NoError(t, err)
	assert.Equal(t, "user-uuid-1", user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.NotEmpty(t, user.PasswordHash)
	userCmd.AssertExpectations(t)
	userQuery.AssertExpectations(t)
}

func TestRegisterUser_EmailTaken(t *testing.T) {
	userCmd := new(mocks.UserCommandRepo)
	userQuery := new(mocks.UserQueryRepo)

	userQuery.On("GetByEmail", mock.Anything, "taken@example.com").
		Return(&domain.User{ID: "user-uuid-1", Email: "taken@example.com"}, nil)

	handler := NewRegisterUserHandler(userCmd, userQuery)
	user, err := handler.Handle(context.Background(), RegisterUserInput{
		Email:    "taken@example.com",
		Password: "password123",
		Name:     "Test",
	})

	assert.True(t, errors.Is(err, domain.ErrEmailTaken))
	assert.Nil(t, user)
}
