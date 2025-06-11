package tools

// Manager handles MCP tool operations
type Manager struct {
	client LoanProClient
}

// NewManager creates a new tool manager
func NewManager(client LoanProClient) *Manager {
	return &Manager{
		client: client,
	}
}

// GetAllTools returns all available MCP tools
func (m *Manager) GetAllTools() []Tool {
	return []Tool{
		GetLoanTool(),
		SearchLoansTool(),
		GetCustomerTool(),
		SearchCustomersTool(),
		GetLoanPaymentsTool(),
	}
}

// ExecuteTool executes the specified tool with given arguments
func (m *Manager) ExecuteTool(toolName string, arguments map[string]any) MCPResponse {
	switch toolName {
	case "get_loan":
		return m.executeGetLoan(arguments)
	case "search_loans":
		return m.executeSearchLoans(arguments)
	case "get_customer":
		return m.executeGetCustomer(arguments)
	case "search_customers":
		return m.executeSearchCustomers(arguments)
	case "get_loan_payments":
		return m.executeGetLoanPayments(arguments)
	default:
		return MCPResponse{
			JSONRPC: "2.0",
			Error:   &MCPError{Code: -32601, Message: "Tool not found"},
		}
	}
}