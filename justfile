# ClipIN Developer Justfile

export PATH := env_var('HOME') + '/.bun/bin:' + env_var('PATH')
export NUXT_TELEMETRY_DISABLED := "1"

# Default target: show available commands
default:
    @just --list

# Start database services (PostgreSQL & Redis)
db-up:
    docker compose up -d

# Stop database services
db-down:
    docker compose down

# View database service logs
db-logs:
    docker compose logs -f

# Run database migrations (action: up, down, status, goto)
migrate action='up':
    cd apps/api && go run ./cmd/migrate/ {{action}}

# Run frontend dev server
dev-frontend:
    @echo "Starting Frontend on http://localhost:3000..."
    cd apps/web && bun run dev

# Run backend dev services (PostgreSQL, Redis, Go API, Go Verifier)
dev-backend: db-up
    @echo "Starting Backend API on http://localhost:8080..."
    @echo "Starting Verifier on http://localhost:8081..."
    (cd apps/api && go run ./cmd/api) & \
    (cd services/verifier && go run ./cmd/verifier) & \
    wait

# Run all dev servers (databases, backend services, frontend)
dev: db-up
    @echo "Starting full dev environment..."
    @echo "Starting API on http://localhost:8080"
    @echo "Starting Verifier on http://localhost:8081"
    @echo "Starting Frontend on http://localhost:3000"
    (cd apps/api && go run ./cmd/api) & \
    (cd services/verifier && go run ./cmd/verifier) & \
    (cd apps/web && bun run dev) & \
    wait

# Shortcut for Nuxt frontend dev server
web:
    cd apps/web && bun run dev

# Shortcut for Go backend API server
api:
    cd apps/api && go run ./cmd/api

# Shortcut for Go verifier service server
verifier:
    cd services/verifier && go run ./cmd/verifier

# Run tests across all projects
test:
    @echo "==> Testing Go API..."
    cd apps/api && go test ./...
    @echo "==> Testing Go Verifier..."
    cd services/verifier && go test ./...
    @echo "==> Testing Web Frontend..."
    cd apps/web && bun run test

# Run linters across all projects
lint:
    @echo "==> Linting Go API..."
    cd apps/api && go vet ./...
    @echo "==> Linting Go Verifier..."
    cd services/verifier && go vet ./...
    @echo "==> Linting Web Frontend..."
    cd apps/web && bun run lint

# Format code across all projects
format:
    @echo "==> Formatting Go API..."
    cd apps/api && go fmt ./...
    @echo "==> Formatting Go Verifier..."
    cd services/verifier && go fmt ./...
    @echo "==> Formatting Web Frontend..."
    cd apps/web && bun run format

# Build production artifacts for all projects
build:
    @echo "==> Building Go API..."
    cd apps/api && go build -o bin/api ./cmd/api
    @echo "==> Building Go Verifier..."
    cd services/verifier && go build -o bin/verifier ./cmd/verifier
    @echo "==> Building Web Frontend..."
    cd apps/web && bun run build

# Clean build artifacts
clean:
    rm -rf apps/api/bin services/verifier/bin apps/web/.output apps/web/.nuxt
