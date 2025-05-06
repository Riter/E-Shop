package repository

import (
	"database/sql"
	"errors"
	"time"

	"comments_service/internal/models"
)

type CommentRepository struct {
	db *sql.DB
}

func NewCommentRepository(db *sql.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) Create(comment models.CreateCommentDTO) (int64, error) {
	query := `
		INSERT INTO comments (user_id, product_id, content, rating, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	var id int64
	now := time.Now()
	err := r.db.QueryRow(
		query,
		comment.UserID,
		comment.ProductID,
		comment.Content,
		comment.Rating,
		now,
		now,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *CommentRepository) GetByID(id int64) (*models.Comment, error) {
	query := `
		SELECT id, user_id, product_id, content, rating, created_at, updated_at
		FROM comments
		WHERE id = $1`

	comment := &models.Comment{}
	err := r.db.QueryRow(query, id).Scan(
		&comment.ID,
		&comment.UserID,
		&comment.ProductID,
		&comment.Content,
		&comment.Rating,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return comment, nil
}

func (r *CommentRepository) GetByProductID(productID int64) ([]*models.Comment, error) {
	query := `
		SELECT id, user_id, product_id, content, rating, created_at, updated_at
		FROM comments
		WHERE product_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*models.Comment
	for rows.Next() {
		comment := &models.Comment{}
		err := rows.Scan(
			&comment.ID,
			&comment.UserID,
			&comment.ProductID,
			&comment.Content,
			&comment.Rating,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

func (r *CommentRepository) Update(id int64, comment models.UpdateCommentDTO) error {
	query := `
		UPDATE comments
		SET content = $1, rating = $2, updated_at = $3
		WHERE id = $4`

	result, err := r.db.Exec(query, comment.Content, comment.Rating, time.Now(), id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("comment not found")
	}

	return nil
}

func (r *CommentRepository) Delete(id int64) error {
	query := `DELETE FROM comments WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("comment not found")
	}

	return nil
}
