package logging

import (
	"context"
	"log/slog"
	"task_vault/internal/ports"
	"time"
)

type CacheLogger struct {
	inner  ports.Cache
	logger *slog.Logger
}

func NewCacheLogger(inner ports.Cache, logger *slog.Logger) *CacheLogger {
	return &CacheLogger{inner: inner, logger: logger}
}

func (c *CacheLogger) Get(ctx context.Context, key string, dest any) error {
	err := c.inner.Get(ctx, key, dest)
	if err != nil {
		c.logger.Debug("Cache.Get", "key", key, "error", err)
	}
	return err
}

func (c *CacheLogger) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	err := c.inner.Set(ctx, key, value, ttl)
	if err != nil {
		c.logger.Error("Cache.Set", "key", key, "error", err)
	}
	return err
}

func (c *CacheLogger) Delete(ctx context.Context, key string) error {
	err := c.inner.Delete(ctx, key)
	if err != nil {
		c.logger.Error("Cache.Delete", "key", key, "error", err)
	}
	return err
}

func (c *CacheLogger) DeleteByPrefix(ctx context.Context, prefix string) error {
	err := c.inner.DeleteByPrefix(ctx, prefix)
	if err != nil {
		c.logger.Error("Cache.DeleteByPrefix", "prefix", prefix, "error", err)
	}
	return err
}
