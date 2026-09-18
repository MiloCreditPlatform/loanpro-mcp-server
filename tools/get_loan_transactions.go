package tools

import "fmt"

// GetLoanTransactionsTool returns the get_loan_transactions tool definition
func GetLoanTransactionsTool() Tool {
	return Tool{
		Name:        "get_loan_transactions",
		Description: "Get detailed transaction history for a loan including payments, charges, credits, and adjustments with payment application breakdown. Supports pagination to retrieve results in batches. Returns pagination metadata (total count, has_more flag). NOTE: LoanPro excludes reversed payments from this endpoint, so a payment that failed and was auto-reversed will be missing entirely rather than shown as failed. Any such payments are listed in a separate section at the end of the output; use get_loan_payments for the full picture.",
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

	text += m.reversedPaymentsNote(loanID)

	return CreateSuccessResponse(text, nil)
}

// reversedPaymentsNote returns a section listing payments that were reversed in
// the LMS. LoanPro omits these from the Transactions endpoint, which makes a
// failed payment look like a payment that was never attempted. Failure to fetch
// them is not fatal - the transaction list is still worth returning - so this
// returns an empty string on error.
func (m *Manager) reversedPaymentsNote(loanID string) string {
	payments, err := m.client.GetLoanPayments(loanID)
	if err != nil {
		LogError("get_loan_transactions", err, fmt.Sprintf("fetching reversed payments for loan ID %s", loanID))
		return ""
	}

	var reversed []Payment
	for _, payment := range payments {
		if isReversed(payment) {
			reversed = append(reversed, payment)
		}
	}
	if len(reversed) == 0 {
		return ""
	}

	note := "\nReversed payments (excluded from the transaction list above):\n"
	for _, payment := range reversed {
		note += fmt.Sprintf("- Date: %s, Amount: $%s, ID: %s, Status: Reversed\n",
			payment.GetDate(), payment.GetAmount(), payment.GetID())
	}
	note += "These payments did not stand. LoanPro does not record why; check the processor's return notification for the ACH return code.\n"
	return note
}
