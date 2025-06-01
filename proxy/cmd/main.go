package main

import (
	
	"log"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	

	"google.golang.org/grpc"

	
	"github.com/Riter/E-Shop/proxy/internal/config"
)

func main() {
    cfg := config.LoadConfig()

    
    conn, err := grpc.Dial(cfg.AuthGRPCAddress(), grpc.WithInsecure())
    if err != nil {
        log.Fatalf("failed to connect to gRPC auth service: %v", err)
    }
    defer conn.Close()
    

    

    

    http.HandleFunc("/proxy/", func(w http.ResponseWriter, r *http.Request) {
        
        originalPath := r.URL.Path
        trimmedProxyPath := strings.TrimPrefix(originalPath, "/proxy") 

        
        r.URL.Path = trimmedProxyPath 

        proxyURL, _, err := GetServiceURL(r, cfg) 
        if err != nil {
            http.Error(w, "Service not found", http.StatusBadGateway)
            return
        }

        
        
        
        
        
        

        
        
        

        
        
        
        
        
        
        

        
        proxy := httputil.NewSingleHostReverseProxy(proxyURL)
        proxy.Director = func(req *http.Request) {
            req.URL.Scheme = proxyURL.Scheme
            req.URL.Host = proxyURL.Host
            req.URL.Path = trimmedProxyPath 
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
                matchedPrefix = prefix 
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