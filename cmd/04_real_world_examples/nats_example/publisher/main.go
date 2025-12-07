package main

import (
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

const (
	NATSUrl = nats.DefaultURL // nats://localhost:4222
)

func main() {
	log.Println("NATS Publisher Demo")
	log.Println("===================")
	log.Printf("Connecting to NATS server at %s", NATSUrl)
	log.Println("Note: Make sure NATS server is running (docker run -p 4222:4222 nats)")

	// Connect to NATS
	nc, err := nats.Connect(NATSUrl)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v\n", err)
		log.Println("Please start a NATS server:")
		log.Println("  docker run -p 4222:4222 nats")
		log.Println("  OR download from: https://nats.io/download/")
		return
	}
	defer nc.Close()

	log.Println("Connected to NATS server")

	// Publish messages to different subjects
	log.Println("\n--- Publishing Messages ---")

	messages := []struct {
		subject string
		message string
	}{
		{"news", "Breaking: Go 1.22 released!"},
		{"updates", "System update: v2.0 available"},
		{"alerts", "Warning: Maintenance scheduled at 2am"},
		{"news", "Tech news: New distributed systems course"},
		{"updates", "Performance improvements in v2.1"},
	}

	for i, msg := range messages {
		log.Printf("[Message %d] Publishing to '%s': %s", i+1, msg.subject, msg.message)

		if err := nc.Publish(msg.subject, []byte(msg.message)); err != nil {
			log.Printf("[Message %d] Failed to publish: %v", i+1, err)
		} else {
			log.Printf("[Message %d] Published successfully", i+1)
		}

		time.Sleep(500 * time.Millisecond)
	}

	// Ensure all messages are sent
	if err := nc.Flush(); err != nil {
		log.Printf("Failed to flush: %v", err)
	}

	// Rapid fire demo
	log.Println("\n--- Rapid Publishing Demo ---")
	for i := 0; i < 10; i++ {
		message := fmt.Sprintf("Rapid message %d", i+1)

		if err := nc.Publish("rapid", []byte(message)); err != nil {
			log.Printf("Failed to publish rapid message %d: %v", i+1, err)
		} else {
			log.Printf("Published rapid message %d", i+1)
		}

		time.Sleep(100 * time.Millisecond)
	}

	// Flush to ensure all messages are sent
	if err := nc.Flush(); err != nil {
		log.Printf("Failed to flush: %v", err)
	}

	log.Println("\n--- All messages published ---")
	log.Printf("Published %d messages total", len(messages)+10)
}
