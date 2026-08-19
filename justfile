# Run Go backend (with air hot reload if available, otherwise go run)
dev-be:
    which air && air --build.cmd "go build -o ./tmp/main ./cmd/event-sensor" --build.bin ./tmp/main || go run ./cmd/event-sensor

# Run frontend dev server
dev-fe:
    cd frontend && bun run dev

# Run backend and frontend together (dev mode: Vite on :5173, API proxied to :8080)
dev:
    #!/usr/bin/env bash
    set -e
    echo "Dev mode: frontend on http://localhost:5173 (HMR), API proxied to :8080..."
    trap 'kill 0' INT TERM EXIT
    just dev-be &
    just dev-fe &
    wait

# Build frontend for production
build-fe:
    cd frontend && bun install && bun run build

# Run Go tests
test:
    go test ./...

# Run frontend unit tests
test-fe:
    cd frontend && bun test

# Run the browser tests (needs `nix develop` for the Playwright browsers)
e2e:
    cd e2e && bun install && bun x playwright test

# Go, frontend units, browser
test-all: test test-fe e2e

# Everything that can fail before a commit, minus the browser suite
check:
    go build ./...
    go vet ./...
    go test ./...
    cd frontend && bun run type-check
    cd frontend && bun run lint
    cd frontend && bun test

# Generate SQLC code
generate:
    sqlc generate -f db/sqlc.yaml

# Run goose migrations manually
migrate:
    goose -dir db/migrations sqlite3 data/event-sensor.db up

# Full production build: frontend + Go binary
build: build-fe
    CGO_ENABLED=0 go build -tags release -o event-sensor ./cmd/event-sensor

# Build Ticketmaster CLI tool
tm-cli:
    go build -o tm-cli ./cmd/tm-cli

# Remove build artifacts
clean:
    rm -f event-sensor
    rm -f tm-cli
    rm -rf data
    rm -rf internal/spa/dist
    rm -rf frontend/node_modules
