.PHONY: help buf-gen

.DEFAULT_GOAL := help

# Colors for terminal output
CYAN   := \033[0;36m
GREEN  := \033[0;32m
YELLOW := \033[0;33m
RESET  := \033[0m

GOBIN := $(shell go env GOPATH)/bin

help: ## Show this help message
	@printf "$(CYAN)Usage:$(RESET) make [target]\n\n"
	@printf "$(CYAN)Available Targets:$(RESET)\n"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-15s$(RESET) %s\n", $$1, $$2}' $(MAKEFILE_LIST)



run: ## Run orchestrator microservice locally
	@printf "$(CYAN)==> Starting orchestrator microservice...$(RESET)\n"
	go run ./cmd/orchestrator --config=./config.yaml

run-binary: ## Run the orchestrator service using binary
	@printf "$(CYAN)==> Starting orchestrator microservice...$(RESET)\n"
	./bin/orchestrator --config=./config.yaml

build: ## Build orchestrator microservice binary
	make remove-binaries
	@printf "$(CYAN)==> Building binary (bin/orchestrator)...$(RESET)\n"
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/orchestrator ./cmd/orchestrator
	@printf "$(GREEN)✓ Binary built successfully at bin/orchestrator$(RESET)\n"

remove-binaries:
	@printf "$(YELLOW)==> Cleaning stale application binaries...$(RESET)\n"
	@find ./bin -mindepth 1 -delete 2>/dev/null || true
	@printf "$(GREEN)✓ Stale application binaries removed.$(RESET)\n"

start-local: ## Start the application locally
	@printf "$(CYAN)==> Starting application locally...$(RESET)\n"
	make remove-binaries
	make build
	make run-binary