package rpc

import (
	"bufio"
	"fmt"
	"net"
	"sync"
	"time"
)

// Client is the RPC client (Stub)
type Client struct {
	addr    string
	conn    net.Conn
	mu      sync.Mutex
	pending map[string]chan *Response
	reader  *bufio.Reader
	closed  bool
}

// NewClient creates a new RPC client
func NewClient(addr string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	client := &Client{
		addr:    addr,
		conn:    conn,
		pending: make(map[string]chan *Response),
		reader:  bufio.NewReader(conn),
	}

	// Start response handler
	go client.handleResponses()

	return client, nil
}

// Call makes a synchronous RPC call
func (c *Client) Call(service, method string, params ...interface{}) (interface{}, error) {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil, fmt.Errorf("client is closed")
	}

	// Generate request ID
	reqID := fmt.Sprintf("%s-%s-%d", service, method, time.Now().UnixNano())
	
	// Create response channel
	respChan := make(chan *Response, 1)
	c.pending[reqID] = respChan
	c.mu.Unlock()

	// Defer cleanup
	defer func() {
		c.mu.Lock()
		delete(c.pending, reqID)
		c.mu.Unlock()
	}()

	// Create request
	req := Request{
		ID:        reqID,
		Service:   service,
		Method:    method,
		Params:    params,
		TimeoutMS: 5000,
	}

	// Encode and send request
	reqData, err := EncodeRequest(req)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	_, err = c.conn.Write(reqData)
	c.mu.Unlock()

	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Wait for response with timeout
	select {
	case resp := <-respChan:
		if resp.Error != "" {
			return nil, fmt.Errorf("remote error: %s", resp.Error)
		}
		return resp.Result, nil
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("request timeout")
	}
}

// handleResponses reads and dispatches responses
func (c *Client) handleResponses() {
	for {
		line, err := c.reader.ReadBytes('\n')
		if err != nil {
			c.mu.Lock()
			c.closed = true
			// Notify all pending requests
			for _, ch := range c.pending {
				close(ch)
			}
			c.pending = make(map[string]chan *Response)
			c.mu.Unlock()
			return
		}

		resp, err := DecodeResponse(line)
		if err != nil {
			continue
		}

		c.mu.Lock()
		respChan, exists := c.pending[resp.ID]
		c.mu.Unlock()

		if exists {
			respChan <- resp
		}
	}
}

// Close closes the client connection
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	return c.conn.Close()
}
