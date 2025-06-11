package loanpro

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"
)

// Client represents a LoanPro API client
type Client struct {
	baseURL  string
	apiKey   string
	tenantID string
	client   *http.Client
}

// NewClient creates a new LoanPro client
func NewClient(baseURL, apiKey, tenantID string) *Client {
	return &Client{
		baseURL:  baseURL,
		apiKey:   apiKey,
		tenantID: tenantID,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// makeRequest makes a GET request to the LoanPro API
func (c *Client) makeRequest(endpoint string, params map[string]string) ([]byte, error) {
	return c.makeRequestWithMethod("GET", endpoint, params, nil)
}

// makePostRequest makes a POST request to the LoanPro API
func (c *Client) makePostRequest(endpoint string, body any) ([]byte, error) {
	return c.makeRequestWithMethod("POST", endpoint, nil, body)
}

// makeRequestWithMethod makes an HTTP request with the specified method
func (c *Client) makeRequestWithMethod(method, endpoint string, params map[string]string, body any) ([]byte, error) {
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

	slog.Debug("Making LoanPro API request", "method", method, "url", u.String())
	if bodyStr != "" {
		slog.Debug("Request body", "data", bodyStr)
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
		slog.Error("LoanPro API request failed", "error", err)
		fmt.Fprintf(os.Stderr, "[ERROR] LoanPro API request failed: %v\n", err)
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	slog.Debug("Response received", "status", resp.StatusCode, "statusText", resp.Status)

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Failed to read LoanPro response body", "error", err)
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to read LoanPro response body: %v\n", err)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	slog.Debug("Response body", "data", string(responseBody))

	if resp.StatusCode != http.StatusOK {
		slog.Error("LoanPro API error", "status", resp.StatusCode, "body", string(responseBody))
		fmt.Fprintf(os.Stderr, "[ERROR] LoanPro API returned status %d: %s\n", resp.StatusCode, string(responseBody))
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	return responseBody, nil
}