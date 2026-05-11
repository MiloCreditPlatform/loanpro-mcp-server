package tools

import "fmt"

// GetPastDueLoansTool returns the get_past_due_loans tool definition
func GetPastDueLoansTool() Tool {
	return Tool{
		Name:        "get_past_due_loans",
		Description: "Get loans that are past due by at least a given number of days, queried directly from the OData API for complete coverage",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"min_days_past_due": map[string]any{
					"type":        "number",
					"description": "Minimum number of days past due (inclusive threshold)",
					"default":     5,
				},
				"limit": map[string]any{
					"type":        "number",
					"description": "Maximum number of results",
					"default":     50,
				},
				"offset": map[string]any{
					"type":        "number",
					"description": "Number of results to skip for pagination",
					"default":     0,
				},
			},
		},
	}
}

// executeGetPastDueLoans handles the get_past_due_loans tool execution
func (m *Manager) executeGetPastDueLoans(arguments map[string]any) MCPResponse {
	minDaysPastDue := 5
	if v, ok := arguments["min_days_past_due"].(float64); ok {
		minDaysPastDue = int(v)
	}
	limit := 50
	if v, ok := arguments["limit"].(float64); ok {
		limit = int(v)
	}
	offset := 0
	if v, ok := arguments["offset"].(float64); ok {
		offset = int(v)
	}

	loans, err := m.client.GetPastDueLoans(minDaysPastDue, limit, offset)
	if err != nil {
		LogError("get_past_due_loans", err, fmt.Sprintf("with min_days_past_due=%d, limit=%d, offset=%d", minDaysPastDue, limit, offset))
		return CreateErrorResponse(-1, err.Error(), nil)
	}

	if len(loans) == 0 {
		return CreateSuccessResponse(fmt.Sprintf("No loans found with more than %d days past due.", minDaysPastDue), nil)
	}

	text := fmt.Sprintf("Past Due Loans (>%d days):\n", minDaysPastDue)
	for _, loan := range loans {
		text += fmt.Sprintf("- ID: %s, Display ID: %s, Customer: %s, Days Past Due: %s, Balance: $%s\n",
			loan.GetID(), loan.GetDisplayID(), loan.GetPrimaryCustomerName(), loan.GetDaysPastDue(), loan.GetPrincipalBalance())
	}

	return CreateSuccessResponse(text, nil)
}
