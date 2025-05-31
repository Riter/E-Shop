package config

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
    AuthGRPCHost string
    AuthGRPCPort string
    SearchServiceUrl   string
	ManageItemCrudUrl   string
	FacadeUrl   string
	ServiceRoutes map[string]string 

    ProxyPort    string
}

func LoadConfig() *Config {
    // Загружаем .env
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found, using system environment")
    }

    cfg := &Config{
        AuthGRPCHost: getEnv("AUTH_GRPC_HOST", "localhost"),
        AuthGRPCPort: getEnv("AUTH_GRPC_PORT", "44044"),
        ServiceRoutes: map[string]string{
            "/search": getEnv("SEARCH_SERVICE_URL", "http://localhost:8081"),
            "/items":  getEnv("MANAGE_ITEM_CRUD_URL", "http://localhost:8081"),
            "/products":   getEnv("FACADE_URL", "http://localhost:8081"),
        },

        SearchServiceUrl:   getEnv("SEARCH_SERVICE_URL", "http://localhost:8081"),
		ManageItemCrudUrl:   getEnv("MANAGE_ITEM_CRUD_URL", "http://localhost:8081"),
		FacadeUrl:   getEnv("FACADE_URL", "http://localhost:8081"),

        ProxyPort:    getEnv("PROXY_PORT", "8000"),
    }

    return cfg
}

func (c *Config) AuthGRPCAddress() string {
    return fmt.Sprintf("%s:%s", c.AuthGRPCHost, c.AuthGRPCPort)
}

func getEnv(key, defaultValue string) string {
    if value, exists := os.LookupEnv(key); exists {
        return value
    }
	slog.Error(fmt.Sprintf("Can't find env %s", key))
    return defaultValue
}
