package tools

import "fmt"

// GetPastDueLoansTool returns the get_past_due_loans tool definition
func GetPastDueLoansTool() Tool {
	return Tool{
		Name:        "get_past_due_loans",
		Description: "Get open loans with more than a given number of days past due, sorted by days past due descending",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"min_days_past_due": map[string]any{
					"type":        "number",
					"description": "Returns loans with MORE than this many days past due (exclusive threshold)",
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

	if minDaysPastDue < 0 {
		return CreateErrorResponse(-1, "min_days_past_due must be non-negative", nil)
	}
	if limit <= 0 {
		return CreateErrorResponse(-1, "limit must be positive", nil)
	}
	if offset < 0 {
		return CreateErrorResponse(-1, "offset must be non-negative", nil)
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
