package loanpro

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// LoanSettings represents loan settings data
type LoanSettings struct {
	ID              json.Number `json:"id"`
	LoanID          json.Number `json:"loanId"`
	LoanStatusID    json.Number `json:"loanStatusId"`
	LoanSubStatusID json.Number `json:"loanSubStatusId"`
	AutopayEnabled  json.Number `json:"autopayEnabled"`
}

// LoanSetup represents loan setup data
type LoanSetup struct {
	ID               json.Number `json:"id"`
	LoanID           json.Number `json:"loanId"`
	ContractDate     string      `json:"contractDate"`
	LoanType         string      `json:"loanType"`
	LoanClass        string      `json:"loanClass"`
	LoanAmount       string      `json:"loanAmount"`
	Payment          string      `json:"payment"`
	FirstPaymentDate string      `json:"firstPaymentDate"`
	LoanRate         string      `json:"loanRate"`
	LoanTerm         string      `json:"loanTerm"`
}

// LoanCustomer represents customer data in loan context
type LoanCustomer struct {
	ID        json.Number `json:"id"`
	FirstName string      `json:"firstName"`
	LastName  string      `json:"lastName"`
	Email     string      `json:"email"`
}

// CustomersWrapper wraps customer results
type CustomersWrapper struct {
	Results []LoanCustomer `json:"results"`
}

// StatusArchiveEntry represents loan status history
type StatusArchiveEntry struct {
	ID                json.Number `json:"id"`
	LoanID            json.Number `json:"loanId"`
	Date              string      `json:"date"`
	PrincipalBalance  string      `json:"principalBalance"`
	Payoff            string      `json:"payoff"`
	AmountDue         string      `json:"amountDue"`
	DaysPastDue       json.Number `json:"daysPastDue"`
	NextPaymentDate   string      `json:"nextPaymentDate"`
	NextPaymentAmount string      `json:"nextPaymentAmount"`
	LoanStatusText    string      `json:"loanStatusText"`
}

// StatusArchiveWrapper wraps status archive results
type StatusArchiveWrapper struct {
	Results []StatusArchiveEntry `json:"results"`
}

// Loan represents loan data
type Loan struct {
	ID            json.Number           `json:"id"`
	DisplayID     string                `json:"displayId"`
	Title         string                `json:"title"`
	Active        json.Number           `json:"active"`
	Archived      json.Number           `json:"archived"`
	Created       string                `json:"created"`
	LoanSettings  *LoanSettings         `json:"LoanSettings,omitempty"`
	LoanSetup     *LoanSetup            `json:"LoanSetup,omitempty"`
	Customers     *CustomersWrapper     `json:"Customers,omitempty"`
	StatusArchive *StatusArchiveWrapper `json:"StatusArchive,omitempty"`
	// Fields that appear in search results but not in individual loan retrieval
	PrimaryCustomerName string      `json:"primaryCustomerName,omitempty"`
	LoanStatusText      string      `json:"loanStatusText,omitempty"`
	PrincipalBalance    string      `json:"principalBalance,omitempty"`
	DaysPastDue         json.Number `json:"daysPastDue,omitempty"`
	NextPaymentAmount   string      `json:"nextPaymentAmount,omitempty"`
	NextPaymentDate     string      `json:"nextPaymentDate,omitempty"`
	// Raw customers array from search results (lowercase)
	CustomersArray []LoanCustomer `json:"customers,omitempty"`
}

