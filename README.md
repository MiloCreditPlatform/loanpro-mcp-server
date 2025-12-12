# LoanPro MCP Server

A Model Context Protocol (MCP) server that exposes read-only access to LoanPro data via multiple transport protocols: HTTP, Server-Sent Events (SSE), and stdio.

## Features

- **Multiple Transport Protocols**: HTTP, SSE, and stdio support
- **Read-only LoanPro Integration**: Secure access to loan and customer data
- **Comprehensive Financial Data**: Balances, payments, status, and payment history
- **Modular Architecture**: Clean separation of concerns with dedicated packages
- **MCP-Compliant**: Full Model Context Protocol implementation
- **Built with Go**: High performance and reliability
- **CORS Support**: Cross-origin requests for web integration

## Architecture

The server is organized into modular packages for maintainability:

```
├── main.go              # Application entry point and transport configuration
├── loanpro/            # LoanPro API integration
│   ├── client.go       # HTTP client implementation
│   ├── types.go        # Data structures and utilities
│   ├── loans.go        # Loan operations
│   ├── customers.go    # Customer operations
│   └── payments.go     # Payment operations
├── tools/              # MCP tool implementations
│   ├── manager.go      # Tool management and execution
│   ├── types.go        # Tool interfaces and types
│   └── *.go           # Individual tool implementations
└── transport/          # Communication protocols
    ├── http.go         # Streamable HTTP transport
    ├── sse.go          # Server-Sent Events transport
    ├── stdio.go        # Stdio transport for MCP clients
    └── types.go        # Protocol types and interfaces
```

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
   
   # Logging configuration (optional)
   LOG_LEVEL=INFO
   LOG_FORMAT=TEXT
   ```

## Running

### HTTP Transport (Default)
```bash
# Default HTTP transport
go run .
# or explicitly
go run . --transport=http
```

The server will provide the following endpoints:
- `POST /mcp` - MCP requests
- `GET /` - Server information
- `GET /health` - Health check

### SSE Transport (for web browsers)
```bash
go run . --transport=sse
```

### Stdio Transport (for MCP clients like Claude Desktop)
```bash
go run . --transport=stdio
# or using the legacy flag
go run . --stdio
```

## Transport Comparison

| Transport | Use Case | Communication | Endpoints |
|-----------|----------|---------------|-----------|
| **HTTP** | REST clients, web apps, testing | Standard HTTP POST | `/mcp`, `/`, `/health` |
| **SSE** | Web browsers, real-time apps | Server-sent events | `/sse`, `/` |
| **Stdio** | MCP clients (Claude Desktop) | Bidirectional stdin/stdout | N/A |

## Available Tools

### get_loan
Retrieve comprehensive loan information by ID including balances, payment schedules, and customer details.

**Parameters:**
- `loan_id` (required): The loan ID to retrieve

**Returns:** Complete loan details with principal balance, payoff amount, next payment info, days past due, status, and customer information.

### search_loans
Search for loans with filters and search terms.

**Parameters:**
- `search_term` (optional): Search term to match against customer name, display ID, or title
- `status` (optional): Filter by loan status
- `limit` (optional): Maximum number of results (default: 10)

**Returns:** List of matching loans with basic information and financial data.

### get_customer
Retrieve customer information by ID.

**Parameters:**
- `customer_id` (required): The customer ID to retrieve

**Returns:** Customer details including name, email, phone, and creation date.

### search_customers
Search for customers with a search term.

**Parameters:**
- `search_term` (optional): Search term to match against customer names, email, or SSN
- `limit` (optional): Maximum number of results (default: 10)

**Returns:** List of matching customers with contact information.

### get_loan_payments
Get payment history for a loan.

**Parameters:**
- `loan_id` (required): The loan ID to get payment history for

**Returns:** Chronological list of payments made on the loan with dates, amounts, payment IDs, and status (Active/Inactive).

### get_loan_transactions
Get detailed transaction history for a loan including payments, charges, credits, and adjustments.

**Parameters:**
- `loan_id` (required): The loan ID to get transaction history for

**Returns:** Comprehensive transaction history including:
- Transaction type (payment, charge, credit, adjustment, etc.)
- Transaction amount, date, ID, and status
- Payment application breakdown (principal, interest, fees, escrow)
- Transaction title and description
- Complete audit trail of all loan activities

## Usage Examples

### HTTP Transport
```bash
# Get server info
curl http://localhost:8080/

# Health check
curl http://localhost:8080/health

# List available tools
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"tools/list","id":1}'

# Get loan details
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"tools/call","params":{"name":"get_loan","arguments":{"loan_id":"123"}},"id":2}'

