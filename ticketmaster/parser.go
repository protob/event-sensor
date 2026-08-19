package ticketmaster

import (
	"regexp"
	"strconv"
	"strings"
)

// FESTIVALS is an optional curated list of known festival name substrings (lowercase).
// Ticketmaster's own festival flag is unreliable - it's set on maybe a quarter of
// festivals (e.g. Austin City Limits) and missing on most (Rock Werchter, Flow,
// Primavera). This is the high-precision extension point: shipped empty for the
// prototype, fill it with names if/when better coverage is wanted, e.g.
//
//	{"rock werchter", "primavera", "roskilde", "open'er", "sziget"}
var FESTIVALS = []string{}

// Matches an explicit "festival" / "fest" word in an event name.
var festivalNameRe = regexp.MustCompile(`(?i)\bfest(ival)?\b`)

// tributeMarkerRe flags titles that present someone's music without that someone being the
// billed artist - tributes, "music of/by X", candlelight covers, etc. Deliberately narrow:
// "plays" is excluded so a real artist's "<Artist> plays <Album>" show isn't mislabelled.
var tributeMarkerRe = regexp.MustCompile(`(?i)\b(tribute|music of|music by|performed by|celebration of|candlelight)\b`)

// ArtistMatch is how a Ticketmaster result relates to the artist that was searched.
type ArtistMatch int

const (
	MatchNone    ArtistMatch = iota // unrelated - keyword coincidence (e.g. "Les Bonobos" for "bonobo")
	MatchArtist                     // the searched artist is the billed attraction
	MatchTribute                    // a tribute / "music of X" referencing the artist, not them
)

// normalizeName lowercases and collapses whitespace for tolerant name comparison.
func normalizeName(s string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(s))), " ")
}

// ClassifyArtistMatch decides whether a parsed event is really the searched artist, a
// tribute to them, or unrelated noise. It matches on the canonical ATTRACTION name rather
// than the event title — that dodges localized/inflected titles (e.g. Polish declension
// "Koncert ... Krzysztofa Pendereckiego") and avoids dropping real shows with long titles
// ("Bonobo: Distance in Static Live ... Tour" whose attraction is just "Bonobo").
func ClassifyArtistMatch(p *ParsedEvent, artistName string) ArtistMatch {
	want := normalizeName(artistName)
	for _, a := range p.Attractions {
		if normalizeName(a.Name) == want {
			return MatchArtist
		}
	}
	// Not billed. A tribute only if the title both names the artist and reads like one.
	if strings.Contains(normalizeName(p.Name), want) && tributeMarkerRe.MatchString(p.Name) {
		return MatchTribute
	}
	return MatchNone
}

// DefaultRegionCodes is the default fetch region: Europe + Turkey, excluding ru/ua/by/md.
// Used only when a user hasn't set their own `region.codes`. Keep in sync with the
// frontend DEFAULT_REGION_CODES (composables/useCountry.ts) - same region, two callers.
var DefaultRegionCodes = CodeSet(
	// EU + EFTA/Schengen + UK + microstates + Western Balkans + Turkey.
	"at,be,bg,hr,cy,cz,dk,ee,fi,fr,de,gr,hu,ie,it,lv,lt,lu,mt,nl,pl,pt,ro,sk,si,es,se," +
		"is,li,no,ch,gb,ad,mc,sm,al,ba,xk,me,mk,rs,tr",
)

// ParsedEvent is a domain representation of a Ticketmaster event.
type ParsedEvent struct {
	TmID        string
	Name        string
	Type        string
	Category    string
	Description string
	StartDate   string
	EndDate     string
	URL         string
	About       string
	ImageURL    string
	Venue       *ParsedVenue
	Attractions []Attraction
	// Alternative ticket products for the same performance (festival day vs combi pass,
	// terrace, …). Populated by MergePerformances; empty for ordinary single-product events.
	TicketOptions []TicketOption
}

// TicketOption is one purchasable product for a performance (a label + its TM buy URL).
type TicketOption struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

// ParsedVenue is a domain representation of a Ticketmaster venue.
type ParsedVenue struct {
	TmID        string
	Name        string
	City        string
	Country     string
	CountryCode string
	Latitude    float64
	Longitude   float64
}

// CodeSet builds a country-code lookup from a comma-separated list (e.g. a user's saved
// `region.codes` setting). Blank entries are ignored; codes are lowercased.
func CodeSet(csv string) map[string]bool {
	out := map[string]bool{}
	for c := range strings.SplitSeq(csv, ",") {
		if c = strings.ToLower(strings.TrimSpace(c)); c != "" {
			out[c] = true
		}
	}
	return out
}

// InRegion reports whether a country code is part of the given region set.
func InRegion(countryCode string, region map[string]bool) bool {
	return region[strings.ToLower(countryCode)]
}

// dateString prefers the full UTC dateTime but falls back to the bare localDate, which TM
// returns (with dateTime: null) for events whose time isn't announced yet. Without this
// fallback those events get an empty start_date and render as "—".
func dateString(d DateInfo) string {
	if d.DateTime != "" {
		return d.DateTime
	}
	return d.LocalDate
}

