package transport

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// MockMCPHandler implements the MCPHandler interface for testing
type MockMCPHandler struct {
	responses map[string]MCPResponse
}

func (m *MockMCPHandler) HandleMCPRequest(req MCPRequest) MCPResponse {
	if response, exists := m.responses[req.Method]; exists {
		response.ID = req.ID
		return response
	}
	
	// Default response for unknown methods
	return MCPResponse{
		JSONRPC: "2.0",
		Error:   &MCPError{Code: -32601, Message: "Method not found"},
		ID:      req.ID,
	}
}

func createMockHandler() *MockMCPHandler {
	return &MockMCPHandler{
		responses: map[string]MCPResponse{
			"tools/list": {
				JSONRPC: "2.0",
				Result: map[string]any{
					"tools": []map[string]any{
						{
							"name":        "test_tool",
							"description": "A test tool",
						},
					},
				},
			},
			"initialize": {
				JSONRPC: "2.0",
				Result: map[string]any{
					"protocolVersion": "2024-11-05",
					"capabilities":    map[string]any{"tools": map[string]any{}},
					"serverInfo":      map[string]any{"name": "test-server", "version": "1.0.0"},
				},
			},
		},
	}
}

func TestHTTPTransport_HandleMCP_ValidRequest(t *testing.T) {
	handler := createMockHandler()
	transport := NewHTTPTransport(handler)
	
	request := MCPRequest{
		JSONRPC: "2.0",
		Method:  "tools/list",
		ID:      1,
	}
	
	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/mcp", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	transport.HandleMCP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response MCPResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
	
	if response.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC 2.0, got %s", response.JSONRPC)
	}
	
	if response.ID != float64(1) {
		t.Errorf("Expected ID 1, got %v", response.ID)
	}
	
	if response.Error != nil {
		t.Errorf("Expected no error, got %v", response.Error)
	}
}

func TestHTTPTransport_HandleMCP_InvalidJSON(t *testing.T) {
	handler := createMockHandler()
	transport := NewHTTPTransport(handler)
	
	req := httptest.NewRequest("POST", "/mcp", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	transport.HandleMCP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response MCPResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
	
	if response.Error == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
	
	if response.Error.Code != -32700 {
		t.Errorf("Expected error code -32700, got %d", response.Error.Code)
	}
}

func TestHTTPTransport_HandleMCP_MethodNotAllowed(t *testing.T) {
	handler := createMockHandler()
	transport := NewHTTPTransport(handler)
	
	req := httptest.NewRequest("GET", "/mcp", nil)
	w := httptest.NewRecorder()
	transport.HandleMCP(w, req)
	
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestHTTPTransport_HandleMCP_OptionsRequest(t *testing.T) {
	handler := createMockHandler()
	transport := NewHTTPTransport(handler)
	
	req := httptest.NewRequest("OPTIONS", "/mcp", nil)
	w := httptest.NewRecorder()
	transport.HandleMCP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	// Check CORS headers
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("Expected CORS header to be set")
	}
}

func TestHTTPTransport_HandleRoot(t *testing.T) {
	handler := createMockHandler()
	transport := NewHTTPTransport(handler)
	
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	transport.HandleRoot(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
	
	if response["name"] != "LoanPro MCP Server" {
		t.Errorf("Expected name 'LoanPro MCP Server', got %s", response["name"])
	}
	
	if response["transport"] != "http" {
		t.Errorf("Expected transport 'http', got %s", response["transport"])
	}
}

func TestHTTPTransport_HandleHealth(t *testing.T) {
	handler := createMockHandler()
	transport := NewHTTPTransport(handler)
	
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	transport.HandleHealth(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
	
	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got %s", response["status"])
	}
	
	if response["transport"] != "http" {
		t.Errorf("Expected transport 'http', got %s", response["transport"])
	}
}

func TestHTTPTransport_HandleMCP_Notification(t *testing.T) {
	handler := &MockMCPHandler{
		responses: map[string]MCPResponse{
			"notifications/initialized": {
				JSONRPC: "", // Empty JSONRPC indicates notification
			},
		},
	}
	transport := NewHTTPTransport(handler)
	
	request := MCPRequest{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
	}
	
	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/mcp", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	transport.HandleMCP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	// For notifications, no response body should be sent
	if w.Body.Len() != 0 {
		t.Errorf("Expected empty response body for notification, got %s", w.Body.String())
	}
}