package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/Riter/E-Shop/internal/config"
	"github.com/Riter/E-Shop/internal/handlers"
	psq "github.com/Riter/E-Shop/internal/storage/postgres"
	"github.com/Riter/E-Shop/internal/storage/redis"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
    "go.opentelemetry.io/otel"

    "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
    "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

    
    semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
)

func InitTracer() func() {
	ctx := context.Background()


	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint("jaeger:4317"),
	)

	if err != nil {
		log.Fatalf("failed to create exporter: %v", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("facade-service"),
		)),
	)
	otel.SetTracerProvider(tp)

	return func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Printf("error shutting down tracer provider: %v", err)
		}
	}
}



func main() {
	
	redisCfg := config.LoadRedisConfig()
    shutdown := InitTracer()
    defer shutdown()
	log.Println("tracing init starting")
	defer log.Println("tracing init done")

	
	rdb := redis.NewRedisClient(redisCfg)
	defer rdb.Close()

	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Successfully connected to Redis")

    cfg := config.LoadPostgresConfigFromEnv()
    dbClient, err := psq.NewPostgres(ctx, cfg)
    if err != nil {
        log.Fatalf("Ошибка подключения к БД: %v", err)
    }

	r := chi.NewRouter()
	r.Use(otelhttp.NewMiddleware("facade-service"))

	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})


	go func() {
		http.Handle("/metrics", promhttp.Handler())
			if err := http.ListenAndServe(":10671", nil); err != nil {
				slog.Error("failed to start metrics server", slog.Any("err", err))
			}
  	}()


	r.Get("/products", handlers.GetProducts(ctx, dbClient, rdb))

	
	log.Println("Listening on :8089")
	if err := http.ListenAndServe(":8089", r); err != nil {
		log.Fatalf("server failed: %v", err)
	}
	log.Println("Server stopped")


}