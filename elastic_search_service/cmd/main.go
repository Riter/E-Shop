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

	"github.com/go-chi/chi/v5"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Инициализация баз данных
	db.InitPsqlDB()
	db.InitMinio()

	// Инициализация Elasticsearch
	elasticClient, err := elasticsearch.NewESClient()
	if err != nil {
		log.Fatalf("ошибка при создании клиента elastic: %v", err)
	}

	// Инициализация репозитория и сервиса
	productRepo := repository.NewProductRepo(db.PsqlDB, db.MinioClient)
	elasticManager := services.NewElasticManager(productRepo, elasticClient)

	// Начальная синхронизация с PostgreSQL
	if err := elasticManager.SyncProductsToElasticSearch(); err != nil {
		log.Printf("Ошибка при начальной синхронизации с PostgreSQL: %v", err)
	}

	// Инициализация Kafka consumer
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

	// Запуск Kafka consumer в отдельной горутине
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := consumer.Start(ctx); err != nil {
			log.Printf("ошибка в Kafka consumer: %v", err)
		}
	}()

	// Настройка HTTP сервера
	r := chi.NewRouter()
	r.Get("/search", elasticManager.ServeHTTP)

	// Настройка graceful shutdown
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.LoadAPIConfig().APIPort),
		Handler: r,
	}

	// Канал для получения сигналов завершения
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Запуск HTTP сервера в отдельной горутине
	go func() {
		log.Printf("Сервер запущен на порту %d\n", config.LoadAPIConfig().APIPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ошибка запуска сервера: %v", err)
		}
	}()

	// Ожидание сигнала завершения
	<-stop
	log.Println("Получен сигнал завершения, начинаем graceful shutdown...")

	// Отмена контекста для Kafka consumer
	cancel()

	// Graceful shutdown HTTP сервера
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Printf("ошибка при graceful shutdown сервера: %v", err)
	}

	log.Println("Сервер успешно остановлен")
}
