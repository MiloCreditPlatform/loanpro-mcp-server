package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"loanpro-mcp-server/loanpro"
	"loanpro-mcp-server/tools"
	"loanpro-mcp-server/transport"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

// configureSlog sets up structured logging based on environment variables
func configureSlog() {
	// Default to INFO level
	level := slog.LevelInfo

	// Parse LOG_LEVEL environment variable
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		switch strings.ToUpper(logLevel) {
		case "DEBUG":
			level = slog.LevelDebug
		case "INFO":
			level = slog.LevelInfo
		case "WARN", "WARNING":
			level = slog.LevelWarn
		case "ERROR":
			level = slog.LevelError
		default:
			fmt.Fprintf(os.Stderr, "Invalid LOG_LEVEL '%s', using INFO\n", logLevel)
		}
	}

	// Configure log format based on LOG_FORMAT environment variable
	format := strings.ToUpper(os.Getenv("LOG_FORMAT"))

	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: level,
	}

	switch format {
	case "JSON":
		handler = slog.NewJSONHandler(os.Stderr, opts)
	case "TEXT", "":
		handler = slog.NewTextHandler(os.Stderr, opts)
	default:
		fmt.Fprintf(os.Stderr, "Invalid LOG_FORMAT '%s', using TEXT\n", format)
		handler = slog.NewTextHandler(os.Stderr, opts)
	}

	// Set the default logger
	slog.SetDefault(slog.New(handler))

	slog.Info("Logger configured",
		"level", level.String(),
		"format", strings.ToLower(format))
}

// MCPServer implements the MCP protocol handler
type MCPServer struct {
	toolManager *tools.Manager
}

// NewMCPServer creates a new MCP server
func NewMCPServer(loanProClient *loanpro.Client) *MCPServer {
	return &MCPServer{
		toolManager: tools.NewManager(&ClientAdapter{client: loanProClient}),
	}
}

// ClientAdapter adapts the loanpro.Client to implement the tools.LoanProClient interface
type ClientAdapter struct {
	client *loanpro.Client
}

func (ca *ClientAdapter) GetLoan(id string) (tools.Loan, error) {
	loan, err := ca.client.GetLoan(id)
	if err != nil {
		return nil, err
	}
	return loan, nil
}

func (ca *ClientAdapter) SearchLoans(searchTerm, status string, limit int) ([]tools.Loan, error) {
	loans, err := ca.client.SearchLoans(searchTerm, status, limit)
	if err != nil {
		return nil, err
	}

	result := make([]tools.Loan, len(loans))
	for i, loan := range loans {
		result[i] = &loan
	}
	return result, nil
}

func (ca *ClientAdapter) GetCustomer(id string) (tools.Customer, error) {
	customer, err := ca.client.GetCustomer(id)
	if err != nil {
		return nil, err
	}
	return customer, nil
}

func (ca *ClientAdapter) SearchCustomers(searchTerm string, limit int) ([]tools.Customer, error) {
	customers, err := ca.client.SearchCustomers(searchTerm, limit)
	if err != nil {
		return nil, err
	}

	result := make([]tools.Customer, len(customers))
	for i, customer := range customers {
		result[i] = &customer
	}
	return result, nil
}

func (ca *ClientAdapter) GetLoanPayments(loanID string) ([]tools.Payment, error) {
	payments, err := ca.client.GetLoanPayments(loanID)
	if err != nil {
		return nil, err
	}

	result := make([]tools.Payment, len(payments))
	for i, payment := range payments {
		result[i] = &payment
	}
	return result, nil
}

func (ca *ClientAdapter) GetLoanTransactions(loanID string) ([]tools.Transaction, error) {
	transactions, err := ca.client.GetLoanTransactions(loanID)
	if err != nil {
		return nil, err
	}

	result := make([]tools.Transaction, len(transactions))
	for i, transaction := range transactions {
		result[i] = &transaction
	}
	return result, nil
}

