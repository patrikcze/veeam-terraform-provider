# Makefile for Veeam Terraform Provider

# Variables
BINARY_NAME=terraform-provider-veeam
VERSION?=dev
LDFLAGS=-ldflags "-X main.version=$(VERSION)"
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)

# Default target
.DEFAULT_GOAL := build

# Build the provider
.PHONY: build
build:
	@echo "Building $(BINARY_NAME) for $(GOOS)/$(GOARCH)..."
	@go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/veeam

# Build for all supported platforms
.PHONY: build-all
build-all:
	@echo "Building for all platforms..."
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)_linux_amd64 ./cmd/veeam
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)_linux_arm64 ./cmd/veeam
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)_darwin_amd64 ./cmd/veeam
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)_darwin_arm64 ./cmd/veeam
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)_windows_amd64.exe ./cmd/veeam

# Install the provider locally
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME) locally..."
	@mkdir -p ~/.terraform.d/plugins/registry.terraform.io/patrikcze/veeam/$(VERSION)/$(GOOS)_$(GOARCH)
	@cp bin/$(BINARY_NAME) ~/.terraform.d/plugins/registry.terraform.io/patrikcze/veeam/$(VERSION)/$(GOOS)_$(GOARCH)/

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.txt -covermode=atomic ./...
	@go tool cover -html=coverage.txt -o coverage.html

# Run acceptance tests
.PHONY: testacc
testacc:
	@echo "Running acceptance tests..."
	@TF_ACC=1 VEEAM_HOST=$(VEEAM_HOST) VEEAM_USERNAME=$(VEEAM_USERNAME) VEEAM_PASSWORD=$(VEEAM_PASSWORD) VEEAM_INSECURE=$(VEEAM_INSECURE) go test -v ./tests -timeout 120m

# Run all tests
.PHONY: testall
testall:
	@echo "Running all tests..."
	@VEEAM_HOST=$(VEEAM_HOST) VEEAM_USERNAME=$(VEEAM_USERNAME) VEEAM_PASSWORD=$(VEEAM_PASSWORD) VEEAM_INSECURE=$(VEEAM_INSECURE) go test -v ./... -timeout 120m

# Lint the code
.PHONY: lint
lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found, installing..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@golangci-lint run ./...

# Format the code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@gofmt -s -w .

# Check code formatting
.PHONY: fmt-check
fmt-check:
	@echo "Checking code formatting..."
	@test -z $$(gofmt -l .)

# Download and vendor dependencies
.PHONY: vendor
vendor:
	@echo "Vendoring dependencies..."
	@go mod tidy
	@go mod vendor

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
docs:
	@echo "Generating documentation..."
	@which tfplugindocs > /dev/null || (echo "tfplugindocs not found, installing..." && go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest)
	@tfplugindocs

# Run all checks (lint, fmt-check, test)
.PHONY: check
check: lint fmt-check test

# Run acceptance tests for specific resource
.PHONY: testacc-credential
testacc-credential:
	@echo "Running credential acceptance tests..."
	@TF_ACC=1 VEEAM_HOST=$(VEEAM_HOST) VEEAM_USERNAME=$(VEEAM_USERNAME) VEEAM_PASSWORD=$(VEEAM_PASSWORD) VEEAM_INSECURE=$(VEEAM_INSECURE) go test -v ./tests -run TestAccCredential -timeout 60m

.PHONY: testacc-repository
testacc-repository:
	@echo "Running repository acceptance tests..."
	@TF_ACC=1 VEEAM_HOST=$(VEEAM_HOST) VEEAM_USERNAME=$(VEEAM_USERNAME) VEEAM_PASSWORD=$(VEEAM_PASSWORD) VEEAM_INSECURE=$(VEEAM_INSECURE) go test -v ./tests -run TestAccRepository -timeout 60m

.PHONY: testacc-backup-job
testacc-backup-job:
	@echo "Running backup job acceptance tests..."
	@TF_ACC=1 VEEAM_HOST=$(VEEAM_HOST) VEEAM_USERNAME=$(VEEAM_USERNAME) VEEAM_PASSWORD=$(VEEAM_PASSWORD) VEEAM_INSECURE=$(VEEAM_INSECURE) go test -v ./tests -run TestAccBackupJob -timeout 60m

.PHONY: testacc-workflow
testacc-workflow:
	@echo "Running workflow acceptance tests..."
	@TF_ACC=1 VEEAM_HOST=$(VEEAM_HOST) VEEAM_USERNAME=$(VEEAM_USERNAME) VEEAM_PASSWORD=$(VEEAM_PASSWORD) VEEAM_INSECURE=$(VEEAM_INSECURE) go test -v ./tests -run TestAccWorkflow -timeout 60m

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
	@echo "  test                - Run unit tests"
	@echo "  test-coverage       - Run tests with coverage report"
	@echo "  testacc             - Run acceptance tests"
	@echo "  testacc-credential  - Run credential acceptance tests"
	@echo "  testacc-repository  - Run repository acceptance tests"
	@echo "  testacc-backup-job  - Run backup job acceptance tests"
	@echo "  testacc-workflow    - Run workflow acceptance tests"
	@echo "  testall             - Run all tests"
	@echo "  setup-test-env      - Set up test environment"
	@echo "  lint                - Run linter"
	@echo "  fmt                 - Format code"
	@echo "  fmt-check           - Check code formatting"
	@echo "  vendor              - Download and vendor dependencies"
	@echo "  clean               - Clean build artifacts"
	@echo "  docs                - Generate documentation"
	@echo "  check               - Run all checks (lint, fmt-check, test)"
	@echo "  help                - Show this help message"
