package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	client *redis.Client
	limit  int64
	window time.Duration
}

func NewRateLimiter(client *redis.Client, limit int64, window time.Duration) *RateLimiter {
	return &RateLimiter{
		client: client,
		limit:  limit,
		window: window,
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
		return true // fail-open: при ошибке Redis пропускаем
	}

	if count == 1 {
		rl.client.Expire(ctx, key, rl.window+time.Second)
	}

	return count <= rl.limit
}
