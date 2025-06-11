package transport

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
)

// HTTPTransport handles MCP communication over streamable HTTP
type HTTPTransport struct {
	handler MCPHandler
}

// NewHTTPTransport creates a new HTTP transport
func NewHTTPTransport(handler MCPHandler) *HTTPTransport {
	return &HTTPTransport{
		handler: handler,
	}
}

// HandleMCP handles HTTP POST requests with MCP messages
func (t *HTTPTransport) HandleMCP(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("Error reading HTTP request body", "error", err)
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to read HTTP request body: %v\n", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	slog.Debug("Received HTTP request", "data", string(body))

	// Parse MCP request
	var req MCPRequest
	if err := json.Unmarshal(body, &req); err != nil {
		slog.Error("JSON parse error", "error", err, "input", string(body))
		fmt.Fprintf(os.Stderr, "[ERROR] JSON parse error: %v\nInput: %s\n", err, string(body))
		t.sendError(w, -32700, "Parse error", nil)
		return
	}

	slog.Debug("Processing HTTP request", "method", req.Method, "id", req.ID)

	// Handle the MCP request
	response := t.handler.HandleMCPRequest(req)

	// Don't send response for notifications (empty JSONRPC means no response)
	if response.JSONRPC == "" {
		slog.Debug("Notification processed, no response sent")
		w.WriteHeader(http.StatusOK)
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	responseData, err := json.Marshal(response)
	if err != nil {
		slog.Error("Marshal error", "error", err)
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to marshal response: %v\n", err)
		t.sendError(w, -32603, "Internal error", req.ID)
		return
	}

	slog.Debug("Sending HTTP response", "data", string(responseData))
	w.WriteHeader(http.StatusOK)
	w.Write(responseData)
}

// HandleRoot handles the root endpoint for server info
func (t *HTTPTransport) HandleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	json.NewEncoder(w).Encode(map[string]string{
		"name":      "LoanPro MCP Server",
		"version":   "1.0.0",
		"transport": "http",
	})
}

// HandleHealth handles health check endpoint
func (t *HTTPTransport) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"transport": "http",
	})
}

// sendError sends an error response
func (t *HTTPTransport) sendError(w http.ResponseWriter, code int, message string, id any) {
	slog.Error("Sending HTTP error response", "code", code, "message", message, "id", id)
	
	errorResponse := MCPResponse{
		JSONRPC: "2.0",
		Error: &MCPError{
			Code:    code,
			Message: message,
		},
		ID: id,
	}

	w.Header().Set("Content-Type", "application/json")
	data, _ := json.Marshal(errorResponse)
	slog.Debug("HTTP error response", "data", string(data))
	w.WriteHeader(http.StatusOK) // Still return 200 for JSON-RPC errors
	w.Write(data)
}