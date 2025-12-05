package main

import (
	"context"
	"log"
	"net"

	pb "github.com/cont1nu1ty/distributed-systems-go-demos/api/proto"
	"google.golang.org/grpc"
)

const (
	Port = ":50051"
)

// server implements the CalculatorService
type server struct {
	pb.UnimplementedCalculatorServiceServer
}

// Add implements CalculatorService.Add
func (s *server) Add(ctx context.Context, req *pb.BinaryOperation) (*pb.Result, error) {
	result := req.A + req.B
	log.Printf("Add(%d, %d) = %d", req.A, req.B, result)
	return &pb.Result{Value: result}, nil
}

// Multiply implements CalculatorService.Multiply
func (s *server) Multiply(ctx context.Context, req *pb.BinaryOperation) (*pb.Result, error) {
	result := req.A * req.B
	log.Printf("Multiply(%d, %d) = %d", req.A, req.B, result)
	return &pb.Result{Value: result}, nil
}

// Subtract implements CalculatorService.Subtract
func (s *server) Subtract(ctx context.Context, req *pb.BinaryOperation) (*pb.Result, error) {
	result := req.A - req.B
	log.Printf("Subtract(%d, %d) = %d", req.A, req.B, result)
	return &pb.Result{Value: result}, nil
}

// Divide implements CalculatorService.Divide
func (s *server) Divide(ctx context.Context, req *pb.BinaryOperation) (*pb.Result, error) {
	if req.B == 0 {
		log.Printf("Divide(%d, %d) - division by zero, returning 0", req.A, req.B)
		return &pb.Result{Value: 0}, nil
	}
	result := req.A / req.B
	log.Printf("Divide(%d, %d) = %d", req.A, req.B, result)
	return &pb.Result{Value: result}, nil
}

func main() {
	log.Println("gRPC Server starting...")

	listener, err := net.Listen("tcp", Port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterCalculatorServiceServer(grpcServer, &server{})

	log.Printf("gRPC Server listening on %s", Port)
	log.Println("Using HTTP/2 and Protocol Buffers")

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
