package services

import (
	"log"
	"online-shop/internal/elasticsearch"
	"online-shop/internal/repository"
)

type IndexService struct {
	Repo    *repository.ProductRepo
	Elastic *elasticsearch.ESClient
}

func NewIndexService(repo *repository.ProductRepo, elastic *elasticsearch.ESClient) *IndexService {
	return &IndexService{Repo: repo, Elastic: elastic}
}

func (s *IndexService) SyncProductsToElasticSearch() error {
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
