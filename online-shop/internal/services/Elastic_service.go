package services

import (
	"encoding/json"
	"log"
	"net/http"
	"online-shop/internal/elasticsearch"
	"online-shop/internal/repository"
)

type ElasticManager struct {
	Repo    *repository.ProductRepo
	Elastic *elasticsearch.ESClient
}

func NewElasticManager(repo *repository.ProductRepo, elastic *elasticsearch.ESClient) *ElasticManager {
	return &ElasticManager{Repo: repo, Elastic: elastic}
}

func (s *ElasticManager) SyncProductsToElasticSearch() error {
	products, err := s.Repo.GetALLProducts()
	if err != nil {
		log.Println("ошибка из функции получения товаров: %w", err)
		return err
	}

	err = s.Elastic.IndexProducts(products)
	if err != nil {
		log.Println("ошибка из функции индексации продуктов: %w", err)
		return err
	}

	log.Println("Все товары успешно загружены в Elasticsearch")
	return nil
}

func (s *ElasticManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "параметр 'q' обязателен", http.StatusBadRequest)
		return
	}

	products, err := s.Elastic.SearchProducts(query)
	if err != nil {
		log.Println("ошибка поиска товаров %w", err)
		http.Error(w, "ошибка поиска товаров", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}
