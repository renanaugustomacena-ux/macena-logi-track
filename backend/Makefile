# LogiTrack Backend Makefile
# Common development and CI targets. Invoke with `make <target>`.

BINARY      := logitrack-server
PKG         := ./...
BUILD_DIR   := bin
VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS     := -s -w -X main.version=$(VERSION)

.PHONY: help
help: ## Show available targets
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9_.-]+:.*?## / {printf "  %-18s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: tidy
tidy: ## Download and tidy module dependencies
	go mod tidy
	go mod verify

.PHONY: fmt
fmt: ## Format source with gofmt and goimports
	gofmt -s -w .

.PHONY: vet
vet: ## Run `go vet`
	go vet $(PKG)

.PHONY: lint
lint: ## Run golangci-lint (must be installed locally)
	golangci-lint run --timeout 5m

.PHONY: test
test: ## Run unit tests with race detector and coverage
	go test -race -cover -covermode=atomic -coverprofile=coverage.out $(PKG)

.PHONY: test-short
test-short: ## Run short tests only
	go test -short -race $(PKG)

.PHONY: build
build: ## Build the production binary
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY) ./cmd/server

.PHONY: run
run: ## Run the server locally (reads .env in cwd)
	go run ./cmd/server

.PHONY: docker
docker: ## Build the Docker image
	docker build -t logitrack/backend:$(VERSION) .

.PHONY: clean
clean: ## Remove build artefacts
	rm -rf $(BUILD_DIR) coverage.out