# Get loan transactions
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"tools/call","params":{"name":"get_loan_transactions","arguments":{"loan_id":"123"}},"id":3}'
```

### MCP Client Configuration (Claude Desktop)

For compiled binary:
```json
{
  "mcpServers": {
    "loanpro": {
      "command": "/path/to/loanpro-mcp-server",
      "args": ["--transport=stdio"],
      "env": {
        "LOANPRO_API_URL": "https://your-loanpro-instance.com/api",
        "LOANPRO_API_KEY": "your_api_key",
        "LOANPRO_TENANT_ID": "your_tenant_id"
      }
    }
  }
}
```

For Go source:
```json
{
  "mcpServers": {
    "loanpro": {
      "command": "go",
      "args": ["run", ".", "--transport=stdio"],
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

### SSE Transport
Start the server with SSE transport:
```bash
go run . --transport=sse
```

Connect to the SSE endpoint:
```
http://localhost:8080/sse
```

## Building

To build a standalone binary:
```bash
go build -o loanpro-mcp-server .
```

## Testing

### Running Tests

Using the provided Makefile:
```bash
make test                # Run tests
make test-verbose        # Run tests with verbose output  
make test-coverage       # Run tests with coverage report
```

Or directly with Go:
```bash
go test ./... -race -coverprofile=coverage.out -covermode=atomic
go test ./... -v
go test ./tools -v       # Test specific package
```

### Test Coverage

Generate and view coverage report:
```bash
make test-coverage       # Generates coverage.out and coverage.html
open coverage.html       # View in browser

# Or manually:
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Continuous Integration

The project includes GitHub Actions workflows that automatically:
- Run tests across multiple Go versions (1.22.x, 1.23.x, 1.24.x)
- Build and test the binary

### Test Structure

- **`tools/`** - Unit tests for MCP tool implementations with mock LoanPro client
- **`transport/`** - Unit tests for HTTP, SSE, and stdio transports with mock handlers  
- **`loanpro/`** - Unit tests for data types, date parsing, and loan methods
- **`main_test.go`** - Integration tests for server initialization and MCP protocol handling

### Mocking

Tests use mock implementations to avoid external dependencies:
- `MockLoanProClient` - Simulates LoanPro API responses
- `MockMCPHandler` - Simulates MCP protocol handling
- Interface-based design enables easy testing and dependency injection

## Development

### Available Make Targets

```bash
make help            # Show all available targets
make build           # Build the binary
make test            # Run tests
make test-coverage   # Run tests with coverage
make lint            # Run linter  
make fmt             # Format code
make clean           # Clean build artifacts
make ci              # Run full CI pipeline
```

### Prerequisites

- Go 1.21 or later
- Optional: golangci-lint for linting
- Optional: gosec for security scanning

### Development Workflow

1. Make changes to the code
2. Run tests: `make test`
3. Format code: `make fmt`
4. Run linter: `make lint`
5. Build binary: `make build`
6. Test manually: `./loanpro-mcp-server --help`

## Example Responses

### Loan Details
```
Loan Details:
ID: 123
Display ID: LN00000456
Status: Open
Customer: John Doe
Balance: $240000.00
```

### Search Results
```
Loans:
- ID: 123, Display ID: LN00000456, Customer: John Doe, Status: Open, Balance: $240000.00
- ID: 124, Display ID: LN00000457, Customer: Jane Smith, Status: Active, Balance: $185000.00
```

### Customer Information
```
Customer Details:
ID: 456
Name: John Doe
Email: john.doe@example.com
Phone: (555) 123-4567
Created: 2024-01-15 10:30:22 UTC
```

### Transaction History
```
Transaction History for Loan 619:
- Date: 2025-11-25, Type: payment, Amount: $75022.40, ID: 2356, Status: Active
  Title: Payment Received
  Applied: Principal: $74500.00 Interest: $522.40
- Date: 2025-11-25, Type: payment, Amount: $500.00, ID: 2357, Status: Active
  Title: Additional Payment
  Applied: Principal: $500.00
- Date: 2025-11-20, Type: charge.latefee, Amount: $50.00, ID: 1234, Status: Active
  Title: Late Fee
- Date: 2025-11-18, Type: credit, Amount: $25.00, ID: 1233, Status: Active
  Title: Fee Waiver
```

## Logging Configuration

The server supports configurable logging via environment variables:

### Log Levels
Set the `LOG_LEVEL` environment variable to control logging verbosity:

- **DEBUG**: Detailed debugging information including request/response data
- **INFO**: General operational messages (default)
- **WARN/WARNING**: Warning messages  
- **ERROR**: Error messages only

### Log Formats
Set the `LOG_FORMAT` environment variable to control output format:

- **TEXT**: Human-readable text format (default)
- **JSON**: Structured JSON format for log aggregation

### Examples

```bash
# Debug level with text format
LOG_LEVEL=DEBUG ./loanpro-mcp-server --transport=stdio

# Info level with JSON format  
LOG_LEVEL=INFO LOG_FORMAT=JSON ./loanpro-mcp-server --transport=http

# Error level only
LOG_LEVEL=ERROR ./loanpro-mcp-server --transport=sse
```

### Sample Output

**Text Format (Default):**
```
time=2025-06-11T13:04:35.886-04:00 level=INFO msg="Starting MCP server" transport=http port=8080
time=2025-06-11T13:04:35.887-04:00 level=DEBUG msg="Processing HTTP request" method=tools/list id=1
```

**JSON Format:**
```json
{"time":"2025-06-11T13:04:35.886-04:00","level":"INFO","msg":"Starting MCP server","transport":"http","port":"8080"}
{"time":"2025-06-11T13:04:35.887-04:00","level":"DEBUG","msg":"Processing HTTP request","method":"tools/list","id":1}
```

## Technical Details

- **JSON-RPC 2.0 Compliance**: Full MCP protocol implementation
- **Structured Logging**: Configurable log levels and formats using Go's slog
- **Error Handling**: Comprehensive error logging to stderr with context
- **Date Parsing**: Supports LoanPro Unix timestamp format (`/Date(1427829732)/`)
- **Flexible Data Mapping**: Handles different API response formats
- **CORS Support**: Cross-origin requests enabled for web integration
- **Modular Design**: Clean separation between transport, tools, and API layers

## License

MIT License - see LICENSE file for details.