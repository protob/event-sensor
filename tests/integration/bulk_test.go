package integration

import (
	"net/http"
	"testing"

	"github.com/protob/event-sensor/api"
)

// Merging artists repoints what pointed at the losers and deletes them, leaving no
// performance row referencing an artist that is gone.
func TestMergeArtistsLeavesNoDanglingRows(t *testing.T) {
	app, stub := newTestAppWithTM(t)

	winner := app.createArtist("Winner")
	loser := app.createArtist("Loser")
	stub.serveRaw("Loser", oneEvent("TM-1", "Loser"))
	app.fetch(loser.ID)

	drain(app.do(http.MethodPost, "/api/artists/merge", map[string]any{
		"from": []string{loser.ID},
		"into": winner.ID,
	}, http.StatusOK))

	if app.countRows(`SELECT count(*) FROM artists WHERE id = ?`, loser.ID) != 0 {
		t.Error("loser artist survived the merge")
	}
	if n := app.countRows(`SELECT count(*) FROM performances WHERE artist_id = ?`, loser.ID); n != 0 {
		t.Errorf("%d performances still point at the merged-away artist", n)
	}
	if n := app.countRows(`SELECT count(*) FROM performances WHERE artist_id = ?`, winner.ID); n != 1 {
		t.Errorf("winner has %d performances, want 1", n)
	}
}

// Merging venues repoints the events and deletes the losers. The two spellings below are
// what the endpoint exists for: resolveManualVenue only folds names that match exactly
// (after trimming), so a stray space or a missing one mints a second row.
func TestMergeVenuesRepointsEvents(t *testing.T) {
	app := newTestApp(t)

	a := app.createManualEvent(api.ManualEventBody{
		Name: "Show A", StartDate: futureDate(20),
		Venue: &api.ManualVenueBody{Name: "Columbiahalle", City: "Berlin", CountryCode: "DE"},
	})
	b := app.createManualEvent(api.ManualEventBody{
		Name: "Show B", StartDate: futureDate(21),
		Venue: &api.ManualVenueBody{Name: "Columbia Halle", City: "Berlin", CountryCode: "DE"},
	})

	venueA, venueB := *app.eventByID(a.ID).VenueID, *app.eventByID(b.ID).VenueID
	if venueA == venueB {
		t.Fatal("both events landed on one venue; there is nothing to merge")
	}

	drain(app.do(http.MethodPost, "/api/venues/merge", map[string]any{
		"from": []string{venueB},
		"into": venueA,
	}, http.StatusOK))

	if *app.eventByID(b.ID).VenueID != venueA {
		t.Error("event was not repointed to the surviving venue")
	}
	if app.countRows(`SELECT count(*) FROM venues WHERE id = ?`, venueB) != 0 {
		t.Error("merged-away venue survived")
	}
}

// A bulk delete keeps the claimed events and says how many it kept.
func TestBulkDeleteKeepsClaimed(t *testing.T) {
	app := newTestApp(t)

	claimed := app.createManualEvent(api.ManualEventBody{Name: "Keep", StartDate: futureDate(5)})
	plain := app.createManualEvent(api.ManualEventBody{Name: "Drop", StartDate: futureDate(6)})
	app.claim(claimed.ID, "going")

	res := decode[struct {
		Deleted     int `json:"deleted"`
		KeptClaimed int `json:"kept_claimed"`
	}](t, app.do(http.MethodPost, "/api/events/bulk-delete", map[string]any{
		"event_ids": []string{claimed.ID, plain.ID},
	}, http.StatusOK))

	if res.Deleted != 1 || res.KeptClaimed != 1 {
		t.Errorf("result = %+v, want one deleted and one kept", res)
	}
	if !app.eventExists(claimed.ID) {
		t.Error("claimed event was bulk-deleted")
	}
	if app.eventExists(plain.ID) {
		t.Error("un-claimed event survived the bulk delete")
	}
}

// A bulk status change claims every event in one transaction, and the empty status
// clears them again.
func TestBulkSetStatusRoundTrip(t *testing.T) {
	app := newTestApp(t)

	first := app.createManualEvent(api.ManualEventBody{Name: "One", StartDate: pastDate(20)})
	second := app.createManualEvent(api.ManualEventBody{Name: "Two", StartDate: pastDate(21)})
	ids := []string{first.ID, second.ID}

	drain(app.do(http.MethodPut, "/api/library/status",
		map[string]any{"event_ids": ids, "status": "attended"}, http.StatusOK))

	if n := app.countRows(`SELECT count(*) FROM library_entries WHERE status = 'attended'`); n != 2 {
		t.Errorf("claimed %d events, want 2", n)
	}

	drain(app.do(http.MethodPut, "/api/library/status",
		map[string]any{"event_ids": ids, "status": ""}, http.StatusOK))

	if n := app.countRows(`SELECT count(*) FROM library_entries`); n != 0 {
		t.Errorf("%d claims left after clearing", n)
	}
}
