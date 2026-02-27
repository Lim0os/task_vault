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

func TestInviteMember_Success(t *testing.T) {
	teamCmd := new(mocks.TeamCommandRepo)
	teamQuery := new(mocks.TeamQueryRepo)
	userQuery := new(mocks.UserQueryRepo)

	teamQuery.On("GetMember", mock.Anything, "team-uuid-1", "user-uuid-100").
		Return(&domain.TeamMember{Role: domain.RoleOwner}, nil)
	userQuery.On("GetByEmail", mock.Anything, "new@example.com").
		Return(&domain.User{ID: "user-uuid-200"}, nil)
	teamQuery.On("GetMember", mock.Anything, "team-uuid-1", "user-uuid-200").
		Return(nil, sql.ErrNoRows)
	teamCmd.On("AddMember", mock.Anything, mock.AnythingOfType("*domain.TeamMember")).
		Return(nil)

	handler := NewInviteMemberHandler(teamCmd, teamQuery, userQuery)
	err := handler.Handle(context.Background(), InviteMemberInput{
		TeamID:      "team-uuid-1",
		InvitedByID: "user-uuid-100",
		UserEmail:   "new@example.com",
	})

	assert.NoError(t, err)
	teamCmd.AssertCalled(t, "AddMember", mock.Anything, mock.MatchedBy(func(m *domain.TeamMember) bool {
		return m.UserID == "user-uuid-200" && m.Role == domain.RoleMember
	}))
}

func TestInviteMember_NoPermission(t *testing.T) {
	teamCmd := new(mocks.TeamCommandRepo)
	teamQuery := new(mocks.TeamQueryRepo)
	userQuery := new(mocks.UserQueryRepo)

	teamQuery.On("GetMember", mock.Anything, "team-uuid-1", "user-uuid-100").
		Return(&domain.TeamMember{Role: domain.RoleMember}, nil)

	handler := NewInviteMemberHandler(teamCmd, teamQuery, userQuery)
	err := handler.Handle(context.Background(), InviteMemberInput{
		TeamID:      "team-uuid-1",
		InvitedByID: "user-uuid-100",
		UserEmail:   "new@example.com",
	})

	assert.True(t, errors.Is(err, domain.ErrNoPermission))
}

func TestInviteMember_AlreadyMember(t *testing.T) {
	teamCmd := new(mocks.TeamCommandRepo)
	teamQuery := new(mocks.TeamQueryRepo)
	userQuery := new(mocks.UserQueryRepo)

	teamQuery.On("GetMember", mock.Anything, "team-uuid-1", "user-uuid-100").
		Return(&domain.TeamMember{Role: domain.RoleOwner}, nil)
	userQuery.On("GetByEmail", mock.Anything, "existing@example.com").
		Return(&domain.User{ID: "user-uuid-200"}, nil)
	teamQuery.On("GetMember", mock.Anything, "team-uuid-1", "user-uuid-200").
		Return(&domain.TeamMember{}, nil)

	handler := NewInviteMemberHandler(teamCmd, teamQuery, userQuery)
	err := handler.Handle(context.Background(), InviteMemberInput{
		TeamID:      "team-uuid-1",
		InvitedByID: "user-uuid-100",
		UserEmail:   "existing@example.com",
	})

	assert.True(t, errors.Is(err, domain.ErrAlreadyMember))
}
