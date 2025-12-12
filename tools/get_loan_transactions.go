package tools

import "fmt"

// GetLoanTransactionsTool returns the get_loan_transactions tool definition
func GetLoanTransactionsTool() Tool {
	return Tool{
		Name:        "get_loan_transactions",
		Description: "Get detailed transaction history for a loan including payments, charges, credits, and adjustments with payment application breakdown. Supports pagination to retrieve results in batches. Returns pagination metadata (total count, has_more flag).",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"loan_id": map[string]any{
					"type":        "string",
					"description": "The loan ID to get transaction history for",
				},
				"limit": map[string]any{
					"type":        "number",
					"description": "Maximum number of transactions to return per page. If not specified, returns all transactions. Recommended: 50-100 for large transaction histories.",
				},
				"offset": map[string]any{
					"type":        "number",
					"description": "Number of transactions to skip (pagination). Use with 'limit' for pagination. For example: offset=0 gets first page, offset=50 gets second page (with limit=50).",
				},
			},
			"required": []string{"loan_id"},
		},
	}
}

// executeGetLoanTransactions handles the get_loan_transactions tool execution
func (m *Manager) executeGetLoanTransactions(arguments map[string]any) MCPResponse {
	loanID := arguments["loan_id"].(string)

	// Get pagination parameters if provided
	var limit, offset int
	if limitValue, ok := arguments["limit"]; ok {
		if limitFloat, ok := limitValue.(float64); ok {
			limit = int(limitFloat)
		}
	}
	if offsetValue, ok := arguments["offset"]; ok {
		if offsetFloat, ok := offsetValue.(float64); ok {
			offset = int(offsetFloat)
		}
	}

	// Call the appropriate method based on whether pagination is requested
	var transactions []Transaction
	var err error

	if limit > 0 || offset > 0 {
		// Use pagination
		opts := &TransactionOptions{
			Limit:  limit,
			Offset: offset,
		}
		transactions, err = m.client.GetLoanTransactionsWithOptions(loanID, opts)
	} else {
		// No pagination
		transactions, err = m.client.GetLoanTransactions(loanID)
	}

	if err != nil {
		LogError("get_loan_transactions", err, fmt.Sprintf("for loan ID %s", loanID))
		return CreateErrorResponse(-1, err.Error(), nil)
	}

	// Build response text with pagination info
	text := fmt.Sprintf("Transaction History for Loan %s:\n", loanID)
	if limit > 0 {
		text += fmt.Sprintf("(Showing up to %d transactions, starting at offset %d)\n\n", limit, offset)
	}
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
