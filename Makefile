APP      := afrisinc-assets
CMD      := ./cmd/server
BIN      := ./bin/$(APP)
GOFLAGS  := -ldflags="-s -w"
GOPATH   := $(shell go env GOPATH)
AIR      := $(GOPATH)/bin/air

.PHONY: build run dev test lint migrate clean

## build: compile a production binary into ./bin/
build:
	@mkdir -p bin
	go build $(GOFLAGS) -o $(BIN) $(CMD)

## run: build and start the server
run: build
	$(BIN)

## dev: run with auto-reload (requires: go install github.com/air-verse/air@latest)
dev:
	@command -v $(AIR) >/dev/null 2>&1 || go install github.com/air-verse/air@latest
	$(AIR)

## test: run all tests with race detector
test:
	go test -race -cover ./...

## lint: run staticcheck + go vet
lint:
	go vet ./...
	staticcheck ./...

## migrate: apply database migrations
migrate:
	psql $$DATABASE_URL -f migrations/001_init.sql

## tidy: tidy and verify modules
tidy:
	go mod tidy
	go mod verify

## clean: remove build artifacts
clean:
	rm -rf bin/
