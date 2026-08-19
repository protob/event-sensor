package api

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/protob/event-sensor/db/sqlc"
	"github.com/protob/event-sensor/internal/uuid"
	"github.com/protob/event-sensor/ticketmaster"
)

func artistToResponse(a sqlc.Artist) ArtistResponse {
	return ArtistResponse{
		ID:        a.ID,
		Name:      a.Name,
		FetchMode: a.FetchMode,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}

// ListArtists returns all artists.
func (h *Handler) ListArtists(ctx context.Context, _ *ListArtistsInput) (*ListArtistsOutput, error) {
	artists, err := h.queries.ListArtists(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list artists: %w", err)
	}

	resp := make([]ArtistResponse, 0, len(artists))
	for _, a := range artists {
		resp = append(resp, artistToResponse(a))
	}

	return &ListArtistsOutput{Body: resp}, nil
}

// GetArtist returns a single artist by ID.
func (h *Handler) GetArtist(ctx context.Context, input *GetArtistInput) (*GetArtistOutput, error) {
	artist, err := h.queries.GetArtist(ctx, input.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, huma404("artist not found")
		}
		return nil, fmt.Errorf("failed to get artist: %w", err)
	}

	return &GetArtistOutput{Body: artistToResponse(artist)}, nil
}

// CreateArtist creates a new artist (fetch_mode defaults to 'auto').
func (h *Handler) CreateArtist(ctx context.Context, input *CreateArtistInput) (*CreateArtistOutput, error) {
	fetchMode := input.Body.FetchMode
	if fetchMode == "" {
		fetchMode = "auto"
	}
	artist, err := h.queries.CreateArtist(ctx, sqlc.CreateArtistParams{
		ID:        uuid.New(),
		Name:      input.Body.Name,
		FetchMode: fetchMode,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create artist: %w", err)
	}

	return &CreateArtistOutput{Body: artistToResponse(artist)}, nil
}

// UpdateArtist updates an artist's name.
func (h *Handler) UpdateArtist(ctx context.Context, input *UpdateArtistInput) (*UpdateArtistOutput, error) {
	artist, err := h.queries.UpdateArtist(ctx, sqlc.UpdateArtistParams{
		Name: input.Body.Name,
		ID:   input.ID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, huma404("artist not found")
		}
		return nil, fmt.Errorf("failed to update artist: %w", err)
	}

	return &UpdateArtistOutput{Body: artistToResponse(artist)}, nil
}

// SetArtistFetchMode toggles an artist between 'auto' (fetched by Fetch All) and 'manual' (fetched only on demand).
func (h *Handler) SetArtistFetchMode(ctx context.Context, input *SetFetchModeInput) (*SetFetchModeOutput, error) {
	artist, err := h.queries.SetArtistFetchMode(ctx, sqlc.SetArtistFetchModeParams{
		FetchMode: input.Body.FetchMode,
		ID:        input.ID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, huma404("artist not found")
		}
		return nil, fmt.Errorf("failed to set fetch mode: %w", err)
	}
	return &SetFetchModeOutput{Body: artistToResponse(artist)}, nil
}

// DeleteArtist drops the artist's un-claimed events, then deletes the artist (its
// performances' artist_id is SET NULL on delete, category memberships cascade). Claimed
// events remain in the diary, now artist-less - a claim never blocks the delete.
func (h *Handler) DeleteArtist(ctx context.Context, input *DeleteArtistInput) (*DeleteArtistOutput, error) {
	if err := h.queries.DeleteUnclaimedArtistEvents(ctx, toNullStr(input.ID)); err != nil {
		return nil, fmt.Errorf("failed to clear artist's un-claimed events: %w", err)
	}
	if err := h.queries.DeleteArtist(ctx, input.ID); err != nil {
		return nil, fmt.Errorf("failed to delete artist: %w", err)
	}

	return &DeleteArtistOutput{
		Body: struct {
			Message string `json:"message"`
		}{Message: "artist deleted"},
	}, nil
}

// ListCategoryArtists returns all artists in a category.
func (h *Handler) ListCategoryArtists(ctx context.Context, input *ListCategoryArtistsInput) (*ListCategoryArtistsOutput, error) {
	artists, err := h.queries.ListArtistsByCategory(ctx, input.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to list artists by category: %w", err)
	}

	resp := make([]ArtistResponse, 0, len(artists))
	for _, a := range artists {
		resp = append(resp, artistToResponse(a))
	}

	return &ListCategoryArtistsOutput{Body: resp}, nil
}

// AddArtistToCategory adds an artist to a category.
func (h *Handler) AddArtistToCategory(ctx context.Context, input *AddArtistToCategoryInput) (*AddArtistToCategoryOutput, error) {
	if err := h.queries.AddArtistToCategory(ctx, sqlc.AddArtistToCategoryParams{
		ArtistID:   input.Body.ArtistID,
		CategoryID: input.ID,
	}); err != nil {
		return nil, fmt.Errorf("failed to add artist to category: %w", err)
	}

	return &AddArtistToCategoryOutput{
		Body: struct {
			Message string `json:"message"`
		}{Message: "artist added to category"},
	}, nil
}

// RemoveArtistFromCategory removes an artist from a category.
func (h *Handler) RemoveArtistFromCategory(ctx context.Context, input *RemoveArtistFromCategoryInput) (*RemoveArtistFromCategoryOutput, error) {
	if err := h.queries.RemoveArtistFromCategory(ctx, sqlc.RemoveArtistFromCategoryParams{
		ArtistID:   input.ArtistID,
		CategoryID: input.CategoryID,
	}); err != nil {
		return nil, fmt.Errorf("failed to remove artist from category: %w", err)
	}

	return &RemoveArtistFromCategoryOutput{
		Body: struct {
			Message string `json:"message"`
		}{Message: "artist removed from category"},
	}, nil
}

// FetchArtistEvents refreshes an artist's Ticketmaster feed via the reconcile (mark-and-
// sweep). Manual artists are fetched on demand (Fetch All skips them client-side). Claimed events are never
// deleted by a refresh.
func (h *Handler) FetchArtistEvents(ctx context.Context, input *FetchArtistEventsInput) (*FetchArtistEventsOutput, error) {
	artist, err := h.queries.GetArtist(ctx, input.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, huma404("artist not found")
		}
		return nil, fmt.Errorf("failed to get artist: %w", err)
	}

	// Resolve per-user overrides: `tm.api_key` (overrides the env/agenix default) and
	// `region.codes` (the countries worth saving - defaults to DefaultRegionCodes when unset).
	tmClient := h.tmClient
	region := ticketmaster.DefaultRegionCodes
	if uid := UserIDFromContext(ctx); uid != "" {
		if s, err := h.queries.GetUserSetting(ctx, sqlc.GetUserSettingParams{UserID: uid, Key: "tm.api_key"}); err == nil && s.Value != "" {
			tmClient = ticketmaster.NewClientWithBase(s.Value, h.config.TicketmasterBaseURL)
		}
		if s, err := h.queries.GetUserSetting(ctx, sqlc.GetUserSettingParams{UserID: uid, Key: "region.codes"}); err == nil && s.Value != "" {
			region = ticketmaster.CodeSet(s.Value)
		}
	}

	adapter := ticketmaster.NewAdapter(tmClient)
	res, err := h.reconcile.Reconcile(ctx, artist, region, adapter)
	if err != nil {
		return nil, fmt.Errorf("ticketmaster reconcile failed: %w", err)
	}

	return &FetchArtistEventsOutput{Body: *res}, nil
}

