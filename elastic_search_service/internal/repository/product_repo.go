package repository

import (
	"database/sql"
	"log"
	"online-shop/internal/models"

	"github.com/minio/minio-go/v7"
)

type ProductRepo struct {
	PsqlDb *sql.DB
	MinIO  *minio.Client
}

func NewProductRepo(PsqlDb *sql.DB, MinIO *minio.Client) *ProductRepo {
	return &ProductRepo{PsqlDb: PsqlDb, MinIO: MinIO}
}

func (r *ProductRepo) GetProductsByName(name string) ([]models.Product, error) {
	rows, err := r.PsqlDb.Query(`
		SELECT id, name, description, price, category
		FROM products
		WHERE name ILIKE '%' || $1 || '%'
	`, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Category); err != nil {
			return nil, err
		}

		images, err := r.getProductImages(p.ID)
		if err != nil {
			return nil, err
		}
		p.Images = images

		products = append(products, p)
	}
	return products, nil
}

func (r *ProductRepo) getProductImages(productID int) ([]string, error) {
	rows, err := r.PsqlDb.Query("SELECT image_url FROM product_images WHERE product_id = $1", productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []string
	for rows.Next() {
		var imageURL string
		if err := rows.Scan(&imageURL); err != nil {
			return nil, err
		}
		images = append(images, imageURL)
	}
	return images, nil
}

func (r *ProductRepo) GetALLProducts() ([]models.Product, error) {
	/*
		Эта функция нужна для того чтобы получать все товары из БД (и все их текстовые характеристики),
		при этом она не возвращает ссылкин на картинки для товаров
	*/
	rows, err := r.PsqlDb.Query("SELECT id, name, description, price, category, created_at FROM products")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Category, &p.CreatedAt); err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	log.Printf("Загружено %d товаров из PostgreSQL", len(products))
	return products, nil
}
