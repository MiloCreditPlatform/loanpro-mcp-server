package loanpro

import (
	"encoding/json"
	"fmt"
	"os"
)

// GetCustomer retrieves a customer by ID
func (c *Client) GetCustomer(customerID string) (*Customer, error) {
	body, err := c.makeRequest("/public/api/1/odata.svc/Customers("+customerID+")", nil)
	if err != nil {
		return nil, err
	}

	var response ODataResponse
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to parse GetCustomer response: %v\nResponse body: %s\n", err, string(body))
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	customerData, err := json.Marshal(response.D)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to marshal customer data: %v\n", err)
		return nil, fmt.Errorf("failed to marshal customer data: %w", err)
	}

	var customer Customer
	if err := json.Unmarshal(customerData, &customer); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to parse customer struct: %v\nCustomer data: %s\n", err, string(customerData))
		return nil, fmt.Errorf("failed to parse customer: %w", err)
	}

	return &customer, nil
}

// SearchCustomers searches for customers using the search API
func (c *Client) SearchCustomers(searchTerm string, limit int) ([]Customer, error) {
	// Build the search query according to LoanPro Customer Search API format
	searchBody := map[string]any{
		"size": limit, // Use 'size' for pagination limit
	}

	// Add query filters if provided
	if searchTerm != "" {
		searchBody["query"] = map[string]any{
			"bool": map[string]any{
				"should": []map[string]any{
					{
						"query_string": map[string]any{
							"query":            "*" + searchTerm + "*",
							"fields":           []string{"firstName", "lastName", "email", "ssn", "companyName"},
							"default_operator": "and",
						},
					},
					{
						"match": map[string]any{
							"firstName": searchTerm,
						},
					},
					{
						"match": map[string]any{
							"lastName": searchTerm,
						},
					},
				},
				"minimum_should_match": 1,
			},
		}
	} else {
		// If no filter, use match_all query
		searchBody["query"] = map[string]any{
			"match_all": map[string]any{},
		}
	}

	body, err := c.makePostRequest("/public/api/1/Customers/Autopal.Search()", searchBody)
	if err != nil {
		return nil, err
	}

	var response CustomerSearchResponse
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to parse SearchCustomers response: %v\nResponse body: %s\n", err, string(body))
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return response.D.Results, nil
}
