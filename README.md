# Event Sensor

A single-user concert tracker. It pulls upcoming shows for your artists from the
Ticketmaster Discovery API, lets you claim events (interested / going / attended /
missed), and keeps a log of past shows, including manually entered ones. Go + SQLite
backend, Vue 3 frontend embedded in the same binary.

- Future events are fetched from the Ticketmaster API and filtered before they are stored
- A claim is durable: claimed events survive refetches and sweeps
- Past shows and festivals can be entered by hand, with a per-act lineup editor
- One binary: frontend embedded, migrations auto-run, SQLite in a single file

## Quick start

`just`, `go` and `bun` are the prerequisites.

On NixOS, `nix develop` supplies the Go and SQL tooling and Playwright;

```sh
cp .env.example .env
just dev           # dev mode: Vite + HMR on :5173, API proxied to :8080
```

The default login is `admin` / `password`, seeded by a migration. Change it if the
instance is reachable beyond localhost. There is no self-registration.

Production build:

```sh
just build && ./event-sensor
```

The binary serves the SPA at `/`, with a fallback to `index.html` for client-side
routing.

Tests: `just test` (Go), `just test-fe` (frontend units), `just e2e` (browser, needs
`nix develop`), or `just test-all` for all three.

## Configuration

Configuration is read from environment variables. The Ticketmaster key is the
exception: the key from agenix is read first, and `TICKETMASTER_API_KEY`
is the fallback.

| Variable               | Default                           | Meaning                                                  |
| ---------------------- | --------------------------------- | -------------------------------------------------------- |
| `ES_BIND`              | `127.0.0.1`                       | Listen address; put Caddy in front for TLS               |
| `PORT`                 | `8080`                            | Listen port                                              |
| `DB_PATH`              | `data/event-sensor.db`            | SQLite path; parent dir created on start (0700)          |
| `TICKETMASTER_API_KEY` | (empty)                           | Server-wide TM key                                       |
| `TM_BASE_URL`          | (empty)                           | Ticketmaster endpoint override; empty means the real API |
| `JWT_SECRET`           | `dev-secret-change-in-production` | JWT signing secret                                       |

## Deployment

A single static binary (`CGO_ENABLED=0`, pure-Go SQLite), for x86_64 and aarch64 Linux.
An example unit file and secrets-file template are in
[deploy/nixos-example/](deploy/nixos-example/):

```sh
just build
install -Dm755 event-sensor /opt/event-sensor/event-sensor-bin
# JWT_SECRET from agenix (or: openssl rand -hex 32 in an EnvironmentFile)
systemctl start event-sensor   # binds 127.0.0.1; put a reverse proxy in front
```

Password reset on the host:

```sh
sqlite3 <DB_PATH> "UPDATE users SET password='$(htpasswd -bnBC 10 "" 'newpass' | tr -d ':\n')' WHERE username='admin';"
```

## Documentation

[docs/reference.md](docs/reference.md): data model, API endpoints, the Ticketmaster
fetch pipeline, the UI pages, and development.
