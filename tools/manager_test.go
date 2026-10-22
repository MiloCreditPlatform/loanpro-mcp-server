package tools

import (
	"errors"
	"strings"
	"testing"
)

// MockLoanProClient implements the LoanProClient interface for testing
type MockLoanProClient struct {
	loans        map[string]MockLoan
	customers    map[string]MockCustomer
	payments     map[string][]MockPayment
	transactions map[string][]MockTransaction
	archiveErr   error
	archivedIDs  []string
}

// MockLoan implements the Loan interface
type MockLoan struct {
	id                  string
	displayID           string
	primaryCustomerName string
	loanStatus          string
	principalBalance    string
	payoffAmount        string
	active              string
	daysPastDue         string
	archived            string
	created             string
	portfolio           string
	subPortfolio        string
}

func (m MockLoan) GetID() string                  { return m.id }
func (m MockLoan) GetDisplayID() string           { return m.displayID }
func (m MockLoan) GetPrimaryCustomerName() string { return m.primaryCustomerName }
func (m MockLoan) GetLoanStatus() string          { return m.loanStatus }
func (m MockLoan) GetPrincipalBalance() string    { return m.principalBalance }
func (m MockLoan) GetPayoffAmount() string        { return m.payoffAmount }
func (m MockLoan) GetActive() string              { return m.active }
func (m MockLoan) GetDaysPastDue() string         { return m.daysPastDue }
func (m MockLoan) GetArchived() string            { return m.archived }
func (m MockLoan) GetCreatedDate() string         { return m.created }

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
	// Check for non-empty and non-zero amounts
	hasData := false
	if m.principalAmount != "" && m.principalAmount != "0" && m.principalAmount != "0.00" {
		hasData = true
	}
	if m.interestAmount != "" && m.interestAmount != "0" && m.interestAmount != "0.00" {
		hasData = true
	}
	if m.feesAmount != "" && m.feesAmount != "0" && m.feesAmount != "0.00" {
		hasData = true
	}
	if m.escrowAmount != "" && m.escrowAmount != "0" && m.escrowAmount != "0.00" {
		hasData = true
	}
	return hasData
}

// MockLoanProClient methods
func (m *MockLoanProClient) GetLoan(id string) (Loan, error) {
	if loan, exists := m.loans[id]; exists {
		return loan, nil
	}
	return nil, nil
}

func (m *MockLoanProClient) SearchLoans(searchTerm, status string, limit, offset int) ([]Loan, error) {
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
	return m.GetLoanTransactionsWithOptions(loanID, nil)
}

func (m *MockLoanProClient) GetPastDueLoans(minDaysPastDue, limit, offset int) ([]Loan, error) {
	var results []Loan
	count := 0
	for _, loan := range m.loans {
		if count >= limit {
			break
		}
		results = append(results, loan)
		count++
	}
	return results, nil
}

func (m *MockLoanProClient) GetLoansByPortfolio(category, portfolio, subPortfolio, createdFrom, createdTo string) ([]Loan, error) {
	var results []Loan
	for _, id := range []string{"100", "101", "102", "103", "104", "123", "456"} {
		loan, ok := m.loans[id]
		if !ok || loan.archived == "1" || !strings.EqualFold(loan.portfolio, portfolio) {
			continue
		}
		if subPortfolio != "" && !strings.EqualFold(loan.subPortfolio, subPortfolio) {
			continue
		}
		if (createdFrom != "" && loan.created < createdFrom) || (createdTo != "" && loan.created > createdTo) {
			continue
		}
		results = append(results, loan)
	}
	return results, nil
}

func (m *MockLoanProClient) ArchiveLoan(id string) error {
	if m.archiveErr != nil {
		return m.archiveErr
	}
	m.archivedIDs = append(m.archivedIDs, id)
	return nil
}

