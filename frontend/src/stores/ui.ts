import { ref } from "vue";
import { useLocalStorage } from "@vueuse/core";
import { defineStore } from "pinia";

export type EventSort = "date-asc" | "date-desc" | "artist";
export type EventView = "list" | "cards";

// Cross-shell UI state the TopBar sets and the Events pane reads. Durable filters/prefs are
// persisted to localStorage (via useLocalStorage) so they survive a reload. Transient bits
// (the search box, edit mode) deliberately are not.
export const useUiStore = defineStore("ui", () => {
  const startDate = useLocalStorage<string>("es-start-date", ""); // yyyy-mm-dd
  const endDate = useLocalStorage<string>("es-end-date", "");
  const sort = useLocalStorage<EventSort>("es-sort", "date-asc");
  const eventQuery = ref<string>(""); // top-bar search box — transient, not persisted
  const lastFetch = useLocalStorage<string | null>("es-last-fetch", null);

  // Quick "what's on in these countries" filter for the events list. Empty = whole region
  // (all). Set from the toolbar multi-select or by clicking country flags (which toggle).
  const countryFilters = useLocalStorage<string[]>("es-country-filters", []);
  function toggleCountry(cc: string) {
    const code = (cc || "").toLowerCase();
    if (!code) return;
    countryFilters.value = countryFilters.value.includes(code)
      ? countryFilters.value.filter((c) => c !== code)
      : [...countryFilters.value, code];
  }
  function clearCountries() {
    countryFilters.value = [];
  }

  // Edit mode: gates the rename/delete/remove management icons across the app. Off by
  // default (and intentionally not persisted) so a fresh load is read-only and clean —
  // destructive controls only appear once the user opts in via the top-bar toggle.
  const editMode = ref(false);
  function toggleEditMode() {
    editMode.value = !editMode.value;
  }

  // List (dense table) vs Cards (roomier grid); persisted across sessions.
  const eventView = useLocalStorage<EventView>("es-event-view", "list");
  function setEventView(v: EventView) {
    eventView.value = v;
  }

  function setLastFetch() {
    lastFetch.value = new Date().toISOString();
  }

  function clearDates() {
    startDate.value = "";
    endDate.value = "";
  }

  return {
    startDate,
    endDate,
    sort,
    eventQuery,
    lastFetch,
    countryFilters,
    toggleCountry,
    clearCountries,
    editMode,
    toggleEditMode,
    eventView,
    setEventView,
    setLastFetch,
    clearDates,
  };
});
