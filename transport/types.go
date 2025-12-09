package transport

// MCPRequest represents a request in the MCP protocol
type MCPRequest struct {
	JSONRPC string         `json:"jsonrpc"`
	Method  string         `json:"method"`
	Params  map[string]any `json:"params,omitempty"`
	ID      any            `json:"id"`
}

// MCPResponse represents a response in the MCP protocol
type MCPResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	Result  any       `json:"result,omitempty"`
	Error   *MCPError `json:"error,omitempty"`
	ID      any       `json:"id"`
}

// MCPError represents an error in the MCP protocol
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// MCPHandler interface for handling MCP requests
type MCPHandler interface {
	HandleMCPRequest(req MCPRequest) MCPResponse
}
