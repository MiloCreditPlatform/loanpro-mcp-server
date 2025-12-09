package loanpro

import (
	"encoding/json"
	"testing"
)

func TestLoan_GetMethods(t *testing.T) {
	loan := &Loan{
		ID:        json.Number("123"),
		DisplayID: "LN00000123",
		Title:     "Test Loan",
		Active:    json.Number("1"),
		Archived:  json.Number("0"),
		Created:   "/Date(1427829732)/",
		LoanSettings: &LoanSettings{
			LoanStatusID: json.Number("2"),
		},
		LoanSetup: &LoanSetup{
			LoanAmount:       "25000.00",
			ContractDate:     "/Date(1427829732)/",
			FirstPaymentDate: "/Date(1430421732)/",
			Payment:          "500.00",
		},
		StatusArchive: &StatusArchiveWrapper{
			Results: []StatusArchiveEntry{
				{
					PrincipalBalance:  "24500.00",
					Payoff:            "24750.00",
					DaysPastDue:       json.Number("0"),
					NextPaymentAmount: "500.00",
					NextPaymentDate:   "/Date(1430421732)/",
					LoanStatusText:    "Current",
				},
			},
		},
		Customers: &CustomersWrapper{
			Results: []LoanCustomer{
				{
					FirstName: "John",
					LastName:  "Doe",
				},
			},
		},
		// Search result fields
		PrimaryCustomerName: "Jane Smith",
		LoanStatusText:      "Active",
		PrincipalBalance:    "23000.00",
		DaysPastDue:         json.Number("5"),
		NextPaymentAmount:   "450.00",
		NextPaymentDate:     "2025-04-15",
	}

	// Test GetID
	if loan.GetID() != "123" {
		t.Errorf("Expected ID 123, got %s", loan.GetID())
	}

	// Test GetDisplayID
	if loan.GetDisplayID() != "LN00000123" {
		t.Errorf("Expected DisplayID LN00000123, got %s", loan.GetDisplayID())
	}

	// Test GetActive
	if loan.GetActive() != "1" {
		t.Errorf("Expected Active 1, got %s", loan.GetActive())
	}

	// Test GetArchived
	if loan.GetArchived() != "0" {
		t.Errorf("Expected Archived 0, got %s", loan.GetArchived())
	}

	// Test GetLoanAmount
	if loan.GetLoanAmount() != "25000.00" {
		t.Errorf("Expected LoanAmount 25000.00, got %s", loan.GetLoanAmount())
	}

	// Test GetPrincipalBalance (should prefer search result)
	if loan.GetPrincipalBalance() != "23000.00" {
		t.Errorf("Expected PrincipalBalance 23000.00, got %s", loan.GetPrincipalBalance())
	}

	// Test GetPayoffAmount
	if loan.GetPayoffAmount() != "24750.00" {
		t.Errorf("Expected PayoffAmount 24750.00, got %s", loan.GetPayoffAmount())
	}

	// Test GetDaysPastDue (should prefer search result)
	if loan.GetDaysPastDue() != "5" {
		t.Errorf("Expected DaysPastDue 5, got %s", loan.GetDaysPastDue())
	}

	// Test GetLoanStatus (should prefer search result)
	if loan.GetLoanStatus() != "Active" {
		t.Errorf("Expected LoanStatus Active, got %s", loan.GetLoanStatus())
	}

	// Test GetPrimaryCustomerName (should prefer search result)
	if loan.GetPrimaryCustomerName() != "Jane Smith" {
		t.Errorf("Expected PrimaryCustomerName Jane Smith, got %s", loan.GetPrimaryCustomerName())
	}

	// Test GetNextPaymentAmount (should prefer search result)
	if loan.GetNextPaymentAmount() != "450.00" {
		t.Errorf("Expected NextPaymentAmount 450.00, got %s", loan.GetNextPaymentAmount())
	}

	// Test GetCreatedDate
	createdDate := loan.GetCreatedDate()
	if createdDate != "2015-03-31 19:22:12 UTC" {
		t.Errorf("Expected parsed created date, got %s", createdDate)
	}

	// Test GetContractDate
	contractDate := loan.GetContractDate()
	if contractDate != "2015-03-31" {
		t.Errorf("Expected parsed contract date, got %s", contractDate)
	}

	// Test GetNextPaymentDate (should prefer search result)
	if loan.GetNextPaymentDate() != "2025-04-15" {
		t.Errorf("Expected NextPaymentDate 2025-04-15, got %s", loan.GetNextPaymentDate())
	}
}

