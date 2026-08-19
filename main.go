package main

import (
	"context"
	"database/sql"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/protob/event-sensor/api"
	"github.com/protob/event-sensor/db"
	"github.com/protob/event-sensor/db/sqlc"
	"github.com/protob/event-sensor/internal/auth"
	"github.com/protob/event-sensor/internal/config"
)

// warnDefaultAdminPassword logs when any account still authenticates with the
// seeded "password". Off-loopback that login is the front door, so the operator
// hears about it at startup instead of after someone else does.
func warnDefaultAdminPassword(db *sql.DB) {
	rows, err := db.Query(`SELECT username, password FROM users`)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var username, hash string
		if err := rows.Scan(&username, &hash); err != nil {
			return
		}
		if auth.CheckPassword(hash, "password") == nil {
			log.Printf("WARNING: user %q still has the seeded default password - change it before exposing this instance", username)
		}
	}
}

func main() {
	cfg := config.Load()

	if cfg.UsingDefaultSecret() {
		log.Println("WARNING: JWT_SECRET is the public default from the repository - set JWT_SECRET for anything beyond loopback")
	}
	// The off-machine case is what makes a public default secret dangerous rather
	// than merely noisy: refuse it instead of warning.
	if !cfg.IsLoopbackBind() {
		if cfg.UsingDefaultSecret() {
			log.Fatalf("refusing to start: non-loopback bind (%s) with the default JWT_SECRET - set JWT_SECRET (e.g. openssl rand -hex 32)", cfg.Bind)
		}
	}

	// SQLite will not create a missing parent directory, and a fresh clone has no data/ yet.
	// 0700: the db file holds bcrypt hashes and the per-user Ticketmaster key - no reason
	// for any other local account to traverse or read it.
	if dir := filepath.Dir(cfg.DBPath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			log.Fatalf("failed to create database directory %s: %v", dir, err)
		}
	}

	conn, err := db.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer conn.Close()

	if err := db.Migrate(conn); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	// Create SQLC queries and handler
	queries := sqlc.New(conn)

	if !cfg.IsLoopbackBind() {
		warnDefaultAdminPassword(conn)
	}

	handler := api.NewHandler(conn, queries, &cfg)

	// Setup router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	api.Mount(r, handler)

	// Serve embedded frontend (SPA fallback)
	frontendDist, err := frontendFS()
	if err != nil {
		log.Fatalf("failed to load embedded frontend: %v", err)
	}
	fileServer := http.FileServer(http.FS(frontendDist))
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		if _, err := fs.Stat(frontendDist, r.URL.Path[1:]); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}
		// SPA fallback: serve index.html for all unmatched routes
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})

	// Start server
	srv := &http.Server{
		Addr:         net.JoinHostPort(cfg.Bind, cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Event Sensor API starting on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-done
	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown failed: %v", err)
	}
	log.Println("server stopped")
}
