package main

import (
	"log"
	"net/http"
	"online-shop/internal/db"
	"online-shop/internal/handlers"
	"online-shop/internal/repository"
	"online-shop/internal/services"

	"github.com/go-chi/chi/v5"
)

func main() {
	db.InitPsqlDB()
	db.InitMinio()

	productRepo := repository.NewProductRepo(db.PsqlDB, db.MinioClient)
	searchService := services.NewSearchService(productRepo)
	searchHandler := handlers.NewSearchHandler(searchService)

	r := chi.NewRouter()

	searchHandler.SetupRoutes(r)

	log.Println("Сервер запущен на порту 8080")
	log.Fatal(http.ListenAndServe(":8080", r))

}
