package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"github.com/cont1nu1ty/distributed-systems-go-demos/internal/broker"
)

const (
	Port = ":9200"
)

// Command represents a broker command
type Command struct {
	Action  string      `json:"action"`  // "subscribe", "publish"
	Topic   string      `json:"topic"`
	Payload interface{} `json:"payload,omitempty"`
}

// BrokerServer wraps the broker and handles network connections
type BrokerServer struct {
	broker *broker.Broker
	mu     sync.Mutex
}

// NewBrokerServer creates a new broker server
func NewBrokerServer() *BrokerServer {
	return &BrokerServer{
		broker: broker.NewBroker(),
	}
}

// handleConnection handles a client connection
func (bs *BrokerServer) handleConnection(conn net.Conn) {
	defer conn.Close()
	
	log.Printf("New connection from %s", conn.RemoteAddr())
	
	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)
	
	for {
		var cmd Command
		if err := decoder.Decode(&cmd); err != nil {
			if err != io.EOF {
				log.Printf("Decode error: %v", err)
			}
			return
		}
		
		log.Printf("Received command: %s on topic '%s'", cmd.Action, cmd.Topic)
		
		switch cmd.Action {
		case "subscribe":
			bs.handleSubscribe(conn, cmd.Topic, encoder)
			return // Subscription is long-lived, exit after handling
			
		case "publish":
			if err := bs.broker.Publish(cmd.Topic, cmd.Payload); err != nil {
				response := map[string]string{"status": "error", "message": err.Error()}
				encoder.Encode(response)
			} else {
				response := map[string]string{"status": "ok"}
				encoder.Encode(response)
			}
			
		default:
			response := map[string]string{"status": "error", "message": "unknown action"}
			encoder.Encode(response)
		}
	}
}

// handleSubscribe handles a subscription request
func (bs *BrokerServer) handleSubscribe(conn net.Conn, topic string, encoder *json.Encoder) {
	msgChan, err := bs.broker.Subscribe(topic)
	if err != nil {
		response := map[string]string{"status": "error", "message": err.Error()}
		encoder.Encode(response)
		return
	}
	
	// Send acknowledgment
	response := map[string]string{"status": "subscribed"}
	if err := encoder.Encode(response); err != nil {
		return
	}
	
	log.Printf("Client subscribed to topic '%s'", topic)
	
	// Stream messages to client
	for msg := range msgChan {
		if err := encoder.Encode(msg); err != nil {
			log.Printf("Failed to send message: %v", err)
			return
		}
	}
	
	log.Printf("Subscription ended for topic '%s'", topic)
}

// Start starts the broker server
func (bs *BrokerServer) Start() error {
	listener, err := net.Listen("tcp", Port)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	defer listener.Close()
	
	log.Printf("Message Broker Server listening on %s", Port)
	log.Println("Waiting for connections...")
	
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}
		
		go bs.handleConnection(conn)
	}
}

func main() {
	log.Println("Message Broker Server starting...")
	
	server := NewBrokerServer()
	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
