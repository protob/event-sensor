package api

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/protob/event-sensor/db/sqlc"
)

func libraryEntryToResponse(le sqlc.LibraryEntry) LibraryEntryResponse {
	return LibraryEntryResponse{
		EventID:      le.EventID,
		Status:       le.Status,
		MissedReason: nullStr(le.MissedReason),
		Note:         nullStr(le.Note),
		AttendedDate: nullStr(le.AttendedDate),
		CreatedAt:    le.CreatedAt,
		UpdatedAt:    le.UpdatedAt,
	}
}

// ListLibrary returns the current user's library entries (the diary index the frontend
// merges onto events).
func (h *Handler) ListLibrary(ctx context.Context, _ *ListLibraryInput) (*ListLibraryOutput, error) {
	uid := UserIDFromContext(ctx)
	if uid == "" {
		return nil, huma400("user not found in context")
	}
	entries, err := h.queries.ListUserLibrary(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("list library: %w", err)
	}
	out := make([]LibraryEntryResponse, 0, len(entries))
	for _, e := range entries {
		out = append(out, libraryEntryToResponse(e))
	}
	return &ListLibraryOutput{Body: out}, nil
}

// SetEventStatus sets or clears the per-user claim for an event. An empty status unclaims
// (deletes the library entry); otherwise it upserts the claim with its diary fields.
func (h *Handler) SetEventStatus(ctx context.Context, in *SetEventStatusInput) (*SetEventStatusOutput, error) {
	uid := UserIDFromContext(ctx)
	if uid == "" {
		return nil, huma400("user not found in context")
	}

	if in.Body.Status == "" {
		if err := h.queries.DeleteLibraryEntry(ctx, sqlc.DeleteLibraryEntryParams{UserID: uid, EventID: in.ID}); err != nil {
			return nil, fmt.Errorf("unclaim event: %w", err)
		}
		return &SetEventStatusOutput{Body: LibraryEntryResponse{EventID: in.ID, Status: ""}}, nil
	}

	// missed_reason only valid for status='missed' (DB CHECK enforces this too).
	missed := toNullStr(in.Body.MissedReason)
	if in.Body.Status != "missed" {
		missed = sql.NullString{}
	}

	le, err := h.queries.UpsertLibraryEntry(ctx, sqlc.UpsertLibraryEntryParams{
		UserID:       uid,
		EventID:      in.ID,
		Status:       in.Body.Status,
		MissedReason: missed,
		Note:         toNullStr(in.Body.Note),
		AttendedDate: toNullStr(in.Body.AttendedDate),
	})
	if err != nil {
		return nil, fmt.Errorf("set status: %w", err)
	}
	return &SetEventStatusOutput{Body: libraryEntryToResponse(le)}, nil
}

// ListNeedsResolution returns past `going` events awaiting a "did you go?" answer (the
// attendance nudge).
func (h *Handler) ListNeedsResolution(ctx context.Context, _ *struct{}) (*NeedsResolutionOutput, error) {
	uid := UserIDFromContext(ctx)
	if uid == "" {
		return nil, huma400("user not found in context")
	}
	rows, err := h.queries.ListNeedsResolution(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("list needs-resolution: %w", err)
	}
	out := make([]NeedsResolutionItem, 0, len(rows))
	for _, r := range rows {
		out = append(out, NeedsResolutionItem{EventID: r.EventID, Name: r.Name, StartDate: r.StartDate})
	}
	return &NeedsResolutionOutput{Body: out}, nil
}
