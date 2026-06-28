.DEFAULT_GOAL := help

## Shell config
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

## Tool Binaries
GO ?= go
GOLANGCI_LINT ?= $(GO) tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: clean
clean: ## Clean up files.
	find . -name .DS_Store -type f -delete

##@ Development

.PHONY: go-tidy
go-tidy: ## Tidy go.mod and go.sum.
	$(GO) mod tidy

.PHONY: go-tidy-check
go-tidy-check: ## Check if go.mod and go.sum are tidy.
	$(GO) mod tidy --diff

.PHONY: go-mod-download
go-mod-download: ## Download dependencies from go.mod and go.sum.
	$(GO) mod download

.PHONY: install-deps
install-deps: go-mod-download ## Install dependencies.

.PHONY: lint
lint: ## Run linters.
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: ## Run linters and perform fixes.
	$(GOLANGCI_LINT) run --fix

.PHONY: test
test: TESTFLAGS := -v -race
test: TESTTARGET := ./...
test: ## Run unit tests.
	$(GO) test $(TESTFLAGS) $(TESTTARGET)

.PHONY: test-bench
test-bench: TESTTARGET := ./...
test-bench: ## Run bench tests
	$(GO) test -bench=$(TESTTARGET) -benchmem $(TESTTARGET)
