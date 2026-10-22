package loanpro

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
