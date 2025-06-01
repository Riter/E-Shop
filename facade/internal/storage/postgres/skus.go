package psq

import (
	"context"
	"encoding/json"

	"github.com/Riter/E-Shop/internal/models"
	pq "github.com/lib/pq"
)

func (p *Postgres) GetProductsByIDs(ctx context.Context, skus []int64) ([]models.ProductResponse, error) {
    query := `
        SELECT
            p.id,
            p.name,
            p.description,
            p.price,
            p.category,
            p.created_at,
            COALESCE(json_agg(pi.image_url) FILTER (WHERE pi.image_url IS NOT NULL), '[]') AS images
        FROM products p
        LEFT JOIN product_images pi ON p.id = pi.product_id
        WHERE p.id = ANY($1)
        GROUP BY p.id;
    `

    rows, err := p.DB.QueryContext(ctx, query, pq.Array(skus))
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var products []models.ProductResponse

    for rows.Next() {
        var p models.ProductResponse
        var imagesRaw []byte

        err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Category, &p.CreatedAt, &imagesRaw)
        if err != nil {
            return nil, err
        }

        
        if err := json.Unmarshal(imagesRaw, &p.Images); err != nil {
            return nil, err
        }

        products = append(products, p)
    }

    return products, nil
}
