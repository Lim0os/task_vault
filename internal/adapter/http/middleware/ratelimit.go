package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	client *redis.Client
	limit  int64
	window time.Duration
	logger *slog.Logger
}

func NewRateLimiter(client *redis.Client, limit int64, window time.Duration, logger *slog.Logger) *RateLimiter {
	return &RateLimiter{
		client: client,
		limit:  limit,
		window: window,
		logger: logger,
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := GetUserID(r.Context())
		if userID == "" {
			next.ServeHTTP(w, r)
			return
		}

		if !rl.allow(r.Context(), userID) {
			http.Error(w, `{"error":"превышен лимит запросов"}`, http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) allow(ctx context.Context, userID string) bool {
	windowSec := int64(rl.window.Seconds())
	windowStart := time.Now().Unix() / windowSec * windowSec
	key := fmt.Sprintf("ratelimit:%s:%d", userID, windowStart)

	count, err := rl.client.Incr(ctx, key).Result()
	if err != nil {
		rl.logger.Error("rate limiter: ошибка Redis, запрос отклонён",
			"user_id", userID,
			"error", err,
		)
		return false
	}

	if count == 1 {
		rl.client.Expire(ctx, key, rl.window+time.Second)
	}

	return count <= rl.limit
}
