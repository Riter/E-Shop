package kafkaclient

import (
    "github.com/segmentio/kafka-go"
)

func NewReader(brokers []string, topic, groupID string) *kafka.Reader {
    return kafka.NewReader(kafka.ReaderConfig{
        Brokers:  brokers,
        Topic:    topic,
        GroupID:  groupID,
        MinBytes: 10e3, 
        MaxBytes: 10e6, 
    })
}
