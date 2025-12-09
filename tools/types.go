package tools

import (
	"fmt"
	"log/slog"
	"os"
)

// Tool represents an MCP tool definition
type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

// MCPResponse represents a response to an MCP request
type MCPResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	Result  any       `json:"result,omitempty"`
	Error   *MCPError `json:"error,omitempty"`
	ID      any       `json:"id"`
}

// MCPError represents an error in MCP protocol
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// LoanProClient interface for dependency injection
type LoanProClient interface {
	GetLoan(id string) (Loan, error)
	SearchLoans(searchTerm, status string, limit int) ([]Loan, error)
	GetCustomer(id string) (Customer, error)
	SearchCustomers(searchTerm string, limit int) ([]Customer, error)
	GetLoanPayments(loanID string) ([]Payment, error)
	GetLoanTransactions(loanID string) ([]Transaction, error)
}

// Loan represents loan data - simplified interface for tools
type Loan interface {
	GetID() string
	GetDisplayID() string
	GetPrimaryCustomerName() string
	GetLoanStatus() string
	GetPrincipalBalance() string
	GetPayoffAmount() string
}

// Customer represents customer data - simplified interface for tools
type Customer interface {
	GetID() int
	GetFirstName() string
	GetLastName() string
	GetEmail() string
	GetPhone() string
	GetCreatedDate() string
}

// Payment represents payment data - simplified interface for tools
type Payment interface {
	GetID() string
	GetAmount() string
	GetDate() string
	GetStatus() string
}

// Transaction represents transaction data - simplified interface for tools
type Transaction interface {
	GetID() string
	GetAmount() string
	GetDate() string
	GetType() string
	GetTitle() string
	GetInfo() string
	GetStatus() string
	GetPrincipalAmount() string
	GetInterestAmount() string
	GetFeesAmount() string
	GetEscrowAmount() string
	HasPaymentBreakdown() bool
}

// Helper function to create error responses
func CreateErrorResponse(code int, message string, id any) MCPResponse {
	return MCPResponse{
		JSONRPC: "2.0",
		Error:   &MCPError{Code: code, Message: message},
		ID:      id,
	}
}

// Helper function to create success responses
func CreateSuccessResponse(text string, id any) MCPResponse {
	return MCPResponse{
		JSONRPC: "2.0",
		Result: map[string]any{
			"content": []map[string]any{
				{
					"type": "text",
					"text": text,
				},
			},
		},
		ID: id,
	}
}

// Helper function to log errors to stderr
func LogError(toolName string, err error, details string) {
	slog.Error("Tool execution failed", "tool", toolName, "error", err, "details", details)
	fmt.Fprintf(os.Stderr, "[ERROR] %s failed %s: %v\n", toolName, details, err)
}
