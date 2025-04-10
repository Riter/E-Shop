package elasticsearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"online-shop/config"
	"online-shop/internal/models"
	"strconv"

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
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{
						"multi_match": map[string]interface{}{
							"query":     query,
							"fields":    []string{"name^3", "description^2", "category"},
							"fuzziness": "AUTO", // Позволяет Elastic прощать опечатки
						},
					},
					// Поиск по префиксу (autocomplete) — находит товары по началу слова
					{
						"prefix": map[string]interface{}{
							"name": map[string]interface{}{
								"value": query,
								"boost": 2, // Усиливаем вес, чтобы совпадения по имени ценились выше
							},
						},
					},
					// Поиск по фразе с перестановкой слов
					{
						"match_phrase": map[string]interface{}{
							"name": map[string]interface{}{
								"query": query,
								"slop":  2, // Разрешает слова стоять рядом, но в разном порядке
							},
						},
					},
				},
				"minimum_should_match": "1",
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

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа от Elasticsearch: %w", err)
	}

	var searchResult struct {
		Hits struct {
			Hits []struct {
				Source models.Product `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.Unmarshal(resBody, &searchResult); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа от Elasticsearch: %w", err)
	}

	var products []models.Product
	for _, hit := range searchResult.Hits.Hits {
		products = append(products, hit.Source)
	}

	return products, nil
}
