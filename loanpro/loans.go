package loanpro

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// GetLoan retrieves a loan by ID with expanded data
func (c *Client) GetLoan(loanID string) (*Loan, error) {
	// Use OData expand to include related data that provides loan amounts, status, and customer info
	params := map[string]string{
		"$expand": "LoanSettings,LoanSetup,Customers,StatusArchive",
	}

	body, err := c.makeRequest("/public/api/1/odata.svc/Loans("+loanID+")", params)
	if err != nil {
		return nil, err
	}

	var response ODataResponse
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

// SearchLoans searches for loans using the search API
func (c *Client) SearchLoans(searchTerm, status string, limit, offset int) ([]Loan, error) {
	// Build the search query according to LoanPro API format
	searchBody := map[string]any{
		"size": limit,
		"from": offset,
	}

	// Build query conditions
	var mustConditions []map[string]any
	var shouldConditions []map[string]any

	// Add search term conditions if provided
	if searchTerm != "" {
		shouldConditions = append(shouldConditions,
			map[string]any{
				"query_string": map[string]any{
					"query":            "*" + searchTerm + "*",
					"fields":           []string{"displayId", "primaryCustomerName", "title"},
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

	var response SearchResponse
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to parse SearchLoans response: %v\nResponse body: %s\n", err, string(body))
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return response.D.Results, nil
}

const maxPastDueLimit = 200
const odataBatchSize = 50

// GetPastDueLoans retrieves open loans with more than minDaysPastDue days past due.
// Elasticsearch is used to filter by daysPastDue (not available as an OData filter),
// then each matched loan is fetched individually via OData with $expand so that
// LoanSetup and other nested entities are fully populated.
func (c *Client) GetPastDueLoans(minDaysPastDue, limit, offset int) ([]Loan, error) {
	if minDaysPastDue < 0 {
		return nil, fmt.Errorf("minDaysPastDue must be non-negative")
	}
	if limit <= 0 {
		return nil, fmt.Errorf("limit must be positive")
	}
	if offset < 0 {
		return nil, fmt.Errorf("offset must be non-negative")
	}
	if limit > maxPastDueLimit {
		limit = maxPastDueLimit
	}

	searchBody := map[string]any{
		"size": limit,
		"query": map[string]any{
			"bool": map[string]any{
				"must": []map[string]any{
					{
						"range": map[string]any{
							"daysPastDue": map[string]any{
								"gt": minDaysPastDue,
							},
						},
					},
					{
						"match": map[string]any{
							"loanStatusText": "Open",
						},
					},
				},
			},
		},
		"sort": []map[string]any{
			{"daysPastDue": map[string]any{"order": "asc"}},
		},
	}
	if offset > 0 {
		searchBody["from"] = offset
	}

	body, err := c.makePostRequest("/public/api/1/Loans/Autopal.Search()", searchBody)
	if err != nil {
		return nil, err
	}

	var searchResp SearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to parse GetPastDueLoans response: %v\nResponse body: %s\n", err, string(body))
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(searchResp.D.Results) == 0 {
		return []Loan{}, nil
	}

	// Validate IDs and chunk into batches to stay within OData URL length limits.
	// Each batch gets its own $filter=id eq X or id eq Y request with $expand.
	results := searchResp.D.Results
	loans := make([]Loan, 0, len(results))
	for start := 0; start < len(results); start += odataBatchSize {
		end := start + odataBatchSize
		if end > len(results) {
			end = len(results)
		}

		filterParts := make([]string, 0, end-start)
		for _, result := range results[start:end] {
			idStr := strings.TrimSpace(string(result.ID))
			if _, err := strconv.Atoi(idStr); err != nil {
				return nil, fmt.Errorf("invalid loan id in search response: %q", idStr)
			}
			filterParts = append(filterParts, "id eq "+idStr)
		}

		params := map[string]string{
			"$filter": strings.Join(filterParts, " or "),
			"$expand": "LoanSettings,LoanSetup,Customers,StatusArchive",
			"$top":    strconv.Itoa(end - start),
		}

		odataBody, err := c.makeRequest("/public/api/1/odata.svc/Loans", params)
		if err != nil {
			return nil, err
		}

		var batch ODataLoansResponse
		if err := json.Unmarshal(odataBody, &batch); err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] Failed to parse GetPastDueLoans OData response: %v\nResponse body: %s\n", err, string(odataBody))
			return nil, fmt.Errorf("failed to parse OData response: %w", err)
		}
		loans = append(loans, batch.D.Results...)
	}

	if len(loans) == 0 {
		return nil, fmt.Errorf("failed to fetch loan details: Elasticsearch returned %d matches but OData returned no results", len(results))
	}

	// OData doesn't preserve Elasticsearch sort order; re-sort by daysPastDue ascending.
	// Loans with unparseable daysPastDue are pushed to the end.
	sort.Slice(loans, func(i, j int) bool {
		dpdI, errI := strconv.Atoi(loans[i].GetDaysPastDue())
		dpdJ, errJ := strconv.Atoi(loans[j].GetDaysPastDue())
		if errI != nil {
			dpdI = int(^uint(0) >> 1)
		}
		if errJ != nil {
			dpdJ = int(^uint(0) >> 1)
		}
		return dpdI < dpdJ
	})

	return loans, nil
}

// ArchiveLoan archives a loan, removing it from live, billable accounts.
// LoanPro charges a one-time fee per archived loan; setting archived back to 0 resurrects it.
func (c *Client) ArchiveLoan(loanID string) error {
	if _, err := strconv.Atoi(loanID); err != nil {
		return fmt.Errorf("invalid loan id: %q", loanID)
	}

	_, err := c.makePutRequest("/public/api/1/odata.svc/Loans("+loanID+")", map[string]any{
		"archived": 1,
	})
	return err
}

// FindPortfolioID resolves a portfolio name, optionally within a named portfolio category,
// to its ID. Names are matched case-insensitively. Portfolio names are not unique across
// categories, so an ambiguous name without a category is an error.
func (c *Client) FindPortfolioID(category, portfolio string) (string, error) {
	body, err := c.makeRequest("/public/api/1/odata.svc/PortfolioCategories", map[string]string{
		"$expand": "Portfolios",
		"$top":    "1000",
	})
	if err != nil {
		return "", err
	}

	var response ODataPortfolioCategoriesResponse
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to parse PortfolioCategories response: %v\nResponse body: %s\n", err, string(body))
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	var ids, descriptions []string
	for _, cat := range response.D.Results {
		if category != "" && !strings.EqualFold(strings.TrimSpace(cat.Title), strings.TrimSpace(category)) {
			continue
		}
		if cat.Portfolios == nil {
			continue
		}
		for _, p := range cat.Portfolios.Results {
			if strings.EqualFold(strings.TrimSpace(p.Title), strings.TrimSpace(portfolio)) {
				ids = append(ids, string(p.ID))
				descriptions = append(descriptions, fmt.Sprintf("%s > %s (id %s)", cat.Title, p.Title, p.ID))
			}
		}
	}

	switch len(ids) {
	case 0:
		if category != "" {
			return "", fmt.Errorf("portfolio %q not found in category %q", portfolio, category)
		}
		return "", fmt.Errorf("portfolio %q not found", portfolio)
	case 1:
		return ids[0], nil
	default:
		return "", fmt.Errorf("portfolio %q is ambiguous, specify its category: %s", portfolio, strings.Join(descriptions, "; "))
	}
}

const portfolioSearchPageSize = 500
const maxPortfolioSearchHits = 10000

// GetLoansByPortfolio returns live (non-archived) loans in the named portfolio (optionally
// within a portfolio category, and optionally also in a named sub-portfolio), created within
// [createdFrom, createdTo] (YYYY-MM-DD, inclusive, UTC). Elasticsearch filters by portfolio ID
// and created date; matches are then fetched via OData to confirm membership and archive state.
func (c *Client) GetLoansByPortfolio(category, portfolio, subPortfolio, createdFrom, createdTo string) ([]Loan, error) {
	if strings.TrimSpace(portfolio) == "" {
		return nil, fmt.Errorf("portfolio is required")
	}

	// The search index stores created as "yyyy-MM-dd HH:mm:ss" (UTC) and rejects date-only values.
	const createdLayout = "2006-01-02 15:04:05"
	createdRange := map[string]any{}
	if createdFrom != "" {
		from, err := time.Parse("2006-01-02", createdFrom)
		if err != nil {
			return nil, fmt.Errorf("createdFrom must be YYYY-MM-DD: %w", err)
		}
		createdRange["gte"] = from.Format(createdLayout)
	}
	if createdTo != "" {
		// Use "< next day" so the whole end date is included.
		to, err := time.Parse("2006-01-02", createdTo)
		if err != nil {
			return nil, fmt.Errorf("createdTo must be YYYY-MM-DD: %w", err)
		}
		createdRange["lt"] = to.AddDate(0, 0, 1).Format(createdLayout)
	}

	portfolioID, err := c.FindPortfolioID(category, portfolio)
	if err != nil {
		return nil, err
	}

	filters := []map[string]any{
		{"terms": map[string]any{"portfolios": []string{portfolioID}}},
	}
	if len(createdRange) > 0 {
		filters = append(filters, map[string]any{"range": map[string]any{"created": createdRange}})
	}

	var ids []string
	for offset := 0; ; offset += portfolioSearchPageSize {
		searchBody := map[string]any{
			"size":  portfolioSearchPageSize,
			"from":  offset,
			"query": map[string]any{"bool": map[string]any{"filter": filters}},
			"sort":  []map[string]any{{"id": map[string]any{"order": "asc"}}},
		}
		body, err := c.makePostRequest("/public/api/1/Loans/Autopal.Search()", searchBody)
		if err != nil {
			return nil, err
		}

		var searchResp SearchIDsResponse
		if err := json.Unmarshal(body, &searchResp); err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] Failed to parse GetLoansByPortfolio response: %v\nResponse body: %s\n", err, string(body))
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}
		if searchResp.Status.Error != nil {
			return nil, fmt.Errorf("loan search failed: %s", searchResp.Status.Error.Message)
		}
		if searchResp.D.Summary.TotalHits > maxPortfolioSearchHits {
			return nil, fmt.Errorf("selection matches %d loans (max %d); narrow created_from/created_to", searchResp.D.Summary.TotalHits, maxPortfolioSearchHits)
		}

		for _, result := range searchResp.D.Results {
			idStr := strings.TrimSpace(string(result.ID))
			if _, err := strconv.Atoi(idStr); err != nil {
				return nil, fmt.Errorf("invalid loan id in search response: %q", idStr)
			}
			ids = append(ids, idStr)
		}
		if len(searchResp.D.Results) < portfolioSearchPageSize {
			break
		}
	}

	loans := []Loan{}
	for start := 0; start < len(ids); start += odataBatchSize {
		end := min(start+odataBatchSize, len(ids))

		filterParts := make([]string, 0, end-start)
		for _, id := range ids[start:end] {
			filterParts = append(filterParts, "id eq "+id)
		}
		params := map[string]string{
			"$filter": strings.Join(filterParts, " or "),
			"$expand": "LoanSetup,Customers,StatusArchive,Portfolios,SubPortfolios",
			"$top":    strconv.Itoa(end - start),
		}

		odataBody, err := c.makeRequest("/public/api/1/odata.svc/Loans", params)
		if err != nil {
			return nil, err
		}

		var batch ODataLoansResponse
		if err := json.Unmarshal(odataBody, &batch); err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] Failed to parse GetLoansByPortfolio OData response: %v\nResponse body: %s\n", err, string(odataBody))
			return nil, fmt.Errorf("failed to parse OData response: %w", err)
		}
		for _, loan := range batch.D.Results {
			if loan.GetArchived() != "1" && loan.InPortfolio(portfolioID, subPortfolio) {
				loans = append(loans, loan)
			}
		}
	}

	sort.Slice(loans, func(i, j int) bool {
		a, _ := strconv.Atoi(loans[i].GetID())
		b, _ := strconv.Atoi(loans[j].GetID())
		return a < b
	})

	return loans, nil
}
