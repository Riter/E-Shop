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
	log.Println("Подключение к базе данных успешно создано")
}
