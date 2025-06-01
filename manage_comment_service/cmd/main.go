package main

import (
	"comments_service/internal/db"
	"comments_service/internal/handler"
	"comments_service/internal/repository"
	"comments_service/internal/service"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

)

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
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	
	db, err := db.InitPsqlDB()
	if err != nil {
		log.Fatalf("ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()

	
	commentRepo := repository.NewCommentRepository(db)
	commentService := service.NewCommentService(commentRepo, db)
	commentHandler := handler.NewCommentHandler(commentService)

	
	r := chi.NewRouter()

	
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)

	
	r.Get("/", welcomeHandler)

	
	commentHandler.RegisterRoutes(r)

	
	server := &http.Server{
		Addr:    ":30333",
		Handler: r,
	}

	
	serverErrors := make(chan error, 1)

	
	go func() {
		log.Printf("Сервис комментариев запущен на порту 30333")
		serverErrors <- server.ListenAndServe()
	}()

	
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	
	select {
	case err := <-serverErrors:
		log.Fatalf("ошибка запуска сервера: %v", err)

	case sig := <-shutdown:
		log.Printf("получен сигнал %v, начинаем graceful shutdown", sig)
	}
}
