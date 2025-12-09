package loanpro

import "fmt"

// Helper methods for Loan to get string values from json.Number fields

// GetID returns the loan ID as string
func (l *Loan) GetID() string {
	return string(l.ID)
}

// GetDisplayID returns the display ID
func (l *Loan) GetDisplayID() string {
	return l.DisplayID
}

// GetActive returns the active status as string
func (l *Loan) GetActive() string {
	return string(l.Active)
}

// GetArchived returns the archived status as string
func (l *Loan) GetArchived() string {
	return string(l.Archived)
}

// GetDaysPastDue returns days past due from various sources
func (l *Loan) GetDaysPastDue() string {
	// Try search result field first
	if l.DaysPastDue != "" {
		return string(l.DaysPastDue)
	}
	// Fallback to StatusArchive for detailed loan view
	if l.StatusArchive != nil && len(l.StatusArchive.Results) > 0 {
		latest := l.StatusArchive.Results[len(l.StatusArchive.Results)-1]
		return string(latest.DaysPastDue)
	}
	return "N/A"
}

// GetLoanAmount returns the loan amount
func (l *Loan) GetLoanAmount() string {
	if l.LoanSetup != nil {
		return l.LoanSetup.LoanAmount
	}
	return ""
}

// GetPrincipalBalance returns the principal balance from various sources
func (l *Loan) GetPrincipalBalance() string {
	// Try search result field first
	if l.PrincipalBalance != "" {
		return l.PrincipalBalance
	}
	// Fallback to StatusArchive for detailed loan view
	if l.StatusArchive != nil && len(l.StatusArchive.Results) > 0 {
		latest := l.StatusArchive.Results[len(l.StatusArchive.Results)-1]
		return latest.PrincipalBalance
	}
	return "N/A"
}

// GetPayoffAmount returns the payoff amount
func (l *Loan) GetPayoffAmount() string {
	if l.StatusArchive != nil && len(l.StatusArchive.Results) > 0 {
		// Get the most recent status archive entry (should be sorted by date)
		latest := l.StatusArchive.Results[len(l.StatusArchive.Results)-1]
		return latest.Payoff
	}
	return "N/A"
}

// GetNextPaymentAmount returns the next payment amount from various sources
func (l *Loan) GetNextPaymentAmount() string {
	// Try search result field first
	if l.NextPaymentAmount != "" {
		return l.NextPaymentAmount
	}
	// Try StatusArchive next (more current)
	if l.StatusArchive != nil && len(l.StatusArchive.Results) > 0 {
		latest := l.StatusArchive.Results[len(l.StatusArchive.Results)-1]
		if latest.NextPaymentAmount != "" {
			return latest.NextPaymentAmount
		}
	}
	// Fallback to LoanSetup
	if l.LoanSetup != nil {
		return l.LoanSetup.Payment
	}
	return ""
}

// GetLoanStatus returns the loan status from various sources
func (l *Loan) GetLoanStatus() string {
	// Try search result field first
	if l.LoanStatusText != "" {
		return l.LoanStatusText
	}
	// Try StatusArchive next (more descriptive)
	if l.StatusArchive != nil && len(l.StatusArchive.Results) > 0 {
		latest := l.StatusArchive.Results[len(l.StatusArchive.Results)-1]
		if latest.LoanStatusText != "" {
			return latest.LoanStatusText
		}
	}
	// Fallback to status ID from LoanSettings
	if l.LoanSettings != nil {
		return fmt.Sprintf("Status ID: %s", string(l.LoanSettings.LoanStatusID))
	}
	return ""
}

// GetPrimaryCustomerName returns the primary customer name from various sources
func (l *Loan) GetPrimaryCustomerName() string {
	// Try search result field first
	if l.PrimaryCustomerName != "" {
		return l.PrimaryCustomerName
	}
	// Try customers array from search results
	if len(l.CustomersArray) > 0 {
		return l.CustomersArray[0].FirstName + " " + l.CustomersArray[0].LastName
	}
	// Fallback to expanded customer data (detailed view)
	if l.Customers != nil && len(l.Customers.Results) > 0 {
		return l.Customers.Results[0].FirstName + " " + l.Customers.Results[0].LastName
	}
	return ""
}

// GetCreatedDate returns the created date in human-readable format
func (l *Loan) GetCreatedDate() string {
	if parsed, err := parseLoanProDateTime(l.Created); err == nil {
		return parsed
	}
	return l.Created
}

// GetContractDate returns the contract date in human-readable format
func (l *Loan) GetContractDate() string {
	if l.LoanSetup != nil {
		if parsed, err := parseLoanProDate(l.LoanSetup.ContractDate); err == nil {
			return parsed
		}
		return l.LoanSetup.ContractDate
	}
	return ""
}

// GetNextPaymentDate returns the next payment date in human-readable format
func (l *Loan) GetNextPaymentDate() string {
	// Try search result field first
	if l.NextPaymentDate != "" {
		if parsed, err := parseLoanProDate(l.NextPaymentDate); err == nil {
			return parsed
		}
		return l.NextPaymentDate
	}
	// Try StatusArchive next (more current)
	if l.StatusArchive != nil && len(l.StatusArchive.Results) > 0 {
		latest := l.StatusArchive.Results[len(l.StatusArchive.Results)-1]
		if latest.NextPaymentDate != "" {
			if parsed, err := parseLoanProDate(latest.NextPaymentDate); err == nil {
				return parsed
			}
			return latest.NextPaymentDate
		}
	}
	// Fallback to LoanSetup
	if l.LoanSetup != nil {
		if parsed, err := parseLoanProDate(l.LoanSetup.FirstPaymentDate); err == nil {
			return parsed
		}
		return l.LoanSetup.FirstPaymentDate
	}
	return ""
}
