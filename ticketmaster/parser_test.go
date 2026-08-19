package ticketmaster

import "testing"

func TestMergePerformances(t *testing.T) {
	venue := &ParsedVenue{Name: "Werchter", City: "Werchter", CountryCode: "be"}
	// Three ticket products for the same Cure appearance at Rock Werchter, 2026-07-05.
	in := []*ParsedEvent{
		{Name: "Rock Werchter 2026 | Sunday", StartDate: "2026-07-05", URL: "u1", Venue: venue,
			Attractions: make([]Attraction, 11)},
		{Name: "Rock Werchter 2026 | Sunday Terras", StartDate: "2026-07-05", URL: "u2", Venue: venue,
			Attractions: make([]Attraction, 11)},
		{Name: "2 July - 5 July 2026 | Rock Werchter Combi", StartDate: "2026-07-05", URL: "u3", Venue: venue,
			Attractions: make([]Attraction, 52)},
		// A genuinely separate gig at another venue/day survives untouched.
		{Name: "The Cure", StartDate: "2026-10-04", URL: "u4",
			Venue: &ParsedVenue{Name: "3Arena", City: "Dublin", CountryCode: "ie"}},
	}

	out := MergePerformances(in)
	if len(out) != 2 {
		t.Fatalf("expected 2 performances, got %d", len(out))
	}
	werchter := out[0]
	if len(werchter.TicketOptions) != 3 {
		t.Fatalf("expected 3 ticket options, got %d", len(werchter.TicketOptions))
	}
	// Canonical = most specific listing (11 attractions, not the 52-attraction combi).
	if werchter.URL != "u1" {
		t.Errorf("canonical URL = %q, want u1 (single day)", werchter.URL)
	}
	if werchter.TicketOptions[2].Label != "Rock Werchter Combi" {
		t.Errorf("combi label = %q", werchter.TicketOptions[2].Label)
	}
	if len(out[1].TicketOptions) != 0 {
		t.Errorf("standalone gig should have no ticket options, got %d", len(out[1].TicketOptions))
	}
}

// Cases mirror real Ticketmaster shapes observed for these keyword searches.
func TestClassifyArtistMatch(t *testing.T) {
	tests := []struct {
		name   string
		artist string
		event  ParsedEvent
		want   ArtistMatch
	}{
		{
			name:   "billed artist, long tour title",
			artist: "bonobo",
			event: ParsedEvent{
				Name:        "Bonobo: Distance in Static Live North American Tour",
				Attractions: []Attraction{{Name: "Bonobo"}},
			},
			want: MatchArtist,
		},
		{
			name:   "coincidental substring, no attraction → rejected",
			artist: "bonobo",
			event: ParsedEvent{
				Name:        "LES BONOBOS",
				Attractions: nil,
			},
			want: MatchNone,
		},
		{
			name:   "tribute to the artist, not billed",
			artist: "howard shore",
			event: ParsedEvent{
				Name:        "The Music of The Lord of The Rings. Tribute to Howard Shore",
				Attractions: []Attraction{{Name: "The Music of The Lord of The Rings. Tribute to Howard Shore"}},
			},
			want: MatchTribute,
		},
		{
			name:   "music of X performed by others",
			artist: "philip glass",
			event: ParsedEvent{
				Name:        "Music of Philip Glass performed by City Symphony Orchestra",
				Attractions: []Attraction{{Name: "City Symphony Orchestra"}},
			},
			want: MatchTribute,
		},
		{
			name:   "real artist whose title says 'plays' is not a tribute",
			artist: "radiohead",
			event: ParsedEvent{
				Name:        "Radiohead plays OK Computer",
				Attractions: []Attraction{{Name: "Radiohead"}},
			},
			want: MatchArtist,
		},
		{
			name:   "co-headliner: artist is one of several attractions",
			artist: "the roots",
			event: ParsedEvent{
				Name:        "Some Festival",
				Attractions: []Attraction{{Name: "Headliner"}, {Name: "The Roots"}},
			},
			want: MatchArtist,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ClassifyArtistMatch(&tc.event, tc.artist); got != tc.want {
				t.Errorf("ClassifyArtistMatch(%q) = %v, want %v", tc.artist, got, tc.want)
			}
		})
	}
}
