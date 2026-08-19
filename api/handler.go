package api

import (
	"database/sql"

	"github.com/protob/event-sensor/db/sqlc"
	"github.com/protob/event-sensor/internal/config"
	"github.com/protob/event-sensor/internal/reconcile"
	"github.com/protob/event-sensor/ticketmaster"
)

// Handler holds dependencies for all API handlers.
type Handler struct {
	db        *sql.DB
	queries   *sqlc.Queries
	config    *config.Config
	tmClient  *ticketmaster.Client
	reconcile *reconcile.Service
}

// NewHandler keeps the *sql.DB alongside the sqlc queries so handlers that need atomic,
// multi-statement work (e.g. the reconcile) can open a transaction and run queries via
// queries.WithTx(tx).
func NewHandler(db *sql.DB, queries *sqlc.Queries, cfg *config.Config) *Handler {
	return &Handler{
		db:        db,
		queries:   queries,
		config:    cfg,
		tmClient:  ticketmaster.NewClientWithBase(cfg.TicketmasterAPIKey, cfg.TicketmasterBaseURL),
		reconcile: reconcile.New(db, queries),
	}
}

func nullStr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

func nullFloat(nf sql.NullFloat64) *float64 {
	if nf.Valid {
		return &nf.Float64
	}
	return nil
}

func toNullStr(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func toNullFloat(f float64) sql.NullFloat64 {
	if f == 0 {
		return sql.NullFloat64{}
	}
	return sql.NullFloat64{Float64: f, Valid: true}
}
