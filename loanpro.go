package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"time"
)

// parseLoanProDate parses LoanPro Unix timestamp format: /Date(1427829732)/
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

// parseLoanProDateTime parses LoanPro Unix timestamp and returns full date-time
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

type LoanProClient struct {
	baseURL  string
	apiKey   string
	tenantID string
	client   *http.Client
}

type LoanSettings struct {
	ID                json.Number `json:"id"`
	LoanID            json.Number `json:"loanId"`
	LoanStatusID      json.Number `json:"loanStatusId"`
	LoanSubStatusID   json.Number `json:"loanSubStatusId"`
	AutopayEnabled    json.Number `json:"autopayEnabled"`
}

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

type LoanCustomer struct {
	ID        json.Number `json:"id"`
	FirstName string      `json:"firstName"`
	LastName  string      `json:"lastName"`
	Email     string      `json:"email"`
}

type CustomersWrapper struct {
	Results []LoanCustomer `json:"results"`
}

type StatusArchiveEntry struct {
	ID               json.Number `json:"id"`
	LoanID           json.Number `json:"loanId"`
	Date             string      `json:"date"`
	PrincipalBalance string      `json:"principalBalance"`
	Payoff           string      `json:"payoff"`
	AmountDue        string      `json:"amountDue"`
	DaysPastDue      json.Number `json:"daysPastDue"`
	NextPaymentDate  string      `json:"nextPaymentDate"`
	NextPaymentAmount string     `json:"nextPaymentAmount"`
	LoanStatusText   string      `json:"loanStatusText"`
}

type StatusArchiveWrapper struct {
	Results []StatusArchiveEntry `json:"results"`
}

type Loan struct {
	ID                   json.Number            `json:"id"`
	DisplayID            string                 `json:"displayId"`
	Title                string                 `json:"title"`
	Active               json.Number            `json:"active"`
	Archived             json.Number            `json:"archived"`
	Created              string                 `json:"created"`
	LoanSettings         *LoanSettings          `json:"LoanSettings,omitempty"`
	LoanSetup            *LoanSetup             `json:"LoanSetup,omitempty"`
	Customers            *CustomersWrapper      `json:"Customers,omitempty"`
	StatusArchive        *StatusArchiveWrapper  `json:"StatusArchive,omitempty"`
	// Fields that appear in search results but not in individual loan retrieval
	PrimaryCustomerName  string                 `json:"primaryCustomerName,omitempty"`
	LoanStatusText       string                 `json:"loanStatusText,omitempty"`
	PrincipalBalance     string                 `json:"principalBalance,omitempty"`
	DaysPastDue          json.Number            `json:"daysPastDue,omitempty"`
	NextPaymentAmount    string                 `json:"nextPaymentAmount,omitempty"`
	NextPaymentDate      string                 `json:"nextPaymentDate,omitempty"`
	// Raw customers array from search results (lowercase)
	CustomersArray       []LoanCustomer         `json:"customers,omitempty"`
}

// Helper methods to get string values from json.Number fields
func (l *Loan) GetID() string {
	return string(l.ID)
}

func (l *Loan) GetActive() string {
	return string(l.Active)
}

func (l *Loan) GetArchived() string {
	return string(l.Archived)
}

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

func (l *Loan) GetLoanAmount() string {
	if l.LoanSetup != nil {
		return l.LoanSetup.LoanAmount
	}
	return ""
}

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

func (l *Loan) GetPayoffAmount() string {
	if l.StatusArchive != nil && len(l.StatusArchive.Results) > 0 {
		// Get the most recent status archive entry (should be sorted by date)
		latest := l.StatusArchive.Results[len(l.StatusArchive.Results)-1]
		return latest.Payoff
	}
	return "N/A"
}

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

