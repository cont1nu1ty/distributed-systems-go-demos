package socket

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
)

// ReadJSON reads a JSON message from the connection
// It expects a newline-delimited JSON message
func ReadJSON(conn net.Conn, v interface{}) error {
	reader := bufio.NewReader(conn)
	line, err := reader.ReadBytes('\n')
	if err != nil {
		if err == io.EOF {
			return fmt.Errorf("connection closed")
		}
		return fmt.Errorf("read error: %w", err)
	}
	
	if err := json.Unmarshal(line, v); err != nil {
		return fmt.Errorf("json unmarshal error: %w", err)
	}
	
	return nil
}

// WriteJSON writes a JSON message to the connection
// It appends a newline for message delimiting
func WriteJSON(conn net.Conn, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}
	
	data = append(data, '\n')
	_, err = conn.Write(data)
	if err != nil {
		return fmt.Errorf("write error: %w", err)
	}
	
	return nil
}
