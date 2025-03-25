package elasticsearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"online-shop/config"
	"strconv"

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
		res, err := es.Client.Index("products", req, es.Client.Index.WithDocumentID(strconv.Itoa(product.ID)))
		if err != nil {
			return fmt.Errorf("ошибка при индексации продукта: %w", err)
		}
		defer res.Body.Close()

		log.Printf("Товар %s добавлен/обновлен в Elasticsearch", product.Name)
	}
	return nil
}

func (es *ESClient) SearchProducts(query string) ([]models.Product, error) {
	searchBody := map[string]interface{}{
		"query": map[string]interface{}{
			"multy_match": map[string]interface{}{
				"query":  query,
				"fields": []string{"name", "description"}, //поля по которым идет поиск
			},
		},
	}

	body, err := json.Marshal(searchBody)
	if err != nil {
		return nil, fmt.Errorf("ошибка сериализации поискового запроса: %w", err)
	}

	res, err := es.Client.Search(
		es.Client.Search.WithIndex("products"),
		es.Client.Search.WithBody(bytes.NewReader(body)),
		es.Client.Search.WithPretty(),
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения поискового запроса: %w", err)
	}
	defer res.Body.Close()

	var searchResult struct {
		Hits struct {
			Hits []struct {
				Source models.Product `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа от Elasticsearch: %w", err)
	}

	var products []models.Product
	for _, hit := range searchResult.Hits.Hits {
		products = append(products, hit.Source)
	}

	return products, nil
}
