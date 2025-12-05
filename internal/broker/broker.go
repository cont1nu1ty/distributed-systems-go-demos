package broker

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// Message represents a message in the broker
type Message struct {
	Topic     string      `json:"topic"`
	Payload   interface{} `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
}

// Broker is an in-memory pub/sub message broker
type Broker struct {
	mu           sync.RWMutex
	subscribers  map[string][]chan Message
	subCounter   int
	closed       bool
	closeChan    chan struct{}
}

// NewBroker creates a new message broker
func NewBroker() *Broker {
	return &Broker{
		subscribers: make(map[string][]chan Message),
		closeChan:   make(chan struct{}),
	}
}

// Subscribe subscribes to a topic and returns a channel for receiving messages
func (b *Broker) Subscribe(topic string) (<-chan Message, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil, fmt.Errorf("broker is closed")
	}

	// Create a buffered channel for the subscriber
	ch := make(chan Message, 100)
	
	b.subscribers[topic] = append(b.subscribers[topic], ch)
	b.subCounter++
	
	log.Printf("New subscriber for topic '%s' (total subscribers: %d)", topic, len(b.subscribers[topic]))
	
	return ch, nil
}

// Publish publishes a message to a topic
func (b *Broker) Publish(topic string, payload interface{}) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return fmt.Errorf("broker is closed")
	}

	msg := Message{
		Topic:     topic,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	subscribers := b.subscribers[topic]
	if len(subscribers) == 0 {
		log.Printf("No subscribers for topic '%s'", topic)
		return nil
	}

	log.Printf("Publishing message to topic '%s' (%d subscribers)", topic, len(subscribers))

	// Send to all subscribers concurrently without blocking
	for _, ch := range subscribers {
		// Use select with default to avoid blocking if subscriber is slow
		select {
		case ch <- msg:
			// Message sent successfully
		default:
			// Channel is full, log warning but don't block
			log.Printf("Warning: subscriber channel full for topic '%s', message dropped", topic)
		}
	}

	return nil
}

// Unsubscribe removes a subscriber channel
func (b *Broker) Unsubscribe(topic string, ch <-chan Message) {
	b.mu.Lock()
	defer b.mu.Unlock()

	subscribers := b.subscribers[topic]
	for i, subCh := range subscribers {
		if subCh == ch {
			// Remove from slice
			b.subscribers[topic] = append(subscribers[:i], subscribers[i+1:]...)
			close(subCh)
			log.Printf("Unsubscribed from topic '%s'", topic)
			break
		}
	}
}

// GetTopics returns all active topics
func (b *Broker) GetTopics() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	topics := make([]string, 0, len(b.subscribers))
	for topic := range b.subscribers {
		topics = append(topics, topic)
	}
	return topics
}

// GetSubscriberCount returns the number of subscribers for a topic
func (b *Broker) GetSubscriberCount(topic string) int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return len(b.subscribers[topic])
}

// Close closes the broker and all subscriber channels
func (b *Broker) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return
	}

	b.closed = true
	close(b.closeChan)

	// Close all subscriber channels
	for topic, subscribers := range b.subscribers {
		for _, ch := range subscribers {
			close(ch)
		}
		delete(b.subscribers, topic)
	}

	log.Println("Broker closed")
}
