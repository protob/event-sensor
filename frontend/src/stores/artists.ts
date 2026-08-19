import { ref } from "vue";
import { defineStore } from "pinia";
import type {
  Artist,
  ArtistSummary,
  CreateArtistPayload,
  UpdateArtistPayload,
  FetchMode,
  FetchArtistEventsResult,
  Event,
} from "@/types";
import { artistsApi, categoriesApi, eventsApi } from "@/api";
import { useToast } from "@/composables/useToast";
import { errMessage } from "@/utils/apiError";
import { useEventsStore } from "./events";
import { useCategoriesStore } from "./categories";
import { useAuthStore } from "./auth";
import { useUiStore } from "./ui";

export const useArtistsStore = defineStore("artists", () => {
  const artists = ref<Artist[]>([]);
  const currentArtist = ref<Artist | null>(null);
  const loading = ref(false);
  const error = ref<string | null>(null);
  const fetchResult = ref<FetchArtistEventsResult | null>(null);
  const summaries = ref<Record<string, ArtistSummary>>({});
  // Events of currentArtist (the detail view's list). Local snapshot, refreshed by
  // fetchAndSync / listArtistEvents / clearArtistEvents.
  const artistEvents = ref<Event[]>([]);

  async function fetchSummaries() {
    try {
      const list = await artistsApi.summaries();
      const next: Record<string, ArtistSummary> = {};
      for (const s of list) next[s.artist_id] = s;
      summaries.value = next;
    } catch (err) {
      // Non-fatal: the views fall back to their own derivation when a summary is missing.
      error.value = errMessage(err, "Failed to fetch artist summaries");
    }
  }

  async function fetchArtists() {
    loading.value = true;
    error.value = null;
    try {
      artists.value = await artistsApi.list();
    } catch (err) {
      error.value = errMessage(err, "Failed to fetch artists");
    } finally {
      loading.value = false;
    }
  }

  async function fetchArtist(id: string) {
    loading.value = true;
    error.value = null;
    try {
      currentArtist.value = await artistsApi.getById(id);
    } catch (err) {
      error.value = errMessage(err, "Failed to fetch artist");
    } finally {
      loading.value = false;
    }
  }

  async function createArtist(payload: CreateArtistPayload) {
    error.value = null;
    try {
      const artist = await artistsApi.create(payload);
      artists.value.push(artist);
      return artist;
    } catch (err) {
      error.value = errMessage(err, "Failed to create artist");
      return null;
    }
  }

  async function updateArtist(id: string, payload: UpdateArtistPayload) {
    error.value = null;
    try {
      const updated = await artistsApi.update(id, payload);
      const idx = artists.value.findIndex((a) => a.id === id);
      if (idx !== -1) {
        artists.value[idx] = updated;
      }
      return updated;
    } catch (err) {
      error.value = errMessage(err, "Failed to update artist");
      return null;
    }
  }

  async function deleteArtist(id: string) {
    error.value = null;
    try {
      await artistsApi.remove(id);
      artists.value = artists.value.filter((a) => a.id !== id);
      return true;
    } catch (err) {
      error.value = errMessage(err, "Failed to delete artist");
      return false;
    }
  }

  async function deleteArtistConfirmed(id: string, name: string) {
    if (!confirm(`Delete artist "${name}"?`)) return false;
    const ok = await deleteArtist(id);
    if (ok) useToast().success("Artist deleted");
    return ok;
  }

  async function setFetchMode(id: string, mode: FetchMode) {
    error.value = null;
    const idx = artists.value.findIndex((a) => a.id === id);
    const prev = idx !== -1 ? artists.value[idx].fetch_mode : undefined;
    const patch = (m?: FetchMode) => {
      if (!m) return;
      if (idx !== -1) artists.value[idx] = { ...artists.value[idx], fetch_mode: m };
      if (currentArtist.value?.id === id)
        currentArtist.value = { ...currentArtist.value, fetch_mode: m };
    };
    patch(mode);
    try {
      const updated = await artistsApi.setFetchMode(id, mode);
      if (idx !== -1) artists.value[idx] = updated;
      if (currentArtist.value?.id === id) currentArtist.value = updated;
      return updated;
    } catch (err) {
      patch(prev);
      error.value = errMessage(err, "Failed to set fetch mode");
      return null;
    }
  }

  // --- Bulk / collection operations ---

  // Bulk set fetch-mode - pending then refresh list (counts/badges follow).
  async function bulkFetchMode(ids: string[], mode: FetchMode) {
    if (ids.length === 0) return;
    error.value = null;
    try {
      const res = await artistsApi.bulkFetchMode(ids, mode);
      await fetchArtists();
      useToast().success(`Set ${res.updated} artist${res.updated === 1 ? "" : "s"} to ${mode}`);
    } catch (err) {
      const msg = errMessage(err, "Failed to set fetch mode");
      error.value = msg;
      useToast().error(msg);
    }
  }

  // Merge artists - destructive (losers deleted). Repointed performances change the
  // events list and category memberships, so both refresh here too.
  async function merge(from: string[], into: string) {
    if (from.length === 0 || !into) return null;
    error.value = null;
    try {
      const res = await artistsApi.merge(from, into);
      await Promise.all([
        fetchArtists(),
        useEventsStore().fetchEvents(),
        useCategoriesStore().fetchCategories(),
      ]);
      useToast().success(`Merged ${res.merged} artist${res.merged === 1 ? "" : "s"}`);
      return res;
    } catch (err) {
      const msg = errMessage(err, "Failed to merge artists");
      error.value = msg;
      useToast().error(msg);
      return null;
    }
  }

  // Bulk add/remove category memberships for a set of artists.
  async function assignCategories(ids: string[], add?: string[], remove?: string[]) {
    if (ids.length === 0) return;
    error.value = null;
    try {
      await categoriesApi.assign(ids, add, remove);
      await fetchArtists();
      const verb = add?.length ? "Added to" : "Removed from";
      useToast().success(`${verb} category for ${ids.length} artist${ids.length === 1 ? "" : "s"}`);
    } catch (err) {
      const msg = errMessage(err, "Failed to update categories");
      error.value = msg;
      useToast().error(msg);
    }
  }

  async function fetchArtistEvents(id: string) {
    loading.value = true;
    error.value = null;
    fetchResult.value = null;
    try {
      fetchResult.value = await artistsApi.fetchEvents(id);
      return fetchResult.value;
    } catch (err) {
      error.value = errMessage(err, "Failed to fetch events from Ticketmaster");
      return null;
    } finally {
      loading.value = false;
    }
  }

  // Shared post-fetch resync: stored events changed, so the feed and the claims map
  // reload, and the "last fetch" marker moves.
  async function refreshAfterFetch() {
    const events = useEventsStore();
    await events.fetchEvents();
    if (useAuthStore().isAuthenticated) await events.loadLibrary();
    useUiStore().setLastFetch();
  }

  // Single-artist fetch with everything that must follow it: feed + claims resync,
  // detail-list reload when open, toast with the outcome.
  async function fetchAndSync(id: string, opts?: { alsoRefreshList?: boolean }) {
    const name =
      artists.value.find((a) => a.id === id)?.name ?? currentArtist.value?.name ?? "artist";
    const res = await fetchArtistEvents(id);
    if (!res) {
      useToast().error(`Fetch failed for ${name}`);
      return null;
    }
    await refreshAfterFetch();
    if (currentArtist.value?.id === id) await listArtistEvents(id);
    if (opts?.alsoRefreshList) await fetchArtists();
    useToast().success(`${name}: ${res.saved_count} saved · ${res.european_count} in region`);
    return res;
  }

  async function listArtistEvents(id: string) {
    const evs = await artistsApi.listEvents(id);
    if (currentArtist.value?.id === id) artistEvents.value = evs;
    return evs;
  }

  async function clearArtistEvents(id: string) {
    await artistsApi.clearEvents(id);
    if (currentArtist.value?.id === id) artistEvents.value = [];
    await useEventsStore().fetchEvents();
    useToast().success("Artist events cleared");
  }

  async function clearAllEvents() {
    await eventsApi.clearAll();
    await Promise.all([fetchArtists(), useEventsStore().fetchEvents()]);
    useToast().success("Un-claimed events cleared");
  }

  return {
    artists,
    currentArtist,
    loading,
    error,
    fetchResult,
    summaries,
    artistEvents,
    fetchSummaries,
    fetchArtists,
    fetchArtist,
    createArtist,
    updateArtist,
    deleteArtist,
    deleteArtistConfirmed,
    setFetchMode,
    bulkFetchMode,
    merge,
    assignCategories,
    fetchArtistEvents,
    refreshAfterFetch,
    fetchAndSync,
    listArtistEvents,
    clearArtistEvents,
    clearAllEvents,
  };
});
