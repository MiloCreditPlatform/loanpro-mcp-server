# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build
go build -o loanpro-mcp-server .

# Test (all packages, with race detector)
make test
# or: go test ./... -race

# Test a single package
go test ./tools -v -race
go test ./loanpro -v -race

# Test a single test by name
go test ./tools -run TestManager_ExecuteTool_GetLoan -v

# Format
go fmt ./...

# Lint (requires golangci-lint)
make lint

# Run locally (HTTP on :8080)
go run .
# stdio mode for MCP clients:
go run . --transport=stdio
```

Credentials are read from `.env` at startup (via `godotenv.Load()`):
```
LOANPRO_API_URL=https://...
LOANPRO_API_KEY=...
LOANPRO_TENANT_ID=...
```

## Architecture

The server has three layers that must stay in sync when adding a new tool:

### 1. `loanpro/` — raw API client
`client.go` provides `makeRequest` (GET) and `makePostRequest` (POST). Individual files (`loans.go`, `customers.go`, etc.) implement domain-specific API calls against two LoanPro endpoints:
- **OData** (`/public/api/1/odata.svc/Loans`) — for fetching loan detail with `$expand`
- **Elasticsearch** (`/public/api/1/Loans/Autopal.Search()`) — for search and filtered queries (e.g. daysPastDue range). Use this for any filter that isn't a direct OData Loan entity property.

`types.go` defines all structs. Important subtlety: `Loan.active` (json: `"active"`) is a soft-delete flag and is always `1` for any fetchable record. `LoanSetup.active` (json: `"active"` on the nested `LoanSetup` object) is the true loan lifecycle active/inactive status shown in the UI.

`loan_methods.go` adds helper methods to `*Loan` for safe access with fallback priority: search-result fields → `StatusArchive` → `LoanSetup`. The `GetActive()` method reads `LoanSetup.active`; if `LoanSetup` is nil, it returns `"0"` (never falls back to `Loan.active`).

Dates from LoanPro arrive as `/Date(unix_seconds)/`; use `parseLoanProDate` (date only) or `parseLoanProDateTime` (with time) from `types.go`.

### 2. `tools/` — MCP tool layer
`types.go` defines the `LoanProClient`, `Loan`, `Customer`, `Payment`, and `Transaction` interfaces used throughout this layer. Every new LoanPro entity type needs a corresponding interface here.

Each tool lives in its own file (`get_loan.go`, `search_loans.go`, etc.) and exports:
- A `XxxTool()` function returning the tool schema (`Tool` struct with `InputSchema`)
- An `(m *Manager) executeXxx(arguments map[string]any) MCPResponse` method

`manager.go` is the only file that needs updating when adding a tool: add it to `GetAllTools()` and `ExecuteTool()`.

`CreateSuccessResponse` / `CreateErrorResponse` in `types.go` are the only correct ways to build responses — they produce the MCP `content[{type, text}]` envelope.

### 3. `main.go` — wiring layer
`ClientAdapter` bridges `*loanpro.Client` to `tools.LoanProClient`. Every method on `LoanProClient` needs a corresponding method here that converts `[]loanpro.Loan` → `[]tools.Loan` (and similarly for other types) by ranging over the slice and wrapping each concrete value as the interface.

### Adding a new tool — checklist
1. Add the API method to `loanpro/` (pick OData vs Elasticsearch as appropriate)
2. Add any new interface methods to `tools/types.go` (both `LoanProClient` and entity interfaces)
3. Add the `ClientAdapter` method to `main.go`
4. Create `tools/get_xxx.go` with the tool schema function and `executeXxx` method
5. Register in `manager.go` (`GetAllTools` + `ExecuteTool` switch)
6. Add mock support in `tools/manager_test.go` (`MockLoanProClient` and `MockLoan`/etc.)

### Transport layer (`transport/`)
Three transports share the same `MCPHandler` interface: `http.go` (POST `/mcp`), `sse.go` (GET `/sse`), `stdio.go`. All delegate to `MCPServer.HandleMCPRequest` in `main.go`. The transport layer is unlikely to need changes when adding tools.
