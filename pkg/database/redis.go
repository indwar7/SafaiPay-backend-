package database

import (
	"context"
	"log/slog"
	"strings"

	"github.com/redis/go-redis/v9"
	"github.com/indwar7/safaipay-backend/config"
)

func NewRedis(cfg *config.RedisConfig) (*redis.Client, error) {
	var client *redis.Client

	// Support Redis URL format (redis://user:pass@host:port) used by Railway
	if strings.HasPrefix(cfg.Addr, "redis://") || strings.HasPrefix(cfg.Addr, "rediss://") {
		opts, err := redis.ParseURL(cfg.Addr)
		if err != nil {
			return nil, err
		}
		client = redis.NewClient(opts)
	} else {
		client = redis.NewClient(&redis.Options{
			Addr:     cfg.Addr,
			Password: cfg.Password,
			DB:       cfg.DB,
		})
	}

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	slog.Info("connected to Redis", "addr", cfg.Addr)
	return client, nil
}
