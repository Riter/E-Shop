package service

import (
	"comments_service/internal/models"
	"comments_service/internal/repository"
	"database/sql"
	"fmt"
)

type CommentService struct {
	repo *repository.CommentRepository
	db   *sql.DB
}

func NewCommentService(repo *repository.CommentRepository, db *sql.DB) *CommentService {
	return &CommentService{repo: repo, db: db}
}

func (s *CommentService) CreateComment(comment models.CreateCommentDTO) (int64, error) {
	return s.repo.Create(comment)
}

func (s *CommentService) GetComment(id int64) (*models.Comment, error) {
	return s.repo.GetByID(id)
}

func (s *CommentService) GetProductComments(productID int64) ([]*models.Comment, error) {
	return s.repo.GetByProductID(productID)
}

func (s *CommentService) UpdateComment(id int64, comment models.UpdateCommentDTO) error {
	return s.repo.Update(id, comment)
}

func (s *CommentService) DeleteComment(id int64) error {
	return s.repo.Delete(id)
}

func (s *CommentService) GetProductRating(productID int64) (*models.ProductRating, error) {
	var rating models.ProductRating
	rating.ProductID = productID

	// Получаем средний рейтинг и количество отзывов
	err := s.db.QueryRow(`
		SELECT COALESCE(AVG(rating), 0), COUNT(*)
		FROM comments
		WHERE product_id = $1
	`, productID).Scan(&rating.AverageRating, &rating.ReviewCount)

	if err != nil {
		return nil, fmt.Errorf("failed to get product rating: %w", err)
	}

	return &rating, nil
}
