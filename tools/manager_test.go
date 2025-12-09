package tools

import (
	"strings"
	"testing"
)

// MockLoanProClient implements the LoanProClient interface for testing
type MockLoanProClient struct {
	loans        map[string]MockLoan
	customers    map[string]MockCustomer
	payments     map[string][]MockPayment
	transactions map[string][]MockTransaction
}

// MockLoan implements the Loan interface
type MockLoan struct {
	id                  string
	displayID           string
	primaryCustomerName string
	loanStatus          string
	principalBalance    string
	payoffAmount        string
}

func (m MockLoan) GetID() string                  { return m.id }
func (m MockLoan) GetDisplayID() string           { return m.displayID }
func (m MockLoan) GetPrimaryCustomerName() string { return m.primaryCustomerName }
func (m MockLoan) GetLoanStatus() string          { return m.loanStatus }
func (m MockLoan) GetPrincipalBalance() string    { return m.principalBalance }
func (m MockLoan) GetPayoffAmount() string        { return m.payoffAmount }

// MockCustomer implements the Customer interface
type MockCustomer struct {
	id        int
	firstName string
	lastName  string
	email     string
	phone     string
}

func (m MockCustomer) GetID() int             { return m.id }
func (m MockCustomer) GetFirstName() string   { return m.firstName }
func (m MockCustomer) GetLastName() string    { return m.lastName }
func (m MockCustomer) GetEmail() string       { return m.email }
func (m MockCustomer) GetPhone() string       { return m.phone }
func (m MockCustomer) GetCreatedDate() string { return "2025-01-01 00:00:00 UTC" }

// MockPayment implements the Payment interface
type MockPayment struct {
	id     string
	amount string
	date   string
	status string
}

func (m MockPayment) GetID() string     { return m.id }
func (m MockPayment) GetAmount() string { return m.amount }
func (m MockPayment) GetDate() string   { return m.date }
func (m MockPayment) GetStatus() string { return m.status }

// MockTransaction implements the Transaction interface
type MockTransaction struct {
	id              string
	amount          string
	date            string
	txnType         string
	title           string
	info            string
	status          string
	principalAmount string
	interestAmount  string
	feesAmount      string
	escrowAmount    string
}

func (m MockTransaction) GetID() string              { return m.id }
func (m MockTransaction) GetAmount() string          { return m.amount }
func (m MockTransaction) GetDate() string            { return m.date }
func (m MockTransaction) GetType() string            { return m.txnType }
func (m MockTransaction) GetTitle() string           { return m.title }
func (m MockTransaction) GetInfo() string            { return m.info }
func (m MockTransaction) GetStatus() string          { return m.status }
func (m MockTransaction) GetPrincipalAmount() string { return m.principalAmount }
func (m MockTransaction) GetInterestAmount() string  { return m.interestAmount }
func (m MockTransaction) GetFeesAmount() string      { return m.feesAmount }
func (m MockTransaction) GetEscrowAmount() string    { return m.escrowAmount }
func (m MockTransaction) HasPaymentBreakdown() bool {
	return m.principalAmount != "" || m.interestAmount != "" ||
		m.feesAmount != "" || m.escrowAmount != ""
}

// MockLoanProClient methods
func (m *MockLoanProClient) GetLoan(id string) (Loan, error) {
	if loan, exists := m.loans[id]; exists {
		return loan, nil
	}
	return nil, nil
}

func (m *MockLoanProClient) SearchLoans(searchTerm, status string, limit int) ([]Loan, error) {
	var results []Loan
	count := 0
	for _, loan := range m.loans {
		if count >= limit {
			break
		}
		if searchTerm == "" || loan.displayID == searchTerm || loan.primaryCustomerName == searchTerm {
			if status == "" || loan.loanStatus == status {
				results = append(results, loan)
				count++
			}
		}
	}
	return results, nil
}

func (m *MockLoanProClient) GetCustomer(id string) (Customer, error) {
	if customer, exists := m.customers[id]; exists {
		return customer, nil
	}
	return nil, nil
}

