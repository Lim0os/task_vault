package http

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	migratemysql "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	mysqltc "github.com/testcontainers/testcontainers-go/modules/mysql"
	redistc "github.com/testcontainers/testcontainers-go/modules/redis"

	"log/slog"
	"time"

	"task_vault/internal/adapter/email"
	"task_vault/internal/adapter/http/handler"
	"task_vault/internal/adapter/logging"
	adaptermysql "task_vault/internal/adapter/mysql"
	adapterredis "task_vault/internal/adapter/redis"
	"task_vault/internal/app/auth"
	"task_vault/internal/app/command"
	"task_vault/internal/app/query"
	"task_vault/internal/config"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
)

var (
	testRouter chi.Router
	testDB     *sql.DB
	testRedis  *redis.Client
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	mysqlContainer, err := mysqltc.Run(ctx, "mysql:8.0",
		mysqltc.WithDatabase("task_vault_test"),
		mysqltc.WithUsername("test"),
		mysqltc.WithPassword("test"),
	)
	if err != nil {
		fmt.Printf("не удалось запустить MySQL контейнер: %v\n", err)
		os.Exit(1)
	}
	defer mysqlContainer.Terminate(ctx)

	redisContainer, err := redistc.Run(ctx, "redis:7-alpine")
	if err != nil {
		fmt.Printf("не удалось запустить Redis контейнер: %v\n", err)
		os.Exit(1)
	}
	defer redisContainer.Terminate(ctx)

	mysqlDSN, err := mysqlContainer.ConnectionString(ctx, "parseTime=true", "multiStatements=true")
	if err != nil {
		fmt.Printf("не удалось получить MySQL DSN: %v\n", err)
		os.Exit(1)
	}

	redisEndpoint, err := redisContainer.Endpoint(ctx, "")
	if err != nil {
		fmt.Printf("не удалось получить Redis endpoint: %v\n", err)
		os.Exit(1)
	}

	testDB, err = sql.Open("mysql", mysqlDSN)
	if err != nil {
		fmt.Printf("не удалось подключиться к MySQL: %v\n", err)
		os.Exit(1)
	}
	defer testDB.Close()

	if err := runMigrations(testDB); err != nil {
		fmt.Printf("ошибка миграций: %v\n", err)
		os.Exit(1)
	}

	testRedis = redis.NewClient(&redis.Options{Addr: redisEndpoint})
	if err := testRedis.Ping(ctx).Err(); err != nil {
		fmt.Printf("не удалось подключиться к Redis: %v\n", err)
		os.Exit(1)
	}
	defer testRedis.Close()

	testRouter = buildRouter(testDB, testRedis)

	os.Exit(m.Run())
}