// ListArtistDBEvents returns events for an artist from the database with venue + claim info.
func (h *Handler) ListArtistDBEvents(ctx context.Context, input *ListArtistEventsInput) (*ListArtistEventsOutput, error) {
	events, err := h.queries.ListEventsByArtist(ctx, toNullStr(input.ID))
	if err != nil {
		return nil, fmt.Errorf("failed to list events for artist: %w", err)
	}

	library := h.userLibraryMap(ctx, UserIDFromContext(ctx))

	resp := make([]EventResponse, 0, len(events))
	for _, e := range events {
		r := eventToResponse(e)
		if le, ok := library[e.ID]; ok {
			applyClaim(&r, le)
		}
		if e.VenueID.Valid {
			venue, err := h.queries.GetVenue(ctx, e.VenueID.String)
			if err == nil {
				v := venueToResponse(venue)
				r.Venue = &v
			}
		}
		resp = append(resp, r)
	}

	return &ListArtistEventsOutput{Body: resp}, nil
}

// ClearArtistEvents deletes an artist's un-claimed events; claims survive.
func (h *Handler) ClearArtistEvents(ctx context.Context, input *ClearArtistEventsInput) (*ClearArtistEventsOutput, error) {
	if err := h.queries.DeleteUnclaimedArtistEvents(ctx, toNullStr(input.ID)); err != nil {
		return nil, fmt.Errorf("failed to clear events for artist: %w", err)
	}

	return &ClearArtistEventsOutput{
		Body: struct {
			Message string `json:"message"`
		}{Message: "artist's un-claimed events cleared"},
	}, nil
}

