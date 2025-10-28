.PHONY: build clean test

GO = $(shell which go)
BIN = ./bin

GO_FLAGS=-ldflags "-X 'github.com/seanmcgary/x402-go/internal/version.Version=$(shell cat VERSION)' -X 'github.com/seanmcgary/x402-go/internal/version.Commit=$(shell git rev-parse --short HEAD 2>/dev/null || echo 'unknown')'"


all: deps/go build/cmd

# -----------------------------------------------------------------------------
# Dependencies
# -----------------------------------------------------------------------------
deps: deps/go


.PHONY: deps/go
deps/go:
	${GO} mod tidy
	$(GO) install github.com/vektra/mockery/v3@v3.5.5
	$(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.5.0


# -----------------------------------------------------------------------------
# Build binaries
# -----------------------------------------------------------------------------

.PHONY: cmd/facilitator
build/cmd/facilitator:
	go build $(GO_FLAGS) -o ${BIN}/facilitator ./cmd/facilitator

.PHONY: build/cmd
build/cmd: build/cmd/facilitator

.PHONY: docker/build
docker/build:
	docker build -t facilitator:latest .
# -----------------------------------------------------------------------------
# Tests and linting
# -----------------------------------------------------------------------------
.PHONY: test
test:
	GOFLAGS="-count=1" ./scripts/goTest.sh -v -p 1 -parallel 1 ./...

.PHONY: lint
lint:
	golangci-lint run --timeout "5m"

.PHONY: fmt
fmt:
	gofmt -w .

.PHONY: fmtcheck
fmtcheck:
	@unformatted_files=$$(gofmt -l .); \
	if [ -n "$$unformatted_files" ]; then \
		echo "The following files are not properly formatted:"; \
		echo "$$unformatted_files"; \
		echo "Please run 'gofmt -w .' to format them."; \
		exit 1; \
	fi

.PHONY: mocks
mocks:
	@echo "Generating mocks..."
	mockery

