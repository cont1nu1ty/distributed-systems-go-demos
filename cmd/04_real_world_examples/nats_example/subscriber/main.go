package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
)

const (
	NATSUrl = nats.DefaultURL // nats://localhost:4222
)

func main() {
	log.Println("NATS Subscriber Demo")
	log.Println("====================")

	if len(os.Args) < 2 {
		log.Println("Usage: go run main.go <subject> [subscriber_id]")
		log.Println("Example: go run main.go news 1")
		log.Println("\nStarting with default subject 'news' and subscriber ID 1")
		os.Args = append(os.Args, "news", "1")
	}

	subject := os.Args[1]
	subscriberID := 1
	if len(os.Args) > 2 {
		fmt.Sscanf(os.Args[2], "%d", &subscriberID)
	}

	log.Printf("Subscriber %d starting...", subscriberID)
	log.Printf("Subscribing to subject: '%s'", subject)
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

	// Message counter
	msgCount := 0

	// Subscribe to subject
	sub, err := nc.Subscribe(subject, func(msg *nats.Msg) {
		msgCount++
		log.Printf("[Subscriber %d] Received message #%d on subject '%s': %s",
			subscriberID, msgCount, msg.Subject, string(msg.Data))
	})
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}
	defer sub.Unsubscribe()

	log.Printf("[Subscriber %d] Subscribed to subject '%s'", subscriberID, subject)
	log.Println("Waiting for messages... (Press Ctrl+C to exit)")

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	<-sigChan

	log.Printf("\n[Subscriber %d] Shutting down...", subscriberID)
	log.Printf("[Subscriber %d] Total messages received: %d", subscriberID, msgCount)

	// Give time for cleanup
	time.Sleep(100 * time.Millisecond)
}
