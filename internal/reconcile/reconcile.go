// Package reconcile refreshes Ticketmaster feeds with a mark-and-sweep, never a
// destructive replace. It upserts fresh facts in place (keeping each event's surrogate id,
// so a claim stays attached), deletes only un-claimed TM events that are now absent from the
// fresh set, and marks claimed-but-gone events `delisted`. Claimed and manual events are
// never deleted by a refresh - guaranteed by the sweep's NOT EXISTS predicate and, as a
// structural backstop, library_entries.event_id ON DELETE RESTRICT.
package reconcile

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/protob/event-sensor/db/sqlc"
	"github.com/protob/event-sensor/internal/provider"
	"github.com/protob/event-sensor/internal/uuid"
)

// Service runs reconciles against the store. It holds the raw *sql.DB (for the transaction +
// the dynamic temp-table sweep/mark statements) alongside the sqlc queries.
type Service struct {
	db *sql.DB
	q  *sqlc.Queries
}

// New constructs a reconcile Service.
func New(db *sql.DB, q *sqlc.Queries) *Service { return &Service{db: db, q: q} }

// Result is the outcome of a reconcile. SavedCount and European are aliases the
// frontend toast reads; the rest are the reconcile counters.
type Result struct {
	TotalFound int  `json:"total_found"`
	Kept       int  `json:"kept"`
	Inserted   int  `json:"inserted"`
	Updated    int  `json:"updated"`
	Deleted    int  `json:"deleted"`
	Delisted   int  `json:"delisted"`
	SavedCount int  `json:"saved_count"`    // = Inserted+Updated
	European   int  `json:"european_count"` // alias of Kept; the frontend toast reads this name
}

// The sweep/mark statements reference a per-transaction TEMP TABLE fresh_ids(source_id) that
// holds the source ids present in this fetch. They are run as raw ExecContext because the
// dynamic fresh-id set is loaded into a temp table rather than a variable-length IN (?), and
// sqlc cannot validate the unknown table at codegen time.
const sweepSQL = `
DELETE FROM events
WHERE source='ticketmaster'
  AND id IN (SELECT event_id FROM performances WHERE artist_id = ?)
  AND NOT EXISTS (SELECT 1 FROM fresh_ids f WHERE f.source_id = events.source_id)
  AND NOT EXISTS (SELECT 1 FROM library_entries le WHERE le.event_id = events.id)
  AND (SELECT COUNT(DISTINCT artist_id) FROM performances x
       WHERE x.event_id = events.id AND x.artist_id IS NOT NULL) = 1`

const markSQL = `
UPDATE events SET listing_state='delisted', updated_at=CURRENT_TIMESTAMP
WHERE source='ticketmaster' AND listing_state='listed'
  AND id IN (SELECT event_id FROM performances WHERE artist_id = ?)
  AND NOT EXISTS (SELECT 1 FROM fresh_ids f WHERE f.source_id = events.source_id)
  AND EXISTS (SELECT 1 FROM library_entries le WHERE le.event_id = events.id)`

// pastSweepSQL removes unclaimed past TM events for an artist regardless of TM feed membership.
// Claimed events are safe: the NOT EXISTS predicate matches the ON DELETE RESTRICT guarantee.
const pastSweepSQL = `
DELETE FROM events
WHERE source='ticketmaster'
  AND id IN (SELECT event_id FROM performances WHERE artist_id = ?)
  AND date(start_date) < date('now')
  AND NOT EXISTS (SELECT 1 FROM library_entries le WHERE le.event_id = events.id)`

