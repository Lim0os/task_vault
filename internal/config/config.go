package config

import (
	"log/slog"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server         ServerConfig         `yaml:"server"`
	MySQL          MySQLConfig          `yaml:"mysql"`
	Redis          RedisConfig          `yaml:"redis"`
	JWT            JWTConfig            `yaml:"jwt"`
	RateLimit      RateLimitConfig      `yaml:"rate_limit"`
	Cache          CacheConfig          `yaml:"cache"`
	CircuitBreaker CircuitBreakerConfig `yaml:"circuit_breaker"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
}

type MySQLConfig struct {
	DSN             string        `yaml:"dsn"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type JWTConfig struct {
	Secret string        `yaml:"secret"`
	TTL    time.Duration `yaml:"ttl"`
}

type RateLimitConfig struct {
	RequestsPerWindow int64         `yaml:"requests_per_window"`
	Window            time.Duration `yaml:"window"`
}

type CacheConfig struct {
	TasksTTL time.Duration `yaml:"tasks_ttl"`
}

type CircuitBreakerConfig struct {
	MaxRequests   uint32        `yaml:"max_requests"`
	Interval      time.Duration `yaml:"interval"`
	Timeout       time.Duration `yaml:"timeout"`
	FailThreshold uint32        `yaml:"fail_threshold"`
}

// Load загружает конфигурацию с приоритетом: ENV > YAML > defaults
func Load() *Config {
	cfg := defaults()
	loadFromYAML(&cfg)
	applyEnv(&cfg)
	return &cfg
}

func defaults() Config {
	return Config{
		Server: ServerConfig{
			Port: "8080",
		},
		MySQL: MySQLConfig{
			DSN:             "root:root@tcp(localhost:3306)/task_vault?parseTime=true",
			MaxOpenConns:    25,
			MaxIdleConns:    10,
			ConnMaxLifetime: 5 * time.Minute,
		},
		Redis: RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
		JWT: JWTConfig{
			Secret: "dev-secret-change-me",
			TTL:    24 * time.Hour,
		},
		RateLimit: RateLimitConfig{
			RequestsPerWindow: 100,
			Window:            time.Minute,
		},
		Cache: CacheConfig{
			TasksTTL: 5 * time.Minute,
		},
		CircuitBreaker: CircuitBreakerConfig{
			MaxRequests:   3,
			Interval:      30 * time.Second,
			Timeout:       10 * time.Second,
			FailThreshold: 5,
		},
	}
}

func loadFromYAML(cfg *Config) {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "config.yaml"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		slog.Warn("ошибка парсинга конфигурации YAML", "path", path, "error", err)
	}
}

func applyEnv(cfg *Config) {
	envString("SERVER_PORT", &cfg.Server.Port)
	envString("MYSQL_DSN", &cfg.MySQL.DSN)
	envInt("MYSQL_MAX_OPEN_CONNS", &cfg.MySQL.MaxOpenConns)
	envInt("MYSQL_MAX_IDLE_CONNS", &cfg.MySQL.MaxIdleConns)
	envDuration("MYSQL_CONN_MAX_LIFETIME", &cfg.MySQL.ConnMaxLifetime)
	envString("REDIS_ADDR", &cfg.Redis.Addr)
	envString("REDIS_PASSWORD", &cfg.Redis.Password)
	envInt("REDIS_DB", &cfg.Redis.DB)
	envString("JWT_SECRET", &cfg.JWT.Secret)
	envDuration("JWT_TTL", &cfg.JWT.TTL)
	envInt64("RATE_LIMIT_REQUESTS", &cfg.RateLimit.RequestsPerWindow)
	envDuration("RATE_LIMIT_WINDOW", &cfg.RateLimit.Window)
	envDuration("CACHE_TASKS_TTL", &cfg.Cache.TasksTTL)
	envUint32("CB_MAX_REQUESTS", &cfg.CircuitBreaker.MaxRequests)
	envDuration("CB_INTERVAL", &cfg.CircuitBreaker.Interval)
	envDuration("CB_TIMEOUT", &cfg.CircuitBreaker.Timeout)
	envUint32("CB_FAIL_THRESHOLD", &cfg.CircuitBreaker.FailThreshold)
}

func envString(key string, dest *string) {
	if v := os.Getenv(key); v != "" {
		*dest = v
	}
}

func envInt(key string, dest *int) {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			*dest = n
		}
	}
}

func envInt64(key string, dest *int64) {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			*dest = n
		}
	}
}

func envUint32(key string, dest *uint32) {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseUint(v, 10, 32); err == nil {
			*dest = uint32(n)
		}
	}
}

func envDuration(key string, dest *time.Duration) {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			*dest = d
		}
	}
}
