# Event Sensor reference

Everything beyond the overview in the [README](../README.md): data model, API, the
Ticketmaster fetch pipeline, the UI pages, and development.

## Contents

- [Data model](#data-model)
- [Startup and bind address](#startup-and-bind-address)
- [API](#api)
- [Ticketmaster pipeline](#ticketmaster-pipeline)
- [UI pages](#ui-pages)
- [Development](#development)

## Data model

SQLite, WAL mode, STRICT tables. `foreign_keys` is enforced via a DSN pragma (it is
per-connection in SQLite); a single writer connection (`SetMaxOpenConns(1)`) serializes
writes so `ON DELETE RESTRICT` always fires. Migrations are embedded and run on startup.

An event has three states that vary independently:

- what the world says: `events.listing_state` - `listed` / `delisted` / `cancelled`
- when it happens: derived from `start_date < now()`, never stored
- what you say: `library_entries.status` - `interested` / `going` / `attended` / `missed`

Tables: `users`, `artists`, `categories`, `artist_categories`, `events`, `venues`,
`performances`, `library_entries`, `fetch_log`, `user_settings`.

How things are meant to work:

- **A claim makes an event permanent.** A `library_entries` row (`event_id REFERENCES
  events(id) ON DELETE RESTRICT`) pins the event. Un-claimed Ticketmaster events are
  ephemeral and may be swept. Manual events are never swept either.
- **A festival is just an event** with `kind='festival'` and its acts as `performances`.
  No separate entity, no festival-to-event link.
- **Performances link what you track.** `artist_id` set = an artist in your catalog;
  NULL = a name-only string. You log only the acts you care about. There is no
  `(event, artist)` unique key - an artist can play twice at one festival on different
  days.
- **Fetch modes.** `fetch_mode='auto'`: included in Fetch All. `'manual'`: fetched only
  when you ask (disbanded or unlikely-to-tour acts you keep for history). Manual
  artists are hidden from the discovery tree. An auto artist with no upcoming listed
  events shows as dormant.
- **Names are stored as typed.** - verbatim,
  display included.

## Startup and bind address

Startup behavior depends on the bind address. On loopback (development) everything stays
open. On a non-loopback bind the server:

- refuses to start while `JWT_SECRET` is the public default
- warns while any account still uses the seeded password
- requires authentication for the reads that are public in development
- refuses password reset until the seeded password is changed, since username and email
  are both public in this repository

## API

OpenAPI spec at `/openapi.json` when running.

**Public**

- `POST /api/auth/login` - login
- `POST /api/auth/reset-password` - reset by username and email
- `GET  /api/artists` - list artists
- `GET  /api/artists/{id}/events` - artist event list
- `GET  /api/events` - list events (`start_date`, `end_date`, `q` params)
- `GET  /api/venues` - list venues with event counts

**Protected (JWT)**

- `GET  /api/auth/me` - current user
- `GET  /api/artists/summary` - per-artist categories, countries and counts in one call
- `POST /api/artists` - create artist (`fetch_mode`: `auto` | `manual`)
- `PUT  /api/artists/{id}` - rename
- `DELETE /api/artists/{id}` - delete (drops its un-claimed events; claims survive)
- `POST /api/artists/{id}/fetch-events` - fetch + reconcile from Ticketmaster
- `PUT  /api/artists/{id}/fetch-mode` - set `fetch_mode`
- `DELETE /api/artists/{id}/events` - clear an artist's un-claimed events
- `POST /api/categories` - create category
- `PUT  /api/categories/{id}` - rename
- `DELETE /api/categories/{id}` - delete
- `POST /api/events` - create manual event (optional artist links + initial claim)
- `PUT  /api/events/{id}` - edit a manual event (409 for Ticketmaster events)
- `DELETE /api/events/{id}` - delete event (409 if claimed)
- `DELETE /api/events` - clear all un-claimed events
- `DELETE /api/events/past` - delete past un-claimed Ticketmaster events
- `PUT  /api/events/{id}/status` - set or clear the per-user claim
- `GET  /api/library` - the user's claims (diary index)
- `GET  /api/library/needs-resolution` - past `going` events awaiting a "did you go?"
- `GET/PUT /api/settings` - read/update settings (`notif.*`, `region.codes`, `tm.api_key`)

**Bulk** (each in a single `BEGIN IMMEDIATE` transaction)

- `POST /api/artists/merge` - repoint performances + memberships to the winner, delete losers
- `PUT  /api/artists/fetch-mode` - bulk set fetch mode
- `POST /api/venues/merge` - repoint events to the winner venue, delete losers
- `POST /api/events/{id}/artists/move` - re-parent performances to another event
- `POST /api/categories/assign` - bulk add/remove artists to/from categories
- `PUT  /api/library/status` - bulk set/clear claims
- `POST /api/events/bulk-delete` - delete many events by ID (claimed kept)

## Ticketmaster pipeline

`POST /api/artists/{id}/fetch-events` runs, per raw TM event:

1. **Junk filter** - drops parking, VIP/hotel packages, add-ons by name.
2. **Region filter** - keeps only events in the user's `region.codes` setting (default:
   Europe + Turkey). Out-of-region events are never stored.
3. **Artist match** - matches the canonical attraction name, not the event title, so it
   survives long and localized titles. Not billed but the title has tribute markers
   (`tribute`, `music of/by`, `performed by`, ...)? Kept as `kind='tribute'`. Otherwise
   rejected as keyword coincidence.
4. **Dedup** - by venue + day: a festival appearance sold as several products (day
   ticket, terrace, combi) folds into one row; the alternatives survive as
   `ticket_options`.
5. **Festival detection** - TM's subType flag, an explicit "fest" word in the name, or a
   curated name list (empty by default).
6. **Date** - TM's full `dateTime` when announced, else the bare `localDate`.

### Reconcile (mark-and-sweep)

A fetch is never a wipe. Inside one transaction:

1. **Upsert** fresh TM facts in place on `(source, source_id)` - the event keeps its
   surrogate id, so claims stay attached.
2. **Sweep** - delete this artist's un-claimed TM events absent from the fresh set.
3. **Mark** - claimed-but-gone events become `delisted` (kept, status untouched).
4. **Past sweep** - delete un-claimed TM events whose date has passed, regardless of
   feed membership.

Claimed and manual events are never deleted by a reconcile - enforced both by the sweep's
`NOT EXISTS (library_entries)` predicate and, structurally, by the `ON DELETE RESTRICT`
backstop. Each reconcile writes a `fetch_log` row.

## UI pages

- **/events - Radar.** Future Ticketmaster events only (a "Show past" toggle reveals
  history). List/cards views, time filters, country multi-select (clicking any flag
  toggles it), grouping by artist/country, multi-select bulk actions. Left pane: the
  category tree with per-artist flags, upcoming counts, dormant markers and per-artist
  fetch.
- **/library - Diary.** Claims across all time in four status tabs. List/grid views,
  sticky month dividers, inline status/missed-reason/note editing, a "did you go?"
  nudge for past `going` events, and a by-artist lens.
- **Event detail.** Meta panel + lineup grid. Ticketmaster lineups are marked "may be
  incomplete"; manual lineups are your own data and get no caveat.
- **/artists.** Dense catalog: inline rename, fetch-mode flip, per-artist event lists,
  "log past show" pre-filled.
- **/venues.** Every venue with its event count; same-name+city duplicates are badged
  and sorted first; multi-select merge picks a survivor and repoints its events.
- **Manual event form.** Festival mode is a two-panel lineup editor (dense list +
  focused-act detail) with CSV paste, day/venue fill-down over a multi-selection, and
  auto-linking of names to the catalog on save.
- **/settings.** Profile, password, per-user TM API key, notification toggles, region
  picker, past-TM-event purge.

Keyboard: `Cmd/Ctrl+A` selects, `Esc` clears/closes, `n` opens the quick-add dialog,
flags click into the country filter.

## Development

Build orchestration lives in the [justfile](../justfile); `just` with no argument lists
every recipe. The commonly used ones:

```sh
just dev         # both, one Ctrl-C (Vite + HMR on :5173, API proxied to :8080)
just dev-be      # Go backend only, air hot reload (or go run)
just dev-fe      # Vite dev server only (bun)
just build       # frontend + binary -> ./event-sensor
just generate    # regenerate sqlc code after editing db/queries/*.sql
just migrate     # run goose migrations manually
just tm-cli      # build the Ticketmaster debug CLI
```

Backend layout:

```
cmd/event-sensor/ the binary: config, router assembly, graceful shutdown
cmd/tm-cli/     Ticketmaster API debug CLI
api/            handlers, routes, middleware, Huma types
db/migrations/  goose migrations
db/queries/     sqlc query definitions
db/sqlc/        generated Go code
internal/       auth, config, reconcile, provider seam, uuid
internal/spa/   SPA embed + fallback handler; Vite writes its dist here
ticketmaster/   TM client + parser (filters, matching, dedup)
frontend/       Vue 3 SPA (see frontend/README.md)
tests/          Go integration tests
e2e/            Playwright specs + the Ticketmaster stub server
```

`tm-cli` is a small helper for looking at raw TM API responses - handy when a fetch
keeps or drops something unexpected, and useful for an AI agent inspecting the API
structure on the fly:

```sh
./tm-cli analyze "Artist Name"          # junk/duplicate analysis
./tm-cli search "Artist Name" -p -d     # -p: no parking, -d: no duplicates
./tm-cli search "Artist Name" -j        # raw JSON
```

### Tests

| Layer | Location | Command |
| --- | --- | --- |
| Go integration | `tests/integration/` | `just test` |
| Ticketmaster pipeline | `ticketmaster/` | `go test ./ticketmaster/` |
| Frontend units | `frontend/src/__tests__/` | `just test-fe` |
| Browser | `e2e/` | `just e2e` (inside `nix develop`) |

The integration tests run the real router against a real SQLite file in a temporary
directory; there are no mocks. The Ticketmaster API is replaced at one point only: the
client's base URL, set through `TM_BASE_URL`, pointed at a server serving recorded
responses from `ticketmaster/testdata/`.

Fixtures are recorded with `./tm-cli search "Artist" -j`, with the API key removed and
absolute dates replaced by tokens (`{{+30d}}`, `{{-10d}}`, `{{+45dT}}` for a value with
a time part) so a fixture never expires.

The browser suite drives the release binary with the SPA embedded, on a temporary
database, with `e2e/tmstub` in place of Ticketmaster. Locators use `data-testid` only.
