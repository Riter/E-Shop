package models

import "time"

type Comment struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	ProductID int64     `json:"product_id"`
	Content   string    `json:"content"`
	Rating    int       `json:"rating"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateCommentDTO struct {
	UserID    int64  `json:"user_id"`
	ProductID int64  `json:"product_id"`
	Content   string `json:"content"`
	Rating    int    `json:"rating"`
}

type UpdateCommentDTO struct {
	Content string `json:"content"`
	Rating  int    `json:"rating"`
}


type ProductRating struct {
	ProductID     int64   `json:"product_id"`
	AverageRating float64 `json:"average_rating"`
	ReviewCount   int64   `json:"review_count"`
}
