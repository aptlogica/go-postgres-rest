# go-postgres-rest Makefile

# Variables
BINARY_NAME=go-postgres-rest
VERSION?=$(shell git describe --tags --always --dirty)
COMMIT?=$(shell git rev-parse --short HEAD)
BUILD_TIME?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildTime=${BUILD_TIME}"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
COVER_PROFILE=coverage.out
COVER_HTML=coverage.html

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[0;33m
BLUE=\033[0;34m
NC=\033[0m # No Color

.PHONY: help build test test-coverage coverage coverage-func clean run deps lint format security docker-build docker-run

# Default target
all: deps format lint test build

# Help target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@echo '  build            - Build the application'
	@echo '  build-all        - Build for multiple platforms'
	@echo '  run              - Build and run the application'
	@echo '  test             - Run all tests'
	@echo '  test-coverage    - Run tests and show coverage'
	@echo '  coverage         - Alias for test-coverage'
	@echo '  coverage-func    - Show coverage by function'
	@echo '  benchmark        - Run benchmarks'
	@echo '  clean            - Clean build artifacts'
	@echo '  deps             - Download and install dependencies'
	@echo '  lint             - Run golangci-lint'
	@echo '  lint-fix         - Run golangci-lint with auto-fix'
	@echo '  format           - Format Go code'
	@echo '  security         - Run security scan'

# Build the application
build: ## Build the application
	@echo "${GREEN}Building ${BINARY_NAME}...${NC}"
	$(GOBUILD) ${LDFLAGS} -o bin/${BINARY_NAME} ./cmd/...
	@echo "${GREEN}Build completed successfully!${NC}"

# Build for multiple platforms
build-all: ## Build for multiple platforms (Linux, Darwin, Windows)
	@echo "${GREEN}Building for multiple platforms...${NC}"
	@mkdir -p dist
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) ${LDFLAGS} -o dist/${BINARY_NAME}-linux-amd64 ./cmd/...
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) ${LDFLAGS} -o dist/${BINARY_NAME}-linux-arm64 ./cmd/...
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) ${LDFLAGS} -o dist/${BINARY_NAME}-darwin-amd64 ./cmd/...
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) ${LDFLAGS} -o dist/${BINARY_NAME}-darwin-arm64 ./cmd/...
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) ${LDFLAGS} -o dist/${BINARY_NAME}-windows-amd64.exe ./cmd/...
	@echo "${GREEN}Multi-platform build completed!${NC}"

# Run tests
test: ## Run all tests
	@echo "${GREEN}Running tests...${NC}"
	$(GOTEST) -v -race -coverprofile=$(COVER_PROFILE) -covermode=atomic ./pkg/...
	@echo "${GREEN}Tests completed!${NC}"

# Run tests with coverage
test-coverage: test ## Run tests and show coverage
	@echo "${GREEN}Generating coverage report...${NC}"
	$(GOCMD) tool cover -html=$(COVER_PROFILE) -o $(COVER_HTML)
	@echo "${BLUE}Coverage report generated: $(COVER_HTML)${NC}"

coverage: test-coverage ## Alias for test-coverage

coverage-func: ## Show coverage by function
	$(GOCMD) tool cover -func=$(COVER_PROFILE)

# Run benchmarks
benchmark: ## Run benchmarks
	@echo "${GREEN}Running benchmarks...${NC}"
	$(GOTEST) -bench=. -benchmem ./...

# Clean build artifacts
clean: ## Clean build artifacts
	@echo "${YELLOW}Cleaning...${NC}"
	$(GOCLEAN)
	rm -rf bin/
	rm -rf dist/
	rm -f coverage.out coverage.html
	@echo "${GREEN}Clean completed!${NC}"

# Run the application
run: build ## Build and run the application
	@echo "${GREEN}Running ${BINARY_NAME}...${NC}"
	./bin/${BINARY_NAME}

# Install dependencies
deps: ## Download and install dependencies
	@echo "${GREEN}Installing dependencies...${NC}"
	$(GOMOD) tidy
	$(GOMOD) download
	@echo "${GREEN}Dependencies installed!${NC}"

# Update dependencies
deps-update: ## Update dependencies to latest versions
	@echo "${GREEN}Updating dependencies...${NC}"
	$(GOGET) -u ./...
	$(GOMOD) tidy
	@echo "${GREEN}Dependencies updated!${NC}"

