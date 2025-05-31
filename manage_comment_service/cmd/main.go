package main

import (
	"comments_service/internal/db"
	"comments_service/internal/handler"
	"comments_service/internal/repository"
	"comments_service/internal/service"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	// "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
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
			semconv.ServiceName("manage-comment-service"),
		)),
	)
	otel.SetTracerProvider(tp)

	return func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Printf("error shutting down tracer provider: %v", err)
		}
	}
}

func welcomeHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"service": "Comments Service",
		"version": "1.0.0",
		"endpoints": map[string]interface{}{
			"public": []string{
				"GET /comments/{id} - получить комментарий по ID",
				"GET /products/{productID}/comments - получить все комментарии продукта",
			},
			"protected": []string{
				"POST /comments - создать новый комментарий",
				"PUT /comments/{id} - обновить комментарий",
				"DELETE /comments/{id} - удалить комментарий",
			},
		},
		"status": "running",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	shutdownTracer := InitTracer()
    defer shutdownTracer()
	log.Println("tracing init starting")
	defer log.Println("tracing init done")
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Initialize database
	db, err := db.InitPsqlDB()
	if err != nil {
		log.Fatalf("ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()

	// Initialize dependencies
	commentRepo := repository.NewCommentRepository(db)
	commentService := service.NewCommentService(commentRepo, db)
	commentHandler := handler.NewCommentHandler(commentService)

	// Initialize router
	r := chi.NewRouter()
	r.Use(otelhttp.NewMiddleware("manage-comment-service"))

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)

	// Register welcome handler
	r.Get("/", welcomeHandler)

	// Register routes
	commentHandler.RegisterRoutes(r)

	// Create server
	server := &http.Server{
		Addr:    ":30333",
		Handler: r,
	}

	// Channel to listen for errors coming from the listener.
	serverErrors := make(chan error, 1)

	// Start the service listening for requests.
	go func() {
		log.Printf("Сервис комментариев запущен на порту 30333")
		serverErrors <- server.ListenAndServe()
	}()

	// Channel to listen for an interrupt or terminate signal from the OS.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		log.Fatalf("ошибка запуска сервера: %v", err)

	case sig := <-shutdown:
		log.Printf("получен сигнал %v, начинаем graceful shutdown", sig)
	}
}
