package tools

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const maxArchiveBatch = 100

// ArchiveLoansTool returns the archive_loans tool definition
func ArchiveLoansTool() Tool {
	return Tool{
		Name: "archive_loans",
		Description: "Archive loans, removing them from live, billable accounts (e.g. test loans). " +
			"Select loans either by loan_ids, or by portfolio (plus optional sub_portfolio and created date range). " +
			"Runs as a dry run by default and only lists the loans that would be archived; set dry_run=false to archive. " +
			"Each archive incurs a one-time LoanPro fee. Archived loans stop being serviced but can be restored by LoanPro (archived=0).",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"loan_ids": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": fmt.Sprintf("Loan IDs (numeric LoanPro IDs, not display IDs) to archive, max %d. Cannot be combined with portfolio.", maxArchiveBatch),
				},
				"portfolio": map[string]any{
					"type":        "string",
					"description": "Archive live loans in this portfolio (name, case-insensitive), e.g. \"Test\"",
				},
				"sub_portfolio": map[string]any{
					"type":        "string",
					"description": "Only loans also in this sub-portfolio (name, case-insensitive), e.g. \"Development\"",
				},
				"created_from": map[string]any{
					"type":        "string",
					"description": "Only loans created on or after this date (YYYY-MM-DD)",
				},
				"created_to": map[string]any{
					"type":        "string",
					"description": "Only loans created on or before this date (YYYY-MM-DD)",
				},
				"exclude_loan_ids": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "Loan IDs to leave untouched even if they match the portfolio selection",
				},
				"limit": map[string]any{
					"type":        "number",
					"description": fmt.Sprintf("Maximum loans to archive in this call when selecting by portfolio (max %d); re-run to continue", maxArchiveBatch),
					"default":     maxArchiveBatch,
				},
				"dry_run": map[string]any{
					"type":        "boolean",
					"description": "When true, only preview the loans that would be archived",
					"default":     true,
				},
			},
		},
	}
}

