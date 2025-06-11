package tools

import "fmt"

// SearchLoansTool returns the search_loans tool definition
func SearchLoansTool() Tool {
	return Tool{
		Name:        "search_loans",
		Description: "Search loans with filters and search terms",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"search_term": map[string]any{
					"type":        "string",
					"description": "Search term to match against customer name, display ID, or title",
				},
				"status": map[string]any{
					"type":        "string",
					"description": "Loan status filter",
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

// executeSearchLoans handles the search_loans tool execution
func (m *Manager) executeSearchLoans(arguments map[string]any) MCPResponse {
	searchTerm := ""
	if term, ok := arguments["search_term"].(string); ok {
		searchTerm = term
	}
	status := ""
	if s, ok := arguments["status"].(string); ok {
		status = s
	}
	limit := 10
	if l, ok := arguments["limit"].(float64); ok {
		limit = int(l)
	}

	loans, err := m.client.SearchLoans(searchTerm, status, limit)
	if err != nil {
		LogError("search_loans", err, fmt.Sprintf("with term='%s', status='%s', limit=%d", searchTerm, status, limit))
		return CreateErrorResponse(-1, err.Error(), nil)
	}

	text := "Loans:\n"
	for _, loan := range loans {
		text += fmt.Sprintf("- ID: %s, Display ID: %s, Customer: %s, Status: %s, Balance: $%s\n", 
			loan.GetID(), loan.GetDisplayID(), loan.GetPrimaryCustomerName(), loan.GetLoanStatus(), loan.GetPrincipalBalance())
	}

	return CreateSuccessResponse(text, nil)
}