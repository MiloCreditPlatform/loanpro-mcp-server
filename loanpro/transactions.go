package loanpro

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

// TransactionOptions contains pagination and filtering options for transactions
type TransactionOptions struct {
	Limit  int // Maximum number of records to return ($top parameter)
	Offset int // Number of records to skip ($skip parameter)
}

// TransactionResult contains transactions and pagination metadata
type TransactionResult struct {
	Transactions []Transaction
	Total        int  // Total number of transactions available
	HasMore      bool // Whether there are more transactions beyond this page
}

// GetLoanTransactions retrieves transaction history for a loan
func (c *Client) GetLoanTransactions(loanID string) ([]Transaction, error) {
	return c.GetLoanTransactionsWithOptions(loanID, nil)
}

// GetLoanTransactionsWithOptions retrieves transaction history for a loan with pagination options
func (c *Client) GetLoanTransactionsWithOptions(loanID string, opts *TransactionOptions) ([]Transaction, error) {
	result, err := c.GetLoanTransactionsWithMetadata(loanID, opts)
	if err != nil {
		return nil, err
	}
	return result.Transactions, nil
}

// GetLoanTransactionsWithMetadata retrieves transaction history with pagination metadata
func (c *Client) GetLoanTransactionsWithMetadata(loanID string, opts *TransactionOptions) (*TransactionResult, error) {
	// Use the Transactions endpoint directly
	endpoint := fmt.Sprintf("/public/api/1/odata.svc/Loans(%s)/Transactions", loanID)

	// Build query parameters
	params := make(map[string]string)
	if opts != nil {
		if opts.Limit > 0 {
			params["$top"] = strconv.Itoa(opts.Limit)
		}
		if opts.Offset > 0 {
			params["$skip"] = strconv.Itoa(opts.Offset)
		}
	}

	body, err := c.makeRequest(endpoint, params)
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
	if err := json.Unmarshal(transactionsData, &transactionsWrapper); err == nil {
		// Convert json.Number to int for Total
		total := 0
		if totalInt, err := transactionsWrapper.Summary.Total.Int64(); err == nil {
			total = int(totalInt)
		}

		fmt.Fprintf(os.Stderr, "[DEBUG] Successfully parsed wrapper. Results count: %d, Summary.Total: %d\n", len(transactionsWrapper.Results), total)

		// Always return the wrapper result if it parsed successfully
		// (even if Results is empty, the wrapper structure was present)
		result := &TransactionResult{
			Transactions: transactionsWrapper.Results,
			Total:        total,
		}

		// Calculate if there are more results
		if opts != nil && opts.Limit > 0 {
			result.HasMore = opts.Offset+len(transactionsWrapper.Results) < total
		}

		return result, nil
	}

	// Try parsing as a direct array
	var transactionsArray []Transaction
	if err := json.Unmarshal(transactionsData, &transactionsArray); err == nil {
		return &TransactionResult{
			Transactions: transactionsArray,
			Total:        len(transactionsArray),
			HasMore:      false,
		}, nil
	}

	// If neither works, log and return empty
	fmt.Fprintf(os.Stderr, "[DEBUG] Could not parse transactions, returning empty array. Data: %s\n", string(transactionsData))
	return &TransactionResult{
		Transactions: []Transaction{},
		Total:        0,
		HasMore:      false,
	}, nil
}
