package rpc

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"reflect"
	"sync"
)

// Server is the RPC server (Skeleton)
type Server struct {
	services map[string]interface{}
	mu       sync.RWMutex
}

// NewServer creates a new RPC server
func NewServer() *Server {
	return &Server{
		services: make(map[string]interface{}),
	}
}

// Register registers a service instance
func (s *Server) Register(name string, service interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.services[name]; exists {
		return fmt.Errorf("service %s already registered", name)
	}

	s.services[name] = service
	log.Printf("Registered service: %s", name)
	return nil
}

// HandleConnection handles a client connection
func (s *Server) HandleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for {
		// Read request
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return
		}

		req, err := DecodeRequest(line)
		if err != nil {
			s.sendError(conn, "", fmt.Sprintf("decode error: %v", err))
			continue
		}

		// Process request
		result, err := s.invoke(req.Service, req.Method, req.Params)

		resp := Response{
			ID: req.ID,
		}

		if err != nil {
			resp.Error = err.Error()
		} else {
			resp.Result = result
		}

		// Send response
		respData, err := EncodeResponse(resp)
		if err != nil {
			log.Printf("Failed to encode response: %v", err)
			continue
		}

		if _, err := conn.Write(respData); err != nil {
			log.Printf("Failed to write response: %v", err)
			return
		}
	}
}

// invoke calls a method on a registered service using reflection
func (s *Server) invoke(serviceName, methodName string, params []interface{}) (interface{}, error) {
	s.mu.RLock()
	service, exists := s.services[serviceName]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("service not found: %s", serviceName)
	}

	// Get service value
	serviceValue := reflect.ValueOf(service)

	// Get method
	method := serviceValue.MethodByName(methodName)
	if !method.IsValid() {
		return nil, fmt.Errorf("method not found: %s.%s", serviceName, methodName)
	}

	// Convert params to reflect.Value
	methodType := method.Type()
	if len(params) != methodType.NumIn() {
		return nil, fmt.Errorf("wrong number of parameters: expected %d, got %d", methodType.NumIn(), len(params))
	}

	args := make([]reflect.Value, len(params))
	for i, param := range params {
		// Convert JSON numbers to proper types
		paramType := methodType.In(i)

		switch paramType.Kind() {
		case reflect.Int:
			// JSON unmarshals numbers as float64
			if v, ok := param.(float64); ok {
				args[i] = reflect.ValueOf(int(v))
			} else {
				return nil, fmt.Errorf("parameter %d: expected number, got %T", i, param)
			}
		case reflect.String:
			if v, ok := param.(string); ok {
				args[i] = reflect.ValueOf(v)
			} else {
				return nil, fmt.Errorf("parameter %d: expected string, got %T", i, param)
			}
		default:
			args[i] = reflect.ValueOf(param)
		}
	}

	// Call method
	results := method.Call(args)

	// For simplicity, assume single return value
	if len(results) == 0 {
		return nil, nil
	}

	return results[0].Interface(), nil
}

// sendError sends an error response
func (s *Server) sendError(conn net.Conn, id, errMsg string) {
	resp := Response{
		ID:    id,
		Error: errMsg,
	}

	respData, err := EncodeResponse(resp)
	if err != nil {
		log.Printf("Failed to encode error response: %v", err)
		return
	}

	if _, err := conn.Write(respData); err != nil {
		log.Printf("Failed to write error response: %v", err)
	}
}

// Serve starts the RPC server
func (s *Server) Serve(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	defer listener.Close()

	log.Printf("RPC Server listening on %s", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}

		go s.HandleConnection(conn)
	}
}
