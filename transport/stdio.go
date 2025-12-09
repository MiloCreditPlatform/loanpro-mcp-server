package transport

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
)

// StdioTransport handles MCP communication over stdin/stdout
type StdioTransport struct {
	handler MCPHandler
	reader  *bufio.Reader
	writer  io.Writer
}

// NewStdioTransport creates a new stdio transport
func NewStdioTransport(handler MCPHandler) *StdioTransport {
	return &StdioTransport{
		handler: handler,
		reader:  bufio.NewReader(os.Stdin),
		writer:  os.Stdout,
	}
}

// Run starts the stdio transport loop
func (t *StdioTransport) Run() error {
	slog.Debug("Starting stdio transport")
	for {
		line, err := t.reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				slog.Debug("EOF received, shutting down")
				return nil
			}
			slog.Error("Error reading from stdin", "error", err)
			fmt.Fprintf(os.Stderr, "[ERROR] Failed to read from stdin: %v\n", err)
			return fmt.Errorf("failed to read from stdin: %w", err)
		}

		slog.Debug("Received message", "data", string(line))

		var req MCPRequest
		if err := json.Unmarshal(line, &req); err != nil {
			slog.Error("JSON parse error", "error", err, "input", string(line))
			fmt.Fprintf(os.Stderr, "[ERROR] JSON parse error: %v\nInput: %s\n", err, string(line))
			t.sendError(-32700, "Parse error", nil)
			continue
		}

		slog.Debug("Processing request", "method", req.Method, "id", req.ID)
		response := t.handler.HandleMCPRequest(req)

		// Don't send response for notifications (empty JSONRPC means no response)
		if response.JSONRPC == "" {
			slog.Debug("Notification processed, no response sent")
			continue
		}

		responseData, err := json.Marshal(response)
		if err != nil {
			slog.Error("Marshal error", "error", err)
			fmt.Fprintf(os.Stderr, "[ERROR] Failed to marshal response: %v\n", err)
			t.sendError(-32603, "Internal error", req.ID)
			continue
		}

		slog.Debug("Sending response", "data", string(responseData))
		fmt.Fprintf(t.writer, "%s\n", responseData)
	}
}

// sendError sends an error response
func (t *StdioTransport) sendError(code int, message string, id any) {
	slog.Error("Sending error response", "code", code, "message", message, "id", id)
	errorResponse := MCPResponse{
		JSONRPC: "2.0",
		Error: &MCPError{
			Code:    code,
			Message: message,
		},
		ID: id,
	}

	data, _ := json.Marshal(errorResponse)
	slog.Debug("Error response", "data", string(data))
	fmt.Fprintf(t.writer, "%s\n", data)
}
