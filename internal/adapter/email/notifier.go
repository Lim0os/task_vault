package email

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sony/gobreaker"
	"task_vault/internal/config"
)

type Notifier struct {
	cb     *gobreaker.CircuitBreaker
	logger *slog.Logger
}

func NewNotifier(logger *slog.Logger, cfg config.CircuitBreakerConfig) *Notifier {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "email-notifier",
		MaxRequests: cfg.MaxRequests,
		Interval:    cfg.Interval,
		Timeout:     cfg.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= cfg.FailThreshold
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			logger.Warn("circuit breaker: смена состояния",
				"name", name,
				"from", from.String(),
				"to", to.String(),
			)
		},
	})

	return &Notifier{cb: cb, logger: logger}
}

func (n *Notifier) SendInvite(ctx context.Context, toEmail, teamName string) error {
	_, err := n.cb.Execute(func() (any, error) {
		return nil, n.doSend(toEmail, teamName)
	})
	return err
}

func (n *Notifier) doSend(toEmail, teamName string) error {
	n.logger.Info("отправка email-приглашения",
		"to", toEmail,
		"team", teamName,
	)
	fmt.Printf("[MOCK EMAIL] Приглашение в команду %q отправлено на %s\n", teamName, toEmail)
	return nil
}
