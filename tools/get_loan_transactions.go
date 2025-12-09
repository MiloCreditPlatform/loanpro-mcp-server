package tools

import "fmt"

// GetLoanTransactionsTool returns the get_loan_transactions tool definition
func GetLoanTransactionsTool() Tool {
	return Tool{
		Name:        "get_loan_transactions",
		Description: "Get detailed transaction history for a loan including payments, charges, credits, and adjustments with payment application breakdown",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"loan_id": map[string]any{
					"type":        "string",
					"description": "The loan ID to get transaction history for",
				},
			},
			"required": []string{"loan_id"},
		},
	}
}

// executeGetLoanTransactions handles the get_loan_transactions tool execution
func (m *Manager) executeGetLoanTransactions(arguments map[string]any) MCPResponse {
	loanID := arguments["loan_id"].(string)
	transactions, err := m.client.GetLoanTransactions(loanID)
	if err != nil {
		LogError("get_loan_transactions", err, fmt.Sprintf("for loan ID %s", loanID))
		return CreateErrorResponse(-1, err.Error(), nil)
	}

	text := fmt.Sprintf("Transaction History for Loan %s:\n", loanID)
	if len(transactions) == 0 {
		text += "No transactions found.\n"
	} else {
		for _, txn := range transactions {
			// Basic transaction info
			text += fmt.Sprintf("- Date: %s, Type: %s, Amount: $%s, ID: %s, Status: %s\n",
				txn.GetDate(), txn.GetType(), txn.GetAmount(), txn.GetID(), txn.GetStatus())
			
			// Add title/description if available
			if txn.GetTitle() != "" {
				text += fmt.Sprintf("  Title: %s\n", txn.GetTitle())
			}
			
			// Add payment breakdown if available
			if txn.HasPaymentBreakdown() {
				text += "  Applied:"
				if txn.GetPrincipalAmount() != "" && txn.GetPrincipalAmount() != "0" && txn.GetPrincipalAmount() != "0.00" {
					text += fmt.Sprintf(" Principal: $%s", txn.GetPrincipalAmount())
				}
				if txn.GetInterestAmount() != "" && txn.GetInterestAmount() != "0" && txn.GetInterestAmount() != "0.00" {
					text += fmt.Sprintf(" Interest: $%s", txn.GetInterestAmount())
				}
				if txn.GetFeesAmount() != "" && txn.GetFeesAmount() != "0" && txn.GetFeesAmount() != "0.00" {
					text += fmt.Sprintf(" Fees: $%s", txn.GetFeesAmount())
				}
				if txn.GetEscrowAmount() != "" && txn.GetEscrowAmount() != "0" && txn.GetEscrowAmount() != "0.00" {
					text += fmt.Sprintf(" Escrow: $%s", txn.GetEscrowAmount())
				}
				text += "\n"
			}
			
			// Add info if available
			if txn.GetInfo() != "" {
				text += fmt.Sprintf("  Info: %s\n", txn.GetInfo())
			}
		}
	}

	return CreateSuccessResponse(text, nil)
}
