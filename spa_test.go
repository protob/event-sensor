package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/go-chi/chi/v5"
)

func spaServer(t *testing.T) *httptest.Server {
	t.Helper()

	dist := fstest.MapFS{
		"index.html":    {Data: []byte("<!doctype html><title>app</title>")},
		"assets/app.js": {Data: []byte("console.log(1)")},
	}

	r := chi.NewRouter()
	r.Route("/api", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})
	})
	mountSPA(r, dist)

	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)
	return srv
}

func get(t *testing.T, url string) (int, string) {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(body)
}

func TestSPAServesAssetsAndFallsBack(t *testing.T) {
	srv := spaServer(t)

	if code, body := get(t, srv.URL+"/assets/app.js"); code != 200 || body != "console.log(1)" {
		t.Errorf("asset: %d %q", code, body)
	}

	// A deep link is a client-side route, so the shell has to come back with 200.
	code, body := get(t, srv.URL+"/events/abc-123")
	if code != 200 {
		t.Errorf("deep link: status %d, want 200", code)
	}
	if !strings.Contains(body, "<title>app</title>") {
		t.Error("deep link did not return the app shell")
	}
}

// A mistyped endpoint must fail loudly. If the catch-all answered it, the client would
// parse HTML as JSON and report something unrelated.
func TestSPADoesNotSwallowAPIRoutes(t *testing.T) {
	srv := spaServer(t)

	code, body := get(t, srv.URL+"/api/no-such-endpoint")
	if code != http.StatusNotFound {
		t.Errorf("status %d, want 404", code)
	}
	if strings.Contains(body, "<title>app</title>") {
		t.Error("the SPA fallback answered an /api path")
	}
}
