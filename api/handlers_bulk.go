package api

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/protob/event-sensor/db/sqlc"
)

// inList builds a "(?,?,?)" clause and the matching []any args for a dynamic IN. Empty → "(NULL)".
func inList(ids []string) (string, []any) {
	if len(ids) == 0 {
		return "(NULL)", nil
	}
	ph := strings.TrimRight(strings.Repeat("?,", len(ids)), ",")
	args := make([]any, len(ids))
	for i, v := range ids {
		args[i] = v
	}
	return "(" + ph + ")", args
}

// dedupeExclude returns ids with duplicates and `exclude` removed (used to derive merge "losers").
func dedupeExclude(ids []string, exclude string) []string {
	seen := map[string]bool{exclude: true}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	return out
}

// MergeArtists merges one or more "loser" artists into a "winner": repoints every reference
// (category memberships, performances), then deletes the losers. One BEGIN IMMEDIATE tx;
// ordering (dedupe → repoint → delete) keeps every intermediate state FK-valid.
func (h *Handler) MergeArtists(ctx context.Context, in *MergeArtistsInput) (*MergeArtistsOutput, error) {
	into := in.Body.Into
	losers := dedupeExclude(in.Body.From, into)
	if into == "" || len(losers) == 0 {
		return nil, huma400("`into` and at least one distinct `from` artist are required")
	}
	if _, err := h.queries.GetArtist(ctx, into); err != nil { // don't repoint into a ghost
		return nil, huma404("target artist not found")
	}

	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	in1, a1 := inList(losers)

	// Drop loser category memberships that would collide with the winner's (composite PK) -
	// repointing them would violate it.
	if _, err := tx.ExecContext(ctx,
		"DELETE FROM artist_categories WHERE artist_id IN "+in1+
			" AND category_id IN (SELECT category_id FROM artist_categories WHERE artist_id = ?)",
		append(append([]any{}, a1...), into)...); err != nil {
		return nil, fmt.Errorf("dedupe memberships: %w", err)
	}
	if _, err := tx.ExecContext(ctx,
		"UPDATE artist_categories SET artist_id = ? WHERE artist_id IN "+in1,
		append([]any{into}, a1...)...); err != nil {
		return nil, fmt.Errorf("repoint memberships: %w", err)
	}
	if _, err := tx.ExecContext(ctx,
		"UPDATE performances SET artist_id = ? WHERE artist_id IN "+in1,
		append([]any{into}, a1...)...); err != nil {
		return nil, fmt.Errorf("repoint performances: %w", err)
	}
	// Delete losers; their fetch_log rows cascade-delete (observability only, acceptable).
	res, err := tx.ExecContext(ctx, "DELETE FROM artists WHERE id IN "+in1, a1...)
	if err != nil {
		return nil, fmt.Errorf("delete losers: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	n, _ := res.RowsAffected()
	out := &MergeArtistsOutput{}
	out.Body.Into = into
	out.Body.Merged = int(n)
	return out, nil
}

// MergeVenues dedupes venues: repoints events from the losers to the winner, then deletes
// the loser venues. One BEGIN IMMEDIATE tx.
func (h *Handler) MergeVenues(ctx context.Context, in *MergeVenuesInput) (*MergeVenuesOutput, error) {
	into := in.Body.Into
	losers := dedupeExclude(in.Body.From, into)
	if into == "" || len(losers) == 0 {
		return nil, huma400("`into` and at least one distinct `from` venue are required")
	}
	if _, err := h.queries.GetVenue(ctx, into); err != nil {
		return nil, huma404("target venue not found")
	}

	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	in1, a1 := inList(losers)

	repointed, err := tx.ExecContext(ctx,
		"UPDATE events SET venue_id = ? WHERE venue_id IN "+in1,
		append([]any{into}, a1...)...)
	if err != nil {
		return nil, fmt.Errorf("repoint events: %w", err)
	}
	deleted, err := tx.ExecContext(ctx, "DELETE FROM venues WHERE id IN "+in1, a1...)
	if err != nil {
		return nil, fmt.Errorf("delete loser venues: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	rep, _ := repointed.RowsAffected()
	del, _ := deleted.RowsAffected()
	out := &MergeVenuesOutput{}
	out.Body.Into = into
	out.Body.RepointedEvents = int(rep)
	out.Body.Merged = int(del)
	return out, nil
}

// MoveArtists re-parents performances from a source event to a target event. The UPDATE is
// scoped to rows that really belong to the source event, so stray IDs are ignored.
func (h *Handler) MoveArtists(ctx context.Context, in *MoveArtistsInput) (*MoveArtistsOutput, error) {
	if in.Body.TargetEventID == "" {
		return nil, huma400("target_event_id is required")
	}
	if _, err := h.queries.GetEvent(ctx, in.Body.TargetEventID); err != nil {
		if err == sql.ErrNoRows {
			return nil, huma404("target event not found")
		}
		return nil, fmt.Errorf("get target event: %w", err)
	}

	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	in1, a1 := inList(in.Body.PerformanceIDs)
	res, err := tx.ExecContext(ctx,
		"UPDATE performances SET event_id = ? WHERE id IN "+in1+" AND event_id = ?",
		append(append([]any{in.Body.TargetEventID}, a1...), in.ID)...)
	if err != nil {
		return nil, fmt.Errorf("move artists: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	n, _ := res.RowsAffected()
	out := &MoveArtistsOutput{}
	out.Body.Moved = int(n)
	return out, nil
}

// AssignCategories bulk add/removes artist↔category memberships. Adds use INSERT OR
// IGNORE - the composite (artist_id, category_id) PK is the dedupe.
func (h *Handler) AssignCategories(ctx context.Context, in *AssignCategoriesInput) (*AssignCategoriesOutput, error) {
	if len(in.Body.ArtistIDs) == 0 {
		return nil, huma400("artist_ids is required")
	}

	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	for _, aid := range in.Body.ArtistIDs {
		for _, cid := range in.Body.Add {
			if _, err := tx.ExecContext(ctx,
				"INSERT OR IGNORE INTO artist_categories (artist_id, category_id) VALUES (?, ?)",
				aid, cid); err != nil {
				return nil, fmt.Errorf("assign: %w", err)
			}
		}
	}
	if len(in.Body.Remove) > 0 {
		aIn, aArgs := inList(in.Body.ArtistIDs)
		cIn, cArgs := inList(in.Body.Remove)
		if _, err := tx.ExecContext(ctx,
			"DELETE FROM artist_categories WHERE artist_id IN "+aIn+" AND category_id IN "+cIn,
			append(aArgs, cArgs...)...); err != nil {
			return nil, fmt.Errorf("remove memberships: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	out := &AssignCategoriesOutput{}
	out.Body.Artists = len(in.Body.ArtistIDs)
	return out, nil
}

// BulkSetStatus sets or clears the per-user library claim for many events at once. An empty
// status unclaims (deletes the entries); otherwise it upserts each claim.
func (h *Handler) BulkSetStatus(ctx context.Context, in *BulkStatusInput) (*BulkStatusOutput, error) {
	userID := UserIDFromContext(ctx)
	if userID == "" {
		return nil, huma400("user not authenticated")
	}
	if len(in.Body.EventIDs) == 0 {
		return nil, huma400("event_ids is required")
	}

	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck
	q := h.queries.WithTx(tx)

	updated := 0
	if in.Body.Status == "" {
		in1, a1 := inList(in.Body.EventIDs)
		res, err := tx.ExecContext(ctx,
			"DELETE FROM library_entries WHERE user_id = ? AND event_id IN "+in1,
			append([]any{userID}, a1...)...)
		if err != nil {
			return nil, fmt.Errorf("bulk unclaim: %w", err)
		}
		n, _ := res.RowsAffected()
		updated = int(n)
	} else {
		// missed_reason only valid for status='missed' (DB CHECK enforces this too).
		missed := toNullStr(in.Body.MissedReason)
		if in.Body.Status != "missed" {
			missed = sql.NullString{}
		}
		for _, eid := range in.Body.EventIDs {
			if _, err := q.UpsertLibraryEntry(ctx, sqlc.UpsertLibraryEntryParams{
				UserID:       userID,
				EventID:      eid,
				Status:       in.Body.Status,
				MissedReason: missed,
				Note:         toNullStr(in.Body.Note),
				AttendedDate: toNullStr(in.Body.AttendedDate),
			}); err != nil {
				return nil, fmt.Errorf("bulk set status: %w", err)
			}
			updated++
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	out := &BulkStatusOutput{}
	out.Body.Updated = updated
	return out, nil
}

// BulkSetFetchMode sets the fetch mode (auto|manual) for many artists at once.
func (h *Handler) BulkSetFetchMode(ctx context.Context, in *BulkFetchModeInput) (*BulkFetchModeOutput, error) {
	if len(in.Body.ArtistIDs) == 0 {
		return nil, huma400("artist_ids is required")
	}

	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	in1, a1 := inList(in.Body.ArtistIDs)
	res, err := tx.ExecContext(ctx,
		"UPDATE artists SET fetch_mode = ?, updated_at = CURRENT_TIMESTAMP WHERE id IN "+in1,
		append([]any{in.Body.FetchMode}, a1...)...)
	if err != nil {
		return nil, fmt.Errorf("bulk fetch-mode: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	n, _ := res.RowsAffected()
	out := &BulkFetchModeOutput{}
	out.Body.Updated = int(n)
	return out, nil
}

// BulkDeleteEvents deletes many events by ID, keeping any that are claimed (protected by both
// the NOT EXISTS predicate and library_entries.event_id ON DELETE RESTRICT).
func (h *Handler) BulkDeleteEvents(ctx context.Context, in *BulkDeleteEventsInput) (*BulkDeleteEventsOutput, error) {
	if len(in.Body.EventIDs) == 0 {
		return nil, huma400("event_ids is required")
	}

	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	in1, a1 := inList(in.Body.EventIDs)
	res, err := tx.ExecContext(ctx,
		"DELETE FROM events WHERE id IN "+in1+
			" AND NOT EXISTS (SELECT 1 FROM library_entries WHERE event_id = events.id)", a1...)
	if err != nil {
		return nil, fmt.Errorf("bulk delete events: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	n, _ := res.RowsAffected()
	out := &BulkDeleteEventsOutput{}
	out.Body.Deleted = int(n)
	out.Body.KeptClaimed = len(in.Body.EventIDs) - int(n)
	return out, nil
}
