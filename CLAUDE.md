# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Development Commands

### Backend (Go) Development

```bash
# Run linting and formatters
make lint
make fmt

# Run all tests with summary
make test

# Run tests with race detector
make test-race

# Run specific tests
make test GOTEST_FLAGS="-run=TestSpecificName"

# Run tests with coverage
make test GOTEST_FLAGS="-cover"

# Generate code assets
make gen-goa      # Generate Goa API assets
make gen-ent      # Generate Ent ORM assets
make gen-enums    # Generate enum types
make gen-mock     # Generate mocks for testing

# Database operations
make db           # Open MySQL shell for enduro database
make atlas-hash   # Recalculate migration hashes

# Check dependencies
make deps         # List available module dependency updates

# Build binaries (for testing compilation)
go build -o dist/enduro ./cmd/enduro
go build -o dist/enduro-am-worker ./cmd/enduro-am-worker
go build -o dist/enduro-a3m-worker ./cmd/enduro-a3m-worker
```

### Frontend (Vue.js) Development

```bash
cd dashboard

# Install dependencies
npm install

# Run development server with hot reload
npm run dev

# Build for production
npm run build

# Run linting
npm run lint

# Format code
npm run format

# Run tests
npm run test

# Run tests with coverage
npm run coverage

# Type checking
npm run type-check
```

### Local Development Environment

```bash
# Start local Kubernetes cluster (k3d recommended)
k3d cluster create sdps-local --registry-create sdps-registry

# Start development environment
tilt up

# Upload a sample transfer
make upload-sample-transfer
```

## High-Level Architecture

Enduro is a preservation workflow application that orchestrates digital preservation activities using Temporal workflows. It integrates with preservation systems like Archivematica and a3m.

### Core Components

1. **API Server** (`cmd/enduro/main.go`)
   - REST API using Goa framework
   - Two instances: public API (with auth) and internal API
   - Handles SIP ingestion, storage operations, and monitoring

2. **Temporal Workflows** (`internal/workflow/`)
   - `ProcessingWorkflow`: Main workflow for SIP processing
   - Handles preservation activities (validation, packaging, storage)
   - Activities for file operations, transformations, and integrations

3. **Storage System** (`internal/storage/`)
   - Manages AIP storage locations and operations
   - Supports multiple storage backends (S3, Azure, filesystem)
   - Handles storage workflows (upload, move, delete)

4. **Ingest Service** (`internal/ingest/`)
   - Manages SIP submission and tracking
   - WebSocket support for real-time updates
   - File upload and validation

5. **Watchers** (`internal/watcher/`)
   - Monitor filesystem and object storage for new SIPs
   - Support for MinIO bucket notifications
   - Configurable polling and event-based watching

6. **Persistence Layer**
   - Uses Ent ORM with MySQL
   - Separate databases for main app and storage module
   - Migrations managed with Atlas

7. **Frontend Dashboard** (`dashboard/`)
   - Vue 3 with TypeScript
   - Real-time updates via WebSocket
   - OpenAPI client generated from backend spec

### Key Integrations

- **Temporal**: Workflow orchestration engine
- **Archivematica/a3m**: Preservation processing backends
- **MinIO/S3**: Object storage for AIPs
- **Keycloak**: SSO and authentication (optional)
- **OpenTelemetry**: Distributed tracing

### Workflow Types

- **Standard**: Full preservation workflow with configurable steps
- **Legacy**: Compatibility mode for existing systems
- **Preprocessing**: Validation and preparation activities

### Data Flow

1. SIP submission (upload, watcher detection, or API)
2. Temporal workflow initialization
3. Preprocessing (validation, unpacking)
4. Preservation processing (via AM/a3m)
5. AIP storage and registration
6. Post-storage activities (cleanup, notifications)

## Key Development Patterns

- Use generated code for API, database models, and enums
- Temporal activities should be idempotent
- Storage operations use location abstraction
- Real-time updates via event service (Redis pub/sub)
- Authentication optional but recommended for production