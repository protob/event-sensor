package api

import "github.com/protob/event-sensor/internal/reconcile"

// --- Auth types ---

type LoginInput struct {
	Body struct {
		Username string `json:"username" required:"true" doc:"Username"`
		Password string `json:"password" required:"true" doc:"Password"`
	}
}

type LoginOutput struct {
	Body AuthResponseBody
}

type AuthResponseBody struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type MeOutput struct {
	Body UserResponse
}

type ChangePasswordInput struct {
	Body struct {
		CurrentPassword string `json:"current_password" required:"true" doc:"Current password"`
		NewPassword     string `json:"new_password" required:"true" minLength:"6" doc:"New password (min 6 characters)"`
	}
}

type ChangePasswordOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}

type UpdateProfileInput struct {
	Body struct {
		Username string `json:"username,omitempty" doc:"New username"`
		Email    string `json:"email,omitempty" doc:"New email"`
	}
}

type UpdateProfileOutput struct {
	Body UserResponse
}

type ResetPasswordInput struct {
	Body struct {
		Username    string `json:"username" required:"true" doc:"Username of the account"`
		Email       string `json:"email" required:"true" doc:"Email address of the account"`
		NewPassword string `json:"new_password" required:"true" minLength:"6" doc:"New password (min 6 characters)"`
	}
}

type ResetPasswordOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}

type UserResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	Active    bool   `json:"active"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// --- Settings types ---

type ListSettingsOutput struct {
	Body []SettingResponse
}

type SettingResponse struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	UpdatedAt string `json:"updated_at"`
}

type UpdateSettingsInput struct {
	Body struct {
		Settings []SettingEntry `json:"settings" required:"true" doc:"List of settings to update"`
	}
}

type SettingEntry struct {
	Key   string `json:"key" required:"true" doc:"Setting key"`
	Value string `json:"value" required:"true" doc:"Setting value"`
}

// --- Category types ---

type ListCategoriesOutput struct {
	Body []CategoryResponse
}

type CategoryResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ArtistCount int    `json:"artist_count"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type CreateCategoryInput struct {
	Body struct {
		Name   string `json:"name" required:"true" minLength:"1" doc:"Category name"`
		UserID string `json:"user_id" required:"true" doc:"User ID"`
	}
}

type CreateCategoryOutput struct {
	Body CategoryResponse
}

type UpdateCategoryInput struct {
	ID   string `path:"id" required:"true" doc:"Category ID"`
	Body struct {
		Name string `json:"name" required:"true" minLength:"1" doc:"New category name"`
	}
}

type UpdateCategoryOutput struct {
	Body CategoryResponse
}

type DeleteCategoryInput struct {
	ID string `path:"id" required:"true" doc:"Category ID"`
}

type DeleteCategoryOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}

// --- Artist types ---

type ListArtistsInput struct{}

type ListArtistsOutput struct {
	Body []ArtistResponse
}

type ArtistResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	FetchMode string `json:"fetch_mode"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type GetArtistInput struct {
	ID string `path:"id" required:"true" doc:"Artist ID"`
}

type GetArtistOutput struct {
	Body ArtistResponse
}

type ListArtistSummariesInput struct{}

type ListArtistSummariesOutput struct {
	Body []ArtistSummaryResponse
}

// ArtistSummaryResponse is per-artist aggregate the list views need: which of the owner's
// categories the artist is in, where their upcoming shows are, and how many. Category is the
// owner's own grouping - no genre data is stored anywhere in this app.
type ArtistSummaryResponse struct {
	ArtistID            string         `json:"artist_id"`
	Categories          []CategoryRef  `json:"categories"`
	Countries           []CountryCount `json:"countries"`
	EventCount          int            `json:"event_count"`
	UpcomingListedCount int            `json:"upcoming_listed_count"`
	ClaimedCount        int            `json:"claimed_count"`
}

type CategoryRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CountryCount struct {
	Code  string `json:"code"`
	Count int    `json:"count"`
}

type CreateArtistInput struct {
	Body struct {
		Name      string `json:"name" required:"true" minLength:"1" doc:"Artist name"`
		FetchMode string `json:"fetch_mode,omitempty" enum:"auto,manual" doc:"auto (fetched by Fetch All) or manual (fetched only on demand); defaults to auto"`
	}
}

type SetFetchModeInput struct {
	ID   string `path:"id" required:"true" doc:"Artist ID"`
	Body struct {
		FetchMode string `json:"fetch_mode" required:"true" enum:"auto,manual" doc:"auto or manual"`
	}
}

type SetFetchModeOutput struct {
	Body ArtistResponse
}

type CreateArtistOutput struct {
	Body ArtistResponse
}

type UpdateArtistInput struct {
	ID   string `path:"id" required:"true" doc:"Artist ID"`
	Body struct {
		Name string `json:"name" required:"true" minLength:"1" doc:"New artist name"`
	}
}

type UpdateArtistOutput struct {
	Body ArtistResponse
}

type DeleteArtistInput struct {
	ID string `path:"id" required:"true" doc:"Artist ID"`
}

type DeleteArtistOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}

type ListCategoryArtistsInput struct {
	ID string `path:"id" required:"true" doc:"Category ID"`
}

type ListCategoryArtistsOutput struct {
	Body []ArtistResponse
}

type AddArtistToCategoryInput struct {
	ID   string `path:"id" required:"true" doc:"Category ID"`
	Body struct {
		ArtistID string `json:"artist_id" required:"true" doc:"Artist ID to add"`
	}
}

type AddArtistToCategoryOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}

type RemoveArtistFromCategoryInput struct {
	CategoryID string `path:"categoryId" required:"true" doc:"Category ID"`
	ArtistID   string `path:"artistId" required:"true" doc:"Artist ID"`
}

type RemoveArtistFromCategoryOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}

// --- Fetch events types ---

type FetchArtistEventsInput struct {
	ID string `path:"id" required:"true" doc:"Artist ID"`
}

type FetchArtistEventsOutput struct {
	Body reconcile.Result
}

type ListArtistEventsInput struct {
	ID string `path:"id" required:"true" doc:"Artist ID"`
}

type ListArtistEventsOutput struct {
	Body []EventResponse
}

type ClearArtistEventsInput struct {
	ID string `path:"id" required:"true" doc:"Artist ID"`
}

type ClearArtistEventsOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}

// --- Event types ---

type ListEventsInput struct {
	StartDate string `query:"start_date" doc:"Filter events starting from this date (ISO 8601)"`
	EndDate   string `query:"end_date" doc:"Filter events up to this date (ISO 8601)"`
	Query     string `query:"q" doc:"Search events by name, venue, or artist"`
}

type ListEventsOutput struct {
	Body []EventResponse
}

type EventResponse struct {
	ID           string  `json:"id"`
	Source       string  `json:"source"`
	SourceID     *string `json:"source_id,omitempty"`
	Name         string  `json:"name"`
	Kind         string  `json:"kind"`
	Description  *string `json:"description,omitempty"`
	StartDate    string  `json:"start_date"`
	EndDate      *string `json:"end_date,omitempty"`
	HasTime      bool    `json:"has_time"`
	ListingState string  `json:"listing_state"`
	VenueID      *string `json:"venue_id,omitempty"`
	URL          *string `json:"url,omitempty"`
	About        *string `json:"about,omitempty"`
	ImageURL     *string `json:"image_url,omitempty"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
	DiscoveredAt string  `json:"discovered_at"`   // first-surfaced timestamp (for "what's new")
	Extra        *string `json:"extra,omitempty"` // raw JSON bag; nil/omitted when "{}"
	// is_past is derived (start_date < now), sent so the UI need not recompute it.
	IsPast bool `json:"is_past"`
	// Per-user library claim, merged in when authenticated. status is null when unclaimed.
	Status        *string               `json:"status,omitempty"`
	Note          *string               `json:"note,omitempty"`
	Venue         *VenueResponse         `json:"venue,omitempty"`
	Performances  []PerformanceResponse  `json:"performances,omitempty"`
	TicketOptions []TicketOptionResponse `json:"ticket_options,omitempty"`
}

// TicketOptionResponse is one purchasable product for an event (label + buy URL).
type TicketOptionResponse struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

