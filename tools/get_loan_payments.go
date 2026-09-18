package tools

import "fmt"

// GetLoanPaymentsTool returns the get_loan_payments tool definition
func GetLoanPaymentsTool() Tool {
	return Tool{
		Name: "get_loan_payments",
		Description: "Get payment history for a loan, including payments that were reversed. " +
			"A payment with status Inactive was reversed in the LMS - most often because Secure Payments " +
			"received a failure from the processor and auto-reversed it (an ACH return), though it can also " +
			"be a manual correction by servicing. Reversed payments do NOT appear in get_loan_transactions, " +
			"so use this tool when asking whether a payment failed. The reversal reason and ACH return code " +
			"are not stored in LoanPro - see the SpeedChex return notification for those.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"loan_id": map[string]any{
					"type":        "string",
					"description": "The loan ID to get payment history for",
				},
				"reversed_only": map[string]any{
					"type":        "boolean",
					"description": "When true, return only reversed (Inactive) payments. Use this to check whether a loan has failed payments without reading the full history.",
				},
			},
			"required": []string{"loan_id"},
		},
	}
}

// executeGetLoanPayments handles the get_loan_payments tool execution
func (m *Manager) executeGetLoanPayments(arguments map[string]any) MCPResponse {
	loanID, ok := arguments["loan_id"].(string)
	if !ok {
		return CreateErrorResponse(-1, "loan_id must be a string", nil)
	}

	reversedOnly := false
	if v, ok := arguments["reversed_only"].(bool); ok {
		reversedOnly = v
	}

	payments, err := m.client.GetLoanPayments(loanID)
	if err != nil {
		LogError("get_loan_payments", err, fmt.Sprintf("for loan ID %s", loanID))
		return CreateErrorResponse(-1, err.Error(), nil)
	}

	reversedCount := 0
	for _, payment := range payments {
		if isReversed(payment) {
			reversedCount++
		}
	}

	text := fmt.Sprintf("Payment History for Loan %s:\n", loanID)
	if reversedOnly {
		text = fmt.Sprintf("Reversed Payments for Loan %s:\n", loanID)
	}

	shown := 0
	for _, payment := range payments {
		reversed := isReversed(payment)
		if reversedOnly && !reversed {
			continue
		}
		shown++

		status := payment.GetStatus()
		if reversed {
			status += " (reversed - payment did not stand)"
		}
		text += fmt.Sprintf("- Date: %s, Amount: $%s, ID: %s, Status: %s\n",
			payment.GetDate(), payment.GetAmount(), payment.GetID(), status)
	}

	if shown == 0 {
		if reversedOnly {
			text += "No reversed payments found.\n"
		} else {
			text += "No payments found.\n"
		}
	}

	if !reversedOnly && reversedCount > 0 {
		text += fmt.Sprintf("\n%d of %d payments were reversed. These are absent from get_loan_transactions.\n",
			reversedCount, len(payments))
	}

	return CreateSuccessResponse(text, nil)
}

// isReversed reports whether a payment was reversed in the LMS (Active = 0).
func isReversed(p Payment) bool {
	return p.GetStatus() != "Active"
}
