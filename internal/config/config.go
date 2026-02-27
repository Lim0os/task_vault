package config

import (
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	MySQL  MySQLConfig  `yaml:"mysql"`
	Redis  RedisConfig  `yaml:"redis"`
	JWT    JWTConfig    `yaml:"jwt"`
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

	_ = yaml.Unmarshal(data, cfg)
}

func applyEnv(cfg *Config) {
	if v := os.Getenv("SERVER_PORT"); v != "" {
		cfg.Server.Port = v
	}
	if v := os.Getenv("MYSQL_DSN"); v != "" {
		cfg.MySQL.DSN = v
	}
	if v := os.Getenv("MYSQL_MAX_OPEN_CONNS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.MySQL.MaxOpenConns = n
		}
	}
	if v := os.Getenv("MYSQL_MAX_IDLE_CONNS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.MySQL.MaxIdleConns = n
		}
	}
	if v := os.Getenv("MYSQL_CONN_MAX_LIFETIME"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.MySQL.ConnMaxLifetime = d
		}
	}
	if v := os.Getenv("REDIS_ADDR"); v != "" {
		cfg.Redis.Addr = v
	}
	if v := os.Getenv("REDIS_PASSWORD"); v != "" {
		cfg.Redis.Password = v
	}
	if v := os.Getenv("REDIS_DB"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Redis.DB = n
		}
	}
	if v := os.Getenv("JWT_SECRET"); v != "" {
		cfg.JWT.Secret = v
	}
	if v := os.Getenv("JWT_TTL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.JWT.TTL = d
		}
	}
}
