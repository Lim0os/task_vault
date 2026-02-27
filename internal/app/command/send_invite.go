package command

import (
	"context"
	"task_vault/internal/domain"
	"task_vault/internal/ports"
)

type SendInviteInput struct {
	TeamID    string
	SenderID  string
	UserEmail string
}

type SendInviteHandler struct {
	teamQuery ports.TeamQueryRepo
	notifier  ports.Notifier
}

func NewSendInviteHandler(
	teamQuery ports.TeamQueryRepo,
	notifier ports.Notifier,
) *SendInviteHandler {
	return &SendInviteHandler{
		teamQuery: teamQuery,
		notifier:  notifier,
	}
}

func (h *SendInviteHandler) Handle(ctx context.Context, input SendInviteInput) error {
	member, err := h.teamQuery.GetMember(ctx, input.TeamID, input.SenderID)
	if err != nil {
		return domain.ErrNoPermission
	}
	if member.Role != domain.RoleOwner && member.Role != domain.RoleAdmin {
		return domain.ErrNoPermission
	}

	team, err := h.teamQuery.GetByID(ctx, input.TeamID)
	if err != nil {
		return domain.ErrTeamNotFound
	}

	return h.notifier.SendInvite(ctx, input.UserEmail, team.Name)
}