func TestLoan_FallbackBehavior(t *testing.T) {
	// Test loan with minimal data to test fallback behavior
	loan := &Loan{
		ID:        json.Number("456"),
		DisplayID: "LN00000456",
		StatusArchive: &StatusArchiveWrapper{
			Results: []StatusArchiveEntry{
				{
					PrincipalBalance:  "15000.00",
					DaysPastDue:       json.Number("30"),
					LoanStatusText:    "Past Due",
					NextPaymentAmount: "300.00",
					NextPaymentDate:   "/Date(1430421732)/",
				},
			},
		},
		LoanSettings: &LoanSettings{
			LoanStatusID: json.Number("3"),
		},
		Customers: &CustomersWrapper{
			Results: []LoanCustomer{
				{
					FirstName: "Alice",
					LastName:  "Johnson",
				},
			},
		},
	}

	// Test fallback to StatusArchive
	if loan.GetPrincipalBalance() != "15000.00" {
		t.Errorf("Expected fallback to StatusArchive PrincipalBalance, got %s", loan.GetPrincipalBalance())
	}

	if loan.GetDaysPastDue() != "30" {
		t.Errorf("Expected fallback to StatusArchive DaysPastDue, got %s", loan.GetDaysPastDue())
	}

	if loan.GetLoanStatus() != "Past Due" {
		t.Errorf("Expected fallback to StatusArchive LoanStatus, got %s", loan.GetLoanStatus())
	}

	// Test fallback to expanded customer data
	if loan.GetPrimaryCustomerName() != "Alice Johnson" {
		t.Errorf("Expected fallback to expanded customer name, got %s", loan.GetPrimaryCustomerName())
	}

	// Test parsed next payment date
	nextPaymentDate := loan.GetNextPaymentDate()
	if nextPaymentDate != "2015-04-30" {
		t.Errorf("Expected parsed next payment date, got %s", nextPaymentDate)
	}
}

func TestLoan_EmptyData(t *testing.T) {
	// Test loan with no data
	loan := &Loan{
		ID:        json.Number("789"),
		DisplayID: "LN00000789",
	}

	// Test N/A fallbacks
	if loan.GetPrincipalBalance() != "N/A" {
		t.Errorf("Expected N/A for empty PrincipalBalance, got %s", loan.GetPrincipalBalance())
	}

	if loan.GetPayoffAmount() != "N/A" {
		t.Errorf("Expected N/A for empty PayoffAmount, got %s", loan.GetPayoffAmount())
	}

	if loan.GetDaysPastDue() != "N/A" {
		t.Errorf("Expected N/A for empty DaysPastDue, got %s", loan.GetDaysPastDue())
	}

	// Test empty string fallbacks
	if loan.GetLoanAmount() != "" {
		t.Errorf("Expected empty string for LoanAmount, got %s", loan.GetLoanAmount())
	}

	if loan.GetLoanStatus() != "" {
		t.Errorf("Expected empty string for LoanStatus, got %s", loan.GetLoanStatus())
	}

	if loan.GetPrimaryCustomerName() != "" {
		t.Errorf("Expected empty string for PrimaryCustomerName, got %s", loan.GetPrimaryCustomerName())
	}
}

func TestLoan_CustomersArray(t *testing.T) {
	// Test loan with customers array (search result format)
	loan := &Loan{
		ID:        json.Number("999"),
		DisplayID: "LN00000999",
		CustomersArray: []LoanCustomer{
			{
				FirstName: "Bob",
				LastName:  "Wilson",
			},
		},
	}

	// Should use customers array
	if loan.GetPrimaryCustomerName() != "Bob Wilson" {
		t.Errorf("Expected customer name from array, got %s", loan.GetPrimaryCustomerName())
	}
}
