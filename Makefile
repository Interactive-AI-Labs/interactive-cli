.PHONY: help docs test test-unit test-coverage kubeconfig start stop clean

help:
	@echo "📝 Available commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
	@echo ""

docs: ## Generate CLI documentation
	@go run main.go gen-docs

test: test-unit ## Run unit tests (default test target)

test-unit: ## Run unit tests
	@echo "Running unit tests..."
	@go test -v ./...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

kubeconfig:
	@./generate-kubeconfig.sh

start: kubeconfig
	@echo ""
	@echo "🐳 Starting Docker containers..."
	@docker compose down 2>/dev/null || true
	@docker compose up

stop:
	@echo "🛑 Stopping Docker containers..."
	@docker compose down

clean:
	@echo "🧹 Cleaning up..."
	@docker compose down 2>/dev/null || true
	@rm -rf tmp/kubeconfig-docker
	@echo "✅ Cleaned up"

.DEFAULT_GOAL := help
