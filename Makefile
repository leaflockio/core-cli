.PHONY: help build build-dev build-prod install tidy lint lint-fix fmt test test-verbose test-coverage test-clean clean

# Variables
GO           := go
BINARY       := leaf
BIN_DIR      := bin
COVERAGE_DIR := coverage
PKG          := ./...
LINTER       := golangci-lint

# Build flags
VERSION        := $(strip $(file <.version))
GCFLAGS_DEV    := all=-N -l
LDFLAGS_COMMON := -X main.version=$(VERSION)
LDFLAGS_DEV    := $(LDFLAGS_COMMON) -X main.env=dev
LDFLAGS_PROD   := $(LDFLAGS_COMMON) -X main.env=prod -s -w

help: ## Show this help message
	@echo 'Leaf CLI Development Commands'
	@echo ''
	@echo 'Usage:'
	@echo '  make <target>'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

build: build-prod ## Build the leaf binary (prod by default)

build-dev: ## Build with dev profile (debug symbols, no optimization)
	@echo "Building $(BINARY) [dev]..."
	@mkdir -p $(BIN_DIR)
	@$(GO) build -gcflags="$(GCFLAGS_DEV)" -ldflags="$(LDFLAGS_DEV)" -o $(BIN_DIR)/$(BINARY) .
	@echo "Binary: $(BIN_DIR)/$(BINARY)"

build-prod: ## Build with prod profile (stripped, optimized)
	@echo "Building $(BINARY) [prod]..."
	@mkdir -p $(BIN_DIR)
	@$(GO) build -ldflags="$(LDFLAGS_PROD)" -o $(BIN_DIR)/$(BINARY) .
	@echo "Binary: $(BIN_DIR)/$(BINARY)"

install: ## Install leaf to GOPATH/bin
	@echo "Installing $(BINARY)..."
	@$(GO) install .
	@echo "$(BINARY) installed"

tidy: ## Tidy go modules
	@echo "Tidying modules..."
	@$(GO) mod tidy
	@echo "Done"

lint: ## Run linter on entire codebase
	@echo "Running linter..."
	@$(LINTER) run

lint-fix: ## Run linter with auto-fix
	@echo "Running linter with auto-fix..."
	@$(LINTER) run --fix

fmt: ## Format code
	@echo "Formatting code..."
	@$(GO) fmt $(PKG)
	@gofumpt -l -w .
	@echo "Code formatted"

test: ## Run tests
	@echo "Running tests..."
	@$(GO) test -race $(PKG)

test-verbose: ## Run tests with verbose output
	@echo "Running tests with verbose output..."
	@$(GO) test -v -race -count=1 $(PKG)

test-coverage: ## Generate test coverage report
	@echo "Generating coverage report..."
	@mkdir -p $(COVERAGE_DIR)
	@$(GO) test -race -coverprofile=$(COVERAGE_DIR)/coverage.out -covermode=atomic $(PKG)
	@$(GO) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/index.html
	@echo "Coverage report: $(COVERAGE_DIR)/index.html"

test-clean: ## Remove test artifacts (coverage reports)
	@echo "Cleaning test artifacts..."
	@rm -rf $(COVERAGE_DIR)
	@echo "Cleaned"

clean: ## Remove build artifacts (bin/)
	@echo "Cleaning build artifacts..."
	@rm -rf $(BIN_DIR)
	@echo "Cleaned"