func (m *MockLoanProClient) GetLoanTransactionsWithOptions(loanID string, opts *TransactionOptions) ([]Transaction, error) {
	if transactions, exists := m.transactions[loanID]; exists {
		var result []Transaction
		for _, transaction := range transactions {
			result = append(result, transaction)
		}

		// Apply pagination if options provided
		if opts != nil {
			start := opts.Offset
			if start > len(result) {
				return []Transaction{}, nil
			}

			end := len(result)
			if opts.Limit > 0 && start+opts.Limit < end {
				end = start + opts.Limit
			}

			return result[start:end], nil
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
				active:              "1",
				daysPastDue:         "15",
			},
			"456": {
				id:                  "456",
				displayID:           "LN00000456",
				primaryCustomerName: "Jane Smith",
				loanStatus:          "Current",
				principalBalance:    "18500.00",
				payoffAmount:        "18650.00",
				active:              "0",
				daysPastDue:         "0",
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

	expectedTools := []string{"get_loan", "search_loans", "get_customer", "search_customers", "get_loan_payments", "get_loan_transactions", "get_past_due_loans", "archive_loans"}

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

func TestManager_ExecuteTool_GetLoanTransactions_WithPagination(t *testing.T) {
	mockClient := createMockClient()
	manager := NewManager(mockClient)

	// Test with limit parameter only
	t.Run("With limit only", func(t *testing.T) {
		arguments := map[string]any{
			"loan_id": "123",
			"limit":   float64(1),
		}

		response := manager.ExecuteTool("get_loan_transactions", arguments)

		if response.Error != nil {
			t.Errorf("Expected no error, got %v", response.Error)
		}

		result, ok := response.Result.(map[string]any)
		if !ok {
			t.Fatal("Expected result to be map")
		}

		content, ok := result["content"].([]map[string]any)
		if !ok {
			t.Fatal("Expected content to be []map[string]any")
		}

		text := content[0]["text"].(string)

		// Should only return 1 transaction
		transactionCount := strings.Count(text, "ID: t")
		if transactionCount != 1 {
			t.Errorf("Expected 1 transaction with limit=1, got %d", transactionCount)
		}
	})

	// Test with limit and offset
	t.Run("With limit and offset", func(t *testing.T) {
		arguments := map[string]any{
			"loan_id": "123",
			"limit":   float64(1),
			"offset":  float64(1),
		}

		response := manager.ExecuteTool("get_loan_transactions", arguments)

		if response.Error != nil {
			t.Errorf("Expected no error, got %v", response.Error)
		}

		result, ok := response.Result.(map[string]any)
		if !ok {
			t.Fatal("Expected result to be map")
		}

		content, ok := result["content"].([]map[string]any)
		if !ok {
			t.Fatal("Expected content to be []map[string]any")
		}

		text := content[0]["text"].(string)

		// Should skip first transaction and return the second one
		if strings.Contains(text, "ID: t1") {
			t.Error("Expected to skip first transaction (t1)")
		}
		if !strings.Contains(text, "ID: t2") {
			t.Error("Expected to include second transaction (t2)")
		}
	})

	// Test with offset beyond available records
	t.Run("Offset beyond available", func(t *testing.T) {
		arguments := map[string]any{
			"loan_id": "123",
			"offset":  float64(10),
		}

		response := manager.ExecuteTool("get_loan_transactions", arguments)

		if response.Error != nil {
			t.Errorf("Expected no error, got %v", response.Error)
		}

		result, ok := response.Result.(map[string]any)
		if !ok {
			t.Fatal("Expected result to be map")
		}

		content, ok := result["content"].([]map[string]any)
		if !ok {
			t.Fatal("Expected content to be []map[string]any")
		}

		text := content[0]["text"].(string)

		// Should return no transactions message
		if !strings.Contains(text, "No transactions found") {
			t.Errorf("Expected 'No transactions found' when offset > available records, got: %s", text)
		}
	})
}

func TestManager_ExecuteTool_GetPastDueLoans(t *testing.T) {
	mockClient := createMockClient()
	manager := NewManager(mockClient)

	arguments := map[string]any{
		"min_days_past_due": float64(5),
		"limit":             float64(10),
	}

	response := manager.ExecuteTool("get_past_due_loans", arguments)

	if response.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC 2.0, got %s", response.JSONRPC)
	}

	if response.Error != nil {
		t.Errorf("Expected no error, got %v", response.Error)
	}

	if response.Result == nil {
		t.Fatal("Expected result, got nil")
	}

	result, ok := response.Result.(map[string]any)
	if !ok {
		t.Fatal("Expected result to be map[string]any")
	}

	content, ok := result["content"].([]map[string]any)
	if !ok {
		t.Fatal("Expected content to be []map[string]any")
	}

	text, ok := content[0]["text"].(string)
	if !ok {
		t.Fatal("Expected text to be string")
	}

	if !strings.Contains(text, "Days Past Due") {
		t.Errorf("Expected response to contain 'Days Past Due', got: %s", text)
	}
}

func TestManager_ExecuteTool_GetPastDueLoans_Defaults(t *testing.T) {
	mockClient := createMockClient()
	manager := NewManager(mockClient)

	// Call with no arguments to exercise defaults
	response := manager.ExecuteTool("get_past_due_loans", map[string]any{})

	if response.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC 2.0, got %s", response.JSONRPC)
	}

	if response.Error != nil {
		t.Errorf("Expected no error, got %v", response.Error)
	}
}

func TestManager_ExecuteTool_GetLoan_ActiveField(t *testing.T) {
	mockClient := createMockClient()
	manager := NewManager(mockClient)

	// Loan 123 is active (active="1")
	resp := manager.ExecuteTool("get_loan", map[string]any{"loan_id": "123"})
	if resp.Error != nil {
		t.Fatalf("Expected no error, got %v", resp.Error)
	}
	text := resp.Result.(map[string]any)["content"].([]map[string]any)[0]["text"].(string)
	if !strings.Contains(text, "Active: 1") {
		t.Errorf("Expected 'Active: 1' for active loan, got: %s", text)
	}
}

func TestManager_ExecuteTool_GetPastDueLoans_ActiveField(t *testing.T) {
	// Single-loan fixture so output is deterministic.
	mockClient := &MockLoanProClient{
		loans: map[string]MockLoan{
			"123": {
				id:                  "123",
				displayID:           "LN00000123",
				primaryCustomerName: "John Doe",
				loanStatus:          "Active",
				principalBalance:    "25000.00",
				payoffAmount:        "25250.00",
				active:              "1",
				daysPastDue:         "15",
			},
		},
	}
	manager := NewManager(mockClient)

	resp := manager.ExecuteTool("get_past_due_loans", map[string]any{
		"min_days_past_due": float64(5),
		"limit":             float64(10),
	})
	if resp.Error != nil {
		t.Fatalf("Expected no error, got %v", resp.Error)
	}
	text := resp.Result.(map[string]any)["content"].([]map[string]any)[0]["text"].(string)
	if !strings.Contains(text, "Active: 1") {
		t.Errorf("Expected 'Active: 1' in past due loans output, got: %s", text)
	}
}

func archiveTestClient() *MockLoanProClient {
	return &MockLoanProClient{
		loans: map[string]MockLoan{
			"123": {id: "123", displayID: "LN00000123", primaryCustomerName: "Test User", loanStatus: "Open", principalBalance: "100.00", archived: "0"},
			"456": {id: "456", displayID: "LN00000456", primaryCustomerName: "Test User", loanStatus: "Closed", principalBalance: "0.00", archived: "1"},
		},
	}
}

func responseText(t *testing.T, resp MCPResponse) string {
	t.Helper()
	if resp.Error != nil {
		t.Fatalf("Expected no error, got %v", resp.Error)
	}
	return resp.Result.(map[string]any)["content"].([]map[string]any)[0]["text"].(string)
}

func TestManager_ExecuteTool_ArchiveLoans_DryRunByDefault(t *testing.T) {
	mockClient := archiveTestClient()
	manager := NewManager(mockClient)

	text := responseText(t, manager.ExecuteTool("archive_loans", map[string]any{
		"loan_ids": []any{"123", "456", "999"},
	}))

	if len(mockClient.archivedIDs) != 0 {
		t.Errorf("Dry run must not archive loans, archived: %v", mockClient.archivedIDs)
	}
	for _, want := range []string{"DRY RUN", "LN00000123", "SKIPPED: already archived", "ID: 999, FAILED: loan not found", "1 to archive, 1 already archived, 1 not found"} {
		if !strings.Contains(text, want) {
			t.Errorf("Expected %q in output, got: %s", want, text)
		}
	}
}

func TestManager_ExecuteTool_ArchiveLoans_Execute(t *testing.T) {
	mockClient := archiveTestClient()
	manager := NewManager(mockClient)

	text := responseText(t, manager.ExecuteTool("archive_loans", map[string]any{
		"loan_ids": []any{"123", "123", float64(456)},
		"dry_run":  false,
	}))

	if len(mockClient.archivedIDs) != 1 || mockClient.archivedIDs[0] != "123" {
		t.Errorf("Expected only loan 123 archived once, got %v", mockClient.archivedIDs)
	}
	if !strings.Contains(text, "1 archived, 1 already archived, 0 failed") {
		t.Errorf("Unexpected summary: %s", text)
	}
}

func TestManager_ExecuteTool_ArchiveLoans_ArchiveError(t *testing.T) {
	mockClient := archiveTestClient()
	mockClient.archiveErr = errors.New("API returned status 400")
	manager := NewManager(mockClient)

	text := responseText(t, manager.ExecuteTool("archive_loans", map[string]any{
		"loan_ids": []any{"123"},
		"dry_run":  false,
	}))

	if !strings.Contains(text, "FAILED: API returned status 400") || !strings.Contains(text, "0 archived, 0 already archived, 1 failed") {
		t.Errorf("Expected failure to be reported, got: %s", text)
	}
}

func TestManager_ExecuteTool_ArchiveLoans_InvalidInput(t *testing.T) {
	manager := NewManager(archiveTestClient())

	tests := []map[string]any{
		{},
		{"loan_ids": []any{}},
		{"loan_ids": "123"},
		{"loan_ids": []any{"123)?$x"}},
		{"loan_ids": make([]any, maxArchiveBatch+1)},
	}
	for _, args := range tests {
		if resp := manager.ExecuteTool("archive_loans", args); resp.Error == nil {
			t.Errorf("Expected error for arguments %v", args)
		}
	}
}

func portfolioTestClient() *MockLoanProClient {
	loan := func(id, created, portfolio, sub, archived string) MockLoan {
		return MockLoan{id: id, displayID: "CL" + id + "dev", primaryCustomerName: "Test User", loanStatus: "Open", created: created, portfolio: portfolio, subPortfolio: sub, archived: archived}
	}
	return &MockLoanProClient{
		loans: map[string]MockLoan{
			"100": loan("100", "2025-01-15", "Test", "Development", "0"),
			"101": loan("101", "2025-06-01", "Test", "Development", "0"),
			"102": loan("102", "2025-06-02", "Test", "Staging", "0"),
			"103": loan("103", "2025-06-03", "Production", "Retail", "0"),
			"104": loan("104", "2025-06-04", "Test", "Development", "1"),
		},
	}
}

func TestManager_ExecuteTool_ArchiveLoans_PortfolioDryRun(t *testing.T) {
	mockClient := portfolioTestClient()
	manager := NewManager(mockClient)

	text := responseText(t, manager.ExecuteTool("archive_loans", map[string]any{
		"portfolio":     "test",
		"sub_portfolio": "development",
		"created_from":  "2025-06-01",
		"created_to":    "2025-12-31",
	}))

	if len(mockClient.archivedIDs) != 0 {
		t.Errorf("Dry run must not archive loans, archived: %v", mockClient.archivedIDs)
	}
	if !strings.Contains(text, "DRY RUN: 1 live loan(s) match portfolio test, sub-portfolio development, created 2025-06-01 to 2025-12-31") {
		t.Errorf("Unexpected dry run header: %s", text)
	}
	if !strings.Contains(text, "ID: 101,") {
		t.Errorf("Expected loan 101 listed, got: %s", text)
	}
	for _, id := range []string{"100", "102", "103", "104"} {
		if strings.Contains(text, "ID: "+id+",") {
			t.Errorf("Loan %s should not match, got: %s", id, text)
		}
	}
}

func TestManager_ExecuteTool_ArchiveLoans_PortfolioExecuteWithLimitAndExclude(t *testing.T) {
	mockClient := portfolioTestClient()
	manager := NewManager(mockClient)

	text := responseText(t, manager.ExecuteTool("archive_loans", map[string]any{
		"portfolio":        "Test",
		"exclude_loan_ids": []any{"100"},
		"limit":            float64(1),
		"dry_run":          false,
	}))

	if len(mockClient.archivedIDs) != 1 || mockClient.archivedIDs[0] != "101" {
		t.Errorf("Expected only loan 101 archived, got %v", mockClient.archivedIDs)
	}
	if !strings.Contains(text, "1 archived, 0 failed, 1 excluded, 1 remaining") {
		t.Errorf("Unexpected summary: %s", text)
	}
}

func TestManager_ExecuteTool_ArchiveLoans_PortfolioInvalidInput(t *testing.T) {
	manager := NewManager(portfolioTestClient())

	tests := []map[string]any{
		{"portfolio": "Test", "loan_ids": []any{"100"}},
		{"portfolio": "Test", "created_from": "06/01/2025"},
		{"portfolio": "Test", "created_from": "2025-07-01", "created_to": "2025-06-01"},
		{"portfolio": "Test", "limit": float64(0)},
		{"portfolio": "Test", "limit": float64(maxArchiveBatch + 1)},
		{"portfolio": "Test", "exclude_loan_ids": []any{"abc"}},
	}
	for _, args := range tests {
		if resp := manager.ExecuteTool("archive_loans", args); resp.Error == nil {
			t.Errorf("Expected error for arguments %v", args)
		}
	}
}
