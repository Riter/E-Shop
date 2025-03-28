package main

import (
	"log"
	"net/http"

	"online-shop/internal/db"
	"online-shop/internal/elasticsearch"

	"online-shop/internal/repository"
	"online-shop/internal/services"

	"github.com/go-chi/chi/v5"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	db.InitPsqlDB()
	db.InitMinio()
	elascticClient, err := elasticsearch.NewESClient()
	if err != nil {
		log.Fatal("ошибка при создании клиента elastic: %w", err)
	}

	productRepo := repository.NewProductRepo(db.PsqlDB, db.MinioClient)
	elasticManager := services.NewElasticManager(productRepo, elascticClient)
	elasticManager.EnablePeriodicSync()

	r := chi.NewRouter()
	r.Get("/search", elasticManager.ServeHTTP)

	log.Println("Сервер запущен на порту 8080")
	log.Fatal(http.ListenAndServe(":8080", r))

}
