package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type EsConfig struct {
	ElasticURL      string
	ElasticUser     string
	ElasticPassword string
	ElasticCluster  string
	ESJavaOpts      string
}

// загрузка конфига для эластика
func LoadEsConfig() *EsConfig {
	err := godotenv.Load("../environment/elastic.env")
	if err != nil {
		log.Fatal("Ошибка при загрузке конфига elastic")
	}

	return &EsConfig{
		ElasticURL:      os.Getenv("ELASTIC_URL"),
		ElasticUser:     os.Getenv("ELASTIC_USER"),
		ElasticPassword: os.Getenv("ELASTIC_PASSWORD"),
		ElasticCluster:  os.Getenv("ELASTIC_CLUSTER_NAME"),
		ESJavaOpts:      os.Getenv("ES_JAVA_OPTS"),
	}
}
