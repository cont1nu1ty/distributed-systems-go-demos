package main

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/cont1nu1ty/distributed-systems-go-demos/pkg/socket"
)

const (
	ServerAddr = "localhost:9001"
)

// Request represents an outgoing request
type Request struct {
	ID        string `json:"id"`
	Service   string `json:"service"`
	Method    string `json:"method"`
	Params    []int  `json:"params"`
	TimeoutMS int    `json:"timeout_ms"`
}

// Response represents an incoming response
type Response struct {
	ID     string      `json:"id"`
	Result interface{} `json:"result"`
	Error  string      `json:"error"`
}

// makeRequest sends a request and waits for a response
func makeRequest(clientID int, method string, params []int) error {
	conn, err := net.Dial("tcp", ServerAddr)
	if err != nil {
		return fmt.Errorf("dial error: %w", err)
	}
	defer conn.Close()

	requestID := fmt.Sprintf("req-%d-%d", clientID, time.Now().UnixNano())

	req := Request{
		ID:        requestID,
		Service:   "CalculatorService",
		Method:    method,
		Params:    params,
		TimeoutMS: 2000,
	}

	log.Printf("[Client %d] Sending request: %s.%s(%v)", clientID, req.Service, req.Method, req.Params)

	// Send request
	if err := socket.WriteJSON(conn, req); err != nil {
		return fmt.Errorf("write error: %w", err)
	}

	// Wait for response
	var resp Response
	if err := socket.ReadJSON(conn, &resp); err != nil {
		return fmt.Errorf("read error: %w", err)
	}

	if resp.Error != "" {
		log.Printf("[Client %d] Error response: %s", clientID, resp.Error)
		return fmt.Errorf("server error: %s", resp.Error)
	}

	log.Printf("[Client %d] Response: %v", clientID, resp.Result)
	return nil
}

func main() {
	log.Println("Raw Socket Client Demo")
	log.Println("======================")

	// Single request demo
	log.Println("\n--- Single Request Demo ---")
	if err := makeRequest(1, "Add", []int{5, 3}); err != nil {
		log.Printf("Request failed: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// Concurrent requests demo
	log.Println("\n--- Concurrent Requests Demo ---")
	var wg sync.WaitGroup
	requests := []struct {
		method string
		params []int
	}{
		{"Add", []int{10, 20}},
		{"Multiply", []int{4, 5}},
		{"Subtract", []int{100, 50}},
		{"Add", []int{7, 8}},
		{"Multiply", []int{3, 3}},
	}

	for i, req := range requests {
		wg.Add(1)
		go func(clientID int, method string, params []int) {
			defer wg.Done()
			if err := makeRequest(clientID, method, params); err != nil {
				log.Printf("[Client %d] Request failed: %v", clientID, err)
			}
		}(i+1, req.method, req.params)
	}

	wg.Wait()
	log.Println("\n--- All requests completed ---")
}
