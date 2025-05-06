package main

import (
	"comments_service/internal/db"
	"log"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	db, err := db.InitPsqlDB()
	if err != nil {
		log.Fatalf("ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()

	// TODO: Инициализация HTTP сервера и других компонентов
	log.Println("Сервис комментариев запущен")
}
