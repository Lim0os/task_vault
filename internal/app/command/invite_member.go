package command

import (
	"context"
	"fmt"
	"task_vault/internal/domain"
	"task_vault/internal/ports"
)

type InviteMemberInput struct {
	TeamID      string
	InvitedByID string
	UserEmail   string
}

type InviteMemberHandler struct {
	teamCmd   ports.TeamCommandRepo
	teamQuery ports.TeamQueryRepo
	userQuery ports.UserQueryRepo
}

func NewInviteMemberHandler(
	teamCmd ports.TeamCommandRepo,
	teamQuery ports.TeamQueryRepo,
	userQuery ports.UserQueryRepo,
) *InviteMemberHandler {
	return &InviteMemberHandler{
		teamCmd:   teamCmd,
		teamQuery: teamQuery,
		userQuery: userQuery,
	}
}

func (h *InviteMemberHandler) Handle(ctx context.Context, input InviteMemberInput) error {
	inviter, err := h.teamQuery.GetMember(ctx, input.TeamID, input.InvitedByID)
	if err != nil {
		return domain.ErrNoPermission
	}
	if inviter.Role != domain.RoleOwner && inviter.Role != domain.RoleAdmin {
		return domain.ErrNoPermission
	}

	invitedUser, err := h.userQuery.GetByEmail(ctx, input.UserEmail)
	if err != nil {
		return domain.ErrUserNotFound
	}

	existing, _ := h.teamQuery.GetMember(ctx, input.TeamID, invitedUser.ID)
	if existing != nil {
		return domain.ErrAlreadyMember
	}

	member := &domain.TeamMember{
		UserID: invitedUser.ID,
		TeamID: input.TeamID,
		Role:   domain.RoleMember,
	}
	if err := h.teamCmd.AddMember(ctx, member); err != nil {
		return fmt.Errorf("приглашение в команду [team_id=%s, user_id=%s]: %w", input.TeamID, invitedUser.ID, err)
	}
	return nil
}
