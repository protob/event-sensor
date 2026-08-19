package integration

import (
	"fmt"
	"net/http"
	"testing"
)

// oneEvent is the smallest response the pipeline accepts: one in-region future concert
// billed to the searched artist.
func oneEvent(sourceID, artist string) string {
	return tmEvent(sourceID, artist, artist, 40)
}

// tmEvent builds that response with the event title and the start-date offset spelled out
// (days is negative for a show that has already happened). The billed attraction stays the
// searched artist, which is what carries the event past the keyword-collision guard.
func tmEvent(sourceID, artist, title string, days int) string {
	return `{"_embedded":{"events":[{
		"id":"` + sourceID + `",
		"name":"` + title + `",
		"type":"event",
		"url":"https://example.invalid/e/` + sourceID + `",
		"dates":{"start":{"localDate":"` + fmt.Sprintf("{{%+dd}}", days) + `","dateTime":"` + fmt.Sprintf("{{%+ddT}}", days) + `"}},
		"_embedded":{
			"venues":[{"id":"V1","name":"Columbiahalle","city":{"name":"Berlin"},
				"country":{"name":"Germany","countryCode":"DE"},
				"location":{"latitude":"52.4938","longitude":"13.3903"}}],
			"attractions":[{"id":"A1","name":"` + artist + `"}]}
	}]},"page":{"totalElements":1,"totalPages":1,"number":0}}`
}

// A claim outlives the listing. The event leaves the feed, the sweep must not delete it,
// and it is marked delisted instead.
func TestClaimSurvivesRefetch(t *testing.T) {
	app, stub := newTestAppWithTM(t)
	const name = "Fixture Artist"

	artist := app.createArtist(name)
	stub.serve(t, name, "artist-simple.json")
	app.fetch(artist.ID)

	events := app.events()
	if len(events) != 2 {
		t.Fatalf("first fetch stored %d events, want 2", len(events))
	}
	claimed, dropped := events[0], events[1]
	app.claim(claimed.ID, "going")

	// Second run: only the claimed event's source id is gone from the feed.
	stub.serveRaw(name, oneEvent(*dropped.SourceID, name))
	res := app.fetch(artist.ID)

	if !app.eventExists(claimed.ID) {
		t.Fatal("claimed event was swept")
	}
	if res.Delisted != 1 {
		t.Errorf("delisted = %d, want 1", res.Delisted)
	}

	after := app.eventByID(claimed.ID)
	if after.ListingState != "delisted" {
		t.Errorf("listing_state = %q, want delisted", after.ListingState)
	}
	if got := app.claimStatus(claimed.ID); got != "going" {
		t.Errorf("claim status = %q, want going", got)
	}
}

// The upsert is keyed on (source, source_id), so the surrogate id - and therefore the
// claim attached to it - survives a refetch that changes the facts.
func TestSurrogateIDSurvivesUpsert(t *testing.T) {
	app, stub := newTestAppWithTM(t)
	const name = "Stable Artist"

	artist := app.createArtist(name)
	stub.serveRaw(name, oneEvent("TM-1", name))
	app.fetch(artist.ID)

	first := app.events()[0]
	app.claim(first.ID, "interested")

	stub.serveRaw(name, tmEvent("TM-1", name, name+" (moved)", 40))
	res := app.fetch(artist.ID)

	if res.Updated != 1 || res.Inserted != 0 {
		t.Errorf("counters = %+v, want one update and no insert", res)
	}

	after := app.events()
	if len(after) != 1 {
		t.Fatalf("stored %d events, want 1", len(after))
	}
	if after[0].ID != first.ID {
		t.Errorf("event id changed from %s to %s", first.ID, after[0].ID)
	}
	if got := app.claimStatus(first.ID); got != "interested" {
		t.Errorf("claim lost: status = %q", got)
	}
}

