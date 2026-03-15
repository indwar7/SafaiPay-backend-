package database

import (
	"context"
	"log/slog"

	"github.com/redis/go-redis/v9"
	"github.com/indwar7/safaipay-backend/config"
)

func NewRedis(cfg *config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	slog.Info("connected to Redis", "addr", cfg.Addr)
	return client, nil
}
