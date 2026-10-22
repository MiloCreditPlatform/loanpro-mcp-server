package tools

import (
	"fmt"
	"strconv"
	"strings"
)

const maxArchiveBatch = 100

// ArchiveLoansTool returns the archive_loans tool definition
func ArchiveLoansTool() Tool {
	return Tool{
		Name: "archive_loans",
		Description: "Archive loans by ID, removing them from live, billable accounts (e.g. test loans). " +
			"Runs as a dry run by default and only lists the loans that would be archived; set dry_run=false to archive. " +
			"Each archive incurs a one-time LoanPro fee. Archived loans stop being serviced but can be restored by LoanPro (archived=0).",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"loan_ids": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": fmt.Sprintf("Loan IDs (numeric LoanPro IDs, not display IDs) to archive, max %d", maxArchiveBatch),
				},
				"dry_run": map[string]any{
					"type":        "boolean",
					"description": "When true, only preview the loans that would be archived",
					"default":     true,
				},
			},
			"required": []string{"loan_ids"},
		},
	}
}

// executeArchiveLoans handles the archive_loans tool execution
func (m *Manager) executeArchiveLoans(arguments map[string]any) MCPResponse {
	rawIDs, ok := arguments["loan_ids"].([]any)
	if !ok || len(rawIDs) == 0 {
		return CreateErrorResponse(-1, "loan_ids must be a non-empty array of loan IDs", nil)
	}
	if len(rawIDs) > maxArchiveBatch {
		return CreateErrorResponse(-1, fmt.Sprintf("at most %d loan_ids may be archived per call", maxArchiveBatch), nil)
	}

	dryRun := true
	if v, ok := arguments["dry_run"].(bool); ok {
		dryRun = v
	}

	loanIDs := make([]string, 0, len(rawIDs))
	seen := make(map[string]bool, len(rawIDs))
	for _, raw := range rawIDs {
		var id string
		switch v := raw.(type) {
		case string:
			id = strings.TrimSpace(v)
		case float64:
			id = strconv.FormatFloat(v, 'f', -1, 64)
		}
		if _, err := strconv.Atoi(id); err != nil {
			return CreateErrorResponse(-1, fmt.Sprintf("invalid loan id: %v", raw), nil)
		}
		if !seen[id] {
			seen[id] = true
			loanIDs = append(loanIDs, id)
		}
	}

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

		details := fmt.Sprintf("ID: %s, Display ID: %s, Customer: %s, Status: %s, Balance: $%s",
			loan.GetID(), loan.GetDisplayID(), loan.GetPrimaryCustomerName(), loan.GetLoanStatus(), loan.GetPrincipalBalance())

		if loan.GetArchived() == "1" {
			text += fmt.Sprintf("- %s, SKIPPED: already archived\n", details)
			skipped++
			continue
		}

		if dryRun {
			text += fmt.Sprintf("- %s\n", details)
			continue
		}

		if err := m.client.ArchiveLoan(id); err != nil {
			LogError("archive_loans", err, fmt.Sprintf("archiving loan %s", id))
			text += fmt.Sprintf("- %s, FAILED: %v\n", details, err)
			failed++
			continue
		}
		text += fmt.Sprintf("- %s, ARCHIVED\n", details)
		archived++
	}

	if dryRun {
		text += fmt.Sprintf("\nSummary: %d to archive, %d already archived, %d not found", len(loanIDs)-skipped-failed, skipped, failed)
	} else {
		text += fmt.Sprintf("\nSummary: %d archived, %d already archived, %d failed", archived, skipped, failed)
	}

	return CreateSuccessResponse(text, nil)
}