// HandleMCPRequest handles MCP protocol requests
func (s *MCPServer) HandleMCPRequest(req transport.MCPRequest) transport.MCPResponse {
	switch req.Method {
	case "initialize":
		slog.Info("Processing initialize request", "method", req.Method)

		// Extract client's protocol version from params
		clientProtocolVersion := "2024-11-05" // fallback default
		if params, ok := req.Params["protocolVersion"].(string); ok {
			clientProtocolVersion = params
			slog.Info("Client protocol version", "version", clientProtocolVersion)
		}

		response := transport.MCPResponse{
			JSONRPC: "2.0",
			Result: map[string]any{
				"protocolVersion": clientProtocolVersion, // Use client's version
				"capabilities": map[string]any{
					"tools": map[string]any{},
				},
				"serverInfo": map[string]any{
					"name":    "loanpro-mcp-server",
					"version": "1.0.0",
				},
			},
			ID: req.ID,
		}

		slog.Info("Responding with protocol version", "version", clientProtocolVersion)
		return response

	case "initialized":
		slog.Debug("Received initialized notification (legacy)")
		// This is a notification, no response needed
		return transport.MCPResponse{} // Empty response indicates no reply

	case "notifications/initialized": // Changed from "initialized"
		slog.Debug("Received initialized notification")
		return transport.MCPResponse{} // No response for notifications

	case "resources/list":
		return transport.MCPResponse{
			JSONRPC: "2.0",
			Result: map[string]any{
				"resources": []any{}, // Empty resources list
			},
			ID: req.ID,
		}

	case "prompts/list":
		return transport.MCPResponse{
			JSONRPC: "2.0",
			Result: map[string]any{
				"prompts": []any{}, // Empty prompts list
			},
			ID: req.ID,
		}

	case "tools/list":
		toolsList := s.toolManager.GetAllTools()
		return transport.MCPResponse{
			JSONRPC: "2.0",
			Result: map[string]any{
				"tools": toolsList,
			},
			ID: req.ID,
		}

	case "tools/call":
		toolName := req.Params["name"].(string)
		arguments := req.Params["arguments"].(map[string]any)

		response := s.toolManager.ExecuteTool(toolName, arguments)
		// Convert tools.MCPResponse to transport.MCPResponse
		return transport.MCPResponse{
			JSONRPC: response.JSONRPC,
			Result:  response.Result,
			Error: func() *transport.MCPError {
				if response.Error != nil {
					return &transport.MCPError{
						Code:    response.Error.Code,
						Message: response.Error.Message,
					}
				}
				return nil
			}(),
			ID: req.ID,
		}

	default:
		return transport.MCPResponse{
			JSONRPC: "2.0",
			Error:   &transport.MCPError{Code: -32601, Message: "Method not found"},
			ID:      req.ID,
		}
	}
}

func main() {
	stdioMode := flag.Bool("stdio", false, "Use stdio transport instead of HTTP/SSE")
	transportType := flag.String("transport", "http", "Transport type: stdio, sse, or http")
	flag.Parse()

	godotenv.Load()

	// Configure structured logging
	configureSlog()

	loanProClient := loanpro.NewClient(
		os.Getenv("LOANPRO_API_URL"),
		os.Getenv("LOANPRO_API_KEY"),
		os.Getenv("LOANPRO_TENANT_ID"),
	)

	server := NewMCPServer(loanProClient)

	// Handle stdio mode for backwards compatibility
	if *stdioMode {
		*transportType = "stdio"
	}

	switch *transportType {
	case "stdio":
		// Run in stdio mode for MCP clients
		slog.Info("Starting MCP server", "transport", "stdio")
		stdioTransport := transport.NewStdioTransport(server)
		if err := stdioTransport.Run(); err != nil {
			slog.Error("Stdio transport failed", "error", err)
			log.Fatal(err)
		}

	case "sse":
		// Run HTTP server with SSE transport
		slog.Info("Starting MCP server", "transport", "sse")
		r := mux.NewRouter()
		sseTransport := transport.NewSSETransport(server)
		r.HandleFunc("/sse", sseTransport.HandleSSE).Methods("GET")
		r.HandleFunc("/", sseTransport.HandleRoot).Methods("GET")

		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}

		slog.Info("MCP Server starting", "transport", "sse", "port", port)
		log.Fatal(http.ListenAndServe(":"+port, r))

	case "http":
		// Run HTTP server with streamable HTTP transport
		slog.Info("Starting MCP server", "transport", "http")
		r := mux.NewRouter()
		httpTransport := transport.NewHTTPTransport(server)

		// MCP endpoints
		r.HandleFunc("/mcp", httpTransport.HandleMCP).Methods("POST", "OPTIONS")

		// Info endpoints
		r.HandleFunc("/", httpTransport.HandleRoot).Methods("GET")
		r.HandleFunc("/health", httpTransport.HandleHealth).Methods("GET")

		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}

		slog.Info("MCP Server starting",
			"transport", "http",
			"port", port,
			"endpoints", map[string]string{
				"POST /mcp":   "MCP requests",
				"GET /":       "Server info",
				"GET /health": "Health check",
			})
		log.Fatal(http.ListenAndServe(":"+port, r))

	default:
		slog.Error("Unknown transport type", "type", *transportType, "valid", []string{"stdio", "sse", "http"})
		log.Fatalf("Unknown transport type: %s. Use stdio, sse, or http", *transportType)
	}
}
