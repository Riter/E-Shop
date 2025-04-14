package db

import (
	"context"
	"log"
	"online-shop/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var MinioClient *minio.Client

// данная функция создает подключение к minio базе данных, и создает бакет для картинок в ней
func InitMinio() {
	cfg := config.LoadConfig()

	client, err := minio.New(cfg.DBMinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.DBMinioRootUser, cfg.DBMinioRootPassw, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatal("Ошибка подключения к MinIO", err)
	}

	MinioClient = client
	log.Println("Подключение к Minio успешно создано")

	ctx := context.Background()
	exists, err := MinioClient.BucketExists(ctx, cfg.DBMinioBucket)
	if err != nil {
		log.Fatal("Ошибка при проверке бакета MinIO", err)
	}
	if !exists {
		err = MinioClient.MakeBucket(ctx, cfg.DBMinioBucket, minio.MakeBucketOptions{})
		if err != nil {
			log.Fatal("Не удалось создать бакет MinIO:", err)
		}
		log.Println("Бакет Minio создан:", cfg.DBMinioBucket)
	} else {
		log.Println("Бакет MinIO уже существует:", cfg.DBMinioBucket)
	}
}
