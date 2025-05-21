package redis

import (
	"fmt"

	"github.com/Riter/E-Shop/internal/config"
	"github.com/redis/go-redis/v9"
)


func NewRedisClient(cfg *config.RedisConfig) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	return rdb
}