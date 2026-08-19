package integration

import (
	"testing"
)

// A fetch stores what the pipeline kept and records the run.
func TestFetchStoresEvents(t *testing.T) {
	app, stub := newTestAppWithTM(t)

	artist := app.createArtist("Fixture Artist")
	stub.serve(t, "Fixture Artist", "artist-fixture-artist.json")

	res := app.fetch(artist.ID)

	if res.Inserted != 2 {
		t.Errorf("inserted = %d, want 2", res.Inserted)
	}
	if got := len(app.events()); got != 2 {
		t.Errorf("stored events = %d, want 2", got)
	}
	if got := app.fetchLogStatuses(); len(got) != 1 || got[0] != "success" {
		t.Errorf("fetch_log = %v, want one success row", got)
	}
	if stub.callCount() != 1 {
		t.Errorf("stub calls = %d, want 1", stub.callCount())
	}
}

// An artist absent from the feed is not an error: the fetch succeeds and stores nothing.
func TestFetchWithNoResults(t *testing.T) {
	app, _ := newTestAppWithTM(t)

	artist := app.createArtist("Unknown Artist")
	res := app.fetch(artist.ID)

	if res.TotalFound != 0 || res.Inserted != 0 {
		t.Errorf("result = %+v, want an empty fetch", res)
	}
	if got := len(app.events()); got != 0 {
		t.Errorf("stored events = %d, want 0", got)
	}
}