// ParseEvent converts a Ticketmaster API event into a ParsedEvent.
// Returns nil if the event has no venue information or is a junk event (parking, add-on, etc).
func ParseEvent(e Event) *ParsedEvent {
	if e.Embedded == nil || len(e.Embedded.Venues) == 0 {
		return nil
	}

	if isJunkEvent(e.Name) {
		return nil
	}

	venue := e.Embedded.Venues[0]
	countryCode := strings.ToLower(venue.Country.CountryCode)

	parsed := &ParsedEvent{
		TmID:      e.ID,
		Name:      e.Name,
		Type:      mapEventType(e),
		Category:  mapEventCategory(e),
		StartDate: dateString(e.Dates.Start),
		URL:       e.URL,
		About:     e.Info,
	}

	if e.Dates.End != nil {
		parsed.EndDate = dateString(*e.Dates.End)
	}

	if len(e.Images) > 0 {
		parsed.ImageURL = e.Images[0].URL
	}

	if e.Embedded != nil {
		parsed.Attractions = e.Embedded.Attractions
	}

	lat, _ := strconv.ParseFloat(venue.Location.Latitude, 64)
	lng, _ := strconv.ParseFloat(venue.Location.Longitude, 64)

	parsed.Venue = &ParsedVenue{
		TmID:        venue.ID,
		Name:        venue.Name,
		City:        venue.City.Name,
		Country:     venue.Country.Name,
		CountryCode: countryCode,
		Latitude:    lat,
		Longitude:   lng,
	}

	return parsed
}

// isJunkEvent returns true if the event is a parking, add-on, or other auxiliary event.
func isJunkEvent(name string) bool {
	n := strings.ToLower(name)
	junkPatterns := []string{
		"parking",
		"ticket & hotel",
		"ticket and hotel",
		"premium ticket",
		"hospitality",
		"wine & dine",
		"wine and dine",
		"add-on",
		"vip package",
		"comfort seats",
		"wheelchair",
		"disabled access",
	}
	for _, p := range junkPatterns {
		if strings.Contains(n, p) {
			return true
		}
	}
	return false
}

// mapEventType classifies an event as Festival or Concert.
//
// There is no reliable concert-vs-festival boolean in the Ticketmaster data, and
// segment/genre names never hold {Concert,Festival,Tour} values, so this is best-effort
// and high-precision: default to Concert, only promote to Festival on a confident signal.
// Output stays within the events.type/category CHECK set.
func mapEventType(e Event) string {
	if isFestival(e) {
		return "Festival"
	}
	return "Concert"
}

// mapEventCategory mirrors the event type. Genre is intentionally not stored (it isn't
// shown); the column's CHECK only allows Concert/Festival/Tour anyway.
func mapEventCategory(e Event) string {
	return mapEventType(e)
}

// isFestival detects festivals from (1) the TM subType flag, (2) a curated name list
// (empty by default), or (3) an explicit "festival"/"fest" word in the name.
func isFestival(e Event) bool {
	for _, c := range e.Classifications {
		if strings.EqualFold(c.SubType.Name, "Festival") {
			return true
		}
	}
	// A festival is usually one of the billed attractions (e.g. "Rock Werchter Festival"),
	// even when the event title is just a ticket name like "Rock Werchter 2026 | Sunday".
	if e.Embedded != nil {
		for _, a := range e.Embedded.Attractions {
			if festivalNameRe.MatchString(a.Name) {
				return true
			}
		}
	}
	name := strings.ToLower(e.Name)
	for _, f := range FESTIVALS {
		if f != "" && strings.Contains(name, f) {
			return true
		}
	}
	return festivalNameRe.MatchString(e.Name)
}

// dayOf returns the YYYY-MM-DD portion of a start date (which may be a full dateTime).
func dayOf(date string) string {
	if len(date) >= 10 {
		return date[:10]
	}
	return date
}

// ticketLabel turns a product event name into a short option label. TM names festival
// products like "Rock Werchter 2026 | Sunday Terras"; the part after the last "|" is the
// distinguishing bit ("Sunday Terras"). Falls back to the whole name.
func ticketLabel(name string) string {
	if i := strings.LastIndex(name, "|"); i >= 0 {
		if lbl := strings.TrimSpace(name[i+1:]); lbl != "" {
			return lbl
		}
	}
	return strings.TrimSpace(name)
}

// MergePerformances folds Ticketmaster's per-ticket-product events into one row per
// performance (same venue + day). A festival appearance is listed as several products -
// day ticket, terrace, multi-day combi - that are the same performance sold differently.
// The most specific listing (fewest attractions -> a single set, not the whole-festival
// pass) becomes the row; the alternatives are attached as TicketOptions so the user can
// still choose which to buy. Input order is preserved by first appearance.
func MergePerformances(events []*ParsedEvent) []*ParsedEvent {
	order := make([]string, 0, len(events))
	groups := make(map[string][]*ParsedEvent)
	for _, e := range events {
		if e.Venue == nil {
			continue
		}
		key := strings.ToLower(e.Venue.Name) + "|" + dayOf(e.StartDate)
		if _, ok := groups[key]; !ok {
			order = append(order, key)
		}
		groups[key] = append(groups[key], e)
	}

	out := make([]*ParsedEvent, 0, len(order))
	for _, key := range order {
		members := groups[key]
		canonical := members[0]
		for _, m := range members[1:] {
			if len(m.Attractions) < len(canonical.Attractions) {
				canonical = m
			}
		}
		if len(members) > 1 {
			seen := make(map[string]bool)
			var opts []TicketOption
			for _, m := range members {
				if m.URL == "" || seen[m.URL] {
					continue
				}
				seen[m.URL] = true
				opts = append(opts, TicketOption{Label: ticketLabel(m.Name), URL: m.URL})
			}
			// Only meaningful when there really are distinct buy options.
			if len(opts) > 1 {
				canonical.TicketOptions = opts
			}
		}
		out = append(out, canonical)
	}
	return out
}
