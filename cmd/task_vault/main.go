package main

import (
	"database/sql"
	"log/slog"
	"os"

	"github.com/golang-migrate/migrate/v4"
	migratemysql "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	_ "task_vault/docs"
	adapter "task_vault/internal/adapter"
	"task_vault/internal/adapter/email"
	adapthttp "task_vault/internal/adapter/http"
	"task_vault/internal/adapter/http/handler"
	"task_vault/internal/adapter/logging"
	adaptermysql "task_vault/internal/adapter/mysql"
	adapterredis "task_vault/internal/adapter/redis"
	"task_vault/internal/app/auth"
	"task_vault/internal/app/command"
	"task_vault/internal/app/query"
	"task_vault/internal/config"
)

// @title           Хранилище задач API
// @version         1.0
// @description     REST API для управления задачами и командами

// @host            localhost:8080
// @BasePath        /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Bearer JWT-токен (формат: "Bearer <token>")

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	cfg := config.Load()

	var db *sql.DB
	if err := adapter.Retry(logger, "MySQL", 10, func() error {
		var err error
		db, err = adaptermysql.NewConnection(cfg.MySQL)
		return err
	}); err != nil {
		logger.Error("не удалось подключиться к MySQL", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	driver, err := migratemysql.WithInstance(db, &migratemysql.Config{})
	if err != nil {
		logger.Error("ошибка инициализации миграций", "error", err)
		os.Exit(1)
	}
	m, err := migrate.NewWithDatabaseInstance("file://migrations", "mysql", driver)
	if err != nil {
		logger.Error("ошибка создания мигратора", "error", err)
		os.Exit(1)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Error("ошибка выполнения миграций", "error", err)
		os.Exit(1)
	}
	logger.Info("миграции применены")

	var cache *adapterredis.Cache
	if err := adapter.Retry(logger, "Redis", 10, func() error {
		var err error
		cache, err = adapterredis.NewCache(cfg.Redis)
		return err
	}); err != nil {
		logger.Error("не удалось подключиться к Redis", "error", err)
		os.Exit(1)
	}
	defer cache.Close()

	cacheLog := logging.NewCacheLogger(cache, logger)

	userRepo := adaptermysql.NewUserRepo(db)
	teamRepo := adaptermysql.NewTeamRepo(db)
	taskRepo := adaptermysql.NewTaskRepo(db)
	transactor := adaptermysql.NewTransactor(db)

	userLog := logging.NewUserRepoLogger(userRepo, userRepo, logger)
	teamLog := logging.NewTeamRepoLogger(teamRepo, teamRepo, logger)
	taskLog := logging.NewTaskRepoLogger(taskRepo, taskRepo, taskRepo, taskRepo, logger)

	jwtManager := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.TTL)

	registerUser := command.NewRegisterUserHandler(userLog, userLog)
	loginUser := command.NewLoginUserHandler(userLog, jwtManager)
	createTeam := command.NewCreateTeamHandler(teamLog, transactor)
	inviteMember := command.NewInviteMemberHandler(teamLog, teamLog, userLog)
	createTask := command.NewCreateTaskHandler(taskLog, teamLog, cacheLog)
	updateTask := command.NewUpdateTaskHandler(taskLog, taskLog, teamLog, taskLog, cacheLog, transactor)

	notifier := email.NewNotifier(logger, cfg.CircuitBreaker)
	sendInvite := command.NewSendInviteHandler(teamLog, notifier)

	getTasks := query.NewGetTasksHandler(taskLog, cacheLog, cfg.Cache.TasksTTL)
	getTaskHistory := query.NewGetTaskHistoryHandler(taskLog)
	getTeams := query.NewGetTeamsHandler(teamLog)
	teamAnalytics := query.NewTeamAnalyticsHandler(taskLog)

	healthHandler := handler.NewHealthHandler(db, cache.Client())
	authHandler := handler.NewAuthHandler(registerUser, loginUser)
	teamHandler := handler.NewTeamHandler(createTeam, inviteMember, sendInvite, getTeams)
	taskHandler := handler.NewTaskHandler(createTask, updateTask, getTasks, getTaskHistory)
	analyticsHandler := handler.NewAnalyticsHandler(teamAnalytics)

	routerCfg := adapthttp.RouterConfig{
		RateLimitRequests: cfg.RateLimit.RequestsPerWindow,
		RateLimitWindow:   cfg.RateLimit.Window,
	}
	router := adapthttp.NewRouter(
		logger, jwtManager, cache.Client(), routerCfg,
		healthHandler, authHandler, teamHandler, taskHandler, analyticsHandler,
	)
	server := adapthttp.NewServer(":"+cfg.Server.Port, router, logger)

	if err := server.Start(); err != nil {
		logger.Error("ошибка завершения сервера", "error", err)
		os.Exit(1)
	}
}
