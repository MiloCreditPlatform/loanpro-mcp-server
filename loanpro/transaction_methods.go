package loanpro

// Helper methods for Transaction

// GetID returns the transaction ID as string
func (t *Transaction) GetID() string {
	return string(t.ID)
}

// GetTxID returns the transaction identifier (e.g., "630-26-pay24360")
func (t *Transaction) GetTxID() string {
	return t.TxID
}

// GetEntityID returns the entity ID (loan ID)
func (t *Transaction) GetEntityID() string {
	return string(t.EntityID)
}

// GetDate returns the transaction date (with date parsing if needed)
func (t *Transaction) GetDate() string {
	if parsed, err := parseLoanProDate(t.Date); err == nil {
		return parsed
	}
	return t.Date
}

// GetPeriodStart returns the period start date (with date parsing if needed)
func (t *Transaction) GetPeriodStart() string {
	if parsed, err := parseLoanProDate(t.PeriodStart); err == nil {
		return parsed
	}
	return t.PeriodStart
}

// GetPeriodEnd returns the period end date (with date parsing if needed)
func (t *Transaction) GetPeriodEnd() string {
	if parsed, err := parseLoanProDate(t.PeriodEnd); err == nil {
		return parsed
	}
	return t.PeriodEnd
}

// GetType returns the transaction type
func (t *Transaction) GetType() string {
	return t.Type
}

// GetTitle returns the transaction title/description
func (t *Transaction) GetTitle() string {
	return t.Title
}

// IsInfoOnly checks if this is an informational transaction only
func (t *Transaction) IsInfoOnly() bool {
	return string(t.InfoOnly) == "1"
}

// GetInfoDetails returns the info details (JSON string)
func (t *Transaction) GetInfoDetails() string {
	return t.InfoDetails
}

// IsFuture checks if this is a future transaction
func (t *Transaction) IsFuture() bool {
	return string(t.Future) == "1"
}

// GetPaymentAmount returns the total payment amount
func (t *Transaction) GetPaymentAmount() string {
	return t.PaymentAmount
}

// GetPaymentPrincipal returns the principal portion of the payment
func (t *Transaction) GetPaymentPrincipal() string {
	return t.PaymentPrincipal
}

// GetPaymentInterest returns the interest portion of the payment
func (t *Transaction) GetPaymentInterest() string {
	return t.PaymentInterest
}

// GetPaymentFees returns the fees portion of the payment
func (t *Transaction) GetPaymentFees() string {
	return t.PaymentFees
}

// GetPaymentEscrow returns the escrow portion of the payment
func (t *Transaction) GetPaymentEscrow() string {
	return t.PaymentEscrow
}

// GetPaymentDiscount returns the discount portion of the payment
func (t *Transaction) GetPaymentDiscount() string {
	return t.PaymentDiscount
}

// GetChargeAmount returns the total charge amount
func (t *Transaction) GetChargeAmount() string {
	return t.ChargeAmount
}

// GetChargePrincipal returns the principal portion of the charge
func (t *Transaction) GetChargePrincipal() string {
	return t.ChargePrincipal
}

// GetChargeInterest returns the interest portion of the charge
func (t *Transaction) GetChargeInterest() string {
	return t.ChargeInterest
}

// GetChargeFees returns the fees portion of the charge
func (t *Transaction) GetChargeFees() string {
	return t.ChargeFees
}

// GetChargeEscrow returns the escrow portion of the charge
func (t *Transaction) GetChargeEscrow() string {
	return t.ChargeEscrow
}

// GetChargeDiscount returns the discount portion of the charge
func (t *Transaction) GetChargeDiscount() string {
	return t.ChargeDiscount
}

// GetPrincipalBalance returns the principal balance after this transaction
func (t *Transaction) GetPrincipalBalance() string {
	return t.PrincipalBalance
}

// HasPaymentBreakdown checks if the transaction has payment breakdown data
func (t *Transaction) HasPaymentBreakdown() bool {
	// Check for non-empty and non-zero amounts
	hasData := false
	if t.PaymentPrincipal != "" && t.PaymentPrincipal != "0" && t.PaymentPrincipal != "0.00" {
		hasData = true
	}
	if t.PaymentInterest != "" && t.PaymentInterest != "0" && t.PaymentInterest != "0.00" {
		hasData = true
	}
	if t.PaymentFees != "" && t.PaymentFees != "0" && t.PaymentFees != "0.00" {
		hasData = true
	}
	if t.PaymentEscrow != "" && t.PaymentEscrow != "0" && t.PaymentEscrow != "0.00" {
		hasData = true
	}
	return hasData
}

// HasChargeBreakdown checks if the transaction has charge breakdown data
func (t *Transaction) HasChargeBreakdown() bool {
	// Check for non-empty and non-zero amounts
	hasData := false
	if t.ChargePrincipal != "" && t.ChargePrincipal != "0" && t.ChargePrincipal != "0.00" {
		hasData = true
	}
	if t.ChargeInterest != "" && t.ChargeInterest != "0" && t.ChargeInterest != "0.00" {
		hasData = true
	}
	if t.ChargeFees != "" && t.ChargeFees != "0" && t.ChargeFees != "0.00" {
		hasData = true
	}
	if t.ChargeEscrow != "" && t.ChargeEscrow != "0" && t.ChargeEscrow != "0.00" {
		hasData = true
	}
	return hasData
}

// Legacy interface compatibility methods
// These methods provide backward compatibility with the old Transaction interface

// GetAmount returns the total transaction amount (payment or charge)
// For payments, returns payment amount; for charges, returns charge amount
func (t *Transaction) GetAmount() string {
	if t.PaymentAmount != "" && t.PaymentAmount != "0" && t.PaymentAmount != "0.00" {
		return t.PaymentAmount
	}
	if t.ChargeAmount != "" && t.ChargeAmount != "0" && t.ChargeAmount != "0.00" {
		return t.ChargeAmount
	}
	return "0"
}

// GetInfo returns transaction info/details
func (t *Transaction) GetInfo() string {
	return t.InfoDetails
}

// GetStatus returns the transaction status
// For the new transaction format, we check if it's info-only or future
func (t *Transaction) GetStatus() string {
	if t.IsInfoOnly() {
		return "Info Only"
	}
	if t.IsFuture() {
		return "Future"
	}
	return "Active"
}

// GetPrincipalAmount returns the principal amount from payment breakdown
func (t *Transaction) GetPrincipalAmount() string {
	return t.PaymentPrincipal
}

// GetInterestAmount returns the interest amount from payment breakdown
func (t *Transaction) GetInterestAmount() string {
	return t.PaymentInterest
}

// GetFeesAmount returns the fees amount from payment breakdown
func (t *Transaction) GetFeesAmount() string {
	return t.PaymentFees
}

// GetEscrowAmount returns the escrow amount from payment breakdown
func (t *Transaction) GetEscrowAmount() string {
	return t.PaymentEscrow
}