// parseLoanIDs validates and de-duplicates a loan ID array argument
func parseLoanIDs(raw any, name string) ([]string, error) {
	if raw == nil {
		return nil, nil
	}
	rawIDs, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("%s must be an array of loan IDs", name)
	}

	ids := make([]string, 0, len(rawIDs))
	seen := make(map[string]bool, len(rawIDs))
	for _, r := range rawIDs {
		var id string
		switch v := r.(type) {
		case string:
			id = strings.TrimSpace(v)
		case float64:
			id = strconv.FormatFloat(v, 'f', -1, 64)
		}
		if _, err := strconv.Atoi(id); err != nil {
			return nil, fmt.Errorf("invalid loan id in %s: %v", name, r)
		}
		if !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func formatArchiveLoan(loan Loan) string {
	return fmt.Sprintf("ID: %s, Display ID: %s, Customer: %s, Status: %s, Created: %s, Balance: $%s",
		loan.GetID(), loan.GetDisplayID(), loan.GetPrimaryCustomerName(), loan.GetLoanStatus(), loan.GetCreatedDate(), loan.GetPrincipalBalance())
}

// executeArchiveLoans handles the archive_loans tool execution
func (m *Manager) executeArchiveLoans(arguments map[string]any) MCPResponse {
	dryRun := true
	if v, ok := arguments["dry_run"].(bool); ok {
		dryRun = v
	}

	loanIDs, err := parseLoanIDs(arguments["loan_ids"], "loan_ids")
	if err != nil {
		return CreateErrorResponse(-1, err.Error(), nil)
	}
	portfolio, _ := arguments["portfolio"].(string)
	portfolio = strings.TrimSpace(portfolio)

	switch {
	case len(loanIDs) > 0 && portfolio != "":
		return CreateErrorResponse(-1, "provide either loan_ids or portfolio, not both", nil)
	case len(loanIDs) > 0:
		if len(loanIDs) > maxArchiveBatch {
			return CreateErrorResponse(-1, fmt.Sprintf("at most %d loan_ids may be archived per call", maxArchiveBatch), nil)
		}
		return m.archiveByIDs(loanIDs, dryRun)
	case portfolio != "":
		return m.archiveByPortfolio(portfolio, arguments, dryRun)
	default:
		return CreateErrorResponse(-1, "loan_ids or portfolio is required", nil)
	}
}

func (m *Manager) archiveByIDs(loanIDs []string, dryRun bool) MCPResponse {
	var text string
	if dryRun {
		text = fmt.Sprintf("DRY RUN: %d loan(s) would be archived. Re-run with dry_run=false to archive.\n", len(loanIDs))
	} else {
		text = fmt.Sprintf("Archive results for %d loan(s):\n", len(loanIDs))
	}

	archived, skipped, failed := 0, 0, 0
	for _, id := range loanIDs {
		loan, err := m.client.GetLoan(id)
		if err != nil || loan == nil {
			LogError("archive_loans", err, fmt.Sprintf("looking up loan %s", id))
			text += fmt.Sprintf("- ID: %s, FAILED: loan not found\n", id)
			failed++
			continue
		}

		if loan.GetArchived() == "1" {
			text += fmt.Sprintf("- %s, SKIPPED: already archived\n", formatArchiveLoan(loan))
			skipped++
			continue
		}

		if dryRun {
			text += fmt.Sprintf("- %s\n", formatArchiveLoan(loan))
			continue
		}

		if m.archiveOne(loan, &text) {
			archived++
		} else {
			failed++
		}
	}

	if dryRun {
		text += fmt.Sprintf("\nSummary: %d to archive, %d already archived, %d not found", len(loanIDs)-skipped-failed, skipped, failed)
	} else {
		text += fmt.Sprintf("\nSummary: %d archived, %d already archived, %d failed", archived, skipped, failed)
	}
	return CreateSuccessResponse(text, nil)
}

func (m *Manager) archiveByPortfolio(portfolio string, arguments map[string]any, dryRun bool) MCPResponse {
	subPortfolio, _ := arguments["sub_portfolio"].(string)
	subPortfolio = strings.TrimSpace(subPortfolio)

	var dates [2]time.Time
	for i, key := range []string{"created_from", "created_to"} {
		v, _ := arguments[key].(string)
		if v == "" {
			continue
		}
		d, err := time.Parse("2006-01-02", v)
		if err != nil {
			return CreateErrorResponse(-1, fmt.Sprintf("%s must be a date in YYYY-MM-DD format", key), nil)
		}
		dates[i] = d
	}
	if !dates[0].IsZero() && !dates[1].IsZero() && dates[0].After(dates[1]) {
		return CreateErrorResponse(-1, "created_from must be on or before created_to", nil)
	}
	createdFrom, _ := arguments["created_from"].(string)
	createdTo, _ := arguments["created_to"].(string)

	excluded, err := parseLoanIDs(arguments["exclude_loan_ids"], "exclude_loan_ids")
	if err != nil {
		return CreateErrorResponse(-1, err.Error(), nil)
	}
	excludeSet := make(map[string]bool, len(excluded))
	for _, id := range excluded {
		excludeSet[id] = true
	}

	limit := maxArchiveBatch
	if v, ok := arguments["limit"].(float64); ok {
		limit = int(v)
	}
	if limit <= 0 || limit > maxArchiveBatch {
		return CreateErrorResponse(-1, fmt.Sprintf("limit must be between 1 and %d", maxArchiveBatch), nil)
	}

	selection := "portfolio " + portfolio
	if subPortfolio != "" {
		selection += " > " + subPortfolio
	}
	if createdFrom != "" || createdTo != "" {
		selection += fmt.Sprintf(", created %s to %s", valueOr(createdFrom, "any"), valueOr(createdTo, "any"))
	}

	matches, err := m.client.GetLoansByPortfolio(portfolio, subPortfolio, createdFrom, createdTo)
	if err != nil {
		LogError("archive_loans", err, "selecting "+selection)
		return CreateErrorResponse(-1, err.Error(), nil)
	}

	var loans []Loan
	for _, loan := range matches {
		if !excludeSet[loan.GetID()] {
			loans = append(loans, loan)
		}
	}
	excludedCount := len(matches) - len(loans)

	if len(loans) == 0 {
		return CreateSuccessResponse(fmt.Sprintf("No live loans found for %s (%d excluded).", selection, excludedCount), nil)
	}

	if dryRun {
		text := fmt.Sprintf("DRY RUN: %d live loan(s) match %s (%d excluded). Re-run with dry_run=false to archive up to %d per call.\n",
			len(loans), selection, excludedCount, limit)
		for _, loan := range loans {
			text += fmt.Sprintf("- %s\n", formatArchiveLoan(loan))
		}
		return CreateSuccessResponse(text, nil)
	}

	batch := loans
	if len(batch) > limit {
		batch = batch[:limit]
	}
	text := fmt.Sprintf("Archive results for %s:\n", selection)
	archived, failed := 0, 0
	for _, loan := range batch {
		if m.archiveOne(loan, &text) {
			archived++
		} else {
			failed++
		}
	}
	text += fmt.Sprintf("\nSummary: %d archived, %d failed, %d excluded, %d remaining (re-run to continue)",
		archived, failed, excludedCount, len(loans)-len(batch))
	return CreateSuccessResponse(text, nil)
}

// archiveOne archives a single loan and appends the outcome to text
func (m *Manager) archiveOne(loan Loan, text *string) bool {
	if err := m.client.ArchiveLoan(loan.GetID()); err != nil {
		LogError("archive_loans", err, fmt.Sprintf("archiving loan %s", loan.GetID()))
		*text += fmt.Sprintf("- %s, FAILED: %v\n", formatArchiveLoan(loan), err)
		return false
	}
	*text += fmt.Sprintf("- %s, ARCHIVED\n", formatArchiveLoan(loan))
	return true
}

func valueOr(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
