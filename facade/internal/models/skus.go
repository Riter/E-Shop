package models

import "time"

type ProductResponse struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Category    string    `json:"category"`
	CreatedAt   time.Time `json:"created_at"`
	Images      []string  `json:"images"`
}


type ProductResponseList struct {
	ProductList []ProductResponse `json:"product_list"`
}


type ProductRequest struct {
	SKUs []int64
}