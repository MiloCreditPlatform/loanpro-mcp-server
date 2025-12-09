package loanpro

// Helper methods for Payment

// GetID returns the payment ID as string
func (p *Payment) GetID() string {
	return string(p.ID)
}

// GetAmount returns the payment amount
func (p *Payment) GetAmount() string {
	return p.Amount
}

// GetDate returns the payment date (with date parsing if needed)
func (p *Payment) GetDate() string {
	if parsed, err := parseLoanProDate(p.Date); err == nil {
		return parsed
	}
	return p.Date
}

// GetStatus returns the payment status as a human-readable string
func (p *Payment) GetStatus() string {
	// Active field: 1 = Active, 0 = Inactive (Reversed/Voided)
	active := string(p.Active)
	if active == "1" {
		return "Active"
	}
	return "Inactive"
}
