package kafka

import (
	"context"
	"encoding/json"
	"log"
	"online-shop/internal/elasticsearch"
	"online-shop/internal/models"

	"github.com/IBM/sarama"
)

type Consumer struct {
	consumer sarama.ConsumerGroup
	elastic  *elasticsearch.ESClient
	topic    string
}

type ProductEvent struct {
	OperationType int             `json:"operation_type"` // 1=DELETE, 2=CHANGE, 3=CREATE
	ItemID        string          `json:"item_id"`
	Item          *models.Product `json:"item"`
}

func NewConsumer(brokers []string, groupID string, topic string, elastic *elasticsearch.ESClient) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		consumer: consumer,
		elastic:  elastic,
		topic:    topic,
	}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	topics := []string{c.topic}
	handler := &consumerGroupHandler{
		elastic: c.elastic,
	}

	for {
		err := c.consumer.Consume(ctx, topics, handler)
		if err != nil {
			log.Printf("Error from consumer: %v", err)
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

type consumerGroupHandler struct {
	elastic *elasticsearch.ESClient
}

func (h *consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		var event ProductEvent
		if err := json.Unmarshal(message.Value, &event); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		switch event.OperationType {
		case 3: // CREATE
			if event.Item == nil {
				log.Printf("Error: CREATE event has nil item")
				continue
			}
			if err := h.elastic.IndexProducts([]models.Product{*event.Item}); err != nil {
				log.Printf("Error indexing product: %v", err)
				continue
			}
			log.Printf("Product %s indexed in Elasticsearch", event.ItemID)

		case 2: // CHANGE
			if event.Item == nil {
				log.Printf("Error: CHANGE event has nil item")
				continue
			}
			if err := h.elastic.IndexProducts([]models.Product{*event.Item}); err != nil {
				log.Printf("Error updating product: %v", err)
				continue
			}
			log.Printf("Product %s updated in Elasticsearch", event.ItemID)

		case 1: // DELETE
			if err := h.elastic.DeleteProduct(event.ItemID); err != nil {
				log.Printf("Error deleting product: %v", err)
				continue
			}
			log.Printf("Product %s deleted from Elasticsearch", event.ItemID)

		default:
			log.Printf("Unknown operation type: %d", event.OperationType)
			continue
		}

		session.MarkMessage(message, "")
	}
	return nil
}
