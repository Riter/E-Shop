package elasticsearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"online-shop/config"

	"online-shop/internal/models"

	"github.com/elastic/go-elasticsearch/v8"
)

type ESClient struct {
	Client *elasticsearch.Client
}

func NewESClient() (*ESClient, error) {
	input_cfg := config.LoadEsConfig()
	cfg := elasticsearch.Config{
		Addresses: []string{input_cfg.ElasticURL},
		Username:  input_cfg.ElasticUser,
		Password:  input_cfg.ElasticPassword,
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("ошибка во время создания клиента elastic: %w", err)
	}

	res, err := client.Info()
	if err != nil {
		return nil, fmt.Errorf("ошибка во время получения информации от elastic: %w", err)
	}
	log.Println(res)

	return &ESClient{Client: client}, nil
}

func (es *ESClient) IndexProducts(products []models.Product) error {
	for _, product := range products {
		data, err := json.Marshal(product)
		if err != nil {
			return fmt.Errorf("ошибка сериализации продукта: %w", err)
		}

		req := bytes.NewReader(data)
		res, err := es.Client.Index("products", req)
		if err != nil {
			return fmt.Errorf("ошибка при индексации продукта: %w", err)
		}
		defer res.Body.Close()

		log.Printf("Товар %s добавлен в Elasticsearch", product.Name)
	}
	return nil
}
