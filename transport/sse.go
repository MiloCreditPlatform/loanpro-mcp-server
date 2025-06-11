package transport

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// SSETransport handles MCP communication over Server-Sent Events
type SSETransport struct {
	handler MCPHandler
}

// NewSSETransport creates a new SSE transport
func NewSSETransport(handler MCPHandler) *SSETransport {
	return &SSETransport{
		handler: handler,
	}
}

// HandleSSE handles SSE connections for MCP communication
func (t *SSETransport) HandleSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "event: ready\n")
	fmt.Fprintf(w, "data: {\"type\":\"ready\"}\n\n")
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		default:
			var req MCPRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				continue
			}

			response := t.handler.HandleMCPRequest(req)
			data, _ := json.Marshal(response)

			fmt.Fprintf(w, "event: message\n")
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}

// HandleRoot handles the root endpoint for server info
func (t *SSETransport) HandleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"name":      "LoanPro MCP Server",
		"version":   "1.0.0",
		"transport": "sse",
	})
}