package loanpro

import (
	"encoding/json"
	"fmt"
	"os"
)

// GetLoanPayments retrieves payment history for a loan
func (c *Client) GetLoanPayments(loanID string) ([]Payment, error) {
	// Use OData expand to get payment history
	params := map[string]string{
		"$expand": "Payments",
	}

	body, err := c.makeRequest("/public/api/1/odata.svc/Loans("+loanID+")", params)
	if err != nil {
		return nil, err
	}

	var response ODataResponse
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to parse GetLoanPayments response: %v\nResponse body: %s\n", err, string(body))
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	loanData, err := json.Marshal(response.D)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to marshal loan data: %v\n", err)
		return nil, fmt.Errorf("failed to marshal loan data: %w", err)
	}

	var loanWithPayments struct {
		Payments *PaymentsWrapper `json:"Payments,omitempty"`
	}

	if err := json.Unmarshal(loanData, &loanWithPayments); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to parse loan payments: %v\nLoan data: %s\n", err, string(loanData))
		return nil, fmt.Errorf("failed to parse loan payments: %w", err)
	}

	if loanWithPayments.Payments != nil {
		return loanWithPayments.Payments.Results, nil
	}

	return []Payment{}, nil
}