// An un-claimed event that leaves the feed is deleted. This is the control for the two
// tests above: without a claim, nothing protects it.
func TestUnclaimedAbsentEventIsSwept(t *testing.T) {
	app, stub := newTestAppWithTM(t)
	const name = "Sweep Artist"

	artist := app.createArtist(name)
	stub.serveRaw(name, oneEvent("TM-1", name))
	app.fetch(artist.ID)
	gone := app.events()[0]

	stub.serveRaw(name, `{"page":{"totalElements":0,"totalPages":0,"number":0}}`)
	res := app.fetch(artist.ID)

	if res.Deleted != 1 {
		t.Errorf("deleted = %d, want 1", res.Deleted)
	}
	if app.eventExists(gone.ID) {
		t.Fatal("un-claimed absent event survived the sweep")
	}
}

// An event billed to two artists in the catalog is not deleted by a sweep run for one of
// them. The second performance is inserted directly: the reconcile owns Ticketmaster
// performances and replaces them per fetch, so no endpoint produces this state.
func TestCoHeadlinerEventSurvivesSweep(t *testing.T) {
	app, stub := newTestAppWithTM(t)
	const name = "Headliner A"

	artistA := app.createArtist(name)
	artistB := app.createArtist("Headliner B")
	stub.serveRaw(name, oneEvent("TM-1", name))
	app.fetch(artistA.ID)

	shared := app.events()[0]
	app.addPerformance(shared.ID, artistB.ID, "Headliner B")

	stub.serveRaw(name, `{"page":{"totalElements":0,"totalPages":0,"number":0}}`)
	res := app.fetch(artistA.ID)

	if res.Deleted != 0 {
		t.Errorf("deleted = %d, want 0", res.Deleted)
	}
	if !app.eventExists(shared.ID) {
		t.Fatal("event with a second billed artist was swept")
	}
}

// The past sweep runs inside the fetch transaction, so a past date in the feed is dropped
// by the very run that inserts it.
func TestPastEventFromFeedIsNotStored(t *testing.T) {
	app, stub := newTestAppWithTM(t)
	const name = "History Artist"

	artist := app.createArtist(name)
	stub.serve(t, name, "artist-past.json")
	res := app.fetch(artist.ID)

	if res.Inserted != 2 {
		t.Errorf("inserted = %d, want both feed entries", res.Inserted)
	}

	stored := app.events()
	if len(stored) != 1 {
		t.Fatalf("stored %d events, want only the future one", len(stored))
	}
	if stored[0].IsPast {
		t.Error("the past event was kept")
	}
}

// A claim survives the show itself: once the date has passed, the past sweep keeps the
// event because it is claimed, and takes it as soon as the claim is dropped.
func TestPastSweepKeepsClaimed(t *testing.T) {
	app, stub := newTestAppWithTM(t)
	const name = "Diary Artist"

	artist := app.createArtist(name)
	stub.serveRaw(name, oneEvent("TM-1", name))
	app.fetch(artist.ID)

	ev := app.events()[0]
	app.claim(ev.ID, "attended")

	// The show happens: the same listing now carries a past date.
	stub.serveRaw(name, tmEvent("TM-1", name, name, -2))
	app.fetch(artist.ID)

	if !app.eventExists(ev.ID) {
		t.Fatal("claimed past event was swept")
	}
	if !app.eventByID(ev.ID).IsPast {
		t.Error("event did not move into the past")
	}

	app.unclaim(ev.ID)
	app.fetch(artist.ID)

	if app.eventExists(ev.ID) {
		t.Error("un-claimed past event survived the past sweep")
	}
}

// A Ticketmaster failure is logged and changes nothing. The previous feed stays as it is.
func TestFetchErrorLeavesDataUntouched(t *testing.T) {
	app, stub := newTestAppWithTM(t)
	const name = "Flaky Artist"

	artist := app.createArtist(name)
	stub.serveRaw(name, oneEvent("TM-1", name))
	app.fetch(artist.ID)
	before := app.events()

	stub.failWith(http.StatusInternalServerError)
	drain(app.do(http.MethodPost, "/api/artists/"+artist.ID+"/fetch-events", nil, http.StatusInternalServerError))

	after := app.events()
	if len(after) != len(before) || after[0].ID != before[0].ID || after[0].UpdatedAt != before[0].UpdatedAt {
		t.Errorf("events changed after a failed fetch: %+v -> %+v", before, after)
	}
	if got := app.fetchLogStatuses(); len(got) != 2 || got[1] != "error" {
		t.Errorf("fetch_log = %v, want a success then an error row", got)
	}
}
