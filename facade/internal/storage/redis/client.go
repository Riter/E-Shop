package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Riter/E-Shop/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

func init() {
  prometheus.MustRegister(
	KeysToGetMetric, SetValuesCount, GotValuesByKey,
	RedisGetErrors, RedisSetErrors, RedisGetDuration, RedisSetDuration,
	RedisGetPayloadSize, RedisSetPipelineLength, RedisCacheMisses,
  )
}

var (
  KeysToGetMetric = prometheus.NewCounter(prometheus.CounterOpts{
    Name: "keys_to_get_metric",
    Help: "requested keys to redis",
  })

  GotValuesByKey = prometheus.NewCounter(prometheus.CounterOpts{
    Name: "got_values_by_key",
    Help: "got keys from redis",
  })

  SetValuesCount = prometheus.NewCounter(prometheus.CounterOpts{
    Name: "set_values_count",
    Help: "count of set values to redis",
  })

	RedisGetErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "redis_get_errors_total",
		Help: "total number of Redis GET errors",
	})

	RedisSetErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "redis_set_errors_total",
		Help: "total number of Redis SET errors",
	})

	RedisGetDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "redis_get_duration_seconds",
		Help:    "duration of Redis GET operations",
		Buckets: prometheus.DefBuckets, // можно задать свои
	})

	RedisSetDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "redis_set_duration_seconds",
		Help:    "duration of Redis SET operations",
		Buckets: prometheus.DefBuckets,
	})

	RedisGetPayloadSize = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "redis_get_payload_size_bytes",
		Help:    "total size of values returned from Redis",
		Buckets: prometheus.ExponentialBuckets(100, 2, 10), // от 100 до ~50К
	})

	RedisSetPipelineLength = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "redis_set_pipeline_length",
		Help:    "number of commands in Redis SET pipeline",
		Buckets: prometheus.LinearBuckets(0, 10, 10),
	})

	RedisCacheMisses = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "redis_cache_misses_total",
		Help: "number of Redis cache misses",
	})

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
	KeysToGetMetric.Add(float64(len(keys)))
	start := time.Now()
	result, err := r.storage.MGet(ctx, keys...).Result()
	RedisGetDuration.Observe(time.Since(start).Seconds())
	for _, v := range result {
		if v == nil {
			RedisCacheMisses.Inc()
		}
	}

	if err != nil {
    	RedisGetErrors.Inc()
	}
	var totalBytes float64
	for _, v := range result {
		if str, ok := v.(string); ok {
			totalBytes += float64(len(str))
		}
	}
	RedisGetPayloadSize.Observe(totalBytes)

	GotValuesByKey.Add(float64(len(result)))
	return result, err
}

func (r *RedisImpl) Set(ctx context.Context, mset map[string]string, expiration time.Duration) {
	if len(mset) > 0 {
		RedisSetPipelineLength.Observe(float64(len(mset)))
		pipe := r.storage.Pipeline()
		for k, v := range mset {
			pipe.Set(ctx, k, v, 5*time.Minute)
		}
		res, err := pipe.Exec(ctx)
		SetValuesCount.Add(float64(len(mset)))
		if err!=nil{
			log.Printf("error set to Redis: %s. %s", err.Error(), res)
			RedisSetErrors.Inc()
		}
	}
}