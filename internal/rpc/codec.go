package rpc

import (
	"encoding/json"
	"fmt"
)

// Request represents an RPC request
type Request struct {
	ID        string        `json:"id"`
	Service   string        `json:"service"`
	Method    string        `json:"method"`
	Params    []interface{} `json:"params"`
	TimeoutMS int           `json:"timeout_ms"`
}

// Response represents an RPC response
type Response struct {
	ID     string      `json:"id"`
	Result interface{} `json:"result"`
	Error  string      `json:"error"`
}

// EncodeRequest encodes a request to JSON
func EncodeRequest(req Request) ([]byte, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}
	return append(data, '\n'), nil
}

// DecodeRequest decodes a request from JSON
func DecodeRequest(data []byte) (*Request, error) {
	var req Request
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("failed to decode request: %w", err)
	}
	return &req, nil
}

// EncodeResponse encodes a response to JSON
func EncodeResponse(resp Response) ([]byte, error) {
	data, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to encode response: %w", err)
	}
	return append(data, '\n'), nil
}

// DecodeResponse decodes a response from JSON
func DecodeResponse(data []byte) (*Response, error) {
	var resp Response
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &resp, nil
}