// Customer represents customer data
type Customer struct {
	ID        int    `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	CreatedAt string `json:"createdAt"`
}

// Payment represents payment data
type Payment struct {
	ID              json.Number `json:"id"`
	LoanID          json.Number `json:"loanId"`
	Date            string      `json:"date"`
	Amount          string      `json:"amount"`
	PaymentTypeID   json.Number `json:"paymentTypeId"`
	PaymentMethodID json.Number `json:"paymentMethodId"`
	Info            string      `json:"info"`
	Active          json.Number `json:"active"`
}

// PaymentsWrapper wraps payment results
type PaymentsWrapper struct {
	Results []Payment `json:"results"`
}

// Transaction represents transaction data from the LoanTransactions endpoint
type Transaction struct {
	ID                     json.Number `json:"id"`
	TxID                   string      `json:"txId"`
	EntityType             string      `json:"entityType"`
	EntityID               json.Number `json:"entityId"`
	ModID                  json.Number `json:"modId"`
	Date                   string      `json:"date"`
	Period                 json.Number `json:"period"`
	PeriodStart            string      `json:"periodStart"`
	PeriodEnd              string      `json:"periodEnd"`
	Title                  string      `json:"title"`
	Type                   string      `json:"type"`
	InfoOnly               json.Number `json:"infoOnly"`
	InfoDetails            string      `json:"infoDetails"`
	PaymentID              json.Number `json:"paymentId"`
	PaymentDisplayID       json.Number `json:"paymentDisplayId"`
	PaymentAmount          string      `json:"paymentAmount"`
	PaymentInterest        string      `json:"paymentInterest"`
	PaymentPrincipal       string      `json:"paymentPrincipal"`
	PaymentDiscount        string      `json:"paymentDiscount"`
	PaymentFees            string      `json:"paymentFees"`
	FeesPaidDetails        string      `json:"feesPaidDetails"`
	PaymentEscrow          string      `json:"paymentEscrow"`
	PaymentEscrowBreakdown string      `json:"paymentEscrowBreakdown"`
	ChargeAmount           string      `json:"chargeAmount"`
	ChargeInterest         string      `json:"chargeInterest"`
	ChargePrincipal        string      `json:"chargePrincipal"`
	ChargeDiscount         string      `json:"chargeDiscount"`
	ChargeFees             string      `json:"chargeFees"`
	ChargeEscrow           string      `json:"chargeEscrow"`
	ChargeEscrowBreakdown  string      `json:"chargeEscrowBreakdown"`
	Future                 json.Number `json:"future"`
	PrincipalOnly          json.Number `json:"principalOnly"`
	Advancement            json.Number `json:"advancement"`
	PayoffFee              json.Number `json:"payoffFee"`
	ChargeOff              json.Number `json:"chargeOff"`
	PaymentType            json.Number `json:"paymentType"`
	AdbDays                json.Number `json:"adbDays"`
	Adb                    string      `json:"adb"`
	PrincipalBalance       string      `json:"principalBalance"`
	DisplayOrder           string      `json:"displayOrder"`
}

// TransactionsWrapper wraps transaction results
type TransactionsWrapper struct {
	Results []Transaction `json:"results"`
	Summary struct {
		Start    json.Number `json:"start"`
		PageSize json.Number `json:"pageSize"`
		Total    json.Number `json:"total"`
	} `json:"summary"`
}

// Response wrapper types
type ODataResponse struct {
	D any `json:"d"`
}

type SearchResponse struct {
	D struct {
		Results []Loan `json:"results"`
		Summary struct {
			TotalHits int `json:"totalHits"`
			TotalTime int `json:"totalTime"`
		} `json:"summary"`
	} `json:"d"`
}

type CustomerSearchResponse struct {
	D struct {
		Results []Customer `json:"results"`
		Summary struct {
			TotalHits int `json:"totalHits"`
			TotalTime int `json:"totalTime"`
		} `json:"summary"`
	} `json:"d"`
}

// Utility functions for date parsing
func parseLoanProDate(dateStr string) (string, error) {
	if dateStr == "" {
		return "", nil
	}

	// Match the LoanPro date format: /Date(1234567890)/
	re := regexp.MustCompile(`/Date\((\d+)\)/`)
	matches := re.FindStringSubmatch(dateStr)

	if len(matches) != 2 {
		// If it doesn't match the Unix format, assume it's already in YYYY-MM-DD format
		return dateStr, nil
	}

	// Parse the Unix timestamp (in seconds)
	timestamp, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return "", fmt.Errorf("failed to parse timestamp: %w", err)
	}

	// Convert to time and format as YYYY-MM-DD
	t := time.Unix(timestamp, 0).UTC()
	return t.Format("2006-01-02"), nil
}

func parseLoanProDateTime(dateStr string) (string, error) {
	if dateStr == "" {
		return "", nil
	}

	// Match the LoanPro date format: /Date(1234567890)/
	re := regexp.MustCompile(`/Date\((\d+)\)/`)
	matches := re.FindStringSubmatch(dateStr)

	if len(matches) != 2 {
		// If it doesn't match the Unix format, assume it's already formatted
		return dateStr, nil
	}

	// Parse the Unix timestamp (in seconds)
	timestamp, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return "", fmt.Errorf("failed to parse timestamp: %w", err)
	}

	// Convert to time and format as YYYY-MM-DD HH:MM:SS UTC
	t := time.Unix(timestamp, 0).UTC()
	return t.Format("2006-01-02 15:04:05 UTC"), nil
}
