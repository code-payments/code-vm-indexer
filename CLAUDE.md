# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

```bash
make test              # Run all tests with coverage
make build-rpc         # Build RPC service binary
make build-geyser      # Build Geyser worker binary
make build-rpc-image   # Build RPC Docker image
make build-geyser-image # Build Geyser Docker image

# Run a single test
go test -run TestName ./path/to/package/...

# Run locally (Docker)
DATA_STORAGE_TYPE=memory make run-rpc
DATA_STORAGE_TYPE=memory GEYSER_WORKER_GRPC_PLUGIN_ENDPOINT=localhost:10000 GEYSER_WORKER_VM_ACCOUNT=<VM_KEY> SOLANA_RPC_ENDPOINT=http://localhost:8899 make run-geyser
```

Proto generation (`make generate`) is currently broken due to optional fields in the Geyser proto.

## Architecture

The indexer has two independently deployable services:

1. **Geyser Worker** (`app/geyser/`) — Subscribes to Yellowstone Geyser plugin via gRPC streaming to receive real-time Code VM account updates. Processes memory account deltas and persists them to storage. Runs a backup worker to periodically re-sync and catch missed events.

2. **RPC Service** (`app/rpc/`) — gRPC server (port 8086) exposing the `Indexer` service with two methods: `GetVirtualTimelockAccounts` and `GetVirtualDurableNonce`. Queries indexed virtual account state from the storage backend.

Both services share a common data layer and are configured via environment variables.

### Data Layer

- **Store interface** (`data/ram/store.go`) — defines `Save`, `GetAllMemoryAccounts`, `GetAllByMemoryAccount`, `GetAllVirtualAccountsByAddressAndType`
- **Implementations**: in-memory (`data/ram/memory/`) and PostgreSQL (`data/ram/postgres/`)
- **Shared test suite** (`data/ram/tests/tests.go`) — `RunTests()` runs the same tests against any Store implementation
- Only memory storage is currently supported for production use

### Geyser Worker Internals

Core processing logic is in `geyser/handler_memory.go`. The worker:
- Caches in-memory account state for delta detection
- Tracks slot progression per memory account to avoid reprocessing
- Uses configurable worker pools (default 1024) and event queue (default 1M)
- Configuration loaded from env vars in `geyser/config.go`

### Proto / Generated Code

- Proto definitions: `proto/indexer/` (service owns) and `proto/geyser/` (external)
- Generated Go code: `generated/indexer/v1/` and `generated/geyser/v1/`

### Key Dependencies

- `github.com/code-payments/ocp-server` — Solana/CVM utilities, gRPC infrastructure, retry logic
- Go 1.26
