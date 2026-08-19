package ticketmaster

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/protob/event-sensor/internal/provider"
)

// fetchFixture runs the whole adapter pipeline against a local server serving one
// recorded response.
func fetchFixture(t *testing.T, fixture, artist string, region map[string]bool) []*provider.ProviderEvent {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join("testdata", fixture))
	if err != nil {
		t.Fatalf("read fixture %s: %v", fixture, err)
	}
	body := expandDates(string(raw))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, body)
	}))
	t.Cleanup(srv.Close)

	events, err := NewAdapter(NewClientWithBase("test-key", srv.URL)).
		FetchByArtist(context.Background(), artist, provider.FetchOpts{Region: region})
	if err != nil {
		t.Fatalf("fetch %s: %v", fixture, err)
	}
	return events
}

var dateToken = regexp.MustCompile(`\{\{([+-])(\d+)d(T?)\}\}`)

// expandDates resolves the relative date tokens in a fixture. The integration package
// carries its own copy: a shared helper would have to live in a non-test package that
// exists for no other reason.
func expandDates(s string) string {
	return dateToken.ReplaceAllStringFunc(s, func(tok string) string {
		m := dateToken.FindStringSubmatch(tok)
		days, _ := strconv.Atoi(m[2])
		if m[1] == "-" {
			days = -days
		}
		d := time.Now().AddDate(0, 0, days)
		if m[3] == "T" {
			return d.Format("2006-01-02T15:04:05Z")
		}
		return d.Format("2006-01-02")
	})
}

func names(events []*provider.ProviderEvent) []string {
	out := make([]string, 0, len(events))
	for _, e := range events {
		out = append(out, e.Name)
	}
	return out
}

// Parking, VIP packages and hotel add-ons are products, not shows.
func TestFetchDropsJunkProducts(t *testing.T) {
	got := fetchFixture(t, "artist-junk.json", "Junk Artist", DefaultRegionCodes)

	if len(got) != 1 {
		t.Fatalf("kept %d events (%v), want only the concert", len(got), names(got))
	}
}

// Out-of-region events are never stored, so they must not cross the provider seam.
func TestFetchAppliesRegionFilter(t *testing.T) {
	all := fetchFixture(t, "artist-region.json", "Region Artist", DefaultRegionCodes)
	if len(all) != 1 {
		t.Fatalf("kept %d events (%v), want the European date only", len(all), names(all))
	}
	if all[0].Venue.CountryCode == "US" {
		t.Error("kept the United States date")
	}

	// With the US added to the region set, both survive.
	both := fetchFixture(t, "artist-region.json", "Region Artist", CodeSet("DE,US"))
	if len(both) != 2 {
		t.Errorf("kept %d events with US in region, want 2", len(both))
	}
}

// One festival sold as several products is one event; the alternatives survive as
// ticket_options.
func TestFetchMergesFestivalProducts(t *testing.T) {
	got := fetchFixture(t, "artist-festival.json", "Festival Artist", DefaultRegionCodes)

	if len(got) != 1 {
		t.Fatalf("kept %d events (%v), want one merged festival", len(got), names(got))
	}
	if got[0].Kind != "festival" {
		t.Errorf("kind = %q, want festival", got[0].Kind)
	}
	if len(got[0].TicketOptions) == 0 {
		t.Error("ticket_options is empty; the alternative products were dropped")
	}
}

// A tribute billing is kept and labelled; a same-name coincidence in another segment is
// rejected.
func TestFetchClassifiesTributeAndCoincidence(t *testing.T) {
	got := fetchFixture(t, "artist-tribute.json", "Tribute Artist", DefaultRegionCodes)

	if len(got) != 1 {
		t.Fatalf("kept %d events (%v), want the tribute only", len(got), names(got))
	}
	if got[0].Kind != "tribute" {
		t.Errorf("kind = %q, want tribute", got[0].Kind)
	}
}

// A date with an announced time keeps it; a date-only listing stays date-only.
func TestFetchKeepsDateGranularity(t *testing.T) {
	got := fetchFixture(t, "artist-simple.json", "Fixture Artist", DefaultRegionCodes)

	for _, e := range got {
		if e.HasTime != (len(e.StartDate) > len("2006-01-02")) {
			t.Errorf("%s: has_time = %v for start_date %q", e.Name, e.HasTime, e.StartDate)
		}
	}
}
