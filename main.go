package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type MCPServer struct {
	loanProClient *LoanProClient
}

type MCPRequest struct {
	JSONRPC string         `json:"jsonrpc"`
	Method  string         `json:"method"`
	Params  map[string]any `json:"params,omitempty"`
	ID      any            `json:"id"`
}

type MCPResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	Result  any       `json:"result,omitempty"`
	Error   *MCPError `json:"error,omitempty"`
	ID      any       `json:"id"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func main() {
	stdioMode := flag.Bool("stdio", false, "Use stdio transport instead of SSE")
	flag.Parse()

	godotenv.Load()

	loanProClient := NewLoanProClient(
		os.Getenv("LOANPRO_API_URL"),
		os.Getenv("LOANPRO_API_KEY"),
		os.Getenv("LOANPRO_TENANT_ID"),
	)

	server := &MCPServer{
		loanProClient: loanProClient,
	}

	if *stdioMode {
		// Run in stdio mode for MCP clients
		log.Println("[MAIN] Starting in stdio mode for MCP client")
		transport := NewStdioTransport(server)
		if err := transport.Run(); err != nil {
			log.Fatal(err)
		}
	} else {
		// Run HTTP server with SSE transport
		r := mux.NewRouter()
		r.HandleFunc("/sse", server.handleSSE).Methods("GET")
		r.HandleFunc("/", server.handleRoot).Methods("GET")

		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}

		fmt.Printf("MCP Server starting on port %s\n", port)
		log.Fatal(http.ListenAndServe(":"+port, r))
	}
}

func (s *MCPServer) handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"name":      "LoanPro MCP Server",
		"version":   "1.0.0",
		"transport": "sse",
	})
}

func (s *MCPServer) handleSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "event: ready\n")
	fmt.Fprintf(w, "data: {\"type\":\"ready\"}\n\n")
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		default:
			var req MCPRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				continue
			}

			response := s.handleMCPRequest(req)
			data, _ := json.Marshal(response)

			fmt.Fprintf(w, "event: message\n")
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}

func (s *MCPServer) handleMCPRequest(req MCPRequest) MCPResponse {
	switch req.Method {
	case "initialize":
		log.Println("[MCP] Processing initialize request")

		// Extract client's protocol version from params
		clientProtocolVersion := "2024-11-05" // fallback default
		if params, ok := req.Params["protocolVersion"].(string); ok {
			clientProtocolVersion = params
			log.Printf("[MCP] Client protocol version: %s", clientProtocolVersion)
		}

		response := MCPResponse{
			JSONRPC: "2.0",
			Result: map[string]any{
				"protocolVersion": clientProtocolVersion, // Use client's version
				"capabilities": map[string]any{
					"tools": map[string]any{},
				},
				"serverInfo": map[string]any{
					"name":    "loanpro-mcp-server",
					"version": "1.0.0",
				},
			},
			ID: req.ID,
		}

		log.Printf("[MCP] Responding with protocol version: %s", clientProtocolVersion)
		return response

	case "initialized":
		// This is a notification, no response needed
		return MCPResponse{} // Empty response indicates no reply

	case "notifications/initialized": // Changed from "initialized"
		log.Println("[MCP] Received initialized notification")
		fmt.Fprintf(os.Stderr, "[MCP] Initialized notification processed\n")
		return MCPResponse{} // No response for notifications

	case "resources/list":
		return MCPResponse{
			JSONRPC: "2.0",
			Result: map[string]any{
				"resources": []any{}, // Empty resources list
			},
			ID: req.ID,
		}

	case "prompts/list":
		return MCPResponse{
			JSONRPC: "2.0",
			Result: map[string]any{
				"prompts": []any{}, // Empty prompts list
			},
			ID: req.ID,
		}

	case "tools/list":
		return MCPResponse{
			JSONRPC: "2.0",
			Result: map[string]any{
				"tools": []map[string]any{
					{
						"name":        "get_loan",
						"description": "Get loan information by ID",
						"inputSchema": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"loan_id": map[string]any{
									"type":        "string",
									"description": "The loan ID to retrieve",
								},
							},
							"required": []string{"loan_id"},
						},
					},
					{
						"name":        "search_loans",
						"description": "Search loans with filters and search terms",
						"inputSchema": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"search_term": map[string]any{
									"type":        "string",
									"description": "Search term to match against customer name, display ID, or title",
								},
								"status": map[string]any{
									"type":        "string",
									"description": "Loan status filter",
								},
								"limit": map[string]any{
									"type":        "number",
									"description": "Maximum number of results",
									"default":     10,
								},
							},
						},
					},
					{
						"name":        "get_customer",
						"description": "Get customer information by ID",
						"inputSchema": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"customer_id": map[string]any{
									"type":        "string",
									"description": "The customer ID to retrieve",
								},
							},
							"required": []string{"customer_id"},
						},
					},
					{
						"name":        "search_customers",
						"description": "Search customers with a search term",
						"inputSchema": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"search_term": map[string]any{
									"type":        "string",
									"description": "Search term to match against customer names, email, or SSN",
								},
								"limit": map[string]any{
									"type":        "number",
									"description": "Maximum number of results",
									"default":     10,
								},
							},
						},
					},
					{
						"name":        "get_loan_payments",
						"description": "Get payment history for a loan",
						"inputSchema": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"loan_id": map[string]any{
									"type":        "string",
									"description": "The loan ID to get payment history for",
								},
							},
							"required": []string{"loan_id"},
						},
					},
				},
			},
			ID: req.ID,
		}

	case "tools/call":
		toolName := req.Params["name"].(string)
		arguments := req.Params["arguments"].(map[string]any)

		switch toolName {
		case "get_loan":
			loanID := arguments["loan_id"].(string)
			loan, err := s.loanProClient.GetLoan(loanID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[ERROR] get_loan failed for ID %s: %v\n", loanID, err)
				return MCPResponse{
					JSONRPC: "2.0",
					Error:   &MCPError{Code: -1, Message: err.Error()},
					ID:      req.ID,
				}
			}
			return MCPResponse{
				JSONRPC: "2.0",
				Result: map[string]any{
					"content": []map[string]any{
						{
							"type": "text",
							"text": fmt.Sprintf("Loan Details:\nID: %s\nDisplay ID: %s\nTitle: %s\nStatus: %s\nCustomer: %s\nAmount: $%s\nBalance: $%s\nPayoff: $%s\nNext Payment: $%s on %s\nDays Past Due: %s\nCreated: %s\nContract Date: %s",
								loan.GetID(), loan.DisplayID, loan.Title, loan.GetLoanStatus(), loan.GetPrimaryCustomerName(), loan.GetLoanAmount(), loan.GetPrincipalBalance(), loan.GetPayoffAmount(), loan.GetNextPaymentAmount(), loan.GetNextPaymentDate(), loan.GetDaysPastDue(), loan.GetCreatedDate(), loan.GetContractDate()),
						},
					},
				},
				ID: req.ID,
			}

		case "search_loans":
			searchTerm := ""
			if term, ok := arguments["search_term"].(string); ok {
				searchTerm = term
			}
			status := ""
			if s, ok := arguments["status"].(string); ok {
				status = s
			}
			limit := 10
			if l, ok := arguments["limit"].(float64); ok {
				limit = int(l)
			}

			loans, err := s.loanProClient.SearchLoans(searchTerm, status, limit)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[ERROR] search_loans failed with term='%s', status='%s', limit=%d: %v\n", searchTerm, status, limit, err)
				return MCPResponse{
					JSONRPC: "2.0",
					Error:   &MCPError{Code: -1, Message: err.Error()},
					ID:      req.ID,
				}
			}

			text := "Loans:\n"
			for _, loan := range loans {
				text += fmt.Sprintf("- ID: %s, Display ID: %s, Customer: %s, Status: %s, Balance: $%s\n", loan.GetID(), loan.DisplayID, loan.GetPrimaryCustomerName(), loan.GetLoanStatus(), loan.GetPrincipalBalance())
			}

			return MCPResponse{
				JSONRPC: "2.0",
				Result: map[string]any{
					"content": []map[string]any{
						{
							"type": "text",
							"text": text,
						},
					},
				},
				ID: req.ID,
			}

		case "get_customer":
			customerID := arguments["customer_id"].(string)
			customer, err := s.loanProClient.GetCustomer(customerID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[ERROR] get_customer failed for ID %s: %v\n", customerID, err)
				return MCPResponse{
					JSONRPC: "2.0",
					Error:   &MCPError{Code: -1, Message: err.Error()},
					ID:      req.ID,
				}
			}
			return MCPResponse{
				JSONRPC: "2.0",
				Result: map[string]any{
					"content": []map[string]any{
						{
							"type": "text",
							"text": fmt.Sprintf("Customer Details:\nID: %d\nName: %s %s\nEmail: %s\nPhone: %s\nCreated: %s",
								customer.ID, customer.FirstName, customer.LastName, customer.Email, customer.Phone, customer.GetCreatedDate()),
						},
					},
				},
				ID: req.ID,
			}
		case "search_customers":
			searchTerm := ""
			if term, ok := arguments["search_term"].(string); ok {
				searchTerm = term
			}
			limit := 10
			if l, ok := arguments["limit"].(float64); ok {
				limit = int(l)
			}

			customers, err := s.loanProClient.SearchCustomers(searchTerm, limit)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[ERROR] search_customers failed with term='%s', limit=%d: %v\n", searchTerm, limit, err)
				return MCPResponse{
					JSONRPC: "2.0",
					Error:   &MCPError{Code: -1, Message: err.Error()},
					ID:      req.ID,
				}
			}

			text := "Customers:\n"
			for _, customer := range customers {
				text += fmt.Sprintf("- ID: %d, Name: %s %s, Email: %s\n", customer.ID, customer.FirstName, customer.LastName, customer.Email)
			}

			return MCPResponse{
				JSONRPC: "2.0",
				Result: map[string]any{
					"content": []map[string]any{
						{
							"type": "text",
							"text": text,
						},
					},
				},
				ID: req.ID,
			}
		
		case "get_loan_payments":
			loanID := arguments["loan_id"].(string)
			payments, err := s.loanProClient.GetLoanPayments(loanID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[ERROR] get_loan_payments failed for loan ID %s: %v\n", loanID, err)
				return MCPResponse{
					JSONRPC: "2.0",
					Error:   &MCPError{Code: -1, Message: err.Error()},
					ID:      req.ID,
				}
			}

			text := fmt.Sprintf("Payment History for Loan %s:\n", loanID)
			if len(payments) == 0 {
				text += "No payments found.\n"
			} else {
				for _, payment := range payments {
					date := payment.Date
					if parsed, err := parseLoanProDate(payment.Date); err == nil {
						date = parsed
					}
					text += fmt.Sprintf("- Date: %s, Amount: $%s, ID: %s\n", 
						date, payment.Amount, string(payment.ID))
				}
			}

			return MCPResponse{
				JSONRPC: "2.0",
				Result: map[string]any{
					"content": []map[string]any{
						{
							"type": "text",
							"text": text,
						},
					},
				},
				ID: req.ID,
			}
		}

	default:
		return MCPResponse{
			JSONRPC: "2.0",
			Error:   &MCPError{Code: -32601, Message: "Method not found"},
			ID:      req.ID,
		}
	}

	return MCPResponse{
		JSONRPC: "2.0",
		Error:   &MCPError{Code: -1, Message: "Unknown error"},
		ID:      req.ID,
	}
}
