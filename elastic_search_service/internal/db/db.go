package db

import (
	"database/sql"
	"fmt"
	"log"
	"online-shop/config"
	"time"

	_ "github.com/lib/pq"
)

var PsqlDB *sql.DB

// данная функция создает connection pool для sql базы данных, используя конфиг из env файла
func InitPsqlDB() {
	cfg := config.LoadConfig()

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)
	var err error
	PsqlDB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}

	PsqlDB.SetMaxOpenConns(cfg.DBMaxOpenConns)
	PsqlDB.SetMaxIdleConns(cfg.DBMaxIdleConns)
	PsqlDB.SetConnMaxLifetime(time.Duration(cfg.DBMaxConnLifeTime) * time.Minute)

	err = PsqlDB.Ping()
	if err != nil {
		log.Fatal("Не удалось подключиться к БД:", err)
	}

	log.Println("Подключение к БД успешно создано")

}
