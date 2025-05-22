package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Riter/E-Shop/facade-consumer/internal/config"
	"github.com/Riter/E-Shop/facade-consumer/internal/kafkaclient"
	"github.com/Riter/E-Shop/facade-consumer/internal/redisclient"
)

// Структура сообщения
type KafkaMessage struct {
    OperationType int             `json:"operation_type"`
    ItemID        int64           `json:"item_id"`
    Item          json.RawMessage `json:"item"` // можно не парсить полностью
}

func main() {
    ctx := context.Background()

    cfg := config.LoadConfig()

    rdb := redisclient.NewClient(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
    kafkaReader := kafkaclient.NewReader(cfg.KafkaBrokers, cfg.KafkaTopic, cfg.KafkaGroupID)

    log.Printf("Starting Kafka consumer for topic %s\n", cfg.KafkaTopic)

    for {
        m, err := kafkaReader.ReadMessage(ctx)
        if err != nil {
            log.Printf("Error reading Kafka message: %v", err)
            time.Sleep(time.Second)
            continue
        }

        log.Printf("Received message at offset %d: %s\n", m.Offset, string(m.Value))

        var msg KafkaMessage
        if err := json.Unmarshal(m.Value, &msg); err != nil {
            log.Printf("Failed to unmarshal message: %v", err)
            continue
        }

        // Инвалидация только при DELETE (1) или CHANGE (2)
        if msg.OperationType == 1 || msg.OperationType == 2 {
            key := fmt.Sprintf("%d", msg.ItemID)
            deleted, err := rdb.Del(ctx, key).Result()
            if err != nil {
                log.Printf("Failed to delete key %s from Redis: %v", key, err)
                continue
            }

            if deleted > 0 {
                log.Printf("Deleted key %s from Redis", key)
            } else {
                log.Printf("Key %s not found in Redis", key)
            }
        }
    }
}
