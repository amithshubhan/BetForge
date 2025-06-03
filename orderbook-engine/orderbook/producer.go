package orderbook

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/IBM/sarama"
)

var producer sarama.SyncProducer

func init() {
    // Configure Sarama producer
    config := sarama.NewConfig()
    config.Producer.RequiredAcks = sarama.WaitForAll          // Wait for all in-sync replicas to acknowledge
    config.Producer.Retry.Max = 5                             // Retry up to 5 times
    config.Producer.Return.Successes = true                  // Return success messages
    config.Producer.Compression = sarama.CompressionSnappy   // Use Snappy compression for better performance
    config.Version = sarama.V2_8_1_0                        // Set Kafka version

    // Create a new Sarama producer
    var err error
    producer, err = sarama.NewSyncProducer([]string{"localhost:9092"}, config)
    if err != nil {
        log.Printf("Failed to start Sarama producer: %v", err)
        return
    }
}

// PublishMatchEvent publishes an order as a Kafka message
func PublishMatchEvent(order Order) {
    if producer == nil {
        log.Println("Kafka producer not initialized")
        return
    }

    payload, err := json.Marshal(order)
    if err != nil {
        log.Printf("Failed to marshal order: %v", err)
        return
    }
    fmt.Println("Publishing event to Kafka:", string(payload))

    // Create a Kafka message
    message := &sarama.ProducerMessage{
        Topic: "match.events",
        Value: sarama.StringEncoder(payload),
    }

    // Send the message
    partition, offset, err := producer.SendMessage(message)
    if err != nil {
        log.Println("Kafka publish failed:", err)
    } else {
        fmt.Printf("Message published to partition %d at offset %d\n", partition, offset)
    }
}

// Close closes the Sarama producer
func Close() {
    if producer != nil {
        if err := producer.Close(); err != nil {
            log.Println("Failed to close Sarama producer:", err)
        }
    }
}