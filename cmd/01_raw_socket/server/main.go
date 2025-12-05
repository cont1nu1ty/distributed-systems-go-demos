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
	Port = ":9001"
)

// Request represents an incoming request
type Request struct {
	ID        string        `json:"id"`
	Service   string        `json:"service"`
	Method    string        `json:"method"`
	Params    []int         `json:"params"`
	TimeoutMS int           `json:"timeout_ms"`
}

// Response represents an outgoing response
type Response struct {
	ID     string      `json:"id"`
	Result interface{} `json:"result"`
	Error  string      `json:"error"`
}

// handleConnection processes a single client connection
func handleConnection(conn net.Conn, connID int) {
	defer func() {
		log.Printf("[Connection %d] Closing connection from %s", connID, conn.RemoteAddr())
		conn.Close()
	}()

	log.Printf("[Connection %d] New connection from %s", connID, conn.RemoteAddr())

	for {
		// Set read deadline to avoid hanging forever
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))

		var req Request
		if err := socket.ReadJSON(conn, &req); err != nil {
			log.Printf("[Connection %d] Read error: %v", connID, err)
			return
		}

		log.Printf("[Connection %d] Received request: %s.%s(%v)", connID, req.Service, req.Method, req.Params)

		// Process the request
		resp := processRequest(req)

		// Send response
		if err := socket.WriteJSON(conn, resp); err != nil {
			log.Printf("[Connection %d] Write error: %v", connID, err)
			return
		}

		log.Printf("[Connection %d] Sent response: %+v", connID, resp)
	}
}

// processRequest handles the business logic
func processRequest(req Request) Response {
	resp := Response{
		ID: req.ID,
	}

	// Manual protocol handling - this is the pain point of raw sockets
	if req.Service != "CalculatorService" {
		resp.Error = fmt.Sprintf("unknown service: %s", req.Service)
		return resp
	}

	switch req.Method {
	case "Add":
		if len(req.Params) != 2 {
			resp.Error = "Add requires exactly 2 parameters"
			return resp
		}
		resp.Result = req.Params[0] + req.Params[1]
	case "Multiply":
		if len(req.Params) != 2 {
			resp.Error = "Multiply requires exactly 2 parameters"
			return resp
		}
		resp.Result = req.Params[0] * req.Params[1]
	case "Subtract":
		if len(req.Params) != 2 {
			resp.Error = "Subtract requires exactly 2 parameters"
			return resp
		}
		resp.Result = req.Params[0] - req.Params[1]
	default:
		resp.Error = fmt.Sprintf("unknown method: %s", req.Method)
	}

	return resp
}

func main() {
	listener, err := net.Listen("tcp", Port)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", Port, err)
	}
	defer listener.Close()

	log.Printf("Raw Socket Server listening on %s", Port)
	log.Println("Waiting for connections...")

	var connCounter int
	var mu sync.Mutex

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}

		mu.Lock()
		connCounter++
		currentID := connCounter
		mu.Unlock()

		// Each connection is handled in its own goroutine for concurrent processing
		go handleConnection(conn, currentID)
	}
}