func runMigrations(db *sql.DB) error {
	driver, err := migratemysql.WithInstance(db, &migratemysql.Config{})
	if err != nil {
		return err
	}

	_, filename, _, _ := runtime.Caller(0)
	migrationsPath := filepath.Join(filepath.Dir(filename), "..", "..", "..", "migrations")
	absPath, _ := filepath.Abs(migrationsPath)

	m, err := migrate.NewWithDatabaseInstance("file://"+absPath, "mysql", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func buildRouter(db *sql.DB, redisClient *redis.Client) chi.Router {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	redisCfg := config.RedisConfig{Addr: redisClient.Options().Addr}
	cacheAdapter, err := adapterredis.NewCache(redisCfg)
	if err != nil {
		panic(fmt.Sprintf("не удалось создать кеш: %v", err))
	}

	cacheLog := logging.NewCacheLogger(cacheAdapter, logger)

	userRepo := adaptermysql.NewUserRepo(db)
	teamRepo := adaptermysql.NewTeamRepo(db)
	taskRepo := adaptermysql.NewTaskRepo(db)
	transactor := adaptermysql.NewTransactor(db)

	userLog := logging.NewUserRepoLogger(userRepo, userRepo, logger)
	teamLog := logging.NewTeamRepoLogger(teamRepo, teamRepo, logger)
	taskLog := logging.NewTaskRepoLogger(taskRepo, taskRepo, taskRepo, taskRepo, logger)

	jwtManager := auth.NewJWTManager("integration-test-secret", 24*time.Hour)

	registerUser := command.NewRegisterUserHandler(userLog, userLog)
	loginUser := command.NewLoginUserHandler(userLog, jwtManager)
	createTeam := command.NewCreateTeamHandler(teamLog, transactor)
	inviteMember := command.NewInviteMemberHandler(teamLog, teamLog, userLog)
	createTask := command.NewCreateTaskHandler(taskLog, teamLog, cacheLog)
	updateTask := command.NewUpdateTaskHandler(taskLog, taskLog, teamLog, taskLog, cacheLog, transactor)

	notifier := email.NewNotifier(logger, config.CircuitBreakerConfig{
		MaxRequests:   3,
		Interval:      30 * time.Second,
		Timeout:       10 * time.Second,
		FailThreshold: 5,
	})
	sendInvite := command.NewSendInviteHandler(teamLog, notifier)

	getTasks := query.NewGetTasksHandler(taskLog, cacheLog, 5*time.Minute)
	getTaskHistory := query.NewGetTaskHistoryHandler(taskLog)
	getTeams := query.NewGetTeamsHandler(teamLog)
	teamAnalytics := query.NewTeamAnalyticsHandler(taskLog)

	healthHandler := handler.NewHealthHandler(db, cacheAdapter.Client())
	authHandler := handler.NewAuthHandler(registerUser, loginUser)
	teamHandler := handler.NewTeamHandler(createTeam, inviteMember, sendInvite, getTeams)
	taskHandler := handler.NewTaskHandler(createTask, updateTask, getTasks, getTaskHistory)
	analyticsHandler := handler.NewAnalyticsHandler(teamAnalytics)

	routerCfg := RouterConfig{
		RateLimitRequests: 1000,
		RateLimitWindow:   time.Minute,
	}
	return NewRouter(
		logger, jwtManager, cacheAdapter.Client(), routerCfg,
		healthHandler, authHandler, teamHandler, taskHandler, analyticsHandler,
	)
}

func cleanTables(t *testing.T) {
	t.Helper()
	tables := []string{"task_comments", "task_history", "tasks", "team_members", "teams", "users"}
	for _, table := range tables {
		_, err := testDB.Exec("DELETE FROM " + table)
		require.NoError(t, err, "не удалось очистить таблицу %s", table)
	}
	testRedis.FlushDB(context.Background())
}

// --- helpers ---

type apiResponse struct {
	Data  json.RawMessage `json:"data"`
	Error string          `json:"error"`
}

func doRequest(t *testing.T, method, path string, body any, token string) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&buf).Encode(body))
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)
	return rec
}

func parseData(t *testing.T, rec *httptest.ResponseRecorder, dest any) {
	t.Helper()
	var resp apiResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp), "не удалось декодировать ответ")
	if dest != nil && resp.Data != nil {
		require.NoError(t, json.Unmarshal(resp.Data, dest), "не удалось декодировать data")
	}
}

func registerUser(t *testing.T, name, emailAddr, password string) map[string]any {
	t.Helper()
	rec := doRequest(t, http.MethodPost, "/api/v1/register", map[string]string{
		"email": emailAddr, "password": password, "name": name,
	}, "")
	require.Equal(t, http.StatusCreated, rec.Code, "register failed: %s", rec.Body.String())
	var data map[string]any
	parseData(t, rec, &data)
	return data
}

func loginUser(t *testing.T, emailAddr, password string) string {
	t.Helper()
	rec := doRequest(t, http.MethodPost, "/api/v1/login", map[string]string{
		"email": emailAddr, "password": password,
	}, "")
	require.Equal(t, http.StatusOK, rec.Code, "login failed: %s", rec.Body.String())
	var data map[string]string
	parseData(t, rec, &data)
	return data["token"]
}

func createTeam(t *testing.T, token, name string) map[string]any {
	t.Helper()
	rec := doRequest(t, http.MethodPost, "/api/v1/teams", map[string]string{"name": name}, token)
	require.Equal(t, http.StatusCreated, rec.Code, "create team failed: %s", rec.Body.String())
	var data map[string]any
	parseData(t, rec, &data)
	return data
}

