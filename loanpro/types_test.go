package loanpro

import (
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