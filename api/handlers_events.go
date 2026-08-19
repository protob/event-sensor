package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/protob/event-sensor/db/sqlc"
	"github.com/protob/event-sensor/internal/provider"
	"github.com/protob/event-sensor/internal/uuid"
)

// addPerformances writes a manual event's artists (the bill / the festival artists you saw) as
// `performances`. ArtistName is required; ArtistID links an entity artist (optional);
// Date is the day (festivals). Empty-named rows are skipped.
func addPerformances(ctx context.Context, q *sqlc.Queries, eventID string, perfs []PerformanceInput) error {
	for _, p := range perfs {
		name := strings.TrimSpace(p.ArtistName)
		if name == "" {
			continue
		}
		var artistID sql.NullString
		if p.ArtistID != nil {
			artistID = toNullStr(strings.TrimSpace(*p.ArtistID))
		}
		var date sql.NullString
		if p.Date != nil {
			date = toNullStr(strings.TrimSpace(*p.Date))
		}
		var venueName sql.NullString
		if p.VenueName != nil {
			venueName = toNullStr(strings.TrimSpace(*p.VenueName))
		}
		headliner := int64(0)
		if p.IsHeadliner {
			headliner = 1
		}
		if err := q.AddPerformance(ctx, sqlc.AddPerformanceParams{
			ID:          uuid.New(),
			EventID:     eventID,
			ArtistID:    artistID,
			ArtistName:  name,
			IsHeadliner: headliner,
			Date:        date,
			VenueName:   venueName,
		}); err != nil {
			return err
		}
	}
	return nil
}

// isPast reports whether an event's start date is before today (UTC), at day granularity -
// matching the SQL date('now') comparison used for the attendance nudge.
func isPast(startDate string) bool {
	if len(startDate) < 10 {
		return false
	}
	return startDate[:10] < time.Now().UTC().Format("2006-01-02")
}

func eventToResponse(e sqlc.Event) EventResponse {
	var ticketOptions []TicketOptionResponse
	if e.TicketOptions != "" {
		_ = json.Unmarshal([]byte(e.TicketOptions), &ticketOptions)
	}
	r := EventResponse{
		ID:            e.ID,
		Source:        e.Source,
		SourceID:      nullStr(e.SourceID),
		Name:          e.Name,
		Kind:          e.Kind,
		Description:   nullStr(e.Description),
		StartDate:     e.StartDate,
		EndDate:       nullStr(e.EndDate),
		HasTime:       e.HasTime != 0,
		ListingState:  e.ListingState,
		VenueID:       nullStr(e.VenueID),
		URL:           nullStr(e.Url),
		About:         nullStr(e.About),
		ImageURL:      nullStr(e.ImageUrl),
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     e.UpdatedAt,
		IsPast:        isPast(e.StartDate),
		TicketOptions: ticketOptions,
		DiscoveredAt:  e.DiscoveredAt,
	}
	if e.Extra != "" && e.Extra != "{}" {
		extra := e.Extra
		r.Extra = &extra
	}
	return r
}

func applyClaim(r *EventResponse, le sqlc.LibraryEntry) {
	status := le.Status
	r.Status = &status
	r.Note = nullStr(le.Note)
}

func venueToResponse(v sqlc.Venue) VenueResponse {
	return VenueResponse{
		ID:          v.ID,
		Name:        v.Name,
		City:        v.City.String,
		Country:     v.Country.String,
		CountryCode: v.CountryCode.String,
		Latitude:    nullFloat(v.Latitude),
		Longitude:   nullFloat(v.Longitude),
		Timezone:    nullStr(v.Timezone),
		Source:      nullStr(v.Source),
		CreatedAt:   v.CreatedAt,
		UpdatedAt:   v.UpdatedAt,
	}
}

// userLibraryMap returns the authenticated user's claims keyed by event_id (empty when
// unauthenticated or on error - reads must not fail because of an absent claim).
func (h *Handler) userLibraryMap(ctx context.Context, userID string) map[string]sqlc.LibraryEntry {
	out := map[string]sqlc.LibraryEntry{}
	if userID == "" {
		return out
	}
	entries, err := h.queries.ListUserLibrary(ctx, userID)
	if err != nil {
		return out
	}
	for _, e := range entries {
		out[e.EventID] = e
	}
	return out
}

