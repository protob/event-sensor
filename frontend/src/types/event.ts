export type EventStatus = "interested" | "going" | "attended" | "missed";
export type ListingState = "listed" | "delisted" | "cancelled";
export type EventKind = "concert" | "festival" | "tribute" | "club" | "other";

// Performance = one act at an event (concert OR festival), as the API sends it. artist_id is set
// for an entity artist (exists in the catalog), null/absent for a name-only act. id is what
// POST /events/{id}/artists/move takes.
export interface Performance {
  id: string;
  artist_id?: string | null;
  artist_name: string;
  is_headliner: boolean;
  date?: string | null; // which day (festivals)
  venue_name?: string | null; // per-act venue (urban festivals: Unsound's halls)
}

// EventArtist / EventLineup are DERIVED client-side from `performances`
// (see api/events.ts normalizeEvent) for the read-side consumers (artist
// filters, category tree, lineup display).
export interface EventArtist {
  artist_id: string;
  artist_name: string;
  is_headliner: boolean;
}

export interface EventLineup {
  id: string;
  artist_name: string;
}

// Alternative ticket products for the same performance (festival day / terrace / combi).
export interface TicketOption {
  label: string;
  url: string;
}

export interface Venue {
  id: string;
  name: string;
  city: string;
  country: string;
  country_code: string;
  latitude?: number;
  longitude?: number;
  timezone?: string;
  source?: string;
  event_count?: number;
  created_at: string;
  updated_at: string;
}

export interface Event {
  id: string;
  source: "ticketmaster" | "manual";
  source_id?: string | null;
  name: string;
  kind: EventKind;
  description?: string;
  start_date: string;
  end_date?: string;
  has_time: boolean;
  listing_state: ListingState;
  url?: string;
  about?: string;
  image_url?: string;
  created_at: string;
  updated_at: string;
  discovered_at: string;
  extra?: string | null;
  venue?: Venue;
  // performances = the canonical "who played" list from the API.
  performances?: Performance[];
  // artists / lineup are DERIVED from performances client-side (normalizeEvent).
  artists?: EventArtist[];
  lineup?: EventLineup[];
  ticket_options?: TicketOption[];
  // Derived (start_date < now); backend sends it, the UI may also recompute.
  is_past?: boolean;
  // Per-user library claim, client-merged from the library map (see stores/events.ts).
  status?: EventStatus | null;
  note?: string | null;
}

// A per-user claim row (the durable diary entry that makes an event persistent).
export interface LibraryEntry {
  event_id: string;
  status: EventStatus;
  missed_reason?: string | null;
  note?: string | null;
  attended_date?: string | null;
}

export interface SetStatusPayload {
  status: EventStatus | null; // null clears the claim
  missed_reason?: string;
  note?: string;
  attended_date?: string;
}

export interface ManualEventPayload {
  name: string;
  kind?: EventKind;
  start_date: string;
  end_date?: string;
  has_time?: boolean;
  about?: string;
  url?: string;
  image_url?: string;
  venue?: { name?: string; city?: string; country?: string; country_code?: string };
  // The acts: the bill (concert) or the acts you saw (festival). artist_id links an entity
  // artist (in catalog); omit it for a name-only act. date = which day (festivals).
  performances?: {
    artist_id?: string | null;
    artist_name: string;
    is_headliner?: boolean;
    date?: string | null;
    venue_name?: string | null;
  }[];
  status?: EventStatus;
}

export interface NeedsResolutionItem {
  event_id: string;
  name: string;
  start_date: string;
}

export interface FetchArtistEventsResult {
  total_found: number;
  kept: number;
  inserted: number;
  updated: number;
  deleted: number;
  delisted: number;
  saved_count: number;
  european_count: number;
}
