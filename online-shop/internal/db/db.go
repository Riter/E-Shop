package db

import (
	"database/sql"
	"fmt"
	"log"
	"online-shop/config"
	"time"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	cfg := config.LoadConfig()

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)
	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}

	DB.SetMaxOpenConns(cfg.DBMaxOpenConns)
	DB.SetMaxIdleConns(cfg.DBMaxIdleConns)
	DB.SetConnMaxLifetime(time.Duration(cfg.DBMaxConnLifeTime) * time.Minute)

	err = DB.Ping()
	if err != nil {
		log.Fatal("Не удалось подключиться к БД:", err)
	}

	log.Println("Подключение к БД успешно создано")

}
