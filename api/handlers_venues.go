package api

import (
	"context"
	"database/sql"
	"fmt"
)

// ListVenues returns all venues with the number of events at each - the count is what makes
// duplicate venues spottable and tells you which one to keep when merging.
func (h *Handler) ListVenues(ctx context.Context, _ *ListVenuesInput) (*ListVenuesOutput, error) {
	venues, err := h.queries.ListVenuesWithCounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list venues: %w", err)
	}

	resp := make([]VenueResponse, 0, len(venues))
	for _, v := range venues {
		resp = append(resp, VenueResponse{
			ID:          v.ID,
			Name:        v.Name,
			City:        v.City.String,
			Country:     v.Country.String,
			CountryCode: v.CountryCode.String,
			Latitude:    nullFloat(v.Latitude),
			Longitude:   nullFloat(v.Longitude),
			Timezone:    nullStr(v.Timezone),
			Source:      nullStr(v.Source),
			EventCount:  int(v.EventCount),
			CreatedAt:   v.CreatedAt,
			UpdatedAt:   v.UpdatedAt,
		})
	}

	return &ListVenuesOutput{Body: resp}, nil
}

// GetVenue returns a single venue by ID.
func (h *Handler) GetVenue(ctx context.Context, input *GetVenueInput) (*GetVenueOutput, error) {
	venue, err := h.queries.GetVenue(ctx, input.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, huma404("venue not found")
		}
		return nil, fmt.Errorf("failed to get venue: %w", err)
	}

	resp := venueToResponse(venue)
	if n, err := h.queries.CountVenueEvents(ctx, toNullStr(input.ID)); err == nil {
		resp.EventCount = int(n)
	}
	return &GetVenueOutput{Body: resp}, nil
}
