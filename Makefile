.PHONY: build test test-verbose test-coverage lint fmt clean help

# Build the binary
build:
	go build -o loanpro-mcp-server .

# Run tests
test:
	go test ./... -race

# Run tests with verbose output
test-verbose:
	go test ./... -v -race

# Run tests with coverage
test-coverage:
	go test ./... -race -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...
	goimports -w .

# Tidy dependencies
tidy:
	go mod tidy

# Clean build artifacts
clean:
	rm -f loanpro-mcp-server coverage.out coverage.html

# Run security scan
security:
	gosec ./...

# Install development tools
install-tools:
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# Full CI pipeline
ci: tidy fmt lint test-coverage security build

# Help
help:
	@echo "Available targets:"
	@echo "  build          - Build the binary"
	@echo "  test           - Run tests"
	@echo "  test-verbose   - Run tests with verbose output"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  lint           - Run linter"
	@echo "  fmt            - Format code"
	@echo "  tidy           - Tidy dependencies"
	@echo "  clean          - Clean build artifacts"
	@echo "  security       - Run security scan"
	@echo "  install-tools  - Install development tools"
	@echo "  ci             - Run full CI pipeline"
	@echo "  help           - Show this help"