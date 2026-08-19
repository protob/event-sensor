package integration

import (
	"net/http"
	"testing"
	"time"

	"github.com/protob/event-sensor/api"
)

func futureDate(days int) string {
	return time.Now().AddDate(0, 0, days).Format("2006-01-02")
}

func pastDate(days int) string {
	return time.Now().AddDate(0, 0, -days).Format("2006-01-02")
}

// A claimed event cannot be deleted: library_entries.event_id is ON DELETE RESTRICT and
// the handler maps that to 409. This also detects a database opened without the
// foreign_keys pragma, where the delete would succeed.
func TestClaimedEventDeleteReturns409(t *testing.T) {
	app := newTestApp(t)

	ev := app.createManualEvent(api.ManualEventBody{Name: "Claimed show", StartDate: futureDate(30)})
	app.claim(ev.ID, "going")

	drain(app.do(http.MethodDelete, "/api/events/"+ev.ID, nil, http.StatusConflict))

	if !app.eventExists(ev.ID) {
		t.Fatal("claimed event was deleted")
	}
}

// Unclaiming removes the only thing pinning the event, so the same delete then succeeds.
func TestUnclaimedEventDeletes(t *testing.T) {
	app := newTestApp(t)

	ev := app.createManualEvent(api.ManualEventBody{Name: "Droppable show", StartDate: futureDate(30)})
	app.claim(ev.ID, "interested")
	app.unclaim(ev.ID)

	drain(app.do(http.MethodDelete, "/api/events/"+ev.ID, nil, http.StatusOK))

	if app.eventExists(ev.ID) {
		t.Fatal("un-claimed event survived a delete")
	}
}

// "Clear all events" drops everything nobody claimed and keeps the rest.
func TestClearAllEventsKeepsClaimed(t *testing.T) {
	app := newTestApp(t)

	kept := app.createManualEvent(api.ManualEventBody{Name: "Kept", StartDate: futureDate(10)})
	dropped := app.createManualEvent(api.ManualEventBody{Name: "Dropped", StartDate: futureDate(11)})
	app.claim(kept.ID, "attended")

	drain(app.do(http.MethodDelete, "/api/events", nil, http.StatusOK))

	if !app.eventExists(kept.ID) {
		t.Error("claimed event was cleared")
	}
	if app.eventExists(dropped.ID) {
		t.Error("un-claimed event survived the clear")
	}
}

// The past purge is scoped to Ticketmaster events; a hand-entered past show is the diary
// and is never touched by it.
func TestPastPurgeKeepsManualEvents(t *testing.T) {
	app := newTestApp(t)

	ev := app.createManualEvent(api.ManualEventBody{Name: "Old gig", StartDate: pastDate(400)})

	drain(app.do(http.MethodDelete, "/api/events/past", nil, http.StatusOK))

	if !app.eventExists(ev.ID) {
		t.Fatal("manual past event was purged")
	}
}

// A claim carries the status the user set and shows up in the diary index.
func TestClaimAppearsInLibrary(t *testing.T) {
	app := newTestApp(t)

	ev := app.createManualEvent(api.ManualEventBody{Name: "Diary entry", StartDate: pastDate(3)})
	app.claim(ev.ID, "attended")

	entries := decode[[]api.LibraryEntryResponse](t,
		app.do(http.MethodGet, "/api/library", nil, http.StatusOK))

	var found bool
	for _, e := range entries {
		if e.EventID == ev.ID {
			found = true
			if e.Status != "attended" {
				t.Errorf("status = %q, want attended", e.Status)
			}
		}
	}
	if !found {
		t.Fatal("claimed event missing from /api/library")
	}
}
