package ports

import "context"

type Notifier interface {
	SendInvite(ctx context.Context, toEmail, teamName string) error
}
