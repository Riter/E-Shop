package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"online-shop/config"
	"online-shop/internal/db"
	"online-shop/internal/elasticsearch"
	"online-shop/internal/kafka"
	"online-shop/internal/repository"
	"online-shop/internal/services"
	"os"
	"os/signal"
	"syscall"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

	"github.com/go-chi/chi/v5"
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
			semconv.ServiceName("elastic-search-service"),
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
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	shutdown := InitTracer()
    defer shutdown()
	log.Println("tracing init starting")
	defer log.Println("tracing init done")

	
	db.InitPsqlDB()
	db.InitMinio()

	
	elasticClient, err := elasticsearch.NewESClient()
	if err != nil {
		log.Fatalf("ошибка при создании клиента elastic: %v", err)
	}

	
	productRepo := repository.NewProductRepo(db.PsqlDB, db.MinioClient)
	elasticManager := services.NewElasticManager(productRepo, elasticClient)

	
	if err := elasticManager.SyncProductsToElasticSearch(); err != nil {
		log.Printf("Ошибка при начальной синхронизации с PostgreSQL: %v", err)
	}

	
	kafkaConfig := config.LoadKafkaConfig()
	consumer, err := kafka.NewConsumer(
		kafkaConfig.Brokers,
		kafkaConfig.GroupID,
		kafkaConfig.Topic,
		elasticClient,
	)
	if err != nil {
		log.Fatalf("ошибка при создании Kafka consumer: %v", err)
	}

	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := consumer.Start(ctx); err != nil {
			log.Printf("ошибка в Kafka consumer: %v", err)
		}
	}()

	
	r := chi.NewRouter()
	r.Use(otelhttp.NewMiddleware("elastic-search-service"))

	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})
	r.Get("/search", elasticManager.ServeHTTP)

	
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.LoadAPIConfig().APIPort),
		Handler: r,
	}

	
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	
	go func() {
		log.Printf("Сервер запущен на порту %d\n", config.LoadAPIConfig().APIPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ошибка запуска сервера: %v", err)
		}
	}()

	
	<-stop
	log.Println("Получен сигнал завершения, начинаем graceful shutdown...")

	
	cancel()

	
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Printf("ошибка при graceful shutdown сервера: %v", err)
	}

	log.Println("Сервер успешно остановлен")
}
