package redisclient

import (
    "context"
    "log"

    "github.com/go-redis/redis/v8"
)

func NewClient(addr, password string, db int) *redis.Client {
    rdb := redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: password,
        DB:       db,
    })

    ctx := context.Background()
    if err := rdb.Ping(ctx).Err(); err != nil {
        log.Fatalf("Failed to connect to Redis: %v", err)
    }

    log.Println("Connected to Redis")
    return rdb
}
