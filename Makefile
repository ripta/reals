.PHONY: help clean test coverage coverage-func coverage-html coverage-clean

COVERAGE_PROFILE := coverage.out
COVERAGE_HTML := coverage.html

help: ## Show this help
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

clean: coverage-clean ## Remove build artifacts

test: ## Run all unit tests
	go test -v ./...

coverage: ## Run tests and write a coverage profile
	go test -coverprofile=$(COVERAGE_PROFILE) ./...

coverage-func: coverage ## Show per-function coverage summary
	go tool cover -func=$(COVERAGE_PROFILE)

coverage-html: coverage ## Generate an HTML coverage report
	go tool cover -html=$(COVERAGE_PROFILE) -o $(COVERAGE_HTML)

coverage-clean: ## Remove generated coverage artifacts
	rm -f $(COVERAGE_PROFILE) $(COVERAGE_HTML)
