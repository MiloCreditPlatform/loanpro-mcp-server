package tools

import "fmt"

// GetLoanTool returns the get_loan tool definition
func GetLoanTool() Tool {
	return Tool{
		Name:        "get_loan",
		Description: "Get comprehensive loan information by ID including balances, payoff amount, and customer details",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"loan_id": map[string]any{
					"type":        "string",
					"description": "The loan ID to retrieve",
				},
			},
			"required": []string{"loan_id"},
		},
	}
}

// executeGetLoan handles the get_loan tool execution
func (m *Manager) executeGetLoan(arguments map[string]any) MCPResponse {
	loanID := arguments["loan_id"].(string)
	loan, err := m.client.GetLoan(loanID)
	if err != nil {
		LogError("get_loan", err, fmt.Sprintf("for ID %s", loanID))
		return CreateErrorResponse(-1, err.Error(), nil)
	}

	text := fmt.Sprintf("Loan Details:\nID: %s\nDisplay ID: %s\nStatus: %s\nCustomer: %s\nBalance: $%s\nPayoff: $%s",
		loan.GetID(), loan.GetDisplayID(), loan.GetLoanStatus(), loan.GetPrimaryCustomerName(), loan.GetPrincipalBalance(), loan.GetPayoffAmount())

	return CreateSuccessResponse(text, nil)
}