func (m *MockLoanProClient) SearchCustomers(searchTerm string, limit int) ([]Customer, error) {
	var results []Customer
	count := 0
	for _, customer := range m.customers {
		if count >= limit {
			break
		}
		if searchTerm == "" || customer.firstName == searchTerm || customer.lastName == searchTerm {
			results = append(results, customer)
			count++
		}
	}
	return results, nil
}

func (m *MockLoanProClient) GetLoanPayments(loanID string) ([]Payment, error) {
	if payments, exists := m.payments[loanID]; exists {
		var result []Payment
		for _, payment := range payments {
			result = append(result, payment)
		}
		return result, nil
	}
	return []Payment{}, nil
}

func (m *MockLoanProClient) GetLoanTransactions(loanID string) ([]Transaction, error) {
	if transactions, exists := m.transactions[loanID]; exists {
		var result []Transaction
		for _, transaction := range transactions {
			result = append(result, transaction)
		}
		return result, nil
	}
	return []Transaction{}, nil
}

// Helper function to create a mock client with test data
func createMockClient() *MockLoanProClient {
	return &MockLoanProClient{
		loans: map[string]MockLoan{
			"123": {
				id:                  "123",
				displayID:           "LN00000123",
				primaryCustomerName: "John Doe",
				loanStatus:          "Active",
				principalBalance:    "25000.00",
				payoffAmount:        "25250.00",
			},
			"456": {
				id:                  "456",
				displayID:           "LN00000456",
				primaryCustomerName: "Jane Smith",
				loanStatus:          "Current",
				principalBalance:    "18500.00",
				payoffAmount:        "18650.00",
			},
		},
		customers: map[string]MockCustomer{
			"789": {
				id:        789,
				firstName: "John",
				lastName:  "Doe",
				email:     "john.doe@example.com",
				phone:     "(555) 123-4567",
			},
		},
		payments: map[string][]MockPayment{
			"123": {
				{id: "p1", amount: "500.00", date: "2025-01-15", status: "Active"},
				{id: "p2", amount: "500.00", date: "2025-02-15", status: "Active"},
			},
		},
		transactions: map[string][]MockTransaction{
			"123": {
				{
					id:              "t1",
					amount:          "500.00",
					date:            "2025-01-15",
					txnType:         "payment",
					title:           "Payment Received",
					status:          "Active",
					principalAmount: "450.00",
					interestAmount:  "50.00",
				},
				{
					id:      "t2",
					amount:  "25.00",
					date:    "2025-01-10",
					txnType: "charge.latefee",
					title:   "Late Fee",
					status:  "Active",
				},
			},
		},
	}
}

func TestManager_GetAllTools(t *testing.T) {
	mockClient := createMockClient()
	manager := NewManager(mockClient)

	tools := manager.GetAllTools()

	expectedTools := []string{"get_loan", "search_loans", "get_customer", "search_customers", "get_loan_payments", "get_loan_transactions"}

	if len(tools) != len(expectedTools) {
		t.Errorf("Expected %d tools, got %d", len(expectedTools), len(tools))
	}

	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Name] = true
	}

	for _, expectedTool := range expectedTools {
		if !toolNames[expectedTool] {
			t.Errorf("Expected tool %s not found", expectedTool)
		}
	}
}

func TestManager_ExecuteTool_GetLoan(t *testing.T) {
	mockClient := createMockClient()
	manager := NewManager(mockClient)

	arguments := map[string]any{
		"loan_id": "123",
	}

	response := manager.ExecuteTool("get_loan", arguments)

	if response.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC 2.0, got %s", response.JSONRPC)
	}

	if response.Error != nil {
		t.Errorf("Expected no error, got %v", response.Error)
	}

	if response.Result == nil {
		t.Error("Expected result, got nil")
	}
}

func TestManager_ExecuteTool_SearchLoans(t *testing.T) {
	mockClient := createMockClient()
	manager := NewManager(mockClient)

	arguments := map[string]any{
		"search_term": "John",
		"limit":       float64(10),
	}

	response := manager.ExecuteTool("search_loans", arguments)

	if response.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC 2.0, got %s", response.JSONRPC)
	}

	if response.Error != nil {
		t.Errorf("Expected no error, got %v", response.Error)
	}
}

