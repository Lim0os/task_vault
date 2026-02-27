package adapter

import (
	"fmt"
	"log/slog"
	"time"
)

func Retry(logger *slog.Logger, name string, maxAttempts int, fn func() error) error {
	delay := time.Second

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := fn(); err != nil {
			if attempt == maxAttempts {
				return fmt.Errorf("%s: все %d попыток исчерпаны: %w", name, maxAttempts, err)
			}
			logger.Warn("попытка подключения не удалась",
				"компонент", name,
				"попытка", attempt,
				"следующая_через", delay.String(),
				"error", err,
			)
			time.Sleep(delay)
			delay *= 2
			if delay > 30*time.Second {
				delay = 30 * time.Second
			}
			continue
		}
		return nil
	}
	return nil
}
