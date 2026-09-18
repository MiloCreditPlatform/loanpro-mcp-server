package tools

import (
	"errors"
	"strings"
	"testing"
)

// reversedPaymentsClient returns a mock client for a loan whose payment history
// contains a reversed payment, and whose transaction history (as LoanPro
// returns it) omits that payment entirely.
func reversedPaymentsClient() *MockLoanProClient {
	return &MockLoanProClient{
		payments: map[string][]MockPayment{
			"456": {
				{id: "p10", amount: "8814.75", date: "2026-08-06", status: "Active"},
				{id: "p11", amount: "8814.75", date: "2026-09-06", status: "Inactive"},
			},
		},
		transactions: map[string][]MockTransaction{
			"456": {},
		},
	}
}

// responseText extracts the single text payload from a successful MCP
// response, failing the test if the response is shaped any other way.
func responseText(t *testing.T, response MCPResponse) string {
	t.Helper()

	if response.Error != nil {
		t.Fatalf("Expected no error, got %v", response.Error)
	}
	result, ok := response.Result.(map[string]any)
	if !ok {
		t.Fatal("Expected result to be map[string]any")
	}
	content, ok := result["content"].([]map[string]any)
	if !ok || len(content) != 1 {
		t.Fatal("Expected exactly one content item")
	}
	text, ok := content[0]["text"].(string)
	if !ok {
		t.Fatal("Expected text to be string")
	}
	return text
}

// TestGetLoanPayments_MarksReversedPayments verifies that a reversed payment is
// labelled as such and counted in the trailing summary.
func TestGetLoanPayments_MarksReversedPayments(t *testing.T) {
	manager := NewManager(reversedPaymentsClient())

	text := responseText(t, manager.ExecuteTool("get_loan_payments", map[string]any{
		"loan_id": "456",
	}))

	if !strings.Contains(text, "ID: p11, Status: Inactive (reversed - payment did not stand)") {
		t.Errorf("Expected the reversed payment to be labelled, got: %s", text)
	}
	if !strings.Contains(text, "1 of 2 payments were reversed") {
		t.Errorf("Expected a reversed-payment summary, got: %s", text)
	}
	if !strings.Contains(text, "ID: p10, Status: Active") {
		t.Errorf("Expected the active payment to stay unannotated, got: %s", text)
	}
}

// TestGetLoanPayments_ReversedOnly verifies that reversed_only narrows the
// history to reversed payments.
func TestGetLoanPayments_ReversedOnly(t *testing.T) {
	manager := NewManager(reversedPaymentsClient())

	text := responseText(t, manager.ExecuteTool("get_loan_payments", map[string]any{
		"loan_id":       "456",
		"reversed_only": true,
	}))

	if !strings.Contains(text, "ID: p11") {
		t.Errorf("Expected the reversed payment, got: %s", text)
	}
	if strings.Contains(text, "ID: p10") {
		t.Errorf("Expected active payments to be filtered out, got: %s", text)
	}
}

// TestGetLoanPayments_ReversedOnly_NoneFound verifies that a loan with no
// reversed payments says so explicitly rather than returning a bare header.
func TestGetLoanPayments_ReversedOnly_NoneFound(t *testing.T) {
	manager := NewManager(createMockClient())

	text := responseText(t, manager.ExecuteTool("get_loan_payments", map[string]any{
		"loan_id":       "123",
		"reversed_only": true,
	}))

	if !strings.Contains(text, "No reversed payments found.") {
		t.Errorf("Expected an explicit empty result, got: %s", text)
	}
}

// A payment that failed and was auto-reversed is absent from LoanPro's
// Transactions endpoint. Without this section it looks like the payment was
// never attempted, which is the failure mode this change exists to prevent.
func TestGetLoanTransactions_ListsReversedPayments(t *testing.T) {
	manager := NewManager(reversedPaymentsClient())

	text := responseText(t, manager.ExecuteTool("get_loan_transactions", map[string]any{
		"loan_id": "456",
	}))

	if !strings.Contains(text, "Reversed payments (excluded from the transaction list above)") {
		t.Errorf("Expected a reversed payments section, got: %s", text)
	}
	if !strings.Contains(text, "ID: p11, Status: Reversed") {
		t.Errorf("Expected the reversed payment to be listed, got: %s", text)
	}
}

// TestGetLoanTransactions_NoSectionWhenNothingReversed verifies that the
// trailing section is omitted when nothing was reversed.
func TestGetLoanTransactions_NoSectionWhenNothingReversed(t *testing.T) {
	manager := NewManager(createMockClient())

	text := responseText(t, manager.ExecuteTool("get_loan_transactions", map[string]any{
		"loan_id": "123",
	}))

	if strings.Contains(text, "Reversed payments") {
		t.Errorf("Expected no reversed payments section, got: %s", text)
	}
}

// failingPaymentsClient is a client whose payment lookup always fails, used to
// check that get_loan_transactions reports the failure instead of silently
// implying that no payments were reversed.
type failingPaymentsClient struct {
	*MockLoanProClient
}

// GetLoanPayments always fails.
func (c *failingPaymentsClient) GetLoanPayments(string) ([]Payment, error) {
	return nil, errors.New("loanpro unavailable")
}

// TestGetLoanTransactions_WarnsWhenReversedPaymentsUnavailable verifies that a
// failed payment lookup is surfaced. Returning nothing would be indistinguishable
// from a loan with no reversed payments.
func TestGetLoanTransactions_WarnsWhenReversedPaymentsUnavailable(t *testing.T) {
	manager := NewManager(&failingPaymentsClient{MockLoanProClient: reversedPaymentsClient()})

	text := responseText(t, manager.ExecuteTool("get_loan_transactions", map[string]any{
		"loan_id": "456",
	}))

	if !strings.Contains(text, "WARNING: Reversed payments could not be retrieved") {
		t.Errorf("Expected a warning when the payment lookup fails, got: %s", text)
	}
	if strings.Contains(text, "Reversed payments (excluded") {
		t.Errorf("Expected no reversed payments section when the lookup failed, got: %s", text)
	}
}
