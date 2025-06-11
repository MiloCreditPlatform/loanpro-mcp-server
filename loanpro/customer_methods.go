package loanpro

// Helper methods for Customer

// GetID returns the customer ID
func (c *Customer) GetID() int {
	return c.ID
}

// GetFirstName returns the first name
func (c *Customer) GetFirstName() string {
	return c.FirstName
}

// GetLastName returns the last name
func (c *Customer) GetLastName() string {
	return c.LastName
}

// GetEmail returns the email
func (c *Customer) GetEmail() string {
	return c.Email
}

// GetPhone returns the phone number
func (c *Customer) GetPhone() string {
	return c.Phone
}

// GetCreatedDate returns the created date in human-readable format
func (c *Customer) GetCreatedDate() string {
	if parsed, err := parseLoanProDateTime(c.CreatedAt); err == nil {
		return parsed
	}
	return c.CreatedAt
}