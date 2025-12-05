package main

import (
	"context"
	"log"
	"sync"
	"time"

	pb "github.com/cont1nu1ty/distributed-systems-go-demos/api/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	ServerAddr = "localhost:50051"
)

func main() {
	log.Println("gRPC Client Demo")
	log.Println("================")

	// Create connection
	conn, err := grpc.NewClient(ServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewCalculatorServiceClient(conn)

	// Single request demo
	log.Println("\n--- Single Request Demo ---")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := client.Add(ctx, &pb.BinaryOperation{A: 5, B: 3})
	if err != nil {
		log.Fatalf("Add failed: %v", err)
	}
	log.Printf("Add(5, 3) = %d", result.Value)

	time.Sleep(500 * time.Millisecond)

	// Concurrent requests demo
	log.Println("\n--- Concurrent Requests Demo ---")
	var wg sync.WaitGroup

	requests := []struct {
		op   string
		a, b int32
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
		go func(id int, op string, a, b int32) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			var result *pb.Result
			var err error

			switch op {
			case "Add":
				result, err = client.Add(ctx, &pb.BinaryOperation{A: a, B: b})
			case "Multiply":
				result, err = client.Multiply(ctx, &pb.BinaryOperation{A: a, B: b})
			case "Subtract":
				result, err = client.Subtract(ctx, &pb.BinaryOperation{A: a, B: b})
			case "Divide":
				result, err = client.Divide(ctx, &pb.BinaryOperation{A: a, B: b})
			}

			if err != nil {
				log.Printf("[Request %d] %s(%d, %d) failed: %v", id, op, a, b, err)
			} else {
				log.Printf("[Request %d] %s(%d, %d) = %d", id, op, a, b, result.Value)
			}
		}(i+1, req.op, req.a, req.b)
	}

	wg.Wait()
	log.Println("\n--- All requests completed ---")

	// Demonstrate type safety
	log.Println("\n--- Type-safe Proto Buffers Demo ---")
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()

	result, err = client.Multiply(ctx2, &pb.BinaryOperation{A: 123, B: 456})
	if err != nil {
		log.Fatalf("Multiply failed: %v", err)
	}
	log.Printf("123 * 456 = %d", result.Value)
}