func TestManager_ExecuteTool_GetCustomer(t *testing.T) {
	mockClient := createMockClient()
	manager := NewManager(mockClient)

	arguments := map[string]any{
		"customer_id": "789",
	}

	response := manager.ExecuteTool("get_customer", arguments)

	if response.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC 2.0, got %s", response.JSONRPC)
	}

	if response.Error != nil {
		t.Errorf("Expected no error, got %v", response.Error)
	}
}

func TestManager_ExecuteTool_InvalidTool(t *testing.T) {
	mockClient := createMockClient()
	manager := NewManager(mockClient)

	arguments := map[string]any{}

	response := manager.ExecuteTool("invalid_tool", arguments)

	if response.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC 2.0, got %s", response.JSONRPC)
	}

	if response.Error == nil {
		t.Error("Expected error for invalid tool, got nil")
	}

	if response.Error.Code != -32601 {
		t.Errorf("Expected error code -32601, got %d", response.Error.Code)
	}
}

func TestCreateSuccessResponse(t *testing.T) {
	text := "Test response"
	id := 123

	response := CreateSuccessResponse(text, id)

	if response.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC 2.0, got %s", response.JSONRPC)
	}

	if response.ID != id {
		t.Errorf("Expected ID %v, got %v", id, response.ID)
	}

	if response.Error != nil {
		t.Errorf("Expected no error, got %v", response.Error)
	}

	// Check result structure
	result, ok := response.Result.(map[string]any)
	if !ok {
		t.Error("Expected result to be map[string]any")
	}

	content, ok := result["content"].([]map[string]any)
	if !ok {
		t.Error("Expected content to be []map[string]any")
	}

	if len(content) != 1 {
		t.Errorf("Expected 1 content item, got %d", len(content))
	}

	if content[0]["text"] != text {
		t.Errorf("Expected text %s, got %s", text, content[0]["text"])
	}
}

func TestCreateErrorResponse(t *testing.T) {
	code := -1
	message := "Test error"
	id := 123

	response := CreateErrorResponse(code, message, id)

	if response.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC 2.0, got %s", response.JSONRPC)
	}

	if response.ID != id {
		t.Errorf("Expected ID %v, got %v", id, response.ID)
	}

	if response.Error == nil {
		t.Error("Expected error, got nil")
	}

	if response.Error.Code != code {
		t.Errorf("Expected error code %d, got %d", code, response.Error.Code)
	}

	if response.Error.Message != message {
		t.Errorf("Expected error message %s, got %s", message, response.Error.Message)
	}
}

func TestManager_ExecuteTool_GetLoanPayments(t *testing.T) {
	mockClient := createMockClient()
	manager := NewManager(mockClient)

	arguments := map[string]any{
		"loan_id": "123",
	}

	response := manager.ExecuteTool("get_loan_payments", arguments)

	if response.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC 2.0, got %s", response.JSONRPC)
	}

	if response.Error != nil {
		t.Errorf("Expected no error, got %v", response.Error)
	}

	if response.Result == nil {
		t.Fatal("Expected result, got nil")
	}

	// Check result structure
	result, ok := response.Result.(map[string]any)
	if !ok {
		t.Fatal("Expected result to be map[string]any")
	}

	content, ok := result["content"].([]map[string]any)
	if !ok {
		t.Fatal("Expected content to be []map[string]any")
	}

	if len(content) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(content))
	}

	text, ok := content[0]["text"].(string)
	if !ok {
		t.Fatal("Expected text to be string")
	}

	// Verify that the response contains payment status
	if !strings.Contains(text, "Status: Active") {
		t.Errorf("Expected response to contain 'Status: Active', got: %s", text)
	}

	// Verify that both payments are listed
	if !strings.Contains(text, "ID: p1") || !strings.Contains(text, "ID: p2") {
		t.Errorf("Expected response to contain payment IDs p1 and p2, got: %s", text)
	}
}

