package tools

import "fmt"

// GetLoanPaymentsTool returns the get_loan_payments tool definition
func GetLoanPaymentsTool() Tool {
	return Tool{
		Name:        "get_loan_payments",
		Description: "Get payment history for a loan",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"loan_id": map[string]any{
					"type":        "string",
					"description": "The loan ID to get payment history for",
				},
			},
			"required": []string{"loan_id"},
		},
	}
}

// executeGetLoanPayments handles the get_loan_payments tool execution
func (m *Manager) executeGetLoanPayments(arguments map[string]any) MCPResponse {
	loanID := arguments["loan_id"].(string)
	payments, err := m.client.GetLoanPayments(loanID)
	if err != nil {
		LogError("get_loan_payments", err, fmt.Sprintf("for loan ID %s", loanID))
		return CreateErrorResponse(-1, err.Error(), nil)
	}

	text := fmt.Sprintf("Payment History for Loan %s:\n", loanID)
	if len(payments) == 0 {
		text += "No payments found.\n"
	} else {
		for _, payment := range payments {
			text += fmt.Sprintf("- Date: %s, Amount: $%s, ID: %s\n", 
				payment.GetDate(), payment.GetAmount(), payment.GetID())
		}
	}

	return CreateSuccessResponse(text, nil)
}