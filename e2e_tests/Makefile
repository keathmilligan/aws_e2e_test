# Makefile for end-to-end tests

# Default API URL (can be overridden via command line)
API_URL ?= $(shell echo $$API_URL)

.PHONY: test
test:
	@if [ -z "$(API_URL)" ]; then \
		echo "Error: API_URL is not set. Use 'make test API_URL=https://your-api-url.com'"; \
		exit 1; \
	fi
	@echo "Running end-to-end tests against API at: $(API_URL)"
	go test -v -api-url="$(API_URL)"

.PHONY: deps
deps:
	go mod tidy

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  test     - Run the end-to-end tests (requires API_URL)"
	@echo "  deps     - Install dependencies"
	@echo "  help     - Show this help message"
	@echo ""
	@echo "Example usage:"
	@echo "  make test API_URL=https://your-api-url.com"
	@echo "  API_URL=https://your-api-url.com make test"