func TestManager_ExecuteTool_GetLoanPayments_NoPayments(t *testing.T) {
	mockClient := createMockClient()
	manager := NewManager(mockClient)

	// Request payments for a loan with no payment history
	arguments := map[string]any{
		"loan_id": "456",
	}

	response := manager.ExecuteTool("get_loan_payments", arguments)

	if response.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC 2.0, got %s", response.JSONRPC)
	}

	if response.Error != nil {
		t.Errorf("Expected no error, got %v", response.Error)
	}

	if response.Result == nil {
		t.Fatal("Expected result, got nil")
	}

	// Check result structure
	result, ok := response.Result.(map[string]any)
	if !ok {
		t.Fatal("Expected result to be map[string]any")
	}

	content, ok := result["content"].([]map[string]any)
	if !ok {
		t.Fatal("Expected content to be []map[string]any")
	}

	if len(content) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(content))
	}

	text, ok := content[0]["text"].(string)
	if !ok {
		t.Fatal("Expected text to be string")
	}

	// Verify that the response indicates no payments found
	if !strings.Contains(text, "No payments found") {
		t.Errorf("Expected response to contain 'No payments found', got: %s", text)
	}
}

func TestManager_ExecuteTool_GetLoanTransactions(t *testing.T) {
	mockClient := createMockClient()
	manager := NewManager(mockClient)

	arguments := map[string]any{
		"loan_id": "123",
	}

	response := manager.ExecuteTool("get_loan_transactions", arguments)

	if response.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC 2.0, got %s", response.JSONRPC)
	}

	if response.Error != nil {
		t.Errorf("Expected no error, got %v", response.Error)
	}

	if response.Result == nil {
		t.Fatal("Expected result, got nil")
	}

	// Check result structure
	result, ok := response.Result.(map[string]any)
	if !ok {
		t.Fatal("Expected result to be map[string]any")
	}

	content, ok := result["content"].([]map[string]any)
	if !ok {
		t.Fatal("Expected content to be []map[string]any")
	}

	if len(content) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(content))
	}

	text, ok := content[0]["text"].(string)
	if !ok {
		t.Fatal("Expected text to be string")
	}

	// Verify that the response contains transaction details
	if !strings.Contains(text, "Type: payment") {
		t.Errorf("Expected response to contain 'Type: payment', got: %s", text)
	}

	if !strings.Contains(text, "Status: Active") {
		t.Errorf("Expected response to contain 'Status: Active', got: %s", text)
	}

	if !strings.Contains(text, "ID: t1") {
		t.Errorf("Expected response to contain 'ID: t1', got: %s", text)
	}

	// Verify payment breakdown is shown
	if !strings.Contains(text, "Principal: $450.00") {
		t.Errorf("Expected response to contain 'Principal: $450.00', got: %s", text)
	}

	if !strings.Contains(text, "Interest: $50.00") {
		t.Errorf("Expected response to contain 'Interest: $50.00', got: %s", text)
	}

	// Verify charge transaction is shown
	if !strings.Contains(text, "Type: charge.latefee") {
		t.Errorf("Expected response to contain 'Type: charge.latefee', got: %s", text)
	}
}

func TestManager_ExecuteTool_GetLoanTransactions_NoTransactions(t *testing.T) {
	mockClient := createMockClient()
	manager := NewManager(mockClient)

	// Request transactions for a loan with no transaction history
	arguments := map[string]any{
		"loan_id": "456",
	}

	response := manager.ExecuteTool("get_loan_transactions", arguments)

	if response.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC 2.0, got %s", response.JSONRPC)
	}

	if response.Error != nil {
		t.Errorf("Expected no error, got %v", response.Error)
	}

	if response.Result == nil {
		t.Fatal("Expected result, got nil")
	}

	// Check result structure
	result, ok := response.Result.(map[string]any)
	if !ok {
		t.Fatal("Expected result to be map[string]any")
	}

	content, ok := result["content"].([]map[string]any)
	if !ok {
		t.Fatal("Expected content to be []map[string]any")
	}

	if len(content) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(content))
	}

	text, ok := content[0]["text"].(string)
	if !ok {
		t.Fatal("Expected text to be string")
	}

	// Verify that the response indicates no transactions found
	if !strings.Contains(text, "No transactions found") {
		t.Errorf("Expected response to contain 'No transactions found', got: %s", text)
	}
}
