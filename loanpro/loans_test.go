package loanpro

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestArchiveLoan(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = w.Write([]byte(`{"d":{"id":123,"archived":1}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "key", "tenant")
	if err := client.ArchiveLoan("123"); err != nil {
		t.Fatalf("ArchiveLoan returned error: %v", err)
	}

	if gotMethod != http.MethodPut {
		t.Errorf("Expected PUT, got %s", gotMethod)
	}
	if gotPath != "/public/api/1/odata.svc/Loans(123)" {
		t.Errorf("Unexpected path: %s", gotPath)
	}
	if gotBody["archived"] != float64(1) {
		t.Errorf("Expected archived=1 in body, got %v", gotBody)
	}
}

func TestArchiveLoan_InvalidID(t *testing.T) {
	client := NewClient("http://unused", "key", "tenant")
	if err := client.ArchiveLoan("123)?$x"); err == nil {
		t.Error("Expected error for non-numeric loan id")
	}
}

const portfolioCategoriesJSON = `{"d":{"results":[
	{"id":1,"title":"Mortgages (Crypto)","Portfolios":{"results":[{"id":20,"title":"Active Mortgages"},{"id":24,"title":"TEST"}]}},
	{"id":6,"title":"Test","Portfolios":{"results":[{"id":35,"title":"Development"}]}},
	{"id":7,"title":"Mortgages (Non-Crypto)","Portfolios":{"results":[{"id":26,"title":"Active Mortgages"}]}},
	{"id":4,"title":"Paid off","Portfolios":{"results":[]}}
]}}`

func portfolioTestServer(t *testing.T, searchBody *map[string]any, odataFilter *string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/PortfolioCategories"):
			_, _ = w.Write([]byte(portfolioCategoriesJSON))
		case r.Method == http.MethodPost:
			_ = json.NewDecoder(r.Body).Decode(searchBody)
			// Search results carry a lowercase "portfolios" array that must not break decoding
			_, _ = w.Write([]byte(`{"d":{"results":[{"id":1,"portfolios":["35"]},{"id":2,"portfolios":["35"]},{"id":3,"portfolios":["35"]},{"id":4,"portfolios":["35"]}],"summary":{"totalHits":4}}}`))
		default:
			*odataFilter = r.URL.Query().Get("$filter")
			loan := func(id, archived, portfolioID int, sub string) string {
				return fmt.Sprintf(`{"id":%d,"archived":%d,"created":"/Date(1735689600)/","Portfolios":{"results":[{"id":%d,"title":"Development","categoryId":6}]},"SubPortfolios":{"results":[{"id":2,"title":%q}]}}`, id, archived, portfolioID, sub)
			}
			_, _ = w.Write([]byte(`{"d":{"results":[` + strings.Join([]string{
				loan(3, 0, 35, "Dev Team"),
				loan(1, 0, 35, "Dev Team"),
				loan(2, 0, 99, "Dev Team"),
				loan(4, 1, 35, "Dev Team"),
			}, ",") + `]}}`))
		}
	}))
}

func TestGetLoansByPortfolio(t *testing.T) {
	var searchBody map[string]any
	var odataFilter string
	server := portfolioTestServer(t, &searchBody, &odataFilter)
	defer server.Close()

	client := NewClient(server.URL, "key", "tenant")
	loans, err := client.GetLoansByPortfolio("test", "DEVELOPMENT", "", "2025-01-01", "2025-12-31")
	if err != nil {
		t.Fatalf("GetLoansByPortfolio returned error: %v", err)
	}

	if len(loans) != 2 || loans[0].GetID() != "1" || loans[1].GetID() != "3" {
		t.Fatalf("Expected live loans in portfolio 35 [1 3], got %v", loans)
	}

	filters := searchBody["query"].(map[string]any)["bool"].(map[string]any)["filter"].([]any)
	terms := filters[0].(map[string]any)["terms"].(map[string]any)["portfolios"].([]any)
	if len(terms) != 1 || terms[0] != "35" {
		t.Errorf("Expected search filtered to portfolio 35, got %v", terms)
	}
	created := filters[1].(map[string]any)["range"].(map[string]any)["created"].(map[string]any)
	if created["gte"] != "2025-01-01 00:00:00" || created["lt"] != "2026-01-01 00:00:00" {
		t.Errorf("Unexpected created range: %v", created)
	}
	if odataFilter != "id eq 1 or id eq 2 or id eq 3 or id eq 4" {
		t.Errorf("Unexpected OData filter: %s", odataFilter)
	}
}

func TestGetLoansByPortfolio_SubPortfolio(t *testing.T) {
	var searchBody map[string]any
	var odataFilter string
	server := portfolioTestServer(t, &searchBody, &odataFilter)
	defer server.Close()

	client := NewClient(server.URL, "key", "tenant")
	loans, err := client.GetLoansByPortfolio("Test", "Development", "Other Team", "", "")
	if err != nil {
		t.Fatalf("GetLoansByPortfolio returned error: %v", err)
	}
	if len(loans) != 0 {
		t.Errorf("Expected no loans in sub-portfolio Other Team, got %d", len(loans))
	}
}

func TestFindPortfolioID(t *testing.T) {
	var searchBody map[string]any
	var odataFilter string
	server := portfolioTestServer(t, &searchBody, &odataFilter)
	defer server.Close()
	client := NewClient(server.URL, "key", "tenant")

	tests := []struct {
		category, portfolio, wantID, wantErr string
	}{
		{"Test", "Development", "35", ""},
		{"", "development", "35", ""},
		{"Mortgages (Non-Crypto)", "Active Mortgages", "26", ""},
		{"", "Active Mortgages", "", "ambiguous"},
		{"Test", "Active Mortgages", "", "not found"},
		{"", "Nope", "", "not found"},
	}
	for _, tt := range tests {
		id, err := client.FindPortfolioID(tt.category, tt.portfolio)
		if tt.wantErr != "" {
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("FindPortfolioID(%q, %q) error = %v, want %q", tt.category, tt.portfolio, err, tt.wantErr)
			}
			continue
		}
		if err != nil || id != tt.wantID {
			t.Errorf("FindPortfolioID(%q, %q) = %q, %v; want %q", tt.category, tt.portfolio, id, err, tt.wantID)
		}
	}
}

func TestGetLoansByPortfolio_TooManyHits(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/PortfolioCategories") {
			_, _ = w.Write([]byte(portfolioCategoriesJSON))
			return
		}
		_, _ = fmt.Fprintf(w, `{"d":{"results":[],"summary":{"totalHits":%d}}}`, maxPortfolioSearchHits+1)
	}))
	defer server.Close()

	client := NewClient(server.URL, "key", "tenant")
	if _, err := client.GetLoansByPortfolio("Test", "Development", "", "", ""); err == nil {
		t.Error("Expected error when selection matches too many loans")
	}
}

func TestGetLoansByPortfolio_SearchError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/PortfolioCategories") {
			_, _ = w.Write([]byte(portfolioCategoriesJSON))
			return
		}
		_, _ = w.Write([]byte(`{"d":{"results":[],"summary":{"totalHits":0}},"status":{"error":{"message":"failed to parse date field"}}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "key", "tenant")
	if _, err := client.GetLoansByPortfolio("Test", "Development", "", "", "2026-07-21"); err == nil || !strings.Contains(err.Error(), "failed to parse date field") {
		t.Errorf("Expected search error to be returned, got %v", err)
	}
}
