# Event Sensor

Single-user concert tracker: pulls upcoming shows for your artists from the Ticketmaster
Discovery API, lets you claim events, and keeps a permanent diary of past
shows including hand-entered ones. Go + SQLite backend, Vue 3 SPA embedded in the same
binary, one static file to deploy.

Module path: `github.com/protob/event-sensor`.

## Commands

Everything goes through the [justfile](justfile) (`just` with no argument lists recipes).

`nix develop` enters a shell with go, gopls, sqlc, goose, air, just, sqlite and
Playwright; bun and node are expected on the host and are never packaged here.
`PLAYWRIGHT_BROWSERS_PATH` is set in the shell, so browsers are never downloaded.

```sh
just dev         # backend + Vite (HMR on :5173, /api proxied to :8080), one Ctrl-C
just dev-be      # backend only (air if installed, else go run)
just dev-fe      # Vite only
just build-fe    # bun install + vite build -> frontend/dist
just build       # build-fe, then CGO_ENABLED=0 go build -o event-sensor .
just generate    # sqlc generate, after editing db/queries/*.sql
just migrate     # goose up against data/event-sensor.db
just tm-cli      # build the Ticketmaster debug CLI
```

Checks: `go build ./...`, `go test ./...`, and in `frontend/`: `bun run type-check`,
`bun run lint`, `bun run format`.

The SPA embed is behind the `release` build tag (`embed_dev.go` / `embed_release.go`),
so a clean tree builds and tests without a frontend. `just build` passes `-tags release`
and embeds `frontend/dist`; a binary built without the tag serves a notice page at `/`.

## Layout

```
main.go          startup: config, router assembly, SPA fallback, graceful shutdown
api/             Huma handlers, routes, middleware, wire types
cmd/tm-cli/      Ticketmaster API debug CLI
db/db.go         Open (pragma DSN, single writer) + Migrate (embedded goose migrations)
db/migrations/   goose migrations (embedded, run on startup)
db/queries/      sqlc query definitions — the hand-written SQL
db/sqlc/         generated Go — never edit by hand
internal/        auth, config, reconcile, provider seam, uuid
ticketmaster/    TM client + parser (junk/region filters, artist match, dedup)
frontend/        Vue 3 SPA (see frontend/README.md)
docs/reference.md  data model, full endpoint list, pipeline, UI pages
```

## Invariants

A change that breaks one of these is a bug, not a refactor.

