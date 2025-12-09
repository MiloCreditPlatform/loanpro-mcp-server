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
				breakdownParts := []struct {
					label  string
					amount string
				}{
					{"Principal", txn.GetPrincipalAmount()},
					{"Interest", txn.GetInterestAmount()},
					{"Fees", txn.GetFeesAmount()},
					{"Escrow", txn.GetEscrowAmount()},
				}
				
				for _, part := range breakdownParts {
					if part.amount != "" && part.amount != "0" && part.amount != "0.00" {
						text += fmt.Sprintf(" %s: $%s", part.label, part.amount)
					}
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
