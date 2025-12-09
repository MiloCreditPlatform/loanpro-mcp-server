package loanpro

// Helper methods for Transaction

// GetID returns the transaction ID as string
func (t *Transaction) GetID() string {
	return string(t.ID)
}

// GetAmount returns the transaction amount
func (t *Transaction) GetAmount() string {
	return t.Amount
}

// GetDate returns the transaction date (with date parsing if needed)
func (t *Transaction) GetDate() string {
	if parsed, err := parseLoanProDate(t.Date); err == nil {
		return parsed
	}
	return t.Date
}

// GetType returns the transaction type
func (t *Transaction) GetType() string {
	return t.Type
}

// GetTitle returns the transaction title/description
func (t *Transaction) GetTitle() string {
	return t.Title
}

// GetInfo returns the transaction info
func (t *Transaction) GetInfo() string {
	return t.Info
}

// GetStatus returns the transaction status as a human-readable string
func (t *Transaction) GetStatus() string {
	// Check if reversed first
	reversed := string(t.Reversed)
	if reversed == "1" {
		return "Reversed"
	}
	
	// Check active field: 1 = Active, 0 = Inactive (Voided)
	active := string(t.Active)
	if active == "1" {
		return "Active"
	}
	return "Voided"
}

// GetPrincipalAmount returns the principal portion of the transaction
func (t *Transaction) GetPrincipalAmount() string {
	return t.PrincipalAmount
}

// GetInterestAmount returns the interest portion of the transaction
func (t *Transaction) GetInterestAmount() string {
	return t.InterestAmount
}

// GetFeesAmount returns the fees portion of the transaction
func (t *Transaction) GetFeesAmount() string {
	return t.FeesAmount
}

// GetEscrowAmount returns the escrow portion of the transaction
func (t *Transaction) GetEscrowAmount() string {
	return t.EscrowAmount
}

// HasPaymentBreakdown checks if the transaction has payment breakdown data
func (t *Transaction) HasPaymentBreakdown() bool {
	return t.PrincipalAmount != "" || t.InterestAmount != "" || 
		t.FeesAmount != "" || t.EscrowAmount != ""
}
