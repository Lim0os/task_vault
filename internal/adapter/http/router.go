package http

import (
	"log/slog"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"task_vault/internal/adapter/http/handler"
	"task_vault/internal/adapter/http/middleware"
	"task_vault/internal/app/auth"
)

type RouterConfig struct {
	RateLimitRequests int64
	RateLimitWindow   time.Duration
}

func NewRouter(
	logger *slog.Logger,
	jwtManager *auth.JWTManager,
	redisClient *redis.Client,
	routerCfg RouterConfig,
	healthHandler *handler.HealthHandler,
	authHandler *handler.AuthHandler,
	teamHandler *handler.TeamHandler,
	taskHandler *handler.TaskHandler,
	analyticsHandler *handler.AnalyticsHandler,
) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger(logger))
	r.Use(middleware.Metrics)

	r.Handle("/metrics", promhttp.Handler())
	r.Get("/swagger/*", httpSwagger.WrapHandler)
	r.Get("/health/live", healthHandler.Live)
	r.Get("/health/ready", healthHandler.Ready)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)

		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuth(jwtManager))
			r.Use(middleware.NewRateLimiter(redisClient, routerCfg.RateLimitRequests, routerCfg.RateLimitWindow, logger).Middleware)

			r.Post("/teams", teamHandler.Create)
			r.Get("/teams", teamHandler.List)
			r.Post("/teams/{id}/invite", teamHandler.Invite)

			r.Post("/tasks", taskHandler.Create)
			r.Get("/tasks", taskHandler.List)
			r.Put("/tasks/{id}", taskHandler.Update)
			r.Get("/tasks/{id}/history", taskHandler.History)

			r.Get("/analytics/teams", analyticsHandler.TeamStats)
			r.Get("/analytics/teams/{id}/top", analyticsHandler.TopContributors)
			r.Get("/analytics/orphaned-assignees", analyticsHandler.OrphanedAssignees)
		})
	})

	return r
}