- **A claim makes an event permanent.** `library_entries.event_id REFERENCES events(id)
  ON DELETE RESTRICT`. Reconcile sweeps only un-claimed Ticketmaster events; claimed and
  manual events are never deleted by a fetch. Both the sweep's `NOT EXISTS
  (library_entries)` predicate and the FK backstop must stay in place.
- **Three independent event states.** `events.listing_state` (`listed`/`delisted`/
  `cancelled`) is what the world says; past/future is derived from `start_date < now()`
  and never stored; `library_entries.status` (`interested`/`going`/`attended`/`missed`)
  is what the user says. Do not collapse them.
- **Fetch is mark-and-sweep, never a wipe.** Upsert on `(source, source_id)` so the
  surrogate id — and therefore the claim — survives; sweep un-claimed absentees; mark
  claimed-but-gone as `delisted`; drop past un-claimed TM events. All in one transaction,
  with a `fetch_log` row per run.
- **Storage is source-agnostic.** `internal/provider.ProviderEvent` is the seam: source
  specific filtering happens before it, and storage never switches on `source`. New
  sources implement `FetchingProvider` or `IntakeProvider` and need no storage change.
- **A festival is just an event** with `kind='festival'` and its acts as `performances`.
  There is no `(event, artist)` unique key — an act can play twice at one festival.
- **Names are stored verbatim.** No normalization or case folding on artist, venue or
  event names.

## Backend conventions

- **SQLite via modernc (pure Go, `CGO_ENABLED=0`).** Pragmas live in the DSN in
  `db/db.go` (`foreign_keys`, `busy_timeout`, WAL, `_txlock=immediate`) because they are
  per-connection; a `db.Exec("PRAGMA ...")` would only reach one pooled connection.
  `SetMaxOpenConns(1)` — single writer. Tables are STRICT.
- **sqlc**: add or edit SQL in `db/queries/*.sql` with a `-- name: X :one|:many|:exec`
  header, then `just generate`. Schema changes need a new numbered goose migration
  (`-- +goose Up` / `-- +goose Down`); migrations are embedded and run at startup, so
  they must be safe to re-apply against an existing database.
- **Handlers** are methods on `api.Handler`, which holds both `*sql.DB` and
  `*sqlc.Queries`.
  Multi-statement atomic work opens a transaction and uses `queries.WithTx(tx)` —
  every bulk endpoint does this in a single `BEGIN IMMEDIATE`.
- **Routes** are registered with `huma.Register` in `api/router.go` with an
  `OperationID`, `Summary` and `Tags`; the OpenAPI spec at `/openapi.json` follows from
  it. Errors go through the `huma404` / `huma400` / `huma409` / `huma403` helpers in
  `api/errors.go`; use `isForeignKeyErr` to map a RESTRICT violation to a 409.
- **Auth**: JWT bearer, `AuthMiddleware` puts the user id in the request context. The
  account is seeded by migration 003. There is no registration endpoint: this is
  single-user software, and a second user runs a second instance. Ids are
  `internal/uuid.New()`.
- **Bind-dependent posture**: loopback keeps read endpoints public (browse before login);
  a non-loopback bind moves them behind auth, refuses to start on the default
  `JWT_SECRET`, warns about seeded passwords, and blocks password reset until the seeded
  password is changed. New public endpoints must follow the same rule.

## Frontend conventions

- Vue 3 `<script setup>` + TypeScript, Tailwind CSS 4, Pinia, Vue Router, VueUse,
  ofetch, unplugin-icons (MDI + circle-flags). `bun` is the package manager.
- **Layering**: `api/` (one ofetch module per resource) is imported only by `stores/`.
  Views and components read stores and never make requests directly. All server state,
  mutation and cross-store resync lives in Pinia stores.
- **Request errors**: everything in `api/` rejects with an `ApiRequestError` - a real
  `Error` (so the stack survives) carrying the server's problem-detail fields. Build it
  with the `apiRequestError` factory, narrow it with `isApiRequestError`, display it
  through `errMessage`. Requests are never retried: 409 is a considered answer from this
  API, not a transient fault, and the write endpoints are not all idempotent.
- `utils/` holds pure functions (no `use*` naming); `composables/` own reactivity or
  lifecycle. Both re-export through an `index.ts`.
- Colors go through semantic tokens (`bg-surface`, `text-muted`, ...) and sizing through
  the fixed density scale in `assets/main.css` — no ad-hoc hex values or one-off spacing.
- The UI is dense and keyboard-first: Shift-click ranges, `Cmd/Ctrl+A` select,
  `Esc` clear/close, `n` quick-add. New list views must support the same bindings.
- `types/` mirrors the wire format; keep it in step with the Huma request/response types
  in `api/types.go`.

## Debugging a fetch

`./tm-cli` (build with `just tm-cli`) inspects raw Ticketmaster responses, which is how
to determine why the pipeline kept or dropped an event:

```sh
./tm-cli analyze "Artist Name"        # junk/duplicate analysis
./tm-cli search "Artist Name" -p -d   # -p: drop parking, -d: drop duplicates
./tm-cli search "Artist Name" -j      # raw JSON
```

Parser behavior is covered by `ticketmaster/parser_test.go`; filter and matching changes
belong there.

## Config

Environment only, loaded in `internal/config`: `ES_BIND` (default `127.0.0.1`), `PORT`
(`8080`), `DB_PATH` (`data/event-sensor.db`), `TICKETMASTER_API_KEY`, `TM_BASE_URL`
(empty: the real Discovery endpoint; tests point it at a local server), `JWT_SECRET`.
`/run/agenix/ticketmaster-api-key` is read first, and the env key is the fallback. A
per-user TM key set in Settings overrides the server-wide one at fetch time. `data/` and
`.env` are gitignored runtime state; never commit a database or a real key.