type GetEventInput struct {
	ID string `path:"id" required:"true" doc:"Event ID"`
}

type GetEventOutput struct {
	Body EventResponse
}

type DeleteEventInput struct {
	ID string `path:"id" required:"true" doc:"Event ID"`
}

type DeleteEventOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}

// --- Manual event creation ---

type CreateManualEventInput struct {
	Body ManualEventBody
}

type ManualEventBody struct {
	Name      string                 `json:"name" required:"true" minLength:"1" doc:"Event name"`
	Kind      string                 `json:"kind,omitempty" enum:"concert,festival,tribute,club,other" doc:"Event kind; defaults to other"`
	StartDate string                 `json:"start_date" required:"true" doc:"ISO-8601 date or datetime"`
	EndDate   string                 `json:"end_date,omitempty"`
	HasTime   bool                   `json:"has_time,omitempty" doc:"true if start_date carries a time"`
	About     string                 `json:"about,omitempty"`
	URL       string                 `json:"url,omitempty"`
	ImageURL  string                 `json:"image_url,omitempty"`
	Venue        *ManualVenueBody   `json:"venue,omitempty"`
	Performances []PerformanceInput `json:"performances,omitempty" doc:"Artists: the bill (concert) or the artists you saw (festival). Link an entity artist via artist_id, or leave it unset for a name-only string."`
	Status       string             `json:"status,omitempty" enum:"interested,going,attended,missed" doc:"Optional initial library claim"`
}

type ManualVenueBody struct {
	Name        string `json:"name,omitempty"`
	City        string `json:"city,omitempty"`
	Country     string `json:"country,omitempty"`
	CountryCode string `json:"country_code,omitempty"`
}

type CreateManualEventOutput struct {
	Body EventResponse
}

type UpdateEventInput struct {
	ID   string `path:"id" required:"true" doc:"Event ID (manual events only)"`
	Body ManualEventBody
}

type UpdateEventOutput struct {
	Body EventResponse
}

type ClearAllEventsInput struct{}

type ClearAllEventsOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}

type PrunePastTmEventsInput struct{}

type PrunePastTmEventsOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}

// PerformanceResponse is one artist at an event. ArtistID is set when it is an entity artist
// (in your catalog); nil when it is a name-only performer. ID is what
// POST /events/{id}/artists/move takes - without it that endpoint is unreachable.
type PerformanceResponse struct {
	ID          string  `json:"id"`
	ArtistID    *string `json:"artist_id,omitempty"`
	ArtistName  string  `json:"artist_name"`
	IsHeadliner bool    `json:"is_headliner"`
	Date        *string `json:"date,omitempty"`       // which day at the festival
	VenueName   *string `json:"venue_name,omitempty"` // per-act venue (urban festivals)
}

// PerformanceInput is one artist on a manual event's bill/lineup. ArtistName is required;
// ArtistID is optional (link to an entity artist in the catalog). Date is the day (festivals).
type PerformanceInput struct {
	ArtistID    *string `json:"artist_id,omitempty" doc:"Link to an entity artist in your catalog; omit for a name-only string"`
	ArtistName  string  `json:"artist_name" required:"true" minLength:"1"`
	IsHeadliner bool    `json:"is_headliner,omitempty"`
	Date        *string `json:"date,omitempty" doc:"Which day at the festival; YYYY-MM-DD"`
	VenueName   *string `json:"venue_name,omitempty" doc:"This artist's venue (urban festivals like Unsound); omit to use the event venue"`
}

// --- Library (per-user claims: interested / going / attended / missed) ---

