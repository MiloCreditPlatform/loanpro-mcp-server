package loanpro

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
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
func (c *Client) SearchLoans(searchTerm, status string, limit, offset int) ([]Loan, error) {
	// Build the search query according to LoanPro API format
	searchBody := map[string]any{
		"size": limit,
		"from": offset,
	}

	// Build query conditions
	var mustConditions []map[string]any
	var shouldConditions []map[string]any

	// Add search term conditions if provided
	if searchTerm != "" {
		shouldConditions = append(shouldConditions,
			map[string]any{
				"query_string": map[string]any{
					"query":            "*" + searchTerm + "*",
					"fields":           []string{"displayId", "primaryCustomerName", "title"},
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

const maxPastDueLimit = 200
const odataBatchSize = 50

// GetPastDueLoans retrieves open loans with more than minDaysPastDue days past due.
// Elasticsearch is used to filter by daysPastDue (not available as an OData filter),
// then each matched loan is fetched individually via OData with $expand so that
// LoanSetup and other nested entities are fully populated.
func (c *Client) GetPastDueLoans(minDaysPastDue, limit, offset int) ([]Loan, error) {
	if minDaysPastDue < 0 {
		return nil, fmt.Errorf("minDaysPastDue must be non-negative")
	}
	if limit <= 0 {
		return nil, fmt.Errorf("limit must be positive")
	}
	if offset < 0 {
		return nil, fmt.Errorf("offset must be non-negative")
	}
	if limit > maxPastDueLimit {
		limit = maxPastDueLimit
	}

	searchBody := map[string]any{
		"size": limit,
		"query": map[string]any{
			"bool": map[string]any{
				"must": []map[string]any{
					{
						"range": map[string]any{
							"daysPastDue": map[string]any{
								"gt": minDaysPastDue,
							},
						},
					},
					{
						"match": map[string]any{
							"loanStatusText": "Open",
						},
					},
				},
			},
		},
		"sort": []map[string]any{
			{"daysPastDue": map[string]any{"order": "asc"}},
		},
	}
	if offset > 0 {
		searchBody["from"] = offset
	}

	body, err := c.makePostRequest("/public/api/1/Loans/Autopal.Search()", searchBody)
	if err != nil {
		return nil, err
	}

	var searchResp SearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to parse GetPastDueLoans response: %v\nResponse body: %s\n", err, string(body))
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(searchResp.D.Results) == 0 {
		return []Loan{}, nil
	}

	// Validate IDs and chunk into batches to stay within OData URL length limits.
	// Each batch gets its own $filter=id eq X or id eq Y request with $expand.
	results := searchResp.D.Results
	loans := make([]Loan, 0, len(results))
	for start := 0; start < len(results); start += odataBatchSize {
		end := start + odataBatchSize
		if end > len(results) {
			end = len(results)
		}

		filterParts := make([]string, 0, end-start)
		for _, result := range results[start:end] {
			idStr := strings.TrimSpace(string(result.ID))
			if _, err := strconv.Atoi(idStr); err != nil {
				return nil, fmt.Errorf("invalid loan id in search response: %q", idStr)
			}
			filterParts = append(filterParts, "id eq "+idStr)
		}

		params := map[string]string{
			"$filter": strings.Join(filterParts, " or "),
			"$expand": "LoanSettings,LoanSetup,Customers,StatusArchive",
			"$top":    strconv.Itoa(end - start),
		}

		odataBody, err := c.makeRequest("/public/api/1/odata.svc/Loans", params)
		if err != nil {
			return nil, err
		}

		var batch ODataLoansResponse
		if err := json.Unmarshal(odataBody, &batch); err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] Failed to parse GetPastDueLoans OData response: %v\nResponse body: %s\n", err, string(odataBody))
			return nil, fmt.Errorf("failed to parse OData response: %w", err)
		}
		loans = append(loans, batch.D.Results...)
	}

	if len(loans) == 0 {
		return nil, fmt.Errorf("failed to fetch loan details: Elasticsearch returned %d matches but OData returned no results", len(results))
	}

	// OData doesn't preserve Elasticsearch sort order; re-sort by daysPastDue ascending.
	// Loans with unparseable daysPastDue are pushed to the end.
	sort.Slice(loans, func(i, j int) bool {
		dpdI, errI := strconv.Atoi(loans[i].GetDaysPastDue())
		dpdJ, errJ := strconv.Atoi(loans[j].GetDaysPastDue())
		if errI != nil {
			dpdI = int(^uint(0) >> 1)
		}
		if errJ != nil {
			dpdJ = int(^uint(0) >> 1)
		}
		return dpdI < dpdJ
	})

	return loans, nil
}
