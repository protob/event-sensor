package integration

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
	"testing"
	"time"
)

// tmStub stands in for the Ticketmaster Discovery API: it answers the same query shape
// with a recorded response chosen by the keyword parameter. An unknown keyword gets an
// empty result, which is what the real API returns for an artist with no listings.
type tmStub struct {
	*httptest.Server

	mu       sync.Mutex
	byArtist map[string]string
	status   int
	calls    int
}

func newTMStub(t *testing.T) *tmStub {
	t.Helper()

	s := &tmStub{byArtist: map[string]string{}}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.calls++

		if s.status != 0 {
			http.Error(w, `{"fault":{"faultstring":"stub failure"}}`, s.status)
			return
		}

		body, ok := s.byArtist[r.URL.Query().Get("keyword")]
		if !ok {
			body = `{"page":{"totalElements":0,"totalPages":0,"number":0}}`
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, body)
	}))
	t.Cleanup(s.Close)
	return s
}

// serve makes the stub answer keyword with the named file from ticketmaster/testdata.
func (s *tmStub) serve(t *testing.T, keyword, fixture string) {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join("..", "..", "ticketmaster", "testdata", fixture))
	if err != nil {
		t.Fatalf("read fixture %s: %v", fixture, err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byArtist[keyword] = expandDates(string(raw))
}

// serveRaw makes the stub answer keyword with an inline body, for a case too small to
// deserve a file.
func (s *tmStub) serveRaw(keyword, body string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byArtist[keyword] = expandDates(body)
}

// forget drops a keyword, so the next fetch sees the artist gone from the feed.
func (s *tmStub) forget(keyword string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.byArtist, keyword)
}

// failWith makes every subsequent request answer with an HTTP error.
func (s *tmStub) failWith(status int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status = status
}

func (s *tmStub) callCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.calls
}

var dateToken = regexp.MustCompile(`\{\{([+-])(\d+)d(T?)\}\}`)

// expandDates resolves {{+30d}} to a date and {{+30dT}} to a datetime, both relative to
// now. Fixtures carry tokens rather than literal dates so that past/future assertions do
// not start failing as the calendar moves.
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
