.PHONY: help build build-dev build-prod install tidy lint lint-fix fmt test test-verbose test-coverage test-coverage-verbose test-clean clean

# Variables
GO           := go
BINARY       := leaf
BIN_DIR      := bin
COVERAGE_DIR := coverage
PKG          := ./...
LINTER       := golangci-lint

# Build flags
VERSION        := $(strip $(shell cat .version))
COMMIT         := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE           := $(shell date -u +%Y-%m-%d)
PKG_VERSION    := github.com/leaflockio/core-cli/internal/version
GCFLAGS_DEV    := all=-N -l
LDFLAGS_COMMON := -X $(PKG_VERSION).Version=$(VERSION) -X $(PKG_VERSION).Commit=$(COMMIT) -X $(PKG_VERSION).Date=$(DATE)
LDFLAGS_DEV    := $(LDFLAGS_COMMON) -X $(PKG_VERSION).Version=$(VERSION)-dev -X main.env=dev
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
	@$(GO) build -gcflags="$(GCFLAGS_DEV)" -ldflags="$(LDFLAGS_DEV)" -o $(BIN_DIR)/$(BINARY) ./cmd
	@echo "Binary: $(BIN_DIR)/$(BINARY)"

build-prod: ## Build with prod profile (stripped, optimized)
	@echo "Building $(BINARY) [prod]..."
	@mkdir -p $(BIN_DIR)
	@$(GO) build -ldflags="$(LDFLAGS_PROD)" -o $(BIN_DIR)/$(BINARY) ./cmd
	@echo "Binary: $(BIN_DIR)/$(BINARY)"

install: ## Install leaf to GOPATH/bin (prod profile)
	@echo "Installing $(BINARY)..."
	@$(GO) install -ldflags="$(LDFLAGS_PROD)" ./cmd
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
	@gofumpt -l -w .
	@echo "Code formatted"

test: ## Run tests
	@echo "Running tests..."
	@mkdir -p $(COVERAGE_DIR)
	@$(GO) test -race -coverprofile=$(COVERAGE_DIR)/coverage.out -covermode=atomic $(PKG)

test-verbose: ## Run tests with verbose output
	@echo "Running tests with verbose output..."
	@$(GO) test -v -race -count=1 $(PKG)

test-coverage: ## Generate test coverage report
	@echo "Generating coverage report..."
	@mkdir -p $(COVERAGE_DIR)
	@$(GO) test -race -coverprofile=$(COVERAGE_DIR)/coverage.out -covermode=atomic $(PKG)
	@$(GO) tool cover -func=$(COVERAGE_DIR)/coverage.out | tail -1
	@$(GO) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/index.html
	@echo "Coverage report: $(COVERAGE_DIR)/index.html"

test-coverage-verbose: ## Generate coverage report with per-test results, totals, and color
	@echo "Generating coverage report..."
	@mkdir -p $(COVERAGE_DIR)
	@$(GO) test -v -race -coverprofile=$(COVERAGE_DIR)/coverage.out -covermode=atomic $(PKG) \
		2>&1 | tee $(COVERAGE_DIR)/test.log | \
		awk '/^--- PASS:/ && $$3 !~ /\// { printf "\033[32m--- PASS\033[0m: %s %s\n", $$3, $$4 } \
		     /^--- FAIL:/ && $$3 !~ /\// { printf "\033[31m--- FAIL\033[0m: %s %s\n", $$3, $$4 } \
		     /^(ok|FAIL)[[:space:]]/ { print ""; print }'
	@echo ""
	@awk 'BEGIN { p=0; f=0 } \
	      /^--- PASS:/ && $$3 !~ /\// { p++ } \
	      /^--- FAIL:/ && $$3 !~ /\// { f++ } \
	      END { printf "\033[32mPassed: %d\033[0m  \033[31mFailed: %d\033[0m  Total: %d\n", p, f, p+f }' \
		$(COVERAGE_DIR)/test.log
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
