BINARY := srest
PREFIX ?= $(HOME)/.local/bin

.PHONY: help build install run test vet lint fmt clean

help: ## Show this help
	@echo "srest - available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary into the current directory
	go build -o $(BINARY) .

install: build ## Build and install (make install PREFIX=/usr/local/bin)
	install -d $(PREFIX)
	install -m 0755 $(BINARY) $(PREFIX)/$(BINARY)

run: ## Run from source
	go run .

test: ## Run unit tests
	go test ./...

vet: ## Run go vet
	go vet ./...

lint: ## Run golangci-lint
	golangci-lint run ./...

fmt: ## Format the code
	gofmt -w .

clean: ## Remove the built binary
	rm -f $(BINARY)

.DEFAULT_GOAL := help