package services

import (
	"log"
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
		return err
	}

	err = s.Elastic.IndexProducts(products)
	if err != nil {
		return err
	}

	log.Println("Все товары успешно загружены в Elasticsearch")
	return nil
}
