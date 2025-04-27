package redis_client

import (
	"context"

	"github.com/redis/go-redis/v9"
)


type RedisClient interface {
    Get(key string) (string, error)
    Set(key string, value interface{}) error
	Ping(ctx context.Context) (error)
}


type redisClient struct {
    client *redis.Client
    ctx    context.Context
}

// NewRedisClient создает новый клиент Redis
func NewRedisClient(client *redis.Client) RedisClient {
    return &redisClient{
        client: client,
        ctx:    context.Background(), // можно сюда передавать ctx извне, если нужно
    }
}

// Set устанавливает значение по ключу
func (r *redisClient) Set(key string, value interface{}) error {
    return r.client.Set(r.ctx, key, value, 0).Err()
}

// Get получает значение по ключу
func (r *redisClient) Get(key string) (string, error) {
    return r.client.Get(r.ctx, key).Result()
}

func (r *redisClient) Ping(ctx context.Context) (error) {
    return r.client.Ping(ctx).Err()
}