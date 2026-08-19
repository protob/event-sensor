import { api } from "./client";
import type {
  Event,
  EventStatus,
  EventFilterParams,
  LibraryEntry,
  SetStatusPayload,
  ManualEventPayload,
  NeedsResolutionItem,
  MessageResponse,
} from "@/types";

// The API sends `performances` (the canonical "who played"). Derive the read-side shapes
// the UI consumes: `artists` = entity performances (artist_id set), `lineup` = the full
// bill (every performer name). One place, applied to every Event the API returns.
export function normalizeEvent(e: Event): Event {
  const perfs = e.performances ?? [];
  e.artists = perfs
    .filter((p) => p.artist_id)
    .map((p) => ({
      artist_id: p.artist_id as string,
      artist_name: p.artist_name,
      is_headliner: p.is_headliner,
    }));
  e.lineup = perfs.map((p, i) => ({ id: `${e.id}-${i}`, artist_name: p.artist_name }));
  return e;
}

export const eventsApi = {
  async list(params?: EventFilterParams): Promise<Event[]> {
    const events = await api.get<Event[]>("/events", { query: params });
    return events.map(normalizeEvent);
  },

  async search(query: string): Promise<Event[]> {
    const events = await api.get<Event[]>("/events", { query: { q: query } });
    return events.map(normalizeEvent);
  },

  async getById(id: string): Promise<Event> {
    return normalizeEvent(await api.get<Event>(`/events/${id}`));
  },

  remove(id: string): Promise<MessageResponse> {
    return api.delete<MessageResponse>(`/events/${id}`);
  },

  clearAll(): Promise<MessageResponse> {
    return api.delete<MessageResponse>("/events");
  },

  prunePastTm(): Promise<MessageResponse> {
    return api.delete<MessageResponse>("/events/past");
  },

  // The diary index: every per-user claim (interested / going / attended / missed).
  listLibrary(): Promise<LibraryEntry[]> {
    return api.get<LibraryEntry[]>("/library");
  },

  // Past `going` events awaiting a "did you go?" answer.
  listNeedsResolution(): Promise<NeedsResolutionItem[]> {
    return api.get<NeedsResolutionItem[]>("/library/needs-resolution");
  },

  // Set or clear (status:null) the per-user claim for an event.
  setStatus(id: string, payload: SetStatusPayload): Promise<LibraryEntry> {
    return api.put<LibraryEntry>(`/events/${id}/status`, {
      ...payload,
      status: payload.status ?? "",
    });
  },

  async createManualEvent(payload: ManualEventPayload): Promise<Event> {
    return normalizeEvent(await api.post<Event>("/events", payload));
  },

  // Edit a manual event (Ticketmaster events are rejected with 409 by the backend).
  async updateEvent(id: string, payload: ManualEventPayload): Promise<Event> {
    return normalizeEvent(await api.put<Event>(`/events/${id}`, payload));
  },

  // --- Bulk / collection operations ---

  // Bulk set/clear (status:null) the per-user claim for many events.
  bulkSetStatus(
    event_ids: string[],
    status: EventStatus | null,
    extra?: { missed_reason?: string; note?: string; attended_date?: string },
  ): Promise<{ updated: number }> {
    return api.put<{ updated: number }>("/library/status", {
      event_ids,
      status: status ?? "",
      ...extra,
    });
  },

  // Bulk delete by IDs (claimed events are kept server-side).
  bulkDelete(event_ids: string[]): Promise<{ deleted: number; kept_claimed: number }> {
    return api.post<{ deleted: number; kept_claimed: number }>("/events/bulk-delete", {
      event_ids,
    });
  },
};
