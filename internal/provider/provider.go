// Package provider is the source-agnostic ingestion seam. Storage (db/queries, reconcile)
// speaks ProviderEvent and never switches on `source`; each concrete source (Ticketmaster,
// manual entry, future scrapers) maps its own shape into a ProviderEvent and applies its own
// filtering before the event crosses this seam.
package provider

import (
	"context"
	"encoding/json"
)

// ProviderEvent is storage-shaped and source-neutral. By the time you hold one, all
// source-specific filtering (region, junk, keyword collisions, festival dedup) has already
// been applied. Names and free text are raw UTF-8 and are stored verbatim.
type ProviderEvent struct {
	Source        string // "ticketmaster" | "manual"
	SourceID      string // TM event id; "" for manual
	Name          string
	Kind          string // "concert" | "festival" | "tribute" | "club" | "other"
	Description   string
	About         string
	StartDate     string // ISO-8601 UTC, or "YYYY-MM-DD"
	EndDate       string
	HasTime       bool
	URL           string
	ImageURL      string
	Venue         *ProviderVenue
	Attractions   []string        // billed performer names (free-text)
	Headliners    []string        // subset of Attractions that headline
	TicketOptions json.RawMessage // JSON [{label,url}] (TM festival products); nil for manual
}

// ProviderVenue is the source-neutral venue shape.
type ProviderVenue struct {
	Source, SourceID                 string
	Name, City, Country, CountryCode string
	Latitude, Longitude              *float64
	Timezone                         string
}

// FetchOpts carries source-specific knobs for a fetch (e.g. the region filter set).
type FetchOpts struct {
	Region map[string]bool // region.codes set; TM-specific filtering uses it
}

// FetchingProvider can search by artist. Today: Ticketmaster only.
type FetchingProvider interface {
	Name() string
	FetchByArtist(ctx context.Context, artistName string, opts FetchOpts) ([]*ProviderEvent, error)
}

// IntakeProvider supplies events one at a time (no search). Manual entry today; future
// scrapers (Bandsintown / Songkick) implement this same interface with zero storage change.
type IntakeProvider interface {
	Name() string
	// Intake validates/normalizes a user- or scraper-supplied event into a ProviderEvent.
	Intake(ctx context.Context, in *ProviderEvent) (*ProviderEvent, error)
}