// ListEvents returns all events, optionally filtered by date range or search query, with the
// user's per-event library claim merged in when authenticated.
func (h *Handler) ListEvents(ctx context.Context, input *ListEventsInput) (*ListEventsOutput, error) {
	var events []sqlc.Event
	var err error

	if input.Query != "" {
		pattern := "%" + input.Query + "%"
		events, err = h.queries.SearchEvents(ctx, sqlc.SearchEventsParams{
			Name:       pattern,
			Name_2:     pattern,
			ArtistName: pattern,
		})
	} else if input.StartDate != "" && input.EndDate != "" {
		events, err = h.queries.ListEventsByDateRange(ctx, sqlc.ListEventsByDateRangeParams{
			StartDate:   input.StartDate,
			StartDate_2: input.EndDate,
		})
	} else {
		events, err = h.queries.ListEvents(ctx)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	userID := UserIDFromContext(ctx)
	library := h.userLibraryMap(ctx, userID)

	// Enrich the list with venues + artists so the dashboard has location and headliner data
	// without a per-row fetch.
	venueByID := map[string]VenueResponse{}
	if venues, err := h.queries.ListVenues(ctx); err == nil {
		for _, v := range venues {
			venueByID[v.ID] = venueToResponse(v)
		}
	}
	perfByEvent := map[string][]PerformanceResponse{}
	if ps, err := h.queries.ListAllPerformances(ctx); err == nil {
		for _, p := range ps {
			perfByEvent[p.EventID] = append(perfByEvent[p.EventID], PerformanceResponse{
				ID:          p.ID,
				ArtistID:    nullStr(p.ArtistID),
				ArtistName:  p.ArtistName,
				IsHeadliner: p.IsHeadliner != 0,
				Date:        nullStr(p.Date),
				VenueName:   nullStr(p.VenueName),
			})
		}
	}

	resp := make([]EventResponse, 0, len(events))
	for _, e := range events {
		r := eventToResponse(e)
		if le, ok := library[e.ID]; ok {
			applyClaim(&r, le)
		}
		if e.VenueID.Valid {
			if v, ok := venueByID[e.VenueID.String]; ok {
				vv := v
				r.Venue = &vv
			}
		}
		if ps, ok := perfByEvent[e.ID]; ok {
			r.Performances = ps
		}
		resp = append(resp, r)
	}

	return &ListEventsOutput{Body: resp}, nil
}

// GetEvent returns a single event with venue, artists, lineup, and the user's claim.
func (h *Handler) GetEvent(ctx context.Context, input *GetEventInput) (*GetEventOutput, error) {
	event, err := h.queries.GetEvent(ctx, input.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, huma404("event not found")
		}
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	resp := eventToResponse(event)

	if userID := UserIDFromContext(ctx); userID != "" {
		if le, err := h.queries.GetLibraryEntry(ctx, sqlc.GetLibraryEntryParams{UserID: userID, EventID: event.ID}); err == nil {
			applyClaim(&resp, le)
		}
	}

	if event.VenueID.Valid {
		venue, err := h.queries.GetVenue(ctx, event.VenueID.String)
		if err == nil {
			v := venueToResponse(venue)
			resp.Venue = &v
		}
	}

	if perfs, err := h.queries.ListPerformancesByEvent(ctx, event.ID); err == nil {
		ps := make([]PerformanceResponse, 0, len(perfs))
		for _, p := range perfs {
			ps = append(ps, PerformanceResponse{
				ID:          p.ID,
				ArtistID:    nullStr(p.ArtistID),
				ArtistName:  p.ArtistName,
				IsHeadliner: p.IsHeadliner != 0,
				Date:        nullStr(p.Date),
				VenueName:   nullStr(p.VenueName),
			})
		}
		resp.Performances = ps
	}

	return &GetEventOutput{Body: resp}, nil
}

// DeleteEvent deletes an event. Claim-safe: a claimed event is protected by ON DELETE
// RESTRICT, so we return 409 ("unclaim it first") instead of a 500.
func (h *Handler) DeleteEvent(ctx context.Context, input *DeleteEventInput) (*DeleteEventOutput, error) {
	if err := h.queries.DeleteEvent(ctx, input.ID); err != nil {
		if isForeignKeyErr(err) {
			return nil, huma409("this event is in a library (interested/going/attended) — unclaim it first")
		}
		return nil, fmt.Errorf("failed to delete event: %w", err)
	}

	return &DeleteEventOutput{
		Body: struct {
			Message string `json:"message"`
		}{Message: "event deleted"},
	}, nil
}

// ClearAllEvents deletes all un-claimed events (claimed events survive via the NOT EXISTS
// predicate + ON DELETE RESTRICT). Artists are kept.
func (h *Handler) ClearAllEvents(ctx context.Context, input *ClearAllEventsInput) (*ClearAllEventsOutput, error) {
	if err := h.queries.DeleteAllUnclaimedEvents(ctx); err != nil {
		return nil, fmt.Errorf("failed to clear events: %w", err)
	}

	return &ClearAllEventsOutput{
		Body: struct {
			Message string `json:"message"`
		}{Message: "all un-claimed events cleared (claimed events kept)"},
	}, nil
}

// PrunePastTmEvents deletes past Ticketmaster events that no user has claimed.
// Claimed events survive via the NOT EXISTS predicate + ON DELETE RESTRICT backstop.
func (h *Handler) PrunePastTmEvents(ctx context.Context, input *PrunePastTmEventsInput) (*PrunePastTmEventsOutput, error) {
	if err := h.queries.DeletePastUnclaimedTmEvents(ctx); err != nil {
		return nil, fmt.Errorf("failed to prune past TM events: %w", err)
	}
	return &PrunePastTmEventsOutput{
		Body: struct {
			Message string `json:"message"`
		}{Message: "past unclaimed Ticketmaster events deleted"},
	}, nil
}

// UpdateEvent edits a manual event's facts in place (name/kind/date/venue/artists/lineup).
// Only manual events are editable - Ticketmaster events are owned by the reconcile and would
// be overwritten on the next fetch, so we reject those with 409.
func (h *Handler) UpdateEvent(ctx context.Context, input *UpdateEventInput) (*UpdateEventOutput, error) {
	userID := UserIDFromContext(ctx)
	if userID == "" {
		return nil, huma400("user not authenticated")
	}
	existing, err := h.queries.GetEvent(ctx, input.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, huma404("event not found")
		}
		return nil, fmt.Errorf("get event: %w", err)
	}
	if existing.Source != "manual" {
		return nil, huma409("only manual events can be edited; Ticketmaster events are managed by fetch")
	}
	b := input.Body
	if b.Name == "" || b.StartDate == "" {
		return nil, huma400("name and start_date are required")
	}
	kind := b.Kind
	if kind == "" {
		kind = existing.Kind
	}

	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck
	q := h.queries.WithTx(tx)

	var venueID sql.NullString
	// Create the event venue when a name OR a city is given (urban festivals carry only a city -
	// the per-artist venues hold the halls; the event venue still gives the row its city + flag).
	if b.Venue != nil && (b.Venue.Name != "" || b.Venue.City != "") {
		resolved, err := resolveManualVenue(ctx, q,
			b.Venue.Name, b.Venue.City, b.Venue.Country, b.Venue.CountryCode)
		if err != nil {
			return nil, err
		}
		venueID = resolved
	}

	hasTime := int64(0)
	if b.HasTime {
		hasTime = 1
	}
	event, err := q.UpdateEvent(ctx, sqlc.UpdateEventParams{
		Name:      b.Name,
		Kind:      kind,
		About:     toNullStr(b.About),
		StartDate: b.StartDate,
		EndDate:   toNullStr(b.EndDate),
		HasTime:   hasTime,
		VenueID:   venueID,
		Url:       toNullStr(b.URL),
		ImageUrl:  toNullStr(b.ImageURL),
		ID:        input.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("update event: %w", err)
	}

	// Replace performances wholesale (the form is the source of truth on edit).
	if err := q.DeletePerformancesByEvent(ctx, event.ID); err != nil {
		return nil, fmt.Errorf("clear performances: %w", err)
	}
	if err := addPerformances(ctx, q, event.ID, b.Performances); err != nil {
		return nil, fmt.Errorf("add performances: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	resp := eventToResponse(event)
	if le, err := h.queries.GetLibraryEntry(ctx, sqlc.GetLibraryEntryParams{UserID: userID, EventID: event.ID}); err == nil {
		applyClaim(&resp, le)
	}
	if venueID.Valid {
		if v, err := h.queries.GetVenue(ctx, venueID.String); err == nil {
			vr := venueToResponse(v)
			resp.Venue = &vr
		}
	}
	return &UpdateEventOutput{Body: resp}, nil
}

// resolveManualVenue returns the id of the manual venue for this name+city, reusing the
// existing row when there is one. Manual venues have no source_id, so the (source, source_id)
// unique index cannot dedupe them - without this every save of a hand-entered event mints
// another row for the same place.
func resolveManualVenue(ctx context.Context, q *sqlc.Queries, name, city, country, countryCode string) (sql.NullString, error) {
	name = strings.TrimSpace(name)
	city = strings.TrimSpace(city)

	existing, err := q.GetManualVenue(ctx, sqlc.GetManualVenueParams{Name: name, City: toNullStr(city)})
	if err == nil {
		return toNullStr(existing.ID), nil
	}
	if err != sql.ErrNoRows {
		return sql.NullString{}, fmt.Errorf("look up manual venue: %w", err)
	}

	created, err := q.UpsertVenue(ctx, sqlc.UpsertVenueParams{
		ID:          uuid.New(),
		Name:        name,
		City:        toNullStr(city),
		Country:     toNullStr(strings.TrimSpace(country)),
		CountryCode: toNullStr(strings.TrimSpace(countryCode)),
		Source:      toNullStr("manual"),
	})
	if err != nil {
		return sql.NullString{}, fmt.Errorf("create manual venue: %w", err)
	}
	return toNullStr(created.ID), nil
}

// CreateManualEvent creates a hand-entered event (source='manual'), optionally linked to
// tracked artists and optionally claimed immediately (e.g. logging a past show as attended).
func (h *Handler) CreateManualEvent(ctx context.Context, input *CreateManualEventInput) (*CreateManualEventOutput, error) {
	userID := UserIDFromContext(ctx)
	if userID == "" {
		return nil, huma400("user not authenticated")
	}
	b := input.Body

	pe := &provider.ProviderEvent{
		Name:      b.Name,
		Kind:      b.Kind,
		StartDate: b.StartDate,
		EndDate:   b.EndDate,
		HasTime:   b.HasTime,
		About:     b.About,
		URL:       b.URL,
		ImageURL:  b.ImageURL,
	}
	if b.Venue != nil {
		pe.Venue = &provider.ProviderVenue{
			Name: b.Venue.Name, City: b.Venue.City,
			Country: b.Venue.Country, CountryCode: b.Venue.CountryCode,
		}
	}
	pe, err := (provider.Manual{}).Intake(ctx, pe)
	if err != nil {
		return nil, huma400(err.Error())
	}

	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck
	q := h.queries.WithTx(tx)

	var venueID sql.NullString
	if pe.Venue != nil && (pe.Venue.Name != "" || pe.Venue.City != "") {
		venueID, err = resolveManualVenue(ctx, q,
			pe.Venue.Name, pe.Venue.City, pe.Venue.Country, pe.Venue.CountryCode)
		if err != nil {
			return nil, err
		}
	}

	hasTime := int64(0)
	if pe.HasTime {
		hasTime = 1
	}
	event, err := q.InsertEvent(ctx, sqlc.InsertEventParams{
		ID:            uuid.New(),
		Source:        "manual",
		SourceID:      sql.NullString{},
		Name:          pe.Name,
		Kind:          pe.Kind,
		About:         toNullStr(pe.About),
		StartDate:     pe.StartDate,
		EndDate:       toNullStr(pe.EndDate),
		HasTime:       hasTime,
		VenueID:       venueID,
		Url:           toNullStr(pe.URL),
		ImageUrl:      toNullStr(pe.ImageURL),
		ListingState:  "listed",
		TicketOptions: "[]", // manual events have no ticket products; valid empty JSON array
	})
	if err != nil {
		return nil, fmt.Errorf("insert manual event: %w", err)
	}

	if err := addPerformances(ctx, q, event.ID, b.Performances); err != nil {
		return nil, fmt.Errorf("add manual performances: %w", err)
	}

	resp := eventToResponse(event)
	if b.Status != "" {
		le, err := q.UpsertLibraryEntry(ctx, sqlc.UpsertLibraryEntryParams{
			UserID: userID, EventID: event.ID, Status: b.Status,
		})
		if err != nil {
			return nil, fmt.Errorf("claim manual event: %w", err)
		}
		applyClaim(&resp, le)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit manual event: %w", err)
	}

	if venueID.Valid {
		if v, err := h.queries.GetVenue(ctx, venueID.String); err == nil {
			vr := venueToResponse(v)
			resp.Venue = &vr
		}
	}
	return &CreateManualEventOutput{Body: resp}, nil
}
