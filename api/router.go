package api

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Mount registers the API under /api with the middleware the binary runs. main.go adds
// the request logger and the SPA fallback around it.
func Mount(r chi.Router, h *Handler) {
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(BodyLimit(1 << 20))

	r.Route("/api", func(r chi.Router) {
		SetupRouter(r, h)
	})
}

// SetupRouter creates the Huma API and registers all routes.
func SetupRouter(r chi.Router, h *Handler) huma.API {
	api := humachi.New(r, huma.DefaultConfig("Event Sensor API", "1.0.0"))

	// The account is seeded (migration 003), never self-registered - there is no
	// register endpoint. Off-loopback the public read routes move behind auth too:
	// a publicly reachable instance is the operator's data, not a browsable feed.
	readsRequireAuth := !h.config.IsLoopbackBind()

	// --- Auth routes (public) ---
	huma.Register(api, huma.Operation{
		OperationID: "login",
		Method:      "POST",
		Path:        "/auth/login",
		Summary:     "Login",
		Tags:        []string{"Auth"},
	}, h.Login)

	huma.Register(api, huma.Operation{
		OperationID: "reset-password",
		Method:      "POST",
		Path:        "/auth/reset-password",
		Summary:     "Reset password using username and email verification",
		Tags:        []string{"Auth"},
	}, h.ResetPassword)

	// --- Auth routes (protected) ---
	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware(h.config.JWTSecret))
		protectedAPI := humachi.New(r, huma.DefaultConfig("Event Sensor API", "1.0.0"))

		huma.Register(protectedAPI, huma.Operation{
			OperationID: "get-me",
			Method:      "GET",
			Path:        "/auth/me",
			Summary:     "Get current user",
			Tags:        []string{"Auth"},
		}, h.Me)

		huma.Register(protectedAPI, huma.Operation{
			OperationID: "change-password",
			Method:      "POST",
			Path:        "/auth/change-password",
			Summary:     "Change password",
			Tags:        []string{"Auth"},
		}, h.ChangePassword)

		huma.Register(protectedAPI, huma.Operation{
			OperationID: "update-profile",
			Method:      "PUT",
			Path:        "/auth/profile",
			Summary:     "Update user profile",
			Tags:        []string{"Auth"},
		}, h.UpdateProfile)
	})

	// --- Read routes ---
	// Loopback: public (browse-before-login). Off-loopback: same handlers, but
	// behind auth like everything else.
	readRouter := r
	if readsRequireAuth {
		readRouter = r.Group(func(r chi.Router) {
			r.Use(AuthMiddleware(h.config.JWTSecret))
		})
	}
	readAPI := humachi.New(readRouter, huma.DefaultConfig("Event Sensor API", "1.0.0"))

	// Categories (read - list artists in a category)
	huma.Register(readAPI, huma.Operation{
		OperationID: "list-category-artists",
		Method:      "GET",
		Path:        "/categories/{id}/artists",
		Summary:     "List artists in category",
		Tags:        []string{"Categories"},
	}, h.ListCategoryArtists)

	// Artists (read)
	huma.Register(readAPI, huma.Operation{
		OperationID: "list-artists",
		Method:      "GET",
		Path:        "/artists",
		Summary:     "List artists",
		Tags:        []string{"Artists"},
	}, h.ListArtists)

	huma.Register(readAPI, huma.Operation{
		OperationID: "get-artist",
		Method:      "GET",
		Path:        "/artists/{id}",
		Summary:     "Get artist",
		Tags:        []string{"Artists"},
	}, h.GetArtist)

	huma.Register(readAPI, huma.Operation{
		OperationID: "list-artist-events",
		Method:      "GET",
		Path:        "/artists/{id}/events",
		Summary:     "List events for artist",
		Tags:        []string{"Artists"},
	}, h.ListArtistDBEvents)

	// Events (read)
	huma.Register(readAPI, huma.Operation{
		OperationID: "list-events",
		Method:      "GET",
		Path:        "/events",
		Summary:     "List events",
		Tags:        []string{"Events"},
	}, h.ListEvents)

	huma.Register(readAPI, huma.Operation{
		OperationID: "get-event",
		Method:      "GET",
		Path:        "/events/{id}",
		Summary:     "Get event details",
		Tags:        []string{"Events"},
	}, h.GetEvent)

	// Venues (read)
	huma.Register(readAPI, huma.Operation{
		OperationID: "list-venues",
		Method:      "GET",
		Path:        "/venues",
		Summary:     "List venues",
		Tags:        []string{"Venues"},
	}, h.ListVenues)

	huma.Register(readAPI, huma.Operation{
		OperationID: "get-venue",
		Method:      "GET",
		Path:        "/venues/{id}",
		Summary:     "Get venue",
		Tags:        []string{"Venues"},
	}, h.GetVenue)

	// --- Protected write routes ---
	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware(h.config.JWTSecret))
		protectedAPI := humachi.New(r, huma.DefaultConfig("Event Sensor API", "1.0.0"))

		// Categories (read, protected)
		huma.Register(protectedAPI, huma.Operation{
			OperationID: "list-categories",
			Method:      "GET",
			Path:        "/categories",
			Summary:     "List categories for authenticated user",
			Tags:        []string{"Categories"},
		}, h.ListCategories)

		// Categories (write)
		huma.Register(protectedAPI, huma.Operation{
			OperationID: "create-category",
			Method:      "POST",
			Path:        "/categories",
			Summary:     "Create category",
			Tags:        []string{"Categories"},
		}, h.CreateCategory)

		huma.Register(protectedAPI, huma.Operation{
			OperationID: "update-category",
			Method:      "PUT",
			Path:        "/categories/{id}",
			Summary:     "Update category",
			Tags:        []string{"Categories"},
		}, h.UpdateCategory)

		huma.Register(protectedAPI, huma.Operation{
			OperationID: "delete-category",
			Method:      "DELETE",
			Path:        "/categories/{id}",
			Summary:     "Delete category",
			Tags:        []string{"Categories"},
		}, h.DeleteCategory)

		huma.Register(protectedAPI, huma.Operation{
			OperationID: "add-artist-to-category",
			Method:      "POST",
			Path:        "/categories/{id}/artists",
			Summary:     "Add artist to category",
			Tags:        []string{"Categories"},
		}, h.AddArtistToCategory)

		huma.Register(protectedAPI, huma.Operation{
			OperationID: "remove-artist-from-category",
			Method:      "DELETE",
			Path:        "/categories/{categoryId}/artists/{artistId}",
			Summary:     "Remove artist from category",
			Tags:        []string{"Categories"},
		}, h.RemoveArtistFromCategory)

		// Artists (write)
		huma.Register(protectedAPI, huma.Operation{
			OperationID: "list-artist-summaries",
			Method:      "GET",
			Path:        "/artists/summary",
			Summary:     "Per-artist categories, countries and counts in one call",
			Tags:        []string{"Artists"},
		}, h.ListArtistSummaries)

		huma.Register(protectedAPI, huma.Operation{
			OperationID: "create-artist",
			Method:      "POST",
			Path:        "/artists",
			Summary:     "Create artist",
			Tags:        []string{"Artists"},
		}, h.CreateArtist)

		huma.Register(protectedAPI, huma.Operation{
			OperationID: "update-artist",
			Method:      "PUT",
			Path:        "/artists/{id}",
			Summary:     "Update artist",
			Tags:        []string{"Artists"},
		}, h.UpdateArtist)

		huma.Register(protectedAPI, huma.Operation{
			OperationID: "delete-artist",
			Method:      "DELETE",
			Path:        "/artists/{id}",
			Summary:     "Delete artist",
			Tags:        []string{"Artists"},
		}, h.DeleteArtist)

		huma.Register(protectedAPI, huma.Operation{
			OperationID: "fetch-artist-events",
			Method:      "POST",
			Path:        "/artists/{id}/fetch-events",
			Summary:     "Fetch events from Ticketmaster for artist",
			Tags:        []string{"Artists"},
		}, h.FetchArtistEvents)

		huma.Register(protectedAPI, huma.Operation{
			OperationID: "clear-artist-events",
			Method:      "DELETE",
			Path:        "/artists/{id}/events",
			Summary:     "Clear an artist's un-claimed events",
			Tags:        []string{"Artists"},
		}, h.ClearArtistEvents)

		huma.Register(protectedAPI, huma.Operation{
			OperationID: "set-artist-fetch-mode",
			Method:      "PUT",
			Path:        "/artists/{id}/fetch-mode",
			Summary:     "Set artist fetch mode (auto | manual)",
			Tags:        []string{"Artists"},
		}, h.SetArtistFetchMode)

		// Library (per-user claims: interested / going / attended / missed)
		huma.Register(protectedAPI, huma.Operation{
			OperationID: "list-library",
			Method:      "GET",
			Path:        "/library",
			Summary:     "List current user's library entries",
			Tags:        []string{"Library"},
		}, h.ListLibrary)

		huma.Register(protectedAPI, huma.Operation{
			OperationID: "list-needs-resolution",
			Method:      "GET",
			Path:        "/library/needs-resolution",
			Summary:     "List past 'going' events awaiting attendance confirmation",
			Tags:        []string{"Library"},
		}, h.ListNeedsResolution)

		huma.Register(protectedAPI, huma.Operation{
			OperationID: "set-event-status",
			Method:      "PUT",
			Path:        "/events/{id}/status",
			Summary:     "Set or clear the per-user library claim for an event",
			Tags:        []string{"Library"},
		}, h.SetEventStatus)

		// Events (write)
		huma.Register(protectedAPI, huma.Operation{
			OperationID: "create-manual-event",
			Method:      "POST",
			Path:        "/events",
			Summary:     "Create a manual (hand-entered) event",
			Tags:        []string{"Events"},
		}, h.CreateManualEvent)

		huma.Register(protectedAPI, huma.Operation{
			OperationID: "update-event",
			Method:      "PUT",
			Path:        "/events/{id}",
			Summary:     "Edit a manual event (409 for Ticketmaster events)",
			Tags:        []string{"Events"},
		}, h.UpdateEvent)

		huma.Register(protectedAPI, huma.Operation{
			OperationID: "delete-event",
			Method:      "DELETE",
			Path:        "/events/{id}",
			Summary:     "Delete event (409 if claimed)",
			Tags:        []string{"Events"},
		}, h.DeleteEvent)

		huma.Register(protectedAPI, huma.Operation{
			OperationID: "clear-all-events",
			Method:      "DELETE",
			Path:        "/events",
			Summary:     "Clear all un-claimed events (keeps artists + claimed)",
			Tags:        []string{"Events"},
		}, h.ClearAllEvents)

		huma.Register(protectedAPI, huma.Operation{
			OperationID: "prune-past-tm-events",
			Method:      "DELETE",
			Path:        "/events/past",
			Summary:     "Delete past unclaimed Ticketmaster events",
			Tags:        []string{"Events"},
		}, h.PrunePastTmEvents)

		// --- Bulk / collection operations (cross-entity repointing) ---
		huma.Register(protectedAPI, huma.Operation{OperationID: "merge-artists", Method: "POST",
			Path: "/artists/merge", Summary: "Merge artists (repoint + delete losers)", Tags: []string{"Artists"}}, h.MergeArtists)
		huma.Register(protectedAPI, huma.Operation{OperationID: "bulk-artist-fetch-mode", Method: "PUT",
			Path: "/artists/fetch-mode", Summary: "Bulk set artist fetch mode", Tags: []string{"Artists"}}, h.BulkSetFetchMode)
		huma.Register(protectedAPI, huma.Operation{OperationID: "merge-venues", Method: "POST",
			Path: "/venues/merge", Summary: "Merge/dedupe venues", Tags: []string{"Venues"}}, h.MergeVenues)
		huma.Register(protectedAPI, huma.Operation{OperationID: "move-artists", Method: "POST",
			Path: "/events/{id}/artists/move", Summary: "Move performances to another event", Tags: []string{"Events"}}, h.MoveArtists)
		huma.Register(protectedAPI, huma.Operation{OperationID: "bulk-delete-events", Method: "POST",
			Path: "/events/bulk-delete", Summary: "Delete many events by ID (claimed kept)", Tags: []string{"Events"}}, h.BulkDeleteEvents)
		huma.Register(protectedAPI, huma.Operation{OperationID: "assign-categories", Method: "POST",
			Path: "/categories/assign", Summary: "Bulk add/remove artists ↔ categories", Tags: []string{"Categories"}}, h.AssignCategories)
		huma.Register(protectedAPI, huma.Operation{OperationID: "bulk-status", Method: "PUT",
			Path: "/library/status", Summary: "Bulk set/clear library claim status", Tags: []string{"Library"}}, h.BulkSetStatus)

		// Settings
		huma.Register(protectedAPI, huma.Operation{
			OperationID: "list-settings",
			Method:      "GET",
			Path:        "/settings",
			Summary:     "List user settings",
			Tags:        []string{"Settings"},
		}, h.ListSettings)

		huma.Register(protectedAPI, huma.Operation{
			OperationID: "update-settings",
			Method:      "PUT",
			Path:        "/settings",
			Summary:     "Update user settings",
			Tags:        []string{"Settings"},
		}, h.UpdateSettings)
	})

	return api
}
