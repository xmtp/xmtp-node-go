# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Core Commands
- `dev/up` - Install dependencies and start databases
- `dev/test` - Run all tests (use `RACE=1 dev/test` for race detection)
- `dev/lint` - Run linters (golangci-lint, gofmt, protolint, shellcheck)
- `dev/build` - Build binaries for all commands in `cmd/`
- `dev/start` - Start local node with WebSocket support on port 9002

### Database Operations
- `dev/migrate-message $MIGRATION_NAME` - Create message database migration
- `dev/migrate-mls $MIGRATION_NAME` - Create MLS database migration  
- `dev/migrate-authz $MIGRATION_NAME` - Create authz database migration
- `dev/sqlc` - Regenerate SQLC code after modifying `pkg/mls/store/queries.sql`

### Docker Operations
- `dev/docker/up` - Start database containers
- `dev/docker/down` - Stop database containers
- `dev/psql` - Connect to PostgreSQL database

### Other Commands
- `dev/generate` - Generate protobuf and other generated code
- `dev/prune` - Run database pruning
- `dev/run --metrics` - Start with metrics (Prometheus on :9090)

## Architecture Overview

### Node Software
This is the current XMTP node implementation written in Go. **No new development is planned** - all new work is focused on `xmtpd` (experimental version). This node software currently powers the XMTP network.

### Core Components

**Main Entry Points:**
- `cmd/xmtpd/main.go` - Primary node daemon
- `cmd/prune/main.go` - Database pruning utility

**Server Layer (`pkg/server/`):**
- `server.go` - Main server implementation with lifecycle management
- `options.go` - Configuration and command-line options
- Handles database connections, Waku node setup, and API server initialization

**API Layer (`pkg/api/`):**
- `server.go` - gRPC and HTTP API server setup
- `message/v1/` - Message API implementation
- `identity/` - Identity service
- `mls/` - MLS (Messaging Layer Security) API
- Supports both gRPC and HTTP/WebSocket protocols

**Storage Layer:**
- `pkg/store/` - Message storage with PostgreSQL
- `pkg/mls/store/` - MLS-specific storage using SQLC for query generation
- `pkg/migrations/` - Database migrations for messages, MLS, and authz

**Core Services:**
- `pkg/authn/` - Authentication (transport-level)
- `pkg/authz/` - Authorization with wallet allowlists
- `pkg/mlsvalidate/` - MLS validation service
- `pkg/crypto/` - Cryptographic utilities (ECDSA, signatures)
- `pkg/metrics/` - Prometheus metrics collection
- `pkg/ratelimiter/` - API rate limiting

### Database Architecture
- **Messages DB**: Stores XMTP messages with topic-based indexing
- **MLS DB**: Stores MLS group state, welcomes, and commit logs
- **Authz DB**: Stores wallet allowlists and transaction history
- Uses PostgreSQL with connection pooling and read replicas

### Network Layer
- Built on Waku v2 protocol with libp2p for peer-to-peer networking
- Supports both relay and direct messaging protocols
- WebSocket support for browser clients

### Protocol Buffers
- Extensive use of protobuf for API definitions
- Generated code in `pkg/proto/` for various XMTP protocols
- OpenAPI documentation generated from protobuf definitions

## Development Notes

### Go Version
Must use exact Go version specified in `go.mod` (currently 1.20). Verify with `go version`.

### Generated Code
- After modifying `.proto` files, run `dev/generate`
- After modifying `pkg/mls/store/queries.sql`, run `dev/sqlc`
- Generated files are committed to the repository

### Testing
- Uses standard Go testing with testify
- Database tests require Docker containers to be running
- Race detection available with `RACE=1 dev/test`

### Linting
- golangci-lint with custom config in `dev/.golangci.yaml`
- gofmt for code formatting
- protolint for protobuf files
- shellcheck for shell scripts

### Environment Variables
Key environment variables for database connections:
- `MESSAGE_DB_CONNECTION_STRING`
- `MESSAGE_DB_READER_CONNECTION_STRING`
- `MLS_DB_CONNECTION_STRING`
- `AUTHZ_DB_CONNECTION_STRING`