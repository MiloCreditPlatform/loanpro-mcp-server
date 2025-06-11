package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

type StdioTransport struct {
	server *MCPServer
	reader *bufio.Reader
	writer io.Writer
}

func NewStdioTransport(server *MCPServer) *StdioTransport {
	return &StdioTransport{
		server: server,
		reader: bufio.NewReader(os.Stdin),
		writer: os.Stdout,
	}
}

func (t *StdioTransport) Run() error {
	log.Println("[STDIO] Starting stdio transport")
	for {
		line, err := t.reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				log.Println("[STDIO] EOF received, shutting down")
				return nil
			}
			log.Printf("[STDIO] Error reading from stdin: %v", err)
			fmt.Fprintf(os.Stderr, "[ERROR] Failed to read from stdin: %v\n", err)
			return fmt.Errorf("failed to read from stdin: %w", err)
		}

		log.Printf("[STDIO] Received: %s", string(line))

		var req MCPRequest
		if err := json.Unmarshal(line, &req); err != nil {
			log.Printf("[STDIO] Parse error: %v", err)
			fmt.Fprintf(os.Stderr, "[ERROR] JSON parse error: %v\nInput: %s\n", err, string(line))
			t.sendError(-32700, "Parse error", nil)
			continue
		}

		log.Printf("[STDIO] Processing request: method=%s, id=%v", req.Method, req.ID)
		response := t.server.handleMCPRequest(req)
		
		// Don't send response for notifications (empty JSONRPC means no response)
		if response.JSONRPC == "" {
			log.Printf("[STDIO] Notification processed, no response sent")
			continue
		}
		
		responseData, err := json.Marshal(response)
		if err != nil {
			log.Printf("[STDIO] Marshal error: %v", err)
			fmt.Fprintf(os.Stderr, "[ERROR] Failed to marshal response: %v\n", err)
			t.sendError(-32603, "Internal error", req.ID)
			continue
		}

		log.Printf("[STDIO] Sending response: %s", string(responseData))
		fmt.Fprintf(t.writer, "%s\n", responseData)
	}
}

func (t *StdioTransport) sendError(code int, message string, id any) {
	log.Printf("[STDIO] Sending error: code=%d, message=%s, id=%v", code, message, id)
	errorResponse := MCPResponse{
		JSONRPC: "2.0",
		Error: &MCPError{
			Code:    code,
			Message: message,
		},
		ID: id,
	}
	
	data, _ := json.Marshal(errorResponse)
	log.Printf("[STDIO] Error response: %s", string(data))
	fmt.Fprintf(t.writer, "%s\n", data)
}