type Customer struct {
	ID        int    `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	CreatedAt string `json:"createdAt"`
}

// GetCreatedDate returns the created date in human-readable format
func (c *Customer) GetCreatedDate() string {
	if parsed, err := parseLoanProDateTime(c.CreatedAt); err == nil {
		return parsed
	}
	return c.CreatedAt
}

type LoanProODataResponse struct {
	D any `json:"d"`
}

type LoanProSearchResponse struct {
	D struct {
		Results []Loan `json:"results"`
		Summary struct {
			TotalHits int `json:"totalHits"`
			TotalTime int `json:"totalTime"`
		} `json:"summary"`
	} `json:"d"`
}

type LoanProCustomerSearchResponse struct {
	D struct {
		Results []Customer `json:"results"`
		Summary struct {
			TotalHits int `json:"totalHits"`
			TotalTime int `json:"totalTime"`
		} `json:"summary"`
	} `json:"d"`
}

func NewLoanProClient(baseURL, apiKey, tenantID string) *LoanProClient {
	return &LoanProClient{
		baseURL:  baseURL,
		apiKey:   apiKey,
		tenantID: tenantID,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *LoanProClient) makeRequest(endpoint string, params map[string]string) ([]byte, error) {
	return c.makeRequestWithMethod("GET", endpoint, params, nil)
}

func (c *LoanProClient) makePostRequest(endpoint string, body any) ([]byte, error) {
	return c.makeRequestWithMethod("POST", endpoint, nil, body)
}

func (c *LoanProClient) makeRequestWithMethod(method, endpoint string, params map[string]string, body any) ([]byte, error) {
	u, err := url.Parse(c.baseURL + endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if len(params) > 0 {
		q := u.Query()
		for key, value := range params {
			q.Add(key, value)
		}
		u.RawQuery = q.Encode()
	}

	var requestBody io.Reader
	var bodyStr string
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] Failed to marshal request body: %v\n", err)
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyStr = string(bodyBytes)
		requestBody = bytes.NewReader(bodyBytes)
	}

	log.Printf("[LOANPRO] %s %s", method, u.String())
	if bodyStr != "" {
		log.Printf("[LOANPRO] Request body: %s", bodyStr)
	}

	req, err := http.NewRequest(method, u.String(), requestBody)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to create HTTP request: %v\n", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Autopal-Instance-Id", c.tenantID)
	req.Header.Set("Authorization", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		log.Printf("[LOANPRO] Request failed: %v", err)
		fmt.Fprintf(os.Stderr, "[ERROR] LoanPro API request failed: %v\n", err)
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("[LOANPRO] Response status: %d %s", resp.StatusCode, resp.Status)

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[LOANPRO] Failed to read response body: %v", err)
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to read LoanPro response body: %v\n", err)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	log.Printf("[LOANPRO] Response body: %s", string(responseBody))

	if resp.StatusCode != http.StatusOK {
		log.Printf("[LOANPRO] API error - status %d, body: %s", resp.StatusCode, string(responseBody))
		fmt.Fprintf(os.Stderr, "[ERROR] LoanPro API returned status %d: %s\n", resp.StatusCode, string(responseBody))
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	return responseBody, nil
}

func (c *LoanProClient) GetLoan(loanID string) (*Loan, error) {
	// Use OData expand to include related data that provides loan amounts, status, and customer info
	params := map[string]string{
		"$expand": "LoanSettings,LoanSetup,Customers,StatusArchive",
	}
	
	body, err := c.makeRequest("/public/api/1/odata.svc/Loans("+loanID+")", params)
	if err != nil {
		return nil, err
	}

	var response LoanProODataResponse
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to parse GetLoan response: %v\nResponse body: %s\n", err, string(body))
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	loanData, err := json.Marshal(response.D)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to marshal loan data: %v\n", err)
		return nil, fmt.Errorf("failed to marshal loan data: %w", err)
	}

	var loan Loan
	if err := json.Unmarshal(loanData, &loan); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to parse loan struct: %v\nLoan data: %s\n", err, string(loanData))
		return nil, fmt.Errorf("failed to parse loan: %w", err)
	}

	return &loan, nil
}

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

type PaymentsWrapper struct {
	Results []Payment `json:"results"`
}

func (c *LoanProClient) GetLoanPayments(loanID string) ([]Payment, error) {
	// Use OData expand to get payment history
	params := map[string]string{
		"$expand": "Payments",
	}
	
	body, err := c.makeRequest("/public/api/1/odata.svc/Loans("+loanID+")", params)
	if err != nil {
		return nil, err
	}

	var response LoanProODataResponse
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to parse GetLoanPayments response: %v\nResponse body: %s\n", err, string(body))
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	loanData, err := json.Marshal(response.D)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to marshal loan data: %v\n", err)
		return nil, fmt.Errorf("failed to marshal loan data: %w", err)
	}

	var loanWithPayments struct {
		Payments *PaymentsWrapper `json:"Payments,omitempty"`
	}
	
	if err := json.Unmarshal(loanData, &loanWithPayments); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to parse loan payments: %v\nLoan data: %s\n", err, string(loanData))
		return nil, fmt.Errorf("failed to parse loan payments: %w", err)
	}

	if loanWithPayments.Payments != nil {
		return loanWithPayments.Payments.Results, nil
	}
	
	return []Payment{}, nil
}

func (c *LoanProClient) SearchLoans(searchTerm, status string, limit int) ([]Loan, error) {
	// Build the search query according to LoanPro API format
	searchBody := map[string]any{
		"size": limit, // Use 'size' for pagination limit
	}
	
	// Build query conditions
	var mustConditions []map[string]any
	var shouldConditions []map[string]any
	
	// Add search term conditions if provided
	if searchTerm != "" {
		shouldConditions = append(shouldConditions, 
			map[string]any{
				"query_string": map[string]any{
					"query":   "*" + searchTerm + "*",
					"fields":  []string{"displayId", "primaryCustomerName", "title"},
					"default_operator": "and",
				},
			},
			map[string]any{
				"match": map[string]any{
					"displayId": searchTerm,
				},
			},
			map[string]any{
				"match": map[string]any{
					"primaryCustomerName": searchTerm,
				},
			},
		)
	}
	
	// Add status filter if provided
	if status != "" {
		mustConditions = append(mustConditions, map[string]any{
			"match": map[string]any{
				"loanStatusText": status,
			},
		})
	}
	
	// Build the final query
	if len(mustConditions) > 0 || len(shouldConditions) > 0 {
		boolQuery := map[string]any{}
		
		if len(mustConditions) > 0 {
			if len(mustConditions) == 1 {
				boolQuery["must"] = mustConditions[0]
			} else {
				boolQuery["must"] = mustConditions
			}
		}
		
		if len(shouldConditions) > 0 {
			boolQuery["should"] = shouldConditions
			boolQuery["minimum_should_match"] = 1
		}
		
		searchBody["query"] = map[string]any{
			"bool": boolQuery,
		}
	} else {
		// If no filters, use match_all query
		searchBody["query"] = map[string]any{
			"match_all": map[string]any{},
		}
	}

	body, err := c.makePostRequest("/public/api/1/Loans/Autopal.Search()", searchBody)
	if err != nil {
		return nil, err
	}

	var response LoanProSearchResponse
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to parse SearchLoans response: %v\nResponse body: %s\n", err, string(body))
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return response.D.Results, nil
}

func (c *LoanProClient) GetCustomer(customerID string) (*Customer, error) {
	body, err := c.makeRequest("/public/api/1/odata.svc/Customers("+customerID+")", nil)
	if err != nil {
		return nil, err
	}

	var response LoanProODataResponse
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to parse GetCustomer response: %v\nResponse body: %s\n", err, string(body))
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	customerData, err := json.Marshal(response.D)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to marshal customer data: %v\n", err)
		return nil, fmt.Errorf("failed to marshal customer data: %w", err)
	}

	var customer Customer
	if err := json.Unmarshal(customerData, &customer); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to parse customer struct: %v\nCustomer data: %s\n", err, string(customerData))
		return nil, fmt.Errorf("failed to parse customer: %w", err)
	}

	return &customer, nil
}

func (c *LoanProClient) SearchCustomers(searchTerm string, limit int) ([]Customer, error) {
	// Build the search query according to LoanPro Customer Search API format
	searchBody := map[string]any{
		"size": limit, // Use 'size' for pagination limit
	}
	
	// Add query filters if provided
	if searchTerm != "" {
		searchBody["query"] = map[string]any{
			"bool": map[string]any{
				"should": []map[string]any{
					{
						"query_string": map[string]any{
							"query":   "*" + searchTerm + "*",
							"fields":  []string{"firstName", "lastName", "email", "ssn", "companyName"},
							"default_operator": "and",
						},
					},
					{
						"match": map[string]any{
							"firstName": searchTerm,
						},
					},
					{
						"match": map[string]any{
							"lastName": searchTerm,
						},
					},
				},
				"minimum_should_match": 1,
			},
		}
	} else {
		// If no filter, use match_all query
		searchBody["query"] = map[string]any{
			"match_all": map[string]any{},
		}
	}

	body, err := c.makePostRequest("/public/api/1/Customers/Autopal.Search()", searchBody)
	if err != nil {
		return nil, err
	}

	var response LoanProCustomerSearchResponse
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to parse SearchCustomers response: %v\nResponse body: %s\n", err, string(body))
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return response.D.Results, nil
}