# Run linting
lint: ## Run golangci-lint
	@echo "${GREEN}Running linter...${NC}"
	@which golangci-lint > /dev/null || (echo "${RED}golangci-lint not found. Install with: 'go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest'${NC}" && exit 1)
	golangci-lint run
	@echo "${GREEN}Linting completed!${NC}"

# Format code
format: ## Format Go code
	@echo "${GREEN}Formatting code...${NC}"
	$(GOFMT) -s -w .
	@echo "${GREEN}Code formatted!${NC}"

# Check code formatting
format-check: ## Check if code is properly formatted
	@echo "${GREEN}Checking code formatting...${NC}"
	@if [ $$($(GOFMT) -d . | wc -l) -gt 0 ]; then \
		echo "${RED}Code is not properly formatted. Run 'make format' to fix.${NC}"; \
		$(GOFMT) -d .; \
		exit 1; \
	fi
	@echo "${GREEN}Code is properly formatted!${NC}"

# Security scan
security: ## Run security scan with gosec
	@echo "${GREEN}Running security scan...${NC}"
	@which gosec > /dev/null || (echo "${RED}gosec not found. Install with: 'go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest'${NC}" && exit 1)
	gosec ./...
	@echo "${GREEN}Security scan completed!${NC}"

# Run vulnerability check
vuln-check: ## Check for vulnerabilities
	@echo "${GREEN}Checking for vulnerabilities...${NC}"
	@which govulncheck > /dev/null || (echo "${RED}govulncheck not found. Install with: 'go install golang.org/x/vuln/cmd/govulncheck@latest'${NC}" && exit 1)
	govulncheck ./...
	@echo "${GREEN}Vulnerability check completed!${NC}"

# Docker build
docker-build: ## Build Docker image
	@echo "${GREEN}Building Docker image...${NC}"
	docker build -t ${BINARY_NAME}:${VERSION} .
	docker tag ${BINARY_NAME}:${VERSION} ${BINARY_NAME}:latest
	@echo "${GREEN}Docker image built: ${BINARY_NAME}:${VERSION}${NC}"

# Docker run
docker-run: ## Run Docker container
	@echo "${GREEN}Running Docker container...${NC}"
	docker run --rm -p 8080:8080 ${BINARY_NAME}:latest

# Start development environment
dev-env: ## Start development environment with Docker Compose
	@echo "${GREEN}Starting development environment...${NC}"
	docker-compose up -d postgres
	@echo "${BLUE}PostgreSQL is starting up. Wait a moment and run 'make dev-env-status' to check status.${NC}"

# Check development environment status
dev-env-status: ## Check status of development environment
	@echo "${GREEN}Checking development environment status...${NC}"
	docker-compose ps

# Stop development environment
dev-env-stop: ## Stop development environment
	@echo "${YELLOW}Stopping development environment...${NC}"
	docker-compose down

# Database migrations
migrate-up: ## Run database migrations up
	@echo "${GREEN}Running database migrations...${NC}"
	$(GOCMD) run ./cmd/migrate up

migrate-down: ## Run database migrations down
	@echo "${YELLOW}Rolling back database migrations...${NC}"
	$(GOCMD) run ./cmd/migrate down

migrate-status: ## Show migration status
	@echo "${GREEN}Checking migration status...${NC}"
	$(GOCMD) run ./cmd/migrate status

# Generate code
generate: ## Run go generate
	@echo "${GREEN}Running code generation...${NC}"
	$(GOCMD) generate ./...
	@echo "${GREEN}Code generation completed!${NC}"

# Install development tools
install-tools: ## Install development tools
	@echo "${GREEN}Installing development tools...${NC}"
	$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOGET) -u github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	$(GOGET) -u golang.org/x/vuln/cmd/govulncheck@latest
	@echo "${GREEN}Development tools installed!${NC}"

# Pre-commit checks (run before committing)
pre-commit: format-check lint test security ## Run all pre-commit checks
	@echo "${GREEN}All pre-commit checks passed!${NC}"

# CI checks (run in CI pipeline)
ci: deps format-check lint test vuln-check build ## Run all CI checks

# Show project info
info: ## Show project information
	@echo "${BLUE}Project: ${BINARY_NAME}${NC}"
	@echo "${BLUE}Version: ${VERSION}${NC}"
	@echo "${BLUE}Commit: ${COMMIT}${NC}"
	@echo "${BLUE}Build Time: ${BUILD_TIME}${NC}"
	@echo "${BLUE}Go Version: $(shell $(GOCMD) version)${NC}"
