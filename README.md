# LoanPro MCP Server

A Model Context Protocol (MCP) server that exposes read-only access to LoanPro data via both Server-Sent Events (SSE) and stdio transports.

## Features

- Read-only access to LoanPro loan and customer data
- Comprehensive loan details with financial data (balances, payments, status)
- Customer search and retrieval
- Payment history tracking
- Both SSE and stdio transport support
- MCP-compliant tool interface
- Built with Go for performance and reliability

## Setup

1. Clone the repository
2. Copy `.env.example` to `.env` and configure your LoanPro API credentials:
   ```bash
   cp .env.example .env
   ```
3. Edit `.env` with your LoanPro API details:
   ```
   LOANPRO_API_URL=https://your-loanpro-instance.com/api
   LOANPRO_API_KEY=your_api_key_here
   LOANPRO_TENANT_ID=your_tenant_id_here
   PORT=8080
   ```

## Running

### Stdio Transport (for MCP clients)
```bash
go run . -stdio
```

### SSE Transport (HTTP server)
```bash
go run .
```

The HTTP server will start on the configured port (default: 8080).

## Available Tools

### get_loan
Retrieve comprehensive loan information by ID including balances, payment schedules, and customer details.

Parameters:
- `loan_id` (required): The loan ID to retrieve

Returns: Complete loan details with principal balance, payoff amount, next payment info, days past due, status, and customer information.

### search_loans
Search for loans with filters and search terms.

Parameters:
- `search_term` (optional): Search term to match against customer name, display ID, or title
- `status` (optional): Filter by loan status
- `limit` (optional): Maximum number of results (default: 10)

Returns: List of matching loans with basic information and financial data.

### get_customer
Retrieve customer information by ID.

Parameters:
- `customer_id` (required): The customer ID to retrieve

Returns: Customer details including name, email, phone, and creation date.

### search_customers
Search for customers with a search term.

Parameters:
- `search_term` (optional): Search term to match against customer names, email, or SSN
- `limit` (optional): Maximum number of results (default: 10)

Returns: List of matching customers with contact information.

### get_loan_payments
Get payment history for a loan.

Parameters:
- `loan_id` (required): The loan ID to get payment history for

Returns: Chronological list of payments made on the loan with dates and amounts.

## MCP Client Usage

### For Stdio Transport (Claude Desktop)
Add to your Claude Desktop config:
```json
{
  "mcpServers": {
    "loanpro": {
      "command": "go",
      "args": ["run", ".", "-stdio"],
      "cwd": "/path/to/loanpro-mcp-server",
      "env": {
        "LOANPRO_API_URL": "https://your-loanpro-instance.com/api",
        "LOANPRO_API_KEY": "your_api_key",
        "LOANPRO_TENANT_ID": "your_tenant_id"
      }
    }
  }
}
```

Alternatively, if you have a compiled binary:
```json
{
  "mcpServers": {
    "loanpro": {
      "command": "/path/to/loanpro-mcp-server",
      "args": ["-stdio"],
      "env": {
        "LOANPRO_API_URL": "https://loanpro.simnang.com/api",
        "LOANPRO_API_KEY": "your_api_key",
        "LOANPRO_TENANT_ID": "your_tenant_id"
      }
    }
  }
}
```

### For SSE Transport
Start the HTTP server:
```bash
go run .
```

Connect to the server using the SSE endpoint:
```
http://localhost:8080/sse
```

## Building

To build a standalone binary:
```bash
go build -o loanpro-mcp-server .
```

## Examples

### Example Loan Details Response
```
Loan Details:
ID: 123
Display ID: LN00000456
Title: Sample Loan
Status: Open
Customer: John Doe
Amount: $250000.00
Balance: $240000.00
Payoff: $242150.75
Next Payment: $1500.00 on 2025-07-15
Days Past Due: 0
Created: 2025-01-15 10:30:22 UTC
Contract Date: 2025-01-10
```

### Example Search Results
```
Loans:
- ID: 123, Display ID: LN00000456, Customer: John Doe, Status: Open, Balance: $240000.00
- ID: 124, Display ID: LN00000457, Customer: Jane Smith, Status: Active, Balance: $185000.00
```

## Architecture

The server implements the MCP protocol and responds to standard MCP requests for:
- Tool listing and execution
- Comprehensive error handling with stderr logging
- Proper JSON-RPC 2.0 compliance
- Unix timestamp parsing for LoanPro date formats
- Flexible struct mapping for search vs. detailed API responses