export type FetchMode = "auto" | "manual";

export interface Artist {
  id: string;
  name: string;
  fetch_mode: FetchMode;
  created_at: string;
  updated_at: string;
}

// Per-artist aggregate from GET /artists/summary. `categories` are the owner's own
// categories - this app stores no genre data.
export interface ArtistSummary {
  artist_id: string;
  categories: { id: string; name: string }[];
  countries: { code: string; count: number }[];
  event_count: number;
  upcoming_listed_count: number;
  claimed_count: number;
}

export interface CreateArtistPayload {
  name: string;
  fetch_mode?: FetchMode;
}

export interface UpdateArtistPayload {
  name: string;
}