// Reconcile refreshes one artist's Ticketmaster feed. fp is the TM adapter built by the
// caller (so a per-user tm.api_key override applies). region is the country-code filter.
//
// A single per-artist fetch always runs regardless of fetch_mode - it's an explicit, on-demand
// request (so you can still check a `manual` artist if you suspect a reunion). "Fetch All" is the
// one that skips `manual` artists, and it does so client-side (it never calls this for them).
// On a TM error nothing is written (the old feed is preserved). An empty result is not an error:
// the sweep clears the un-claimed feed and claimed-but-gone events become `delisted`.
func (s *Service) Reconcile(ctx context.Context, artist sqlc.Artist, region map[string]bool, fp provider.FetchingProvider) (*Result, error) {
	started := time.Now().UTC().Format(time.RFC3339)

	// Fetch first - the only step outside the transaction. On TM error, write a log row
	// and return; touch no event data.
	kept, err := fp.FetchByArtist(ctx, artist.Name, provider.FetchOpts{Region: region})
	if err != nil {
		msg := err.Error()
		s.writeLog(ctx, fetchLog{artistID: artist.ID, source: "ticketmaster", startedAt: started, status: "error", errorMessage: msg})
		return nil, err
	}

	res := &Result{TotalFound: len(kept), Kept: len(kept), European: len(kept)}

	// One transaction from here to commit; _txlock=immediate (DSN) makes this BEGIN IMMEDIATE.
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() //nolint:errcheck // no-op once Commit succeeds
	q := s.q.WithTx(tx)

	// Per-tx temp table of the fresh source ids. Dropped before commit; a rollback also
	// unwinds the CREATE, so a failed reconcile never leaves it behind for the pooled conn.
	if _, err := tx.ExecContext(ctx, `CREATE TEMP TABLE fresh_ids(source_id TEXT)`); err != nil {
		return nil, err
	}
	for _, e := range kept {
		if _, err := tx.ExecContext(ctx, `INSERT INTO fresh_ids(source_id) VALUES (?)`, e.SourceID); err != nil {
			return nil, err
		}
	}

	// Upsert facts in place.
	for _, e := range kept {
		var venueID sql.NullString
		if e.Venue != nil {
			v, err := q.UpsertVenue(ctx, sqlc.UpsertVenueParams{
				ID:          uuid.New(),
				Name:        e.Venue.Name,
				City:        nstr(e.Venue.City),
				Country:     nstr(e.Venue.Country),
				CountryCode: nstr(e.Venue.CountryCode),
				Latitude:    nfloat(e.Venue.Latitude),
				Longitude:   nfloat(e.Venue.Longitude),
				Timezone:    nstr(e.Venue.Timezone),
				Source:      nstr(e.Venue.Source),
				SourceID:    nstr(e.Venue.SourceID),
			})
			if err != nil {
				return nil, err
			}
			venueID = nstr(v.ID)
		}

		// Classify insert vs update for the counters (pre-upsert existence probe).
		_, getErr := q.GetEventBySource(ctx, sqlc.GetEventBySourceParams{Source: "ticketmaster", SourceID: nstr(e.SourceID)})
		isNew := getErr == sql.ErrNoRows

		ev, err := q.UpsertTmEvent(ctx, sqlc.UpsertTmEventParams{
			ID:            uuid.New(),
			SourceID:      nstr(e.SourceID),
			Name:          e.Name,
			Kind:          e.Kind,
			Description:   nstr(e.Description),
			About:         nstr(e.About),
			StartDate:     e.StartDate,
			EndDate:       nstr(e.EndDate),
			HasTime:       b2i(e.HasTime),
			VenueID:       venueID,
			Url:           nstr(e.URL),
			ImageUrl:      nstr(e.ImageURL),
			TicketOptions: jsonArr(string(e.TicketOptions)),
		})
		if err != nil {
			return nil, err
		}
		if isNew {
			res.Inserted++
		} else {
			res.Updated++
		}

		// Performances are feed-owned: replace wholesale. The fetched artist is the linked
		// headliner; the remaining billed attractions become name-only artists.
		if err := q.DeletePerformancesByEvent(ctx, ev.ID); err != nil {
			return nil, err
		}
		if err := q.AddPerformance(ctx, sqlc.AddPerformanceParams{
			ID: uuid.New(), EventID: ev.ID,
			ArtistID: nstr(artist.ID), ArtistName: artist.Name, IsHeadliner: 1,
		}); err != nil {
			return nil, err
		}
		for _, name := range e.Attractions {
			if strings.EqualFold(strings.TrimSpace(name), strings.TrimSpace(artist.Name)) {
				continue // already added as the linked headliner
			}
			if err := q.AddPerformance(ctx, sqlc.AddPerformanceParams{
				ID: uuid.New(), EventID: ev.ID,
				ArtistID: sql.NullString{}, ArtistName: name, IsHeadliner: 0,
			}); err != nil {
				return nil, err
			}
		}
	}

// TODO(design): the claimed-event-survives-refetch invariant is verified by a manual
// sqlite3 recipe. Owner decides whether to add a single Go test for it (the one place
// worth breaking the repo no-tests rule).
	//
	// Sweep: delete un-claimed, now-absent TM events for this artist.
	sweepRes, err := tx.ExecContext(ctx, sweepSQL, artist.ID)
	if err != nil {
		return nil, err
	}
	if n, err := sweepRes.RowsAffected(); err == nil {
		res.Deleted = int(n)
	}

	// Mark claimed-but-gone events delisted (kept, status untouched).
	markRes, err := tx.ExecContext(ctx, markSQL, artist.ID)
	if err != nil {
		return nil, err
	}
	if n, err := markRes.RowsAffected(); err == nil {
		res.Delisted = int(n)
	}

	// Past sweep: delete unclaimed events whose date has passed, regardless of TM feed.
	if _, err := tx.ExecContext(ctx, pastSweepSQL, artist.ID); err != nil {
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, `DROP TABLE fresh_ids`); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	res.SavedCount = res.Inserted + res.Updated

	status := "success"
	if len(kept) == 0 {
		status = "empty"
	}
	s.writeLog(ctx, fetchLog{
		artistID: artist.ID, source: "ticketmaster", startedAt: started,
		finishedAt: time.Now().UTC().Format(time.RFC3339), status: status,
		fetched: res.TotalFound, kept: res.Kept, inserted: res.Inserted,
		updated: res.Updated, deleted: res.Deleted, delisted: res.Delisted,
	})
	return res, nil
}

// fetchLog mirrors the fetch_log row; written outside the main tx so a logged failure
// survives a rollback of the data changes.
type fetchLog struct {
	artistID, source, startedAt, finishedAt, status, errorMessage string
	fetched, kept, inserted, updated, deleted, delisted           int
}

func (s *Service) writeLog(ctx context.Context, l fetchLog) {
	_, _ = s.db.ExecContext(ctx, `
INSERT INTO fetch_log (id, artist_id, source, started_at, finished_at, status, error_message,
                       events_fetched, events_kept, events_inserted, events_updated, events_deleted, events_delisted)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		uuid.New(), nstr(l.artistID), l.source, l.startedAt, nstr(l.finishedAt), l.status, nstr(l.errorMessage),
		l.fetched, l.kept, l.inserted, l.updated, l.deleted, l.delisted)
}

func nstr(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func nfloat(f *float64) sql.NullFloat64 {
	if f == nil {
		return sql.NullFloat64{}
	}
	return sql.NullFloat64{Float64: *f, Valid: true}
}

func b2i(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

// jsonArr returns s, or "[]" when s is blank, so json_valid(ticket_options) always holds.
func jsonArr(s string) string {
	if strings.TrimSpace(s) == "" {
		return "[]"
	}
	return s
}
