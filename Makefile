GO_OS := $(shell go env GOOS)
GO_ARCH := $(shell go env GOARCH)

GIT_BRANCH := $(shell git symbolic-ref -q --short HEAD || git describe --tags --exact-match)

.PHONY: all
all: generate test

.PHONY: test
test:
	@go test -cover ./...

# todo: Currently broken with Geyser proto definitions due to optional fields
.PHONY: generate
generate:
	@rm -rf generated/*
	@docker run --rm -v $(PWD)/proto:/proto -v $(PWD)/generated:/generated/go code-protobuf-api-builder-go
	@mv $(PWD)/generated/github.com/code-payments/code-vm-indexer/generated/go/* $(PWD)/generated
	@rm -rf $(PWD)/generated/github.com

.PHONY: build-rpc
build-rpc:
	@GOOS=$(GO_OS) GOARCH=$(GO_ARCH) CGO_ENABLED=0 go build -o build/$(GO_OS)-$(GO_ARCH)/rpc github.com/code-payments/code-vm-indexer/app/rpc

.PHONY: build-geyser
build-geyser:
	@GOOS=$(GO_OS) GOARCH=$(GO_ARCH) CGO_ENABLED=0 go build -o build/$(GO_OS)-$(GO_ARCH)/geyser github.com/code-payments/code-vm-indexer/app/geyser

.PHONY: build-rpc-image
build-rpc-image: GO_OS := linux
build-rpc-image: GO_ARCH := amd64
build-rpc-image: build-rpc
	@docker build -f Dockerfile.rpc --platform linux/amd64 -t code-vm-indexer-rpc-service:$(GIT_BRANCH) .

.PHONY: build-geyser-image
build-geyser-image: GO_OS := linux
build-geyser-image: GO_ARCH := amd64
build-geyser-image: build-geyser
	@docker build -f Dockerfile.geyser --platform linux/amd64 -t code-vm-indexer-geyser-service:$(GIT_BRANCH) .

.PHONY: run-rpc
run-rpc: GO_OS := linux
run-rpc: GO_ARCH := amd64
run-rpc: build-rpc build-rpc-image
	@docker run \
		--rm \
		-e APP_NAME=code-vm-indexer-rpc-service \
		-e DATA_STORAGE_TYPE=$(DATA_STORAGE_TYPE) \
		-e INSECURE_LISTEN_ADDRESS=:8086 \
		-e LOG_LEVEL=trace \
		-e POSTGRES_USER=$(POSTGRES_USER) \
		-e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
		-e POSTGRES_HOST=$(POSTGRES_HOST) \
		-e POSTGRES_PORT=$(POSTGRES_PORT) \
		-e POSTGRES_DB_NAME=$(POSTGRES_DB_NAME) \
		-e RAM_TABLE_NAME=$(RAM_TABLE_NAME) \
		-p 8086:8086 \
		code-vm-indexer-rpc-service:$(GIT_BRANCH)

.PHONY: run-geyser
run-geyser: GO_OS := linux
run-geyser: GO_ARCH := amd64
run-geyser: build-geyser build-geyser-image
	@docker run \
		--rm \
		-e APP_NAME=code-vm-indexer-geyser-service \
		-e DATA_STORAGE_TYPE=$(DATA_STORAGE_TYPE) \
		-e GEYSER_WORKER_GRPC_PLUGIN_ENDPOINT=$(GEYSER_WORKER_GRPC_PLUGIN_ENDPOINT) \
		-e GEYSER_WORKER_PROGRAM_UPDATE_WORKER_COUNT=4 \
		-e GEYSER_WORKER_VM_ACCOUNT=$(GEYSER_WORKER_VM_ACCOUNT) \
		-e POSTGRES_USER=$(POSTGRES_USER) \
		-e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
		-e POSTGRES_HOST=$(POSTGRES_HOST) \
		-e POSTGRES_PORT=$(POSTGRES_PORT) \
		-e POSTGRES_DB_NAME=$(POSTGRES_DB_NAME) \
		-e RAM_TABLE_NAME=$(RAM_TABLE_NAME) \
		-e SOLANA_RPC_ENDPOINT=$(SOLANA_RPC_ENDPOINT) \
		-e LOG_LEVEL=trace \
		code-vm-indexer-geyser-service:$(GIT_BRANCH)
