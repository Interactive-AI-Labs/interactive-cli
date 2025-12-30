.PHONY: help docs test test-unit test-integration test-all test-coverage

help:
	@echo "📝 Available commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
	@echo ""

docs: ## Generate CLI documentation
	@go run main.go gen-docs

test: test-unit ## Run unit tests (default test target)

test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	@go test -v ./...

test-integration: ## Run integration tests only
	@echo "Running integration tests..."
	@go test -v -tags=integration ./...

test-all: ## Run all tests (unit + integration)
	@echo "Running all tests..."
	@go test -v ./...
	@go test -v -tags=integration ./...

test-coverage: ## Run unit tests with coverage report
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-coverage-all: ## Run all tests (unit + integration) with coverage report
	@echo "Running all tests with coverage (including integration tests)..."
	@go test -coverprofile=coverage-all.out -tags=integration ./...
	@go tool cover -html=coverage-all.out -o coverage-all.html
	@echo "Combined coverage report generated: coverage-all.html"
	@go tool cover -func=coverage-all.out | grep total:

.DEFAULT_GOAL := help
