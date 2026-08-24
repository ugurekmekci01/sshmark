.PHONY: build test test-race lint fmt fmt-check vet hooks setup clean

BINARY := sshmark
BUILD_DIR := bin
MAIN := ./cmd/sshmark

build:
	go build -o $(BUILD_DIR)/$(BINARY) $(MAIN)

test:
	go test -count=1 ./...

test-race:
	go test -race -count=1 ./...

test-integration:
	go test -tags=integration -count=1 ./...

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .
	goimports -w .

fmt-check:
	@test -z "$$(gofmt -l .)" || (echo "gofmt needed on:"; gofmt -l .; exit 1)
	@command -v goimports >/dev/null 2>&1 && test -z "$$(goimports -l .)" || true

vet:
	go vet ./...

hooks:
	./scripts/setup-hooks.sh

setup: hooks
	@echo "Run: go mod tidy && make test"

ci: fmt-check vet lint test-race build

clean:
	rm -rf $(BUILD_DIR) dist coverage.out
