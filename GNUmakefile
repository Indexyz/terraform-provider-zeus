default: fmt lint install generate

CACHE_DIR := $(CURDIR)/.cache
GO_BUILD_CACHE_DIR := $(CACHE_DIR)/go-build

GO_ENV := GOCACHE=$(GO_BUILD_CACHE_DIR) XDG_CACHE_HOME=$(CACHE_DIR)

build:
	mkdir -p $(GO_BUILD_CACHE_DIR)
	$(GO_ENV) go build -v ./...

install: build
	mkdir -p $(GO_BUILD_CACHE_DIR)
	$(GO_ENV) go install -v ./...

lint:
	mkdir -p $(GO_BUILD_CACHE_DIR) $(CACHE_DIR)/golangci-lint
	$(GO_ENV) golangci-lint run

generate:
	mkdir -p $(GO_BUILD_CACHE_DIR)
	$(GO_ENV) cd tools; go generate ./...

fmt:
	gofmt -s -w -e .

test:
	mkdir -p $(GO_BUILD_CACHE_DIR)
	$(GO_ENV) go test -v -cover -timeout=120s -parallel=10 ./...

testacc:
	mkdir -p $(GO_BUILD_CACHE_DIR)
	TF_ACC=1 $(GO_ENV) go test -v -cover -timeout 120m ./...

.PHONY: fmt lint test testacc build install generate
