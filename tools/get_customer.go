package tools

import "fmt"

// GetCustomerTool returns the get_customer tool definition
func GetCustomerTool() Tool {
	return Tool{
		Name:        "get_customer",
		Description: "Get customer information by ID",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"customer_id": map[string]any{
					"type":        "string",
					"description": "The customer ID to retrieve",
				},
			},
			"required": []string{"customer_id"},
		},
	}
}

// executeGetCustomer handles the get_customer tool execution
func (m *Manager) executeGetCustomer(arguments map[string]any) MCPResponse {
	customerID := arguments["customer_id"].(string)
	customer, err := m.client.GetCustomer(customerID)
	if err != nil {
		LogError("get_customer", err, fmt.Sprintf("for ID %s", customerID))
		return CreateErrorResponse(-1, err.Error(), nil)
	}

	text := fmt.Sprintf("Customer Details:\nID: %d\nName: %s %s\nEmail: %s\nPhone: %s\nCreated: %s",
		customer.GetID(), customer.GetFirstName(), customer.GetLastName(), customer.GetEmail(), customer.GetPhone(), customer.GetCreatedDate())

	return CreateSuccessResponse(text, nil)
}