func createTask(t *testing.T, token, title, teamID string) map[string]any {
	t.Helper()
	rec := doRequest(t, http.MethodPost, "/api/v1/tasks", map[string]string{
		"title": title, "team_id": teamID,
	}, token)
	require.Equal(t, http.StatusCreated, rec.Code, "create task failed: %s", rec.Body.String())
	var data map[string]any
	parseData(t, rec, &data)
	return data
}

// =====================
// Auth integration tests
// =====================

func TestIntegration_Register(t *testing.T) {
	cleanTables(t)

	rec := doRequest(t, http.MethodPost, "/api/v1/register", map[string]string{
		"email": "alice@test.com", "password": "password123", "name": "Alice",
	}, "")

	assert.Equal(t, http.StatusCreated, rec.Code)
	var data map[string]any
	parseData(t, rec, &data)
	assert.Equal(t, "alice@test.com", data["email"])
	assert.Equal(t, "Alice", data["name"])
	assert.NotEmpty(t, data["id"])
}

func TestIntegration_Register_DuplicateEmail(t *testing.T) {
	cleanTables(t)
	registerUser(t, "Alice", "alice@test.com", "password123")

	rec := doRequest(t, http.MethodPost, "/api/v1/register", map[string]string{
		"email": "alice@test.com", "password": "password456", "name": "Alice2",
	}, "")

	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestIntegration_Register_Validation(t *testing.T) {
	cleanTables(t)

	rec := doRequest(t, http.MethodPost, "/api/v1/register", map[string]string{
		"email": "bad-email", "password": "short", "name": "",
	}, "")

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestIntegration_Login(t *testing.T) {
	cleanTables(t)
	registerUser(t, "Alice", "alice@test.com", "password123")

	rec := doRequest(t, http.MethodPost, "/api/v1/login", map[string]string{
		"email": "alice@test.com", "password": "password123",
	}, "")

	assert.Equal(t, http.StatusOK, rec.Code)
	var data map[string]string
	parseData(t, rec, &data)
	assert.NotEmpty(t, data["token"])
}

func TestIntegration_Login_WrongPassword(t *testing.T) {
	cleanTables(t)
	registerUser(t, "Alice", "alice@test.com", "password123")

	rec := doRequest(t, http.MethodPost, "/api/v1/login", map[string]string{
		"email": "alice@test.com", "password": "wrongpass",
	}, "")

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestIntegration_Login_NonexistentUser(t *testing.T) {
	cleanTables(t)

	rec := doRequest(t, http.MethodPost, "/api/v1/login", map[string]string{
		"email": "ghost@test.com", "password": "password123",
	}, "")

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// =====================
// Team integration tests
// =====================

func TestIntegration_CreateTeam(t *testing.T) {
	cleanTables(t)
	registerUser(t, "Alice", "alice@test.com", "password123")
	token := loginUser(t, "alice@test.com", "password123")

	rec := doRequest(t, http.MethodPost, "/api/v1/teams", map[string]string{"name": "Alpha Team"}, token)

	assert.Equal(t, http.StatusCreated, rec.Code)
	var data map[string]any
	parseData(t, rec, &data)
	assert.Equal(t, "Alpha Team", data["name"])
	assert.NotEmpty(t, data["id"])
}

func TestIntegration_ListTeams(t *testing.T) {
	cleanTables(t)
	registerUser(t, "Alice", "alice@test.com", "password123")
	token := loginUser(t, "alice@test.com", "password123")
	createTeam(t, token, "Team A")
	createTeam(t, token, "Team B")

	rec := doRequest(t, http.MethodGet, "/api/v1/teams", nil, token)

	assert.Equal(t, http.StatusOK, rec.Code)
	var teams []map[string]any
	parseData(t, rec, &teams)
	assert.Len(t, teams, 2)
}

func TestIntegration_InviteMember(t *testing.T) {
	cleanTables(t)
	registerUser(t, "Alice", "alice@test.com", "password123")
	registerUser(t, "Bob", "bob@test.com", "password123")
	aliceToken := loginUser(t, "alice@test.com", "password123")
	bobToken := loginUser(t, "bob@test.com", "password123")

	team := createTeam(t, aliceToken, "Alpha Team")
	teamID := team["id"].(string)

	// Alice invites Bob
	rec := doRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/teams/%s/invite", teamID),
		map[string]string{"email": "bob@test.com"}, aliceToken)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Bob should now see the team
	rec = doRequest(t, http.MethodGet, "/api/v1/teams", nil, bobToken)
	assert.Equal(t, http.StatusOK, rec.Code)
	var teams []map[string]any
	parseData(t, rec, &teams)
	assert.Len(t, teams, 1)
	assert.Equal(t, "Alpha Team", teams[0]["name"])
}

func TestIntegration_InviteMember_AlreadyMember(t *testing.T) {
	cleanTables(t)
	registerUser(t, "Alice", "alice@test.com", "password123")
	registerUser(t, "Bob", "bob@test.com", "password123")
	token := loginUser(t, "alice@test.com", "password123")

	team := createTeam(t, token, "Team")
	teamID := team["id"].(string)

	doRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/teams/%s/invite", teamID),
		map[string]string{"email": "bob@test.com"}, token)

	// Try to invite Bob again
	rec := doRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/teams/%s/invite", teamID),
		map[string]string{"email": "bob@test.com"}, token)
	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestIntegration_InviteMember_NoPermission(t *testing.T) {
	cleanTables(t)
	registerUser(t, "Alice", "alice@test.com", "password123")
	registerUser(t, "Bob", "bob@test.com", "password123")
	registerUser(t, "Charlie", "charlie@test.com", "password123")
	aliceToken := loginUser(t, "alice@test.com", "password123")
	bobToken := loginUser(t, "bob@test.com", "password123")

	team := createTeam(t, aliceToken, "Team")
	teamID := team["id"].(string)

	// Alice invites Bob as member
	doRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/teams/%s/invite", teamID),
		map[string]string{"email": "bob@test.com"}, aliceToken)

	// Bob (member) tries to invite Charlie — should fail
	rec := doRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/teams/%s/invite", teamID),
		map[string]string{"email": "charlie@test.com"}, bobToken)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestIntegration_Unauthorized_NoToken(t *testing.T) {
	cleanTables(t)

	rec := doRequest(t, http.MethodGet, "/api/v1/teams", nil, "")
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestIntegration_Unauthorized_InvalidToken(t *testing.T) {
	cleanTables(t)

	rec := doRequest(t, http.MethodGet, "/api/v1/teams", nil, "invalid.jwt.token")
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// =====================
// Task integration tests
// =====================

func TestIntegration_CreateTask(t *testing.T) {
	cleanTables(t)
	registerUser(t, "Alice", "alice@test.com", "password123")
	token := loginUser(t, "alice@test.com", "password123")
	team := createTeam(t, token, "Team")
	teamID := team["id"].(string)

	rec := doRequest(t, http.MethodPost, "/api/v1/tasks", map[string]string{
		"title": "My Task", "team_id": teamID,
	}, token)

	assert.Equal(t, http.StatusCreated, rec.Code)
	var data map[string]any
	parseData(t, rec, &data)
	assert.Equal(t, "My Task", data["title"])
	assert.Equal(t, "todo", data["status"])
	assert.NotEmpty(t, data["id"])
}

func TestIntegration_CreateTask_NotTeamMember(t *testing.T) {
	cleanTables(t)
	registerUser(t, "Alice", "alice@test.com", "password123")
	registerUser(t, "Bob", "bob@test.com", "password123")
	aliceToken := loginUser(t, "alice@test.com", "password123")
	bobToken := loginUser(t, "bob@test.com", "password123")

	team := createTeam(t, aliceToken, "Alice Team")
	teamID := team["id"].(string)

	// Bob tries to create task in Alice's team
	rec := doRequest(t, http.MethodPost, "/api/v1/tasks", map[string]string{
		"title": "Hacked", "team_id": teamID,
	}, bobToken)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestIntegration_CreateTask_Validation(t *testing.T) {
	cleanTables(t)
	registerUser(t, "Alice", "alice@test.com", "password123")
	token := loginUser(t, "alice@test.com", "password123")

	// Missing title and team_id
	rec := doRequest(t, http.MethodPost, "/api/v1/tasks", map[string]string{
		"title": "", "team_id": "",
	}, token)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestIntegration_ListTasks_WithFilters(t *testing.T) {
	cleanTables(t)
	registerUser(t, "Alice", "alice@test.com", "password123")
	token := loginUser(t, "alice@test.com", "password123")
	team := createTeam(t, token, "Team")
	teamID := team["id"].(string)

	createTask(t, token, "Task 1", teamID)
	createTask(t, token, "Task 2", teamID)
	createTask(t, token, "Task 3", teamID)

	// List all tasks for team
	rec := doRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/tasks?team_id=%s", teamID), nil, token)
	assert.Equal(t, http.StatusOK, rec.Code)

	var listResp struct {
		Tasks []map[string]any `json:"tasks"`
		Total float64          `json:"total"`
	}
	parseData(t, rec, &listResp)
	assert.Equal(t, float64(3), listResp.Total)
	assert.Len(t, listResp.Tasks, 3)
}

func TestIntegration_ListTasks_Pagination(t *testing.T) {
	cleanTables(t)
	registerUser(t, "Alice", "alice@test.com", "password123")
	token := loginUser(t, "alice@test.com", "password123")
	team := createTeam(t, token, "Team")
	teamID := team["id"].(string)

	for i := 1; i <= 5; i++ {
		createTask(t, token, fmt.Sprintf("Task %d", i), teamID)
	}

	// Page 1: limit=2, offset=0
	rec := doRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/tasks?team_id=%s&limit=2&offset=0", teamID), nil, token)
	assert.Equal(t, http.StatusOK, rec.Code)
	var page1 struct {
		Tasks []map[string]any `json:"tasks"`
		Total float64          `json:"total"`
	}
	parseData(t, rec, &page1)
	assert.Equal(t, float64(5), page1.Total)
	assert.Len(t, page1.Tasks, 2)

	// Page 2: limit=2, offset=2
	rec = doRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/tasks?team_id=%s&limit=2&offset=2", teamID), nil, token)
	var page2 struct {
		Tasks []map[string]any `json:"tasks"`
		Total float64          `json:"total"`
	}
	parseData(t, rec, &page2)
	assert.Len(t, page2.Tasks, 2)

	// Ensure pages have different tasks
	assert.NotEqual(t, page1.Tasks[0]["id"], page2.Tasks[0]["id"])
}

func TestIntegration_ListTasks_FilterByStatus(t *testing.T) {
	cleanTables(t)
	registerUser(t, "Alice", "alice@test.com", "password123")
	token := loginUser(t, "alice@test.com", "password123")
	team := createTeam(t, token, "Team")
	teamID := team["id"].(string)

	task := createTask(t, token, "Task To Complete", teamID)
	taskID := task["id"].(string)
	createTask(t, token, "Task Still Todo", teamID)

	// Update task to done
	doRequest(t, http.MethodPut, fmt.Sprintf("/api/v1/tasks/%s", taskID), map[string]any{
		"status": "done",
	}, token)

	// Filter by status=done
	rec := doRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/tasks?team_id=%s&status=done", teamID), nil, token)
	assert.Equal(t, http.StatusOK, rec.Code)
	var listResp struct {
		Tasks []map[string]any `json:"tasks"`
		Total float64          `json:"total"`
	}
	parseData(t, rec, &listResp)
	assert.Equal(t, float64(1), listResp.Total)
	assert.Equal(t, "done", listResp.Tasks[0]["status"])
}

func TestIntegration_UpdateTask(t *testing.T) {
	cleanTables(t)
	registerUser(t, "Alice", "alice@test.com", "password123")
	token := loginUser(t, "alice@test.com", "password123")
	team := createTeam(t, token, "Team")
	teamID := team["id"].(string)

	task := createTask(t, token, "Original", teamID)
	taskID := task["id"].(string)

	rec := doRequest(t, http.MethodPut, fmt.Sprintf("/api/v1/tasks/%s", taskID), map[string]any{
		"title":  "Updated Title",
		"status": "in_progress",
	}, token)

	assert.Equal(t, http.StatusOK, rec.Code)
	var updated map[string]any
	parseData(t, rec, &updated)
	assert.Equal(t, "Updated Title", updated["title"])
	assert.Equal(t, "in_progress", updated["status"])
}

func TestIntegration_UpdateTask_NotFound(t *testing.T) {
	cleanTables(t)
	registerUser(t, "Alice", "alice@test.com", "password123")
	token := loginUser(t, "alice@test.com", "password123")

	rec := doRequest(t, http.MethodPut, "/api/v1/tasks/nonexistent-id", map[string]any{
		"title": "X",
	}, token)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestIntegration_UpdateTask_NoPermission(t *testing.T) {
	cleanTables(t)
	registerUser(t, "Alice", "alice@test.com", "password123")
	registerUser(t, "Bob", "bob@test.com", "password123")
	aliceToken := loginUser(t, "alice@test.com", "password123")
	bobToken := loginUser(t, "bob@test.com", "password123")

	team := createTeam(t, aliceToken, "Team")
	teamID := team["id"].(string)
	task := createTask(t, aliceToken, "Alice Task", teamID)
	taskID := task["id"].(string)

	// Invite Bob as member
	doRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/teams/%s/invite", teamID),
		map[string]string{"email": "bob@test.com"}, aliceToken)

	// Bob (regular member, not creator/assignee) tries to update
	rec := doRequest(t, http.MethodPut, fmt.Sprintf("/api/v1/tasks/%s", taskID), map[string]any{
		"title": "Hacked",
	}, bobToken)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestIntegration_UpdateTask_InvalidStatus(t *testing.T) {
	cleanTables(t)
	registerUser(t, "Alice", "alice@test.com", "password123")
	token := loginUser(t, "alice@test.com", "password123")
	team := createTeam(t, token, "Team")
	teamID := team["id"].(string)
	task := createTask(t, token, "Task", teamID)
	taskID := task["id"].(string)

	rec := doRequest(t, http.MethodPut, fmt.Sprintf("/api/v1/tasks/%s", taskID), map[string]any{
		"status": "invalid_status",
	}, token)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestIntegration_TaskHistory(t *testing.T) {
	cleanTables(t)
	registerUser(t, "Alice", "alice@test.com", "password123")
	token := loginUser(t, "alice@test.com", "password123")
	team := createTeam(t, token, "Team")
	teamID := team["id"].(string)

	task := createTask(t, token, "Original", teamID)
	taskID := task["id"].(string)

	// Make two updates
	doRequest(t, http.MethodPut, fmt.Sprintf("/api/v1/tasks/%s", taskID), map[string]any{
		"title": "Renamed",
	}, token)
	doRequest(t, http.MethodPut, fmt.Sprintf("/api/v1/tasks/%s", taskID), map[string]any{
		"status": "done",
	}, token)

	// Get history
	rec := doRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/tasks/%s/history", taskID), nil, token)
	assert.Equal(t, http.StatusOK, rec.Code)

	var history []map[string]any
	parseData(t, rec, &history)
	assert.Len(t, history, 2)

	// Verify history entries contain correct field names
	fields := make(map[string]bool)
	for _, h := range history {
		fields[h["field_name"].(string)] = true
	}
	assert.True(t, fields["title"])
	assert.True(t, fields["status"])
}

// =====================
// Analytics integration tests
// =====================

func TestIntegration_Analytics_TeamStats(t *testing.T) {
	cleanTables(t)
	registerUser(t, "Alice", "alice@test.com", "password123")
	registerUser(t, "Bob", "bob@test.com", "password123")
	token := loginUser(t, "alice@test.com", "password123")

	team := createTeam(t, token, "Analytics Team")
	teamID := team["id"].(string)

	// Invite Bob
	doRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/teams/%s/invite", teamID),
		map[string]string{"email": "bob@test.com"}, token)

	// Create a task and mark it done
	task := createTask(t, token, "Done Task", teamID)
	taskID := task["id"].(string)
	doRequest(t, http.MethodPut, fmt.Sprintf("/api/v1/tasks/%s", taskID), map[string]any{
		"status": "done",
	}, token)

	// Create another task (still todo)
	createTask(t, token, "Todo Task", teamID)

	rec := doRequest(t, http.MethodGet, "/api/v1/analytics/teams", nil, token)
	assert.Equal(t, http.StatusOK, rec.Code)

	var stats []map[string]any
	parseData(t, rec, &stats)
	assert.NotEmpty(t, stats)

	var found bool
	for _, s := range stats {
		if s["team_id"] == teamID {
			found = true
			assert.Equal(t, "Analytics Team", s["team_name"])
			assert.Equal(t, float64(2), s["members_count"]) // Alice + Bob
			assert.Equal(t, float64(1), s["done_last_week"])
		}
	}
	assert.True(t, found, "team not found in stats")
}

func TestIntegration_Analytics_TopContributors(t *testing.T) {
	cleanTables(t)
	registerUser(t, "Alice", "alice@test.com", "password123")
	registerUser(t, "Bob", "bob@test.com", "password123")
	aliceToken := loginUser(t, "alice@test.com", "password123")
	bobToken := loginUser(t, "bob@test.com", "password123")

	team := createTeam(t, aliceToken, "Team")
	teamID := team["id"].(string)
	doRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/teams/%s/invite", teamID),
		map[string]string{"email": "bob@test.com"}, aliceToken)

	// Alice creates 3 tasks, Bob creates 1
	for i := 0; i < 3; i++ {
		createTask(t, aliceToken, fmt.Sprintf("Alice Task %d", i), teamID)
	}
	createTask(t, bobToken, "Bob Task", teamID)

	rec := doRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/analytics/teams/%s/top", teamID), nil, aliceToken)
	assert.Equal(t, http.StatusOK, rec.Code)

	var ranks []map[string]any
	parseData(t, rec, &ranks)
	assert.NotEmpty(t, ranks)
	// First should be Alice (3 tasks)
	assert.Equal(t, float64(3), ranks[0]["tasks_created"])
}

func TestIntegration_Analytics_OrphanedAssignees(t *testing.T) {
	cleanTables(t)
	registerUser(t, "Alice", "alice@test.com", "password123")
	registerUser(t, "Bob", "bob@test.com", "password123")
	aliceToken := loginUser(t, "alice@test.com", "password123")

	team := createTeam(t, aliceToken, "Team")
	teamID := team["id"].(string)

	// Invite Bob so he can be assigned
	doRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/teams/%s/invite", teamID),
		map[string]string{"email": "bob@test.com"}, aliceToken)

	// Create task assigned to Bob
	bobData := registerUser(t, "Charlie", "charlie@test.com", "password123")
	_ = bobData

	// We need Bob's user ID. Let's get it via the task approach.
	// Actually, let's assign a task to Bob, then remove Bob from the team manually to create orphan.
	// For simplicity, let's use a direct DB insert to assign to a non-member.
	// Actually the API doesn't validate assignee membership on create, so we can assign to Charlie who is not in team.
	rec := doRequest(t, http.MethodPost, "/api/v1/register", map[string]string{
		"email": "outsider@test.com", "password": "password123", "name": "Outsider",
	}, "")
	var outsiderData map[string]any
	parseData(t, rec, &outsiderData)
	outsiderID := outsiderData["id"].(string)

	// Create task with assignee who is not a team member
	rec = doRequest(t, http.MethodPost, "/api/v1/tasks", map[string]any{
		"title": "Orphaned Task", "team_id": teamID, "assignee_id": outsiderID,
	}, aliceToken)
	require.Equal(t, http.StatusCreated, rec.Code)

	rec = doRequest(t, http.MethodGet, "/api/v1/analytics/orphaned-assignees", nil, aliceToken)
	assert.Equal(t, http.StatusOK, rec.Code)

	var tasks []map[string]any
	parseData(t, rec, &tasks)
	assert.NotEmpty(t, tasks)
	assert.Equal(t, "Orphaned Task", tasks[0]["title"])
}

// =====================
// Full flow integration test
// =====================

func TestIntegration_FullWorkflow(t *testing.T) {
	cleanTables(t)

	// 1. Register two users
	registerUser(t, "Alice", "alice@test.com", "password123")
	registerUser(t, "Bob", "bob@test.com", "password123")

	// 2. Login
	aliceToken := loginUser(t, "alice@test.com", "password123")
	bobToken := loginUser(t, "bob@test.com", "password123")

	// 3. Alice creates a team
	team := createTeam(t, aliceToken, "Project X")
	teamID := team["id"].(string)

	// 4. Alice invites Bob
	rec := doRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/teams/%s/invite", teamID),
		map[string]string{"email": "bob@test.com"}, aliceToken)
	require.Equal(t, http.StatusOK, rec.Code)

	// 5. Alice creates tasks
	task1 := createTask(t, aliceToken, "Design API", teamID)
	task1ID := task1["id"].(string)
	createTask(t, aliceToken, "Write tests", teamID)

	// 6. Bob creates a task too (he's now a member)
	task3 := createTask(t, bobToken, "Setup CI", teamID)
	task3ID := task3["id"].(string)

	// 7. Alice updates task1
	rec = doRequest(t, http.MethodPut, fmt.Sprintf("/api/v1/tasks/%s", task1ID), map[string]any{
		"status": "in_progress",
	}, aliceToken)
	assert.Equal(t, http.StatusOK, rec.Code)

	rec = doRequest(t, http.MethodPut, fmt.Sprintf("/api/v1/tasks/%s", task1ID), map[string]any{
		"status": "done",
	}, aliceToken)
	assert.Equal(t, http.StatusOK, rec.Code)

	// 8. Bob updates his task
	rec = doRequest(t, http.MethodPut, fmt.Sprintf("/api/v1/tasks/%s", task3ID), map[string]any{
		"status": "done",
	}, bobToken)
	assert.Equal(t, http.StatusOK, rec.Code)

	// 9. List tasks filtered by status=done
	rec = doRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/tasks?team_id=%s&status=done", teamID), nil, aliceToken)
	assert.Equal(t, http.StatusOK, rec.Code)
	var doneList struct {
		Tasks []map[string]any `json:"tasks"`
		Total float64          `json:"total"`
	}
	parseData(t, rec, &doneList)
	assert.Equal(t, float64(2), doneList.Total)

	// 10. Check task history
	rec = doRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/tasks/%s/history", task1ID), nil, aliceToken)
	assert.Equal(t, http.StatusOK, rec.Code)
	var history []map[string]any
	parseData(t, rec, &history)
	assert.Len(t, history, 2) // in_progress + done

	// 11. Check team stats
	rec = doRequest(t, http.MethodGet, "/api/v1/analytics/teams", nil, aliceToken)
	assert.Equal(t, http.StatusOK, rec.Code)
	var stats []map[string]any
	parseData(t, rec, &stats)
	var teamStat map[string]any
	for _, s := range stats {
		if s["team_id"] == teamID {
			teamStat = s
		}
	}
	require.NotNil(t, teamStat)
	assert.Equal(t, float64(2), teamStat["members_count"])
	assert.Equal(t, float64(2), teamStat["done_last_week"])

	// 12. Check top contributors
	rec = doRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/analytics/teams/%s/top", teamID), nil, aliceToken)
	assert.Equal(t, http.StatusOK, rec.Code)
	var ranks []map[string]any
	parseData(t, rec, &ranks)
	assert.NotEmpty(t, ranks)
}

// =====================
// Health endpoint
// =====================

func TestIntegration_HealthLive(t *testing.T) {
	rec := doRequest(t, http.MethodGet, "/health/live", nil, "")
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestIntegration_HealthReady(t *testing.T) {
	rec := doRequest(t, http.MethodGet, "/health/ready", nil, "")
	assert.Equal(t, http.StatusOK, rec.Code)
}
