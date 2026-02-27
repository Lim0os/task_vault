package mysql

import (
	"context"
	"task_vault/internal/domain"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepo_CreateAndGetByID(t *testing.T) {
	cleanTables(t)
	repo := NewUserRepo(testDB)
	ctx := context.Background()

	user := &domain.User{
		Email:        "test@example.com",
		PasswordHash: "hash123",
		Name:         "Test User",
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)
	assert.NotEmpty(t, user.ID)

	found, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.Email, found.Email)
	assert.Equal(t, user.Name, found.Name)
	assert.Equal(t, user.PasswordHash, found.PasswordHash)
}

func TestUserRepo_GetByEmail(t *testing.T) {
	cleanTables(t)
	repo := NewUserRepo(testDB)
	ctx := context.Background()

	user := &domain.User{
		Email:        "find@example.com",
		PasswordHash: "hash",
		Name:         "Find Me",
	}
	require.NoError(t, repo.Create(ctx, user))

	found, err := repo.GetByEmail(ctx, "find@example.com")
	require.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
}

func TestUserRepo_GetByEmail_NotFound(t *testing.T) {
	cleanTables(t)
	repo := NewUserRepo(testDB)

	found, err := repo.GetByEmail(context.Background(), "none@example.com")
	assert.Error(t, err)
	assert.Nil(t, found)
}
