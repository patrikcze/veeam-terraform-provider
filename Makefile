# Makefile for Veeam Terraform Provider

# Variables
GO ?= go
GOPATH ?= $(shell $(GO) env GOPATH)
GOBIN ?= $(GOPATH)/bin
BINARY_NAME=terraform-provider-veeam
VERSION?=dev
LDFLAGS=-ldflags "-X main.version=$(VERSION)"
GOOS?=$(shell $(GO) env GOOS)
GOARCH?=$(shell $(GO) env GOARCH)

UNIT_PACKAGES=./internal/... ./pkg/...
ALL_PACKAGES=./...
ACC_TEST_PACKAGES=./tests

GOLANGCI_LINT_VERSION?=v2.11.3
GOLANGCI_LINT=$(GOBIN)/golangci-lint
TFPLUGINDOCS_VERSION?=v0.24.0
TFPLUGINDOCS=$(GOBIN)/tfplugindocs

# Default target
.DEFAULT_GOAL := build

# Validate Go toolchain consistency (prevents go/go tool version mismatch issues)
.PHONY: toolchain-check
toolchain-check:
	@echo "Checking Go toolchain consistency..."
	@go_ver="$$( $(GO) version | awk '{print $$3}' )"; \
	tool_ver="$$( $(GO) tool compile -V | awk '{print $$3}' )"; \
	if [ "$$go_ver" != "$$tool_ver" ]; then \
		echo "Go toolchain mismatch detected: go=$$go_ver, compile=$$tool_ver"; \
		echo "Fix GOROOT/Go installation and retry."; \
		exit 1; \
	fi

# Build the provider
.PHONY: build
build: toolchain-check
	@echo "Building $(BINARY_NAME) for $(GOOS)/$(GOARCH)..."
	@$(GO) build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/veeam

# Build for all supported platforms
.PHONY: build-all
build-all: toolchain-check
	@echo "Building for all platforms..."
	@GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o bin/$(BINARY_NAME)_linux_amd64 ./cmd/veeam
	@GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -o bin/$(BINARY_NAME)_linux_arm64 ./cmd/veeam
	@GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o bin/$(BINARY_NAME)_darwin_amd64 ./cmd/veeam
	@GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o bin/$(BINARY_NAME)_darwin_arm64 ./cmd/veeam
	@GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o bin/$(BINARY_NAME)_windows_amd64.exe ./cmd/veeam

# Install the provider locally
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME) locally..."
	@mkdir -p ~/.terraform.d/plugins/registry.terraform.io/patrikcze/veeam/$(VERSION)/$(GOOS)_$(GOARCH)
	@cp bin/$(BINARY_NAME) ~/.terraform.d/plugins/registry.terraform.io/patrikcze/veeam/$(VERSION)/$(GOOS)_$(GOARCH)/

# Run tests
.PHONY: test
test: test-unit

# Run unit tests only (internal + pkg)
.PHONY: test-unit
test-unit: toolchain-check
	@echo "Running unit tests..."
	@$(GO) test -v $(UNIT_PACKAGES)

# Run all tests including tests/ package (acceptance tests still require TF_ACC=1)
.PHONY: test-all
test-all: toolchain-check
	@echo "Running tests..."
	@$(GO) test -v $(ALL_PACKAGES)

# Run tests with coverage
.PHONY: test-coverage
test-coverage: toolchain-check
	@echo "Running tests with coverage..."
	@$(GO) test -v -coverprofile=coverage.txt -covermode=atomic $(UNIT_PACKAGES)
	@$(GO) tool cover -html=coverage.txt -o coverage.html

# Run acceptance tests
.PHONY: testacc
testacc: toolchain-check
	@echo "Running acceptance tests..."
	@TF_ACC=1 VEEAM_HOST=$(VEEAM_HOST) VEEAM_USERNAME=$(VEEAM_USERNAME) VEEAM_PASSWORD=$(VEEAM_PASSWORD) VEEAM_INSECURE=$(VEEAM_INSECURE) $(GO) test -v $(ACC_TEST_PACKAGES) -timeout 120m

# Run all tests
.PHONY: testall
testall: test-all

# Lint the code
.PHONY: lint
lint: toolchain-check
	@echo "Running linter..."
	@if [ ! -x "$(GOLANGCI_LINT)" ]; then \
		echo "golangci-lint not found, installing $(GOLANGCI_LINT_VERSION)..."; \
		GOBIN="$(GOBIN)" $(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION); \
	fi
	@if ! "$(GOLANGCI_LINT)" --version | grep -q "$(GOLANGCI_LINT_VERSION)"; then \
		echo "Updating golangci-lint to $(GOLANGCI_LINT_VERSION)..."; \
		GOBIN="$(GOBIN)" $(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION); \
	fi
	@$(GOLANGCI_LINT) run $(UNIT_PACKAGES)

# Run go vet on provider implementation packages
.PHONY: vet
vet: toolchain-check
	@echo "Running go vet..."
	@$(GO) vet $(UNIT_PACKAGES)

# Format the code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@$(GO) fmt ./...
	@gofmt -s -w .

# Check code formatting
.PHONY: fmt-check
fmt-check:
	@echo "Checking code formatting..."
	@test -z "$$(find . -type f -name '*.go' -not -path './vendor/*' -exec gofmt -l {} +)"

# Download and vendor dependencies
.PHONY: vendor
vendor: toolchain-check
	@echo "Vendoring dependencies..."
	@$(GO) mod tidy
	@$(GO) mod vendor

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -rf dist/
	@rm -rf vendor/
	@rm -f coverage.txt coverage.html

