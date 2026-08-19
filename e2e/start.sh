#!/usr/bin/env bash
# Playwright webServer command: build the release binary and run it against a fresh
# temporary database, with the Ticketmaster stub in place of the real API.
set -euo pipefail

E2E="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$E2E/.." && pwd)"

# One directory per port instead of a fresh mktemp: Playwright kills this script hard
# enough that a cleanup trap cannot be relied on, and a run should not leave a binary and
# a database behind every time. Wiping it here is also what makes each run start empty.
WORK="/tmp/es-e2e-${ES_E2E_PORT:-8099}"
rm -rf "$WORK"
mkdir -p "$WORK"

STUB=""
trap '[ -n "$STUB" ] && kill "$STUB" 2>/dev/null || true' EXIT INT TERM

(cd "$E2E" && bun install --silent) >&2
(cd "$ROOT/frontend" && bun install --silent && bun run build) >&2
(cd "$ROOT" && CGO_ENABLED=0 go build -tags release -o "$WORK/event-sensor" ./cmd/event-sensor) >&2
(cd "$ROOT" && go build -o "$WORK/tmstub" ./e2e/tmstub) >&2

TM_PORT="${ES_E2E_TM_PORT:-8098}"
"$WORK/tmstub" -addr "127.0.0.1:$TM_PORT" -dir "$ROOT/ticketmaster/testdata" &
STUB=$!

export ES_BIND=127.0.0.1
export PORT="${ES_E2E_PORT:-8099}"
export DB_PATH="$WORK/e2e.db"
export JWT_SECRET="e2e-secret"
export TICKETMASTER_API_KEY="e2e-key"
export TM_BASE_URL="http://127.0.0.1:$TM_PORT"

# Foreground, not exec: the trap has to outlive the build so an interrupted run leaves
# the stub port free.
"$WORK/event-sensor"
