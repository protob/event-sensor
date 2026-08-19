package integration

import (
	"encoding/base64"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/protob/event-sensor/internal/auth"
	"github.com/protob/event-sensor/internal/config"
)

// On loopback the read endpoints stay open so the app can be browsed before login.
func TestLoopbackKeepsReadsPublic(t *testing.T) {
	app := newTestApp(t)

	drain(app.doAnon(http.MethodGet, "/api/events", nil, http.StatusOK))
	drain(app.doAnon(http.MethodGet, "/api/artists", nil, http.StatusOK))
}

// Off loopback the same routes are the operator's data on a reachable interface.
func TestNonLoopbackBindMovesReadsBehindAuth(t *testing.T) {
	app := newTestApp(t, func(c *config.Config) { c.Bind = "0.0.0.0" })

	drain(app.doAnon(http.MethodGet, "/api/events", nil, http.StatusUnauthorized))
	drain(app.doAnon(http.MethodGet, "/api/artists", nil, http.StatusUnauthorized))

	// Login stays public, otherwise the instance is unusable.
	drain(app.doAnon(http.MethodPost, "/api/auth/login",
		map[string]string{"username": "admin", "password": "password"}, http.StatusOK))

	// With a token the same reads work again.
	drain(app.do(http.MethodGet, "/api/events", nil, http.StatusOK))
}

// Username and email are both public in this repository, so a reachable instance with
// the seeded password must not hand out a reset.
func TestNonLoopbackBlocksResetWhileSeededPassword(t *testing.T) {
	app := newTestApp(t, func(c *config.Config) { c.Bind = "0.0.0.0" })

	drain(app.doAnon(http.MethodPost, "/api/auth/reset-password", map[string]string{
		"username":     "admin",
		"email":        "admin@app.localdev",
		"new_password": "irrelevant",
	}, http.StatusForbidden))
}

func TestTokenRejection(t *testing.T) {
	app := newTestApp(t)

	expired, err := auth.GenerateToken("some-id", "admin", app.cfg.JWTSecret, -time.Hour)
	if err != nil {
		t.Fatalf("generate expired token: %v", err)
	}
	foreign, err := auth.GenerateToken("some-id", "admin", "another-secret", time.Hour)
	if err != nil {
		t.Fatalf("generate foreign token: %v", err)
	}

	cases := []struct {
		name  string
		authz string
	}{
		{"no header", ""},
		{"wrong scheme", "Token " + app.token},
		{"bearer without a token", "Bearer"},
		{"garbage", "Bearer not.a.token"},
		{"expired", "Bearer " + expired},
		{"signed with another secret", "Bearer " + foreign},
		{"alg none", "Bearer " + algNoneToken("some-id")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			drain(app.request(tc.authz, http.MethodGet, "/api/auth/me", nil, http.StatusUnauthorized))
		})
	}
}

// algNoneToken builds an unsigned token claiming the given subject: the classic attempt
// to have the server trust the header instead of the signature.
func algNoneToken(subject string) string {
	enc := func(s string) string {
		return base64.RawURLEncoding.EncodeToString([]byte(s))
	}
	return strings.Join([]string{
		enc(`{"alg":"none","typ":"JWT"}`),
		enc(`{"sub":"` + subject + `","username":"admin","exp":9999999999}`),
		"",
	}, ".")
}

// An oversized body is refused before a handler runs, so nothing lands in the database.
// Both BodyLimit and huma's own MaxBodyBytes cap at 1 MiB, so this pins the behaviour
// rather than either of the two layers that produce it.
func TestOversizedBodyIsRefused(t *testing.T) {
	app := newTestApp(t)

	resp := app.do(http.MethodPost, "/api/artists",
		map[string]string{"name": strings.Repeat("a", 2<<20)}, 0)
	defer drain(resp)

	if resp.StatusCode < 400 {
		t.Fatalf("status %d, want a client error", resp.StatusCode)
	}
	if n := app.countRows(`SELECT count(*) FROM artists`); n != 0 {
		t.Errorf("%d artists created from a rejected request", n)
	}
}
