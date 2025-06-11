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