# Generate documentation
.PHONY: docs
docs: toolchain-check
	@echo "Generating documentation..."
	@if [ ! -x "$(TFPLUGINDOCS)" ]; then \
		echo "tfplugindocs not found, installing $(TFPLUGINDOCS_VERSION)..."; \
		GOBIN="$(GOBIN)" $(GO) install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@$(TFPLUGINDOCS_VERSION); \
	fi
	@$(TFPLUGINDOCS) generate --provider-dir=./cmd/veeam --provider-name=veeam --rendered-website-dir=../../docs --website-source-dir=../../templates

# Run all checks (lint, fmt-check, test)
.PHONY: check
check: fmt-check vet lint test-unit

# Run acceptance tests for specific resource
.PHONY: testacc-credential
testacc-credential: toolchain-check
	@echo "Running credential acceptance tests..."
	@TF_ACC=1 VEEAM_HOST=$(VEEAM_HOST) VEEAM_USERNAME=$(VEEAM_USERNAME) VEEAM_PASSWORD=$(VEEAM_PASSWORD) VEEAM_INSECURE=$(VEEAM_INSECURE) $(GO) test -v $(ACC_TEST_PACKAGES) -run TestAccCredential -timeout 60m

.PHONY: testacc-repository
testacc-repository: toolchain-check
	@echo "Running repository acceptance tests..."
	@TF_ACC=1 VEEAM_HOST=$(VEEAM_HOST) VEEAM_USERNAME=$(VEEAM_USERNAME) VEEAM_PASSWORD=$(VEEAM_PASSWORD) VEEAM_INSECURE=$(VEEAM_INSECURE) $(GO) test -v $(ACC_TEST_PACKAGES) -run TestAccRepository -timeout 60m

.PHONY: testacc-backup-job
testacc-backup-job: toolchain-check
	@echo "Running backup job acceptance tests..."
	@TF_ACC=1 VEEAM_HOST=$(VEEAM_HOST) VEEAM_USERNAME=$(VEEAM_USERNAME) VEEAM_PASSWORD=$(VEEAM_PASSWORD) VEEAM_INSECURE=$(VEEAM_INSECURE) $(GO) test -v $(ACC_TEST_PACKAGES) -run TestAccBackupJob -timeout 60m

.PHONY: testacc-proxy
testacc-proxy: toolchain-check
	@echo "Running proxy acceptance tests..."
	@TF_ACC=1 VEEAM_HOST=$(VEEAM_HOST) VEEAM_USERNAME=$(VEEAM_USERNAME) VEEAM_PASSWORD=$(VEEAM_PASSWORD) VEEAM_INSECURE=$(VEEAM_INSECURE) $(GO) test -v $(ACC_TEST_PACKAGES) -run TestAccProxy -timeout 60m

.PHONY: testacc-scale-out-repository
testacc-scale-out-repository: toolchain-check
	@echo "Running scale-out repository acceptance tests..."
	@TF_ACC=1 VEEAM_HOST=$(VEEAM_HOST) VEEAM_USERNAME=$(VEEAM_USERNAME) VEEAM_PASSWORD=$(VEEAM_PASSWORD) VEEAM_INSECURE=$(VEEAM_INSECURE) $(GO) test -v $(ACC_TEST_PACKAGES) -run TestAccScaleOutRepository -timeout 60m

.PHONY: testacc-workflow
testacc-workflow: toolchain-check
	@echo "Running workflow acceptance tests..."
	@TF_ACC=1 VEEAM_HOST=$(VEEAM_HOST) VEEAM_USERNAME=$(VEEAM_USERNAME) VEEAM_PASSWORD=$(VEEAM_PASSWORD) VEEAM_INSECURE=$(VEEAM_INSECURE) $(GO) test -v $(ACC_TEST_PACKAGES) -run TestAccWorkflow -timeout 60m

# Set up test environment
.PHONY: setup-test-env
setup-test-env:
	@echo "Setting up test environment..."
	@if [ ! -f .env.test ]; then \
		cp .env.test.example .env.test; \
		echo "Created .env.test file. Please edit it with your Veeam server details."; \
	fi

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build               - Build the provider binary"
	@echo "  build-all           - Build for all supported platforms"
	@echo "  install             - Install the provider locally"
	@echo "  toolchain-check     - Validate go and compile tool versions match"
	@echo "  test                - Run unit tests (alias for test-unit)"
	@echo "  test-unit           - Run unit tests for internal/ and pkg/"
	@echo "  test-all            - Run all go tests (including tests/)"
	@echo "  test-coverage       - Run tests with coverage report"
	@echo "  testacc             - Run acceptance tests"
	@echo "  testacc-credential             - Run credential acceptance tests"
	@echo "  testacc-repository             - Run repository acceptance tests (needs TF_VAR_test_host_id)"
	@echo "  testacc-proxy                  - Run proxy acceptance tests (needs TF_VAR_test_host_id)"
	@echo "  testacc-scale-out-repository   - Run scale-out repository acceptance tests"
	@echo "  testacc-backup-job             - Run backup job acceptance tests"
	@echo "  testacc-workflow               - Run workflow acceptance tests"
	@echo "  testall             - Run all tests"
	@echo "  setup-test-env      - Set up test environment"
	@echo "  lint                - Run linter"
	@echo "  vet                 - Run go vet"
	@echo "  fmt                 - Format code"
	@echo "  fmt-check           - Check code formatting"
	@echo "  vendor              - Download and vendor dependencies"
	@echo "  clean               - Clean build artifacts"
	@echo "  docs                - Generate documentation"
	@echo "  check               - Run all checks (fmt-check, vet, lint, test-unit)"
	@echo "  help                - Show this help message"
