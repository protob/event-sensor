import { ref, computed } from "vue";
import { defineStore } from "pinia";
import type {
  Event,
  EventFilterParams,
  EventStatus,
  LibraryEntry,
  SetStatusPayload,
  ManualEventPayload,
  NeedsResolutionItem,
} from "@/types";
import { eventsApi } from "@/api";
import { useToast } from "@/composables/useToast";
import { errMessage } from "@/utils/apiError";

export const useEventsStore = defineStore("events", () => {
  const events = ref<Event[]>([]);
  const currentEvent = ref<Event | null>(null);
  const loading = ref(false);
  const error = ref<string | null>(null);

  // event_id -> per-user claim. The durable side of the Feed/Library split.
  const libraryMap = ref<Record<string, LibraryEntry>>({});
  const needsResolution = ref<NeedsResolutionItem[]>([]);
  // id -> full event (about/lineup/tickets), for expanded rows. List payloads omit them.
  const detailCache = ref<Record<string, Event>>({});

  // Per-status counts (status bar + library tabs).
  const claimedCount = computed(() => Object.keys(libraryMap.value).length);
  const countByStatus = (s: EventStatus) =>
    Object.values(libraryMap.value).filter((e) => e.status === s).length;
  const interestedCount = computed(() => countByStatus("interested"));
  const goingCount = computed(() => countByStatus("going"));
  const attendedCount = computed(() => countByStatus("attended"));
  const missedCount = computed(() => countByStatus("missed"));

  // Events the user has claimed in any status (the Events "Interested" pill = claimed).
  const claimedEvents = computed(() => events.value.filter((e) => e.status));

  function statusOf(id: string): EventStatus | null {
    return libraryMap.value[id]?.status ?? null;
  }

  // Merge the claim map onto loaded events so rows render the right toggle state.
  function applyLibrary() {
    for (const e of events.value) {
      const le = libraryMap.value[e.id];
      e.status = le?.status ?? null;
      e.note = le?.note ?? null;
    }
    if (currentEvent.value) {
      const le = libraryMap.value[currentEvent.value.id];
      currentEvent.value.status = le?.status ?? null;
      currentEvent.value.note = le?.note ?? null;
    }
  }

  async function fetchEvents(params?: EventFilterParams) {
    loading.value = true;
    error.value = null;
    try {
      events.value = await eventsApi.list(params);
      applyLibrary();
    } catch (err) {
      error.value = errMessage(err, "Failed to fetch events");
    } finally {
      loading.value = false;
    }
  }

  async function searchEvents(query: string) {
    loading.value = true;
    error.value = null;
    try {
      events.value = await eventsApi.search(query);
      applyLibrary();
    } catch (err) {
      error.value = errMessage(err, "Failed to search events");
    } finally {
      loading.value = false;
    }
  }

  async function fetchEvent(id: string) {
    loading.value = true;
    error.value = null;
    try {
      currentEvent.value = await eventsApi.getById(id);
      applyLibrary();
    } catch (err) {
      error.value = errMessage(err, "Failed to fetch event");
    } finally {
      loading.value = false;
    }
  }

  // Full event for an expanded row; cached so re-expanding is free.
  async function fetchEventDetail(id: string): Promise<Event> {
    if (detailCache.value[id]) return detailCache.value[id];
    const ev = await eventsApi.getById(id);
    const le = libraryMap.value[id];
    if (le) {
      ev.status = le.status ?? null;
      ev.note = le.note ?? null;
    }
    detailCache.value = { ...detailCache.value, [id]: ev };
    return ev;
  }

  // Load the user's library (claims) and merge onto loaded events. Auth only.
  async function loadLibrary() {
    try {
      const list = await eventsApi.listLibrary();
      const map: Record<string, LibraryEntry> = {};
      for (const e of list) map[e.event_id] = e;
      libraryMap.value = map;
      applyLibrary();
    } catch {
      // not authenticated / empty - leave as-is
    }
  }

  async function loadNeedsResolution() {
    try {
      needsResolution.value = await eventsApi.listNeedsResolution();
    } catch {
      needsResolution.value = [];
    }
  }

  // Generic optimistic setter. status=null clears the claim.
  async function setStatus(
    id: string,
    status: EventStatus | null,
    extra?: Omit<SetStatusPayload, "status">,
  ) {
    const prev = libraryMap.value[id] ?? null;

    if (status === null) {
      const { [id]: _drop, ...rest } = libraryMap.value;
      libraryMap.value = rest;
    } else {
      libraryMap.value = {
        ...libraryMap.value,
        [id]: { event_id: id, status, ...extra },
      };
    }
    applyLibrary();

    try {
      await eventsApi.setStatus(id, { status, ...extra });
      if (status !== null) {
        // refresh needs-resolution lazily; a going->attended/missed resolves the nudge
        needsResolution.value = needsResolution.value.filter(
          (n) => n.event_id !== id || status === "going",
        );
      }
    } catch (err) {
      if (prev) libraryMap.value = { ...libraryMap.value, [id]: prev };
      else {
        const { [id]: _drop, ...rest } = libraryMap.value;
        libraryMap.value = rest;
      }
      applyLibrary();
      useToast().error(errMessage(err, "Failed to update status"));
    }
  }

  // Dense row toggles: bookmark = interested (claimed on/off), check = going (going<->interested).
  function toggleInterested(id: string) {
    return setStatus(id, statusOf(id) ? null : "interested");
  }
  function toggleGoing(id: string) {
    return setStatus(id, statusOf(id) === "going" ? "interested" : "going");
  }

  async function createManualEvent(payload: ManualEventPayload) {
    error.value = null;
    try {
      const ev = await eventsApi.createManualEvent(payload);
      await Promise.all([fetchEvents(), loadLibrary()]);
      return ev;
    } catch (err) {
      error.value = errMessage(err, "Failed to create event");
      useToast().error(error.value);
      return null;
    }
  }

  async function updateEvent(id: string, payload: ManualEventPayload) {
    error.value = null;
    try {
      const ev = await eventsApi.updateEvent(id, payload);
      await Promise.all([fetchEvents(), loadLibrary()]);
      return ev;
    } catch (err) {
      error.value = errMessage(err, "Failed to update event");
      useToast().error(error.value);
      return null;
    }
  }

  async function deleteEvent(id: string) {
    try {
      await eventsApi.remove(id);
      events.value = events.value.filter((e) => e.id !== id);
      const { [id]: _drop, ...rest } = detailCache.value;
      detailCache.value = rest;
      return true;
    } catch (err) {
      const msg = errMessage(err, "Failed to delete event");
      error.value = msg;
      useToast().error(msg);
      return false;
    }
  }

  async function deleteEventConfirmed(ev: { id: string; name: string }) {
    if (!confirm(`Delete "${ev.name}"?`)) return false;
    const ok = await deleteEvent(ev.id);
    if (ok) useToast().success("Event deleted");
    return ok;
  }

  async function prunePastTm() {
    await eventsApi.prunePastTm();
    await fetchEvents();
  }

  // Bulk set/clear claim status — optimistic, mirrors setStatus. Rolls all back on error.
  async function bulkSetStatus(
    ids: string[],
    status: EventStatus | null,
    extra?: Omit<SetStatusPayload, "status">,
  ) {
    if (ids.length === 0) return;
    const prev = { ...libraryMap.value };

    const next = { ...libraryMap.value };
    for (const id of ids) {
      if (status === null) delete next[id];
      else next[id] = { event_id: id, status, ...extra };
    }
    libraryMap.value = next;
    applyLibrary();

    try {
      await eventsApi.bulkSetStatus(ids, status, extra);
      if (status !== "going") {
        needsResolution.value = needsResolution.value.filter((n) => !ids.includes(n.event_id));
      }
    } catch (err) {
      libraryMap.value = prev;
      applyLibrary();
      useToast().error(errMessage(err, "Failed to update status"));
    }
  }

  // Bulk delete — destructive, so pending (not optimistic): call API, then refresh.
  async function bulkDelete(ids: string[]) {
    if (ids.length === 0) return;
    try {
      const res = await eventsApi.bulkDelete(ids);
      await Promise.all([fetchEvents(), loadLibrary()]);
      const kept = res.kept_claimed > 0 ? ` (${res.kept_claimed} claimed kept)` : "";
      useToast().success(`Deleted ${res.deleted} event${res.deleted === 1 ? "" : "s"}${kept}`);
    } catch (err) {
      const msg = errMessage(err, "Failed to delete events");
      error.value = msg;
      useToast().error(msg);
    }
  }

  return {
    events,
    currentEvent,
    loading,
    error,
    libraryMap,
    needsResolution,
    claimedCount,
    interestedCount,
    goingCount,
    attendedCount,
    missedCount,
    claimedEvents,
    statusOf,
    fetchEvents,
    searchEvents,
    fetchEvent,
    fetchEventDetail,
    loadLibrary,
    loadNeedsResolution,
    setStatus,
    toggleInterested,
    toggleGoing,
    createManualEvent,
    updateEvent,
    deleteEvent,
    deleteEventConfirmed,
    prunePastTm,
    bulkSetStatus,
    bulkDelete,
  };
});
