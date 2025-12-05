package main

import (
	"log"

	"github.com/cont1nu1ty/distributed-systems-go-demos/internal/rpc"
)

const (
	Port = ":9100"
)

// CalculatorService defines the calculator service interface
type CalculatorService interface {
	Add(a, b int) int
	Multiply(a, b int) int
	Subtract(a, b int) int
	Divide(a, b int) int
}

// CalculatorImpl is the implementation of CalculatorService
type CalculatorImpl struct{}

// Add adds two numbers
func (c *CalculatorImpl) Add(a, b int) int {
	log.Printf("Add(%d, %d) called", a, b)
	return a + b
}

// Multiply multiplies two numbers
func (c *CalculatorImpl) Multiply(a, b int) int {
	log.Printf("Multiply(%d, %d) called", a, b)
	return a * b
}

// Subtract subtracts b from a
func (c *CalculatorImpl) Subtract(a, b int) int {
	log.Printf("Subtract(%d, %d) called", a, b)
	return a - b
}

// Divide divides a by b
func (c *CalculatorImpl) Divide(a, b int) int {
	log.Printf("Divide(%d, %d) called", a, b)
	if b == 0 {
		log.Println("Warning: division by zero, returning 0")
		return 0
	}
	return a / b
}

func main() {
	// Create RPC server
	server := rpc.NewServer()

	// Create and register calculator service
	calc := &CalculatorImpl{}
	if err := server.Register("CalculatorService", calc); err != nil {
		log.Fatalf("Failed to register service: %v", err)
	}

	log.Println("Simple RPC Server starting...")
	log.Printf("Listening on %s", Port)

	// Start server (blocking)
	if err := server.Serve(Port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
