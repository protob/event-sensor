package integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/protob/event-sensor/api"
	"github.com/protob/event-sensor/db"
	"github.com/protob/event-sensor/db/sqlc"
	"github.com/protob/event-sensor/internal/config"
	"github.com/protob/event-sensor/internal/reconcile"
)

// testApp is the binary without main(): the real router over a real migrated SQLite file
// in t.TempDir(), reached over real HTTP. Every test gets its own database and server.
type testApp struct {
	t       *testing.T
	server  *httptest.Server
	conn    *sql.DB
	queries *sqlc.Queries
	token   string
}

// newTestApp starts an app and logs in as the seeded admin. mutate adjusts the config
// before anything is opened - a non-loopback Bind or a Ticketmaster base URL, say.
func newTestApp(t *testing.T, mutate ...func(*config.Config)) *testApp {
	t.Helper()

	cfg := config.Config{
		Bind:      "127.0.0.1",
		Port:      "0",
		DBPath:    filepath.Join(t.TempDir(), "test.db"),
		JWTSecret: "test-secret",
	}
	for _, m := range mutate {
		m(&cfg)
	}

	conn, err := db.Open(cfg.DBPath)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	if err := db.Migrate(conn); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	queries := sqlc.New(conn)
	r := chi.NewRouter()
	api.Mount(r, api.NewHandler(conn, queries, &cfg))

	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)

	app := &testApp{t: t, server: srv, conn: conn, queries: queries}
	app.login("admin", "password")
	return app
}

// newTestAppWithTM starts an app whose Ticketmaster client points at a local stub.
func newTestAppWithTM(t *testing.T) (*testApp, *tmStub) {
	t.Helper()
	stub := newTMStub(t)
	app := newTestApp(t, func(c *config.Config) {
		c.TicketmasterAPIKey = "test-key"
		c.TicketmasterBaseURL = stub.URL
	})
	return app, stub
}

// do sends an authenticated JSON request and fails the test unless the status matches.
// The caller closes the body, or passes the response to decode/drain.
func (a *testApp) do(method, path string, body any, want int) *http.Response {
	a.t.Helper()
	return a.request(a.token, method, path, body, want)
}

// doAnon sends the same request without an Authorization header.
func (a *testApp) doAnon(method, path string, body any, want int) *http.Response {
	a.t.Helper()
	return a.request("", method, path, body, want)
}

func (a *testApp) request(token, method, path string, body any, want int) *http.Response {
	a.t.Helper()

	var payload io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			a.t.Fatalf("marshal %s %s body: %v", method, path, err)
		}
		payload = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, a.server.URL+path, payload)
	if err != nil {
		a.t.Fatalf("build %s %s: %v", method, path, err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := a.server.Client().Do(req)
	if err != nil {
		a.t.Fatalf("%s %s: %v", method, path, err)
	}
	if resp.StatusCode != want {
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		a.t.Fatalf("%s %s: got status %d, want %d, body: %s", method, path, resp.StatusCode, want, b)
	}
	return resp
}

func (a *testApp) login(username, password string) {
	a.t.Helper()
	resp := a.request("", http.MethodPost, "/api/auth/login",
		map[string]string{"username": username, "password": password}, http.StatusOK)
	a.token = decode[api.AuthResponseBody](a.t, resp).Token
}

// decode reads and closes a JSON response body.
func decode[T any](t *testing.T, resp *http.Response) T {
	t.Helper()
	defer resp.Body.Close()
	var out T
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return out
}

// drain closes a response whose body the test does not read.
func drain(resp *http.Response) {
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
}

// --- fixtures -----------------------------------------------------------------

func (a *testApp) createArtist(name string) api.ArtistResponse {
	a.t.Helper()
	resp := a.do(http.MethodPost, "/api/artists", map[string]any{"name": name}, http.StatusOK)
	return decode[api.ArtistResponse](a.t, resp)
}

func (a *testApp) createManualEvent(body api.ManualEventBody) api.EventResponse {
	a.t.Helper()
	resp := a.do(http.MethodPost, "/api/events", body, http.StatusOK)
	return decode[api.EventResponse](a.t, resp)
}

func (a *testApp) claim(eventID, status string) {
	a.t.Helper()
	drain(a.do(http.MethodPut, "/api/events/"+eventID+"/status",
		map[string]string{"status": status}, http.StatusOK))
}

func (a *testApp) unclaim(eventID string) {
	a.t.Helper()
	drain(a.do(http.MethodPut, "/api/events/"+eventID+"/status",
		map[string]string{"status": ""}, http.StatusOK))
}

func (a *testApp) events() []api.EventResponse {
	a.t.Helper()
	return decode[[]api.EventResponse](a.t, a.do(http.MethodGet, "/api/events", nil, http.StatusOK))
}

// eventExists reports whether an event id is still in the store, without going through
// the list endpoint's filters.
func (a *testApp) eventExists(id string) bool {
	a.t.Helper()
	var n int
	if err := a.conn.QueryRow(`SELECT count(*) FROM events WHERE id = ?`, id).Scan(&n); err != nil {
		a.t.Fatalf("count events: %v", err)
	}
	return n == 1
}

// fetch runs a reconcile for one artist and returns the counters.
func (a *testApp) fetch(artistID string) reconcile.Result {
	a.t.Helper()
	resp := a.do(http.MethodPost, "/api/artists/"+artistID+"/fetch-events", nil, http.StatusOK)
	return decode[reconcile.Result](a.t, resp)
}

// fetchLogStatuses returns the fetch_log status column in insertion order.
func (a *testApp) fetchLogStatuses() []string {
	a.t.Helper()
	rows, err := a.conn.Query(`SELECT status FROM fetch_log ORDER BY started_at, rowid`)
	if err != nil {
		a.t.Fatalf("read fetch_log: %v", err)
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			a.t.Fatalf("scan fetch_log: %v", err)
		}
		out = append(out, s)
	}
	return out
}
