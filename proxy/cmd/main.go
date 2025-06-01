package main

import (
	// "context"
	"log"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	// "time"

	"google.golang.org/grpc"

	// pb "github.com/GGiovanni9152/protos/gen/go/sso" // замените на свой путь
	"github.com/Riter/E-Shop/proxy/internal/config"
)

func main() {
    cfg := config.LoadConfig()

    // Подключение к gRPC авторизации
    conn, err := grpc.Dial(cfg.AuthGRPCAddress(), grpc.WithInsecure())
    if err != nil {
        log.Fatalf("failed to connect to gRPC auth service: %v", err)
    }
    defer conn.Close()
    // authClient := pb.NewAuthClient(conn)

    // Парсим URL бэкенда

    // HTTP обработчик

    http.HandleFunc("/proxy/", func(w http.ResponseWriter, r *http.Request) {
        // 1. Удаляем только /proxy из пути
        originalPath := r.URL.Path
        trimmedProxyPath := strings.TrimPrefix(originalPath, "/proxy") // /search...

        // 2. Получаем целевой сервис по trimmedProxyPath
        r.URL.Path = trimmedProxyPath // например: "/search"

        proxyURL, _, err := GetServiceURL(r, cfg) // matchedPrefix = "/search"
        if err != nil {
            http.Error(w, "Service not found", http.StatusBadGateway)
            return
        }

        // 3. Проверка JWT
        // cookie, err := r.Cookie("jwt")
        // if err != nil {
        //     http.Error(w, "Unauthorized: no jwt cookie", http.StatusUnauthorized)
        //     return
        // }

        // 4. gRPC проверка токена
        // ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        // defer cancel()

        // resp, err := authClient.ValidateToken(ctx, &pb.ValidateTokenRequest{
        //     Token: cookie.Value,
        // })
        // if err != nil || !resp.IsValid {
        //     http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
        //     return
        // }

        // 5. Настраиваем прокси
        proxy := httputil.NewSingleHostReverseProxy(proxyURL)
        proxy.Director = func(req *http.Request) {
            req.URL.Scheme = proxyURL.Scheme
            req.URL.Host = proxyURL.Host
            req.URL.Path = trimmedProxyPath // передаем как есть: /search
            req.URL.RawQuery = r.URL.RawQuery
            req.Header = r.Header
        }

        proxy.ServeHTTP(w, r)
    })

    
    log.Printf("Proxy server started on port :%s", cfg.ProxyPort)
    if err := http.ListenAndServe(":"+cfg.ProxyPort, nil); err != nil {
        log.Fatalf("server failed: %v", err)
    }
}




func GetServiceURL(r *http.Request, cfg *config.Config) (*url.URL, string, error) {
    path := r.URL.Path

    var matchedPrefix string
    for prefix := range cfg.ServiceRoutes {
        if strings.HasPrefix(path, prefix) {
            if len(prefix) > len(matchedPrefix) {
                matchedPrefix = prefix // выбираем самый длинный (наиболее конкретный) префикс
            }
        }
    }

    if matchedPrefix == "" {
		slog.Error("Can't find matchedPrefix")
        return nil, "", http.ErrAbortHandler
    }

    serviceBase := cfg.ServiceRoutes[matchedPrefix]
    parsedURL, err := url.Parse(serviceBase)
    return parsedURL, matchedPrefix, err
}