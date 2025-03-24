package main

import (
	"log"

	//"net/http"
	"online-shop/internal/db"
	"online-shop/internal/elasticsearch"
	"online-shop/internal/handlers"
	"online-shop/internal/repository"
	"online-shop/internal/services"

	"github.com/go-chi/chi/v5"
)

func main() {
	db.InitPsqlDB()
	db.InitMinio()
	elscticClient, err := elasticsearch.NewESClient("http://localhost:9200")
	if err != nil {
		log.Fatal("ошибка при создании клиента elastic: %w", err)
	}

	productRepo := repository.NewProductRepo(db.PsqlDB, db.MinioClient)
	searchService := services.NewSearchService(productRepo)
	searchHandler := handlers.NewSearchHandler(searchService)

	indexServese := services.NewElasticManager(productRepo, elscticClient)
	indexServese.SyncProductsToElasticSearch()

	r := chi.NewRouter()

	searchHandler.SetupRoutes(r)

	// log.Println("Сервер запущен на порту 8080")
	// log.Fatal(http.ListenAndServe(":8080", r))

}
