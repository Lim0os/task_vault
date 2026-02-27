package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type HealthHandler struct {
	db    *sql.DB
	redis *redis.Client
}

func NewHealthHandler(db *sql.DB, redisClient *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, redis: redisClient}
}

func (h *HealthHandler) Live(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *HealthHandler) Ready(w http.ResponseWriter, _ *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	details := make(map[string]string)
	healthy := true

	if err := h.db.PingContext(ctx); err != nil {
		details["mysql"] = err.Error()
		healthy = false
	} else {
		details["mysql"] = "ok"
	}

	if err := h.redis.Ping(ctx).Err(); err != nil {
		details["redis"] = err.Error()
		healthy = false
	} else {
		details["redis"] = "ok"
	}

	w.Header().Set("Content-Type", "application/json")
	if !healthy {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]any{"status": "unavailable", "details": details})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"status": "ok", "details": details})
}
