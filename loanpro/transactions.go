package loanpro

import (
	"encoding/json"
	"fmt"
	"os"
)

// GetLoanTransactions retrieves transaction history for a loan
func (c *Client) GetLoanTransactions(loanID string) ([]Transaction, error) {
	// Use the Transactions endpoint directly
	endpoint := fmt.Sprintf("/public/api/1/Loans(%s)/Transactions", loanID)
	
	body, err := c.makeRequest(endpoint, nil)
	if err != nil {
		return nil, err
	}

	var response ODataResponse
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to parse GetLoanTransactions response: %v\nResponse body: %s\n", err, string(body))
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// The response.D could be a single object or an array
	// Try to marshal and unmarshal to handle both cases
	transactionsData, err := json.Marshal(response.D)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to marshal transaction data: %v\n", err)
		return nil, fmt.Errorf("failed to marshal transaction data: %w", err)
	}

	// Try parsing as a wrapper first
	var transactionsWrapper TransactionsWrapper
	if err := json.Unmarshal(transactionsData, &transactionsWrapper); err == nil && len(transactionsWrapper.Results) > 0 {
		return transactionsWrapper.Results, nil
	}

	// Try parsing as a direct array
	var transactionsArray []Transaction
	if err := json.Unmarshal(transactionsData, &transactionsArray); err == nil {
		return transactionsArray, nil
	}

	// If neither works, log and return empty
	fmt.Fprintf(os.Stderr, "[DEBUG] Could not parse transactions, returning empty array. Data: %s\n", string(transactionsData))
	return []Transaction{}, nil
}