// ListArtistSummaries returns per-artist aggregates in one round trip. The category tree
// otherwise fetches every category's membership separately on mount - one request per
// category, on the app's landing page.
func (h *Handler) ListArtistSummaries(ctx context.Context, _ *ListArtistSummariesInput) (*ListArtistSummariesOutput, error) {
	userID := UserIDFromContext(ctx)
	if userID == "" {
		return nil, huma400("user not authenticated")
	}

	artists, err := h.queries.ListArtists(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list artists: %w", err)
	}

	byArtist := make(map[string]*ArtistSummaryResponse, len(artists))
	out := make([]ArtistSummaryResponse, 0, len(artists))
	for _, a := range artists {
		byArtist[a.ID] = &ArtistSummaryResponse{
			ArtistID:   a.ID,
			Categories: []CategoryRef{},
			Countries:  []CountryCount{},
		}
	}

	pairs, err := h.queries.ListArtistCategoryPairs(ctx, toNullStr(userID))
	if err != nil {
		return nil, fmt.Errorf("failed to list category memberships: %w", err)
	}
	for _, p := range pairs {
		if s := byArtist[p.ArtistID]; s != nil {
			s.Categories = append(s.Categories, CategoryRef{ID: p.CategoryID, Name: p.CategoryName})
		}
	}

	counts, err := h.queries.ListArtistEventCounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count artist events: %w", err)
	}
	for _, c := range counts {
		if s := byArtist[c.ArtistID.String]; s != nil {
			s.EventCount = int(c.EventCount)
			s.UpcomingListedCount = int(c.UpcomingListedCount)
		}
	}

	countries, err := h.queries.ListArtistCountryCounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count artist countries: %w", err)
	}
	for _, c := range countries {
		if s := byArtist[c.ArtistID.String]; s != nil {
			s.Countries = append(s.Countries, CountryCount{Code: c.CountryCode.String, Count: int(c.N)})
		}
	}

	claimed, err := h.queries.ListArtistClaimedCounts(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count claimed events: %w", err)
	}
	for _, c := range claimed {
		if s := byArtist[c.ArtistID.String]; s != nil {
			s.ClaimedCount = int(c.ClaimedCount)
		}
	}

	// Preserve ListArtists' name order rather than map order.
	for _, a := range artists {
		out = append(out, *byArtist[a.ID])
	}
	return &ListArtistSummariesOutput{Body: out}, nil
}
