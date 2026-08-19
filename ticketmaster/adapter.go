package ticketmaster

import (
	"context"
	"encoding/json"

	"github.com/protob/event-sensor/internal/provider"
)

// Adapter wraps the Ticketmaster client + parsing pipeline behind the neutral
// provider.FetchingProvider seam. All of TM's valuable filtering (junk drop, region,
// keyword-collision guard, festival per-product dedup) stays here, inside the adapter; what
// crosses the seam is already-filtered, storage-shaped provider.ProviderEvent values.
type Adapter struct{ client *Client }

// NewAdapter builds a TM adapter over a client (the handler picks the client so a per-user
// `tm.api_key` override still applies).
func NewAdapter(c *Client) *Adapter { return &Adapter{client: c} }

func (a *Adapter) Name() string { return "ticketmaster" }

// FetchByArtist runs the pipeline: search -> parse+junk-drop -> region filter ->
// keyword-collision guard -> festival per-product merge -> map to ProviderEvent.
func (a *Adapter) FetchByArtist(ctx context.Context, artist string, opts provider.FetchOpts) ([]*provider.ProviderEvent, error) {
	resp, err := a.client.SearchEvents(artist)
	if err != nil {
		return nil, err
	}
	if resp.Embedded == nil {
		return nil, nil
	}

	// Parse + filter first (pure). MergePerformances operates on []*ParsedEvent, so keep
	// working in that shape and map to ProviderEvent only after the merge.
	var kept []*ParsedEvent
	for _, raw := range resp.Embedded.Events {
		p := ParseEvent(raw) // junk-drop + venue
		if p == nil || p.Venue == nil {
			continue
		}
		if !InRegion(p.Venue.CountryCode, opts.Region) {
			continue
		}
		if ClassifyArtistMatch(p, artist) == MatchNone {
			continue
		}
		kept = append(kept, p)
	}

	merged := MergePerformances(kept)

	out := make([]*provider.ProviderEvent, 0, len(merged))
	for _, p := range merged {
		out = append(out, toProviderEvent(p, artist))
	}
	return out, nil
}

// toProviderEvent maps a TM ParsedEvent to the source-neutral ProviderEvent. Free text is
// passed through verbatim (UTF-8).
func toProviderEvent(p *ParsedEvent, artist string) *provider.ProviderEvent {
	pe := &provider.ProviderEvent{
		Source:      "ticketmaster",
		SourceID:    p.TmID,
		Name:        p.Name,
		Kind:        mapKind(p, artist),
		About:       p.About,
		Description: p.Description,
		StartDate:   p.StartDate,
		EndDate:     p.EndDate,
		HasTime:     hasTime(p.StartDate),
		URL:         p.URL,
		ImageURL:    p.ImageURL,
	}

	if p.Venue != nil {
		pe.Venue = &provider.ProviderVenue{
			Source:      "ticketmaster",
			SourceID:    p.Venue.TmID,
			Name:        p.Venue.Name,
			City:        p.Venue.City,
			Country:     p.Venue.Country,
			CountryCode: p.Venue.CountryCode,
			Latitude:    floatPtr(p.Venue.Latitude),
			Longitude:   floatPtr(p.Venue.Longitude),
		}
	}

	for _, at := range p.Attractions {
		pe.Attractions = append(pe.Attractions, at.Name)
	}
	// The searched artist headlines the result (that is why it was kept).
	pe.Headliners = []string{artist}

	if len(p.TicketOptions) > 0 {
		if b, err := json.Marshal(p.TicketOptions); err == nil {
			pe.TicketOptions = b
		}
	}
	return pe
}

// mapKind maps TM's coarse Type ("Concert"/"Festival") + tribute detection onto the
// events.kind CHECK set.
func mapKind(p *ParsedEvent, artist string) string {
	if p.Type == "Festival" {
		return "festival"
	}
	if ClassifyArtistMatch(p, artist) == MatchTribute {
		return "tribute"
	}
	return "concert"
}

// hasTime reports whether a TM start date carries a time component. dateString() returns a
// full ISO dateTime ("2026-06-20T19:00:00Z") when TM announced a time, or a bare localDate
// ("2026-06-20") when it hasn't - so a length past the date prefix means a time is present.
func hasTime(start string) bool { return len(start) > len("2006-01-02") }

// floatPtr returns nil for a zero coordinate, else a pointer to the value.
func floatPtr(f float64) *float64 {
	if f == 0 {
		return nil
	}
	return &f
}
