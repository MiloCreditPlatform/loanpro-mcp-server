package loanpro

import (
	"encoding/json"
	"fmt"
	"os"
)

// GetLoan retrieves a loan by ID with expanded data
func (c *Client) GetLoan(loanID string) (*Loan, error) {
	// Use OData expand to include related data that provides loan amounts, status, and customer info
	params := map[string]string{
		"$expand": "LoanSettings,LoanSetup,Customers,StatusArchive",
	}
	
	body, err := c.makeRequest("/public/api/1/odata.svc/Loans("+loanID+")", params)
	if err != nil {
		return nil, err
	}

	var response ODataResponse
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to parse GetLoan response: %v\nResponse body: %s\n", err, string(body))
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	loanData, err := json.Marshal(response.D)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to marshal loan data: %v\n", err)
		return nil, fmt.Errorf("failed to marshal loan data: %w", err)
	}

	var loan Loan
	if err := json.Unmarshal(loanData, &loan); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to parse loan struct: %v\nLoan data: %s\n", err, string(loanData))
		return nil, fmt.Errorf("failed to parse loan: %w", err)
	}

	return &loan, nil
}

// SearchLoans searches for loans using the search API
func (c *Client) SearchLoans(searchTerm, status string, limit int) ([]Loan, error) {
	// Build the search query according to LoanPro API format
	searchBody := map[string]any{
		"size": limit, // Use 'size' for pagination limit
	}
	
	// Build query conditions
	var mustConditions []map[string]any
	var shouldConditions []map[string]any
	
	// Add search term conditions if provided
	if searchTerm != "" {
		shouldConditions = append(shouldConditions, 
			map[string]any{
				"query_string": map[string]any{
					"query":   "*" + searchTerm + "*",
					"fields":  []string{"displayId", "primaryCustomerName", "title"},
					"default_operator": "and",
				},
			},
			map[string]any{
				"match": map[string]any{
					"displayId": searchTerm,
				},
			},
			map[string]any{
				"match": map[string]any{
					"primaryCustomerName": searchTerm,
				},
			},
		)
	}
	
	// Add status filter if provided
	if status != "" {
		mustConditions = append(mustConditions, map[string]any{
			"match": map[string]any{
				"loanStatusText": status,
			},
		})
	}
	
	// Build the final query
	if len(mustConditions) > 0 || len(shouldConditions) > 0 {
		boolQuery := map[string]any{}
		
		if len(mustConditions) > 0 {
			if len(mustConditions) == 1 {
				boolQuery["must"] = mustConditions[0]
			} else {
				boolQuery["must"] = mustConditions
			}
		}
		
		if len(shouldConditions) > 0 {
			boolQuery["should"] = shouldConditions
			boolQuery["minimum_should_match"] = 1
		}
		
		searchBody["query"] = map[string]any{
			"bool": boolQuery,
		}
	} else {
		// If no filters, use match_all query
		searchBody["query"] = map[string]any{
			"match_all": map[string]any{},
		}
	}

	body, err := c.makePostRequest("/public/api/1/Loans/Autopal.Search()", searchBody)
	if err != nil {
		return nil, err
	}

	var response SearchResponse
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to parse SearchLoans response: %v\nResponse body: %s\n", err, string(body))
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return response.D.Results, nil
}