package services

import (
	"online-shop/internal/models"
	"online-shop/internal/repository"
)

type SearchService struct {
	Repo *repository.ProductRepo
}

func NewSearchService(repo *repository.ProductRepo) *SearchService {
	return &SearchService{Repo: repo}
}

func (s *SearchService) SearchProducts(query string) ([]models.Product, error) {
	return s.Repo.GetProductsByName(query)
}
