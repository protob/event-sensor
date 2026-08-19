import { api } from "./client";
import type { Venue } from "@/types";

export const venuesApi = {
  list(): Promise<Venue[]> {
    return api.get<Venue[]>("/venues");
  },

  getById(id: string): Promise<Venue> {
    return api.get<Venue>(`/venues/${id}`);
  },

  // Merge/dedupe venues: repoint events from the losers to the winner, delete the losers.
  merge(
    from: string[],
    into: string,
  ): Promise<{ into: string; repointed_events: number; merged: number }> {
    return api.post<{ into: string; repointed_events: number; merged: number }>("/venues/merge", {
      from,
      into,
    });
  },
};
