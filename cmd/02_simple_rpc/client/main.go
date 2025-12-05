package main

import (
	"log"
	"sync"
	"time"

	"github.com/cont1nu1ty/distributed-systems-go-demos/internal/rpc"
)

const (
	ServerAddr = "localhost:9100"
)

func main() {
	log.Println("Simple RPC Client Demo")
	log.Println("======================")

	// Create RPC client
	client, err := rpc.NewClient(ServerAddr)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Single request demo
	log.Println("\n--- Single Request Demo ---")
	result, err := client.Call("CalculatorService", "Add", 5, 3)
	if err != nil {
		log.Printf("Call failed: %v", err)
	} else {
		log.Printf("Add(5, 3) = %v", result)
	}

	time.Sleep(500 * time.Millisecond)

	// Concurrent requests demo
	log.Println("\n--- Concurrent Requests Demo ---")
	var wg sync.WaitGroup

	requests := []struct {
		method string
		a, b   int
	}{
		{"Add", 10, 20},
		{"Multiply", 4, 5},
		{"Subtract", 100, 50},
		{"Add", 7, 8},
		{"Multiply", 3, 3},
		{"Divide", 100, 4},
	}

	for i, req := range requests {
		wg.Add(1)
		go func(id int, method string, a, b int) {
			defer wg.Done()
			
			result, err := client.Call("CalculatorService", method, a, b)
			if err != nil {
				log.Printf("[Request %d] %s(%d, %d) failed: %v", id, method, a, b, err)
			} else {
				log.Printf("[Request %d] %s(%d, %d) = %v", id, method, a, b, result)
			}
		}(i+1, req.method, req.a, req.b)
	}

	wg.Wait()
	log.Println("\n--- All requests completed ---")

	// Demonstrate type-safe interface-like calling
	log.Println("\n--- Interface-like Calling Demo ---")
	result, err = client.Call("CalculatorService", "Add", 100, 200)
	if err != nil {
		log.Printf("Call failed: %v", err)
	} else {
		log.Printf("100 + 200 = %v", result)
	}
}
