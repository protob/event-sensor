import { api } from "./client";
import { normalizeEvent } from "./events";
import type {
  Artist,
  ArtistSummary,
  CreateArtistPayload,
  UpdateArtistPayload,
  FetchMode,
  Event,
  FetchArtistEventsResult,
  MessageResponse,
} from "@/types";

export const artistsApi = {
  list(): Promise<Artist[]> {
    return api.get<Artist[]>("/artists");
  },

  summaries(): Promise<ArtistSummary[]> {
    return api.get<ArtistSummary[]>("/artists/summary");
  },

  getById(id: string): Promise<Artist> {
    return api.get<Artist>(`/artists/${id}`);
  },

  create(payload: CreateArtistPayload): Promise<Artist> {
    return api.post<Artist>("/artists", payload);
  },

  update(id: string, payload: UpdateArtistPayload): Promise<Artist> {
    return api.put<Artist>(`/artists/${id}`, payload);
  },

  remove(id: string): Promise<MessageResponse> {
    return api.delete<MessageResponse>(`/artists/${id}`);
  },

  fetchEvents(id: string): Promise<FetchArtistEventsResult> {
    return api.post<FetchArtistEventsResult>(`/artists/${id}/fetch-events`);
  },

  async listEvents(id: string): Promise<Event[]> {
    const events = await api.get<Event[]>(`/artists/${id}/events`);
    return events.map(normalizeEvent);
  },

  clearEvents(id: string): Promise<MessageResponse> {
    return api.delete<MessageResponse>(`/artists/${id}/events`);
  },

  setFetchMode(id: string, mode: FetchMode): Promise<Artist> {
    return api.put<Artist>(`/artists/${id}/fetch-mode`, { fetch_mode: mode });
  },

  // --- Bulk / collection operations ---

  merge(from: string[], into: string): Promise<{ into: string; merged: number }> {
    return api.post<{ into: string; merged: number }>("/artists/merge", { from, into });
  },

  bulkFetchMode(artist_ids: string[], fetch_mode: FetchMode): Promise<{ updated: number }> {
    return api.put<{ updated: number }>("/artists/fetch-mode", { artist_ids, fetch_mode });
  },
};
