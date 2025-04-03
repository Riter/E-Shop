package main

import (
	"fmt"
	"log"
	"net/http"
	"online-shop/config"

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
		log.Fatalf("ошибка при создании клиента elastic: %v", err)
	}

	productRepo := repository.NewProductRepo(db.PsqlDB, db.MinioClient)
	elasticManager := services.NewElasticManager(productRepo, elascticClient)
	elasticManager.EnablePeriodicSync(15)

	r := chi.NewRouter()
	r.Get("/search", elasticManager.ServeHTTP)

	api_port := config.LoadAPIConfig().APIPort
	log.Printf("Сервер запущен на порту %d\n", api_port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", api_port), r)) // Добавлены закрывающие скобки

}