// LibraryEntryResponse is one per-user claim row (the diary entry for an event).
type LibraryEntryResponse struct {
	EventID      string  `json:"event_id"`
	Status       string  `json:"status"`
	MissedReason *string `json:"missed_reason,omitempty"`
	Note         *string `json:"note,omitempty"`
	AttendedDate *string `json:"attended_date,omitempty"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

type ListLibraryInput struct{}

type ListLibraryOutput struct {
	Body []LibraryEntryResponse
}

// SetEventStatusInput sets/clears the per-user claim. status="" (or null) unclaims the event.
type SetEventStatusInput struct {
	ID   string `path:"id" required:"true" doc:"Event ID"`
	Body struct {
		Status       string `json:"status" enum:"interested,going,attended,missed," doc:"Claim status; empty string unclaims"`
		MissedReason string `json:"missed_reason,omitempty"`
		Note         string `json:"note,omitempty"`
		AttendedDate string `json:"attended_date,omitempty"`
	}
}

type SetEventStatusOutput struct {
	Body LibraryEntryResponse
}

type NeedsResolutionItem struct {
	EventID   string `json:"event_id"`
	Name      string `json:"name"`
	StartDate string `json:"start_date"`
}

type NeedsResolutionOutput struct {
	Body []NeedsResolutionItem
}

// --- Venue types ---

type ListVenuesInput struct{}

type ListVenuesOutput struct {
	Body []VenueResponse
}

type VenueResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	City        string   `json:"city"`
	Country     string   `json:"country"`
	CountryCode string   `json:"country_code"`
	Latitude    *float64 `json:"latitude,omitempty"`
	Longitude   *float64 `json:"longitude,omitempty"`
	Timezone    *string  `json:"timezone,omitempty"`
	Source      *string  `json:"source,omitempty"`
	EventCount  int      `json:"event_count"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

type GetVenueInput struct {
	ID string `path:"id" required:"true" doc:"Venue ID"`
}

type GetVenueOutput struct {
	Body VenueResponse
}

// --- Bulk / collection types (cross-entity operations) ---

type MergeArtistsInput struct {
	Body struct {
		From []string `json:"from" required:"true" minItems:"1" doc:"Artist IDs to merge away (deleted)"`
		Into string   `json:"into" required:"true" doc:"Surviving artist ID"`
	}
}
type MergeArtistsOutput struct {
	Body struct {
		Into   string `json:"into"`
		Merged int    `json:"merged"` // losers deleted
	}
}

type MergeVenuesInput struct {
	Body struct {
		From []string `json:"from" required:"true" minItems:"1"`
		Into string   `json:"into" required:"true"`
	}
}
type MergeVenuesOutput struct {
	Body struct {
		Into            string `json:"into"`
		RepointedEvents int    `json:"repointed_events"`
		Merged          int    `json:"merged"`
	}
}

type MoveArtistsInput struct {
	ID   string `path:"id" required:"true" doc:"Source event ID"`
	Body struct {
		PerformanceIDs []string `json:"performance_ids" required:"true" minItems:"1"`
		TargetEventID  string   `json:"target_event_id" required:"true"`
	}
}
type MoveArtistsOutput struct {
	Body struct {
		Moved int `json:"moved"`
	}
}

type AssignCategoriesInput struct {
	Body struct {
		ArtistIDs []string `json:"artist_ids" required:"true" minItems:"1"`
		Add       []string `json:"add,omitempty"`    // category IDs to add
		Remove    []string `json:"remove,omitempty"` // category IDs to remove
	}
}
type AssignCategoriesOutput struct {
	Body struct {
		Artists int `json:"artists"`
	}
}

type BulkStatusInput struct {
	Body struct {
		EventIDs     []string `json:"event_ids" required:"true" minItems:"1"`
		Status       string   `json:"status" enum:"interested,going,attended,missed," doc:"empty unclaims"`
		MissedReason string   `json:"missed_reason,omitempty"`
		Note         string   `json:"note,omitempty"`
		AttendedDate string   `json:"attended_date,omitempty"`
	}
}
type BulkStatusOutput struct {
	Body struct {
		Updated int `json:"updated"`
	}
}

type BulkFetchModeInput struct {
	Body struct {
		ArtistIDs []string `json:"artist_ids" required:"true" minItems:"1"`
		FetchMode string   `json:"fetch_mode" required:"true" enum:"auto,manual"`
	}
}
type BulkFetchModeOutput struct {
	Body struct {
		Updated int `json:"updated"`
	}
}

type BulkDeleteEventsInput struct {
	Body struct {
		EventIDs []string `json:"event_ids" required:"true" minItems:"1"`
	}
}
type BulkDeleteEventsOutput struct {
	Body struct {
		Deleted     int `json:"deleted"`
		KeptClaimed int `json:"kept_claimed"`
	}
}
