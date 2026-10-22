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

func TestGetLoansByPortfolio(t *testing.T) {
	var searchBody map[string]any
	var odataFilter, odataExpand string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			_ = json.NewDecoder(r.Body).Decode(&searchBody)
			_, _ = w.Write([]byte(`{"d":{"results":[{"id":1},{"id":2},{"id":3},{"id":4}],"summary":{"totalHits":4}}}`))
			return
		}
		odataFilter = r.URL.Query().Get("$filter")
		odataExpand = r.URL.Query().Get("$expand")
		loan := func(id int, archived int, portfolio, sub string) string {
			return fmt.Sprintf(`{"id":%d,"archived":%d,"created":"/Date(1735689600)/","Portfolios":{"results":[{"id":1,"title":%q}]},"SubPortfolios":{"results":[{"id":2,"title":%q}]}}`, id, archived, portfolio, sub)
		}
		_, _ = w.Write([]byte(`{"d":{"results":[` + strings.Join([]string{
			loan(3, 0, "Test", "Development"),
			loan(1, 0, "Test", "Development"),
			loan(2, 0, "Test", "Staging"),
			loan(4, 1, "Test", "Development"),
		}, ",") + `]}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "key", "tenant")
	loans, err := client.GetLoansByPortfolio("test", "DEVELOPMENT", "2025-01-01", "2025-12-31")
	if err != nil {
		t.Fatalf("GetLoansByPortfolio returned error: %v", err)
	}

	if len(loans) != 2 || loans[0].GetID() != "1" || loans[1].GetID() != "3" {
		t.Fatalf("Expected live Test>Development loans [1 3], got %v", loans)
	}
	if loans[0].GetCreatedDate() == "" {
		t.Error("Expected created date to be parsed")
	}

	created := searchBody["query"].(map[string]any)["bool"].(map[string]any)["filter"].(map[string]any)["range"].(map[string]any)["created"].(map[string]any)
	if created["gte"] != "2025-01-01" || created["lt"] != "2026-01-01" {
		t.Errorf("Unexpected created range: %v", created)
	}
	if odataFilter != "id eq 1 or id eq 2 or id eq 3 or id eq 4" {
		t.Errorf("Unexpected OData filter: %s", odataFilter)
	}
	if !strings.Contains(odataExpand, "Portfolios") || !strings.Contains(odataExpand, "SubPortfolios") {
		t.Errorf("Expected Portfolios and SubPortfolios expanded, got %s", odataExpand)
	}
}

func TestGetLoansByPortfolio_TooManyHits(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, `{"d":{"results":[],"summary":{"totalHits":%d}}}`, maxPortfolioSearchHits+1)
	}))
	defer server.Close()

	client := NewClient(server.URL, "key", "tenant")
	if _, err := client.GetLoansByPortfolio("Test", "", "", ""); err == nil {
		t.Error("Expected error when date range matches too many loans")
	}
}
