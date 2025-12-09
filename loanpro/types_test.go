package loanpro

import (
	"encoding/json"
	"testing"
)

func TestParseLoanProDate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		{
			name:     "Valid Unix timestamp",
			input:    "/Date(1427829732)/",
			expected: "2015-03-31",
			hasError: false,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
			hasError: false,
		},
		{
			name:     "Already formatted date",
			input:    "2025-01-15",
			expected: "2025-01-15",
			hasError: false,
		},
		{
			name:     "Invalid timestamp",
			input:    "/Date(invalid)/",
			expected: "/Date(invalid)/",
			hasError: false,
		},
		{
			name:     "Zero timestamp",
			input:    "/Date(0)/",
			expected: "1970-01-01",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseLoanProDate(tt.input)

			if tt.hasError && err == nil {
				t.Errorf("Expected error but got none")
			}

			if !tt.hasError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestParseLoanProDateTime(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		{
			name:     "Valid Unix timestamp",
			input:    "/Date(1427829732)/",
			expected: "2015-03-31 19:22:12 UTC",
			hasError: false,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
			hasError: false,
		},
		{
			name:     "Already formatted datetime",
			input:    "2025-01-15 10:30:00 UTC",
			expected: "2025-01-15 10:30:00 UTC",
			hasError: false,
		},
		{
			name:     "Invalid timestamp",
			input:    "/Date(invalid)/",
			expected: "/Date(invalid)/",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseLoanProDateTime(tt.input)

			if tt.hasError && err == nil {
				t.Errorf("Expected error but got none")
			}

			if !tt.hasError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}
func TestPayment_GetStatus(t *testing.T) {
	tests := []struct {
		name     string
		active   string
		expected string
	}{
		{
			name:     "Active payment",
			active:   "1",
			expected: "Active",
		},
		{
			name:     "Inactive payment",
			active:   "0",
			expected: "Inactive",
		},
		{
			name:     "Empty active field",
			active:   "",
			expected: "Inactive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payment := &Payment{
				Active: json.Number(tt.active),
			}

			result := payment.GetStatus()

			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestPayment_GetMethods(t *testing.T) {
	payment := &Payment{
		ID:     json.Number("12345"),
		Amount: "500.00",
		Date:   "/Date(1427829732)/",
		Active: json.Number("1"),
	}

	// Test GetID
	if payment.GetID() != "12345" {
		t.Errorf("Expected ID 12345, got %s", payment.GetID())
	}

	// Test GetAmount
	if payment.GetAmount() != "500.00" {
		t.Errorf("Expected Amount 500.00, got %s", payment.GetAmount())
	}

	// Test GetDate
	expectedDate := "2015-03-31"
	if payment.GetDate() != expectedDate {
		t.Errorf("Expected Date %s, got %s", expectedDate, payment.GetDate())
	}

	// Test GetStatus
	if payment.GetStatus() != "Active" {
		t.Errorf("Expected Status Active, got %s", payment.GetStatus())
	}
}
