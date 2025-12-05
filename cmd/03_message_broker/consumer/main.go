package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	BrokerAddr = "localhost:9200"
)

// Command represents a broker command
type Command struct {
	Action  string      `json:"action"`
	Topic   string      `json:"topic"`
	Payload interface{} `json:"payload,omitempty"`
}

// Response represents a broker response
type Response struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// Message represents a received message
type Message struct {
	Topic     string                 `json:"topic"`
	Payload   map[string]interface{} `json:"payload"`
	Timestamp time.Time              `json:"timestamp"`
}

func subscribe(topic string, consumerID int) error {
	conn, err := net.Dial("tcp", BrokerAddr)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)

	// Send subscribe command
	cmd := Command{
		Action: "subscribe",
		Topic:  topic,
	}

	if err := encoder.Encode(cmd); err != nil {
		return fmt.Errorf("failed to send command: %w", err)
	}

	// Read subscription acknowledgment
	var resp Response
	if err := decoder.Decode(&resp); err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.Status != "subscribed" {
		return fmt.Errorf("subscription failed: %s", resp.Message)
	}

	log.Printf("[Consumer %d] Subscribed to topic '%s'", consumerID, topic)

	// Receive messages
	msgCount := 0
	for {
		var msg Message
		if err := decoder.Decode(&msg); err != nil {
			log.Printf("[Consumer %d] Connection closed: %v", consumerID, err)
			return nil
		}

		msgCount++
		log.Printf("[Consumer %d] Received message #%d on topic '%s': %v",
			consumerID, msgCount, msg.Topic, msg.Payload)
	}
}

func main() {
	log.Println("Message Broker Consumer Demo")
	log.Println("=============================")

	if len(os.Args) < 2 {
		log.Println("Usage: go run main.go <topic> [consumer_id]")
		log.Println("Example: go run main.go news 1")
		log.Println("\nStarting with default topic 'news' and consumer ID 1")
		os.Args = append(os.Args, "news", "1")
	}

	topic := os.Args[1]
	consumerID := 1
	if len(os.Args) > 2 {
		fmt.Sscanf(os.Args[2], "%d", &consumerID)
	}

	log.Printf("Consumer %d starting...", consumerID)
	log.Printf("Subscribing to topic: '%s'", topic)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	errChan := make(chan error, 1)

	// Start subscription in a goroutine
	go func() {
		errChan <- subscribe(topic, consumerID)
	}()

	// Wait for error or shutdown signal
	select {
	case err := <-errChan:
		if err != nil {
			log.Printf("Subscription error: %v", err)
		}
	case sig := <-sigChan:
		log.Printf("\nReceived signal %v, shutting down...", sig)
	}

	log.Printf("[Consumer %d] Stopped", consumerID)
}
