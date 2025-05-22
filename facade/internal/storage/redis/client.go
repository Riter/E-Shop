package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Riter/E-Shop/internal/config"
	"github.com/redis/go-redis/v9"
)


type RedisImpl struct{
	storage *redis.Client
}


func NewRedisClient(cfg *config.RedisConfig) *RedisImpl {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	return &RedisImpl{storage: rdb}
}

func (r *RedisImpl) Ping(ctx context.Context) (string, error) {
	return r.storage.Ping(ctx).Result()
}

func (r *RedisImpl) Close() {
	r.storage.Close()
}

func  (r *RedisImpl) Get(ctx context.Context, keys ...string) ([]interface{}, error) {
	return r.storage.MGet(ctx, keys...).Result()
}

func (r *RedisImpl) Set(ctx context.Context, mset map[string]string, expiration time.Duration) {
	if len(mset) > 0 {
		pipe := r.storage.Pipeline()
		for k, v := range mset {
			pipe.Set(ctx, k, v, 5*time.Minute)
		}
		res, err := pipe.Exec(ctx) // можно логировать ошибку
		if err!=nil{
			log.Printf("error set to Redis: %s. %s", err.Error(), res)
		}
	}
}