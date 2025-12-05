package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
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

func publish(topic string, payload interface{}) error {
	conn, err := net.Dial("tcp", BrokerAddr)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)

	// Send publish command
	cmd := Command{
		Action:  "publish",
		Topic:   topic,
		Payload: payload,
	}

	if err := encoder.Encode(cmd); err != nil {
		return fmt.Errorf("failed to send command: %w", err)
	}

	// Read response
	var resp Response
	if err := decoder.Decode(&resp); err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.Status != "ok" {
		return fmt.Errorf("publish failed: %s", resp.Message)
	}

	return nil
}

func main() {
	log.Println("Message Broker Producer Demo")
	log.Println("=============================")

	// Publish messages to different topics
	log.Println("\n--- Publishing Messages ---")

	messages := []struct {
		topic   string
		payload interface{}
	}{
		{"news", map[string]interface{}{"title": "Breaking News", "content": "Go 1.22 released!"}},
		{"updates", map[string]interface{}{"version": "2.0", "changes": "Bug fixes and improvements"}},
		{"alerts", map[string]interface{}{"level": "warning", "message": "System maintenance at 2am"}},
		{"news", map[string]interface{}{"title": "Tech Update", "content": "New distributed systems course"}},
		{"updates", map[string]interface{}{"version": "2.1", "changes": "Performance enhancements"}},
	}

	for i, msg := range messages {
		log.Printf("[Message %d] Publishing to topic '%s': %v", i+1, msg.topic, msg.payload)

		if err := publish(msg.topic, msg.payload); err != nil {
			log.Printf("[Message %d] Failed to publish: %v", i+1, err)
		} else {
			log.Printf("[Message %d] Published successfully", i+1)
		}

		time.Sleep(500 * time.Millisecond)
	}

	// Rapid fire demo
	log.Println("\n--- Rapid Publishing Demo ---")
	for i := 0; i < 10; i++ {
		payload := map[string]interface{}{
			"sequence": i + 1,
			"data":     fmt.Sprintf("Rapid message %d", i+1),
		}

		if err := publish("rapid", payload); err != nil {
			log.Printf("Failed to publish rapid message %d: %v", i+1, err)
		} else {
			log.Printf("Published rapid message %d", i+1)
		}

		time.Sleep(100 * time.Millisecond)
	}

	log.Println("\n--- All messages published ---")
}
