# Demo Shop Makefile

.PHONY: run down test lint build clean help

## Podman operations
run: ## Start the application with podman-compose
	podman-compose up --build --force-recreate

rund: ## Start the application with podman-compose (detached mode)
	podman-compose up --build --force-recreate -d

down: ## Stop and clean up containers
	podman-compose down --remove-orphans
	podman rmi $(shell podman images -q --filter dangling=true) --force 2>/dev/null || true

## Development operations
test: ## Run all tests with coverage
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-short: ## Run tests without coverage
	go test -v ./...

lint: ## Run linter
	golangci-lint run

clean: ## Clean build artifacts
	rm -rf bin/ coverage.out coverage.html

install-tools: ## Install development tools
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/goreleaser/goreleaser@latest

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
