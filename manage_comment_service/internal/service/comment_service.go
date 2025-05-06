package service

import (
	"comments_service/internal/models"
	"comments_service/internal/repository"
)

type CommentService struct {
	repo *repository.CommentRepository
}

func NewCommentService(repo *repository.CommentRepository) *CommentService {
	return &CommentService{repo: repo}
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
