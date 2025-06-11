package main

import (
	"os"
	"strings"
	"testing"

	"loanpro-mcp-server/loanpro"
	"loanpro-mcp-server/tools"
	"loanpro-mcp-server/transport"
)

func TestConfigureSlog(t *testing.T) {
	// Test default configuration
	configureSlog()
	
	// Test DEBUG level
	os.Setenv("LOG_LEVEL", "DEBUG")
	configureSlog()
	
	// Test WARN level
	os.Setenv("LOG_LEVEL", "WARN")
	configureSlog()
	
	// Test ERROR level
	os.Setenv("LOG_LEVEL", "ERROR")
	configureSlog()
	
	// Test invalid level
	os.Setenv("LOG_LEVEL", "INVALID")
	configureSlog()
	
	// Test JSON format
	os.Setenv("LOG_FORMAT", "JSON")
	configureSlog()
	
	// Test invalid format
	os.Setenv("LOG_FORMAT", "INVALID")
	configureSlog()
	
	// Clean up
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("LOG_FORMAT")
}

func TestNewMCPServer(t *testing.T) {
	mockClient := &loanpro.Client{}
	server := NewMCPServer(mockClient)
	
	if server == nil {
		t.Error("Expected server to be created")
	}
	
	if server.toolManager == nil {
		t.Error("Expected tool manager to be initialized")
	}
}

func TestClientAdapter_Interface(t *testing.T) {
	mockClient := &loanpro.Client{}
	adapter := &ClientAdapter{client: mockClient}
	
	// Test that adapter implements the interface correctly
	if adapter.client == nil {
		t.Error("Expected client to be set")
	}
}

func TestMCPServer_HandleMCPRequest_Initialize(t *testing.T) {
	mockClient := &loanpro.Client{}
	server := NewMCPServer(mockClient)
	
	req := transport.MCPRequest{
		JSONRPC: "2.0",
		Method:  "initialize",
		Params: map[string]any{
			"protocolVersion": "2024-11-05",
		},
		ID: 1,
	}
	
	response := server.HandleMCPRequest(req)
	
	if response.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC 2.0, got %s", response.JSONRPC)
	}
	
	if response.Error != nil {
		t.Errorf("Expected no error, got %v", response.Error)
	}
	
	result, ok := response.Result.(map[string]any)
	if !ok {
		t.Error("Expected result to be map[string]any")
	}
	
	if result["protocolVersion"] != "2024-11-05" {
		t.Errorf("Expected protocol version 2024-11-05, got %v", result["protocolVersion"])
	}
}

func TestMCPServer_HandleMCPRequest_ToolsList(t *testing.T) {
	mockClient := &loanpro.Client{}
	server := NewMCPServer(mockClient)
	
	req := transport.MCPRequest{
		JSONRPC: "2.0",
		Method:  "tools/list",
		ID:      1,
	}
	
	response := server.HandleMCPRequest(req)
	
	if response.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC 2.0, got %s", response.JSONRPC)
	}
	
	if response.Error != nil {
		t.Errorf("Expected no error, got %v", response.Error)
	}
	
	result, ok := response.Result.(map[string]any)
	if !ok {
		t.Error("Expected result to be map[string]any")
	}
	
	toolsList, ok := result["tools"].([]tools.Tool)
	if !ok {
		t.Error("Expected tools to be []tools.Tool")
	}
	
	if len(toolsList) == 0 {
		t.Error("Expected at least one tool")
	}
}

func TestMCPServer_HandleMCPRequest_ResourcesList(t *testing.T) {
	mockClient := &loanpro.Client{}
	server := NewMCPServer(mockClient)
	
	req := transport.MCPRequest{
		JSONRPC: "2.0",
		Method:  "resources/list",
		ID:      1,
	}
	
	response := server.HandleMCPRequest(req)
	
	if response.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC 2.0, got %s", response.JSONRPC)
	}
	
	if response.Error != nil {
		t.Errorf("Expected no error, got %v", response.Error)
	}
}

func TestMCPServer_HandleMCPRequest_PromptsList(t *testing.T) {
	mockClient := &loanpro.Client{}
	server := NewMCPServer(mockClient)
	
	req := transport.MCPRequest{
		JSONRPC: "2.0",
		Method:  "prompts/list",
		ID:      1,
	}
	
	response := server.HandleMCPRequest(req)
	
	if response.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC 2.0, got %s", response.JSONRPC)
	}
	
	if response.Error != nil {
		t.Errorf("Expected no error, got %v", response.Error)
	}
}

func TestMCPServer_HandleMCPRequest_Initialized(t *testing.T) {
	mockClient := &loanpro.Client{}
	server := NewMCPServer(mockClient)
	
	req := transport.MCPRequest{
		JSONRPC: "2.0",
		Method:  "initialized",
		ID:      1,
	}
	
	response := server.HandleMCPRequest(req)
	
	// Should return empty response for notifications
	if response.JSONRPC != "" {
		t.Errorf("Expected empty JSONRPC for notification, got %s", response.JSONRPC)
	}
}

func TestMCPServer_HandleMCPRequest_NotificationsInitialized(t *testing.T) {
	mockClient := &loanpro.Client{}
	server := NewMCPServer(mockClient)
	
	req := transport.MCPRequest{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
	}
	
	response := server.HandleMCPRequest(req)
	
	// Should return empty response for notifications
	if response.JSONRPC != "" {
		t.Errorf("Expected empty JSONRPC for notification, got %s", response.JSONRPC)
	}
}

func TestMCPServer_HandleMCPRequest_UnknownMethod(t *testing.T) {
	mockClient := &loanpro.Client{}
	server := NewMCPServer(mockClient)
	
	req := transport.MCPRequest{
		JSONRPC: "2.0",
		Method:  "unknown/method",
		ID:      1,
	}
	
	response := server.HandleMCPRequest(req)
	
	if response.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC 2.0, got %s", response.JSONRPC)
	}
	
	if response.Error == nil {
		t.Error("Expected error for unknown method")
	}
	
	if response.Error.Code != -32601 {
		t.Errorf("Expected error code -32601, got %d", response.Error.Code)
	}
	
	if !strings.Contains(response.Error.Message, "Method not found") {
		t.Errorf("Expected 'Method not found' in error message, got %s", response.Error.Message)
	}
}

// Mock implementations for testing ClientAdapter
type MockLoan struct {
	id string
}

func (m MockLoan) GetID() string                  { return m.id }
func (m MockLoan) GetDisplayID() string           { return "LN123" }
func (m MockLoan) GetPrimaryCustomerName() string { return "Test Customer" }
func (m MockLoan) GetLoanStatus() string          { return "Active" }
func (m MockLoan) GetPrincipalBalance() string    { return "1000.00" }

type MockCustomer struct {
	id int
}

func (m MockCustomer) GetID() int             { return m.id }
func (m MockCustomer) GetFirstName() string   { return "John" }
func (m MockCustomer) GetLastName() string    { return "Doe" }
func (m MockCustomer) GetEmail() string       { return "john@example.com" }
func (m MockCustomer) GetPhone() string       { return "555-1234" }
func (m MockCustomer) GetCreatedDate() string { return "2025-01-01" }

type MockPayment struct {
	id string
}

func (m MockPayment) GetID() string     { return m.id }
func (m MockPayment) GetAmount() string { return "100.00" }
func (m MockPayment) GetDate() string   { return "2025-01-01" }