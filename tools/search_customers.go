package tools

import "fmt"

// SearchCustomersTool returns the search_customers tool definition
func SearchCustomersTool() Tool {
	return Tool{
		Name:        "search_customers",
		Description: "Search customers with a search term",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"search_term": map[string]any{
					"type":        "string",
					"description": "Search term to match against customer names, email, or SSN",
				},
				"limit": map[string]any{
					"type":        "number",
					"description": "Maximum number of results",
					"default":     10,
				},
			},
		},
	}
}

// executeSearchCustomers handles the search_customers tool execution
func (m *Manager) executeSearchCustomers(arguments map[string]any) MCPResponse {
	searchTerm := ""
	if term, ok := arguments["search_term"].(string); ok {
		searchTerm = term
	}
	limit := 10
	if l, ok := arguments["limit"].(float64); ok {
		limit = int(l)
	}

	customers, err := m.client.SearchCustomers(searchTerm, limit)
	if err != nil {
		LogError("search_customers", err, fmt.Sprintf("with term='%s', limit=%d", searchTerm, limit))
		return CreateErrorResponse(-1, err.Error(), nil)
	}

	text := "Customers:\n"
	for _, customer := range customers {
		text += fmt.Sprintf("- ID: %d, Name: %s %s, Email: %s\n", customer.GetID(), customer.GetFirstName(), customer.GetLastName(), customer.GetEmail())
	}

	return CreateSuccessResponse(text, nil)
}
