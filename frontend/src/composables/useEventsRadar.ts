import { computed, ref } from "vue";
import { useLocalStorage } from "@vueuse/core";
import { storeToRefs } from "pinia";
import { useEventsStore } from "@/stores/events";
import { useUiStore } from "@/stores/ui";
import { countryName } from "@/utils";
import { matchesQuery, inDateRange, sortEvents, withinDays } from "@/composables/useEventFilters";

type TimeFilter = "all" | "week" | "month";
type GroupBy = "none" | "artist" | "country";

// The Events page's filter model: the radar set (forward-looking TM discovery), the
// filter controls above it, and the visible-events pipeline. State that is a durable
// preference persists via useLocalStorage; the transient bits (time filter, artist
// pick) do not.
export function useEventsRadar() {
  const store = useEventsStore();
  const ui = useUiStore();
  const { events } = storeToRefs(store);

  const showPast = useLocalStorage("es-events-show-past", false);
  const onlyClaimed = useLocalStorage("es-events-only-claimed", false);
  const groupBy = useLocalStorage<GroupBy>("es-events-group-by", "none");
  // Transient on purpose: a reload resets the time window and the artist pick.
  const timeFilter = ref<TimeFilter>("all");
  const selectedArtistId = ref<string | null>(null);

  // The radar set: discovery is forward-looking only - past and delisted TM events stay
  // hidden unless "Show past" is on.
  const radarEvents = computed(() =>
    events.value.filter((e) => {
      if (e.source !== "ticketmaster") return false;
      if (!showPast.value && (e.is_past || e.listing_state !== "listed")) return false;
      return true;
    }),
  );

  // Countries present in the radar events, with per-country event counts, for the country
  // filter. Counts come from the full radar set (not the filtered view) so they stay stable
  // as countries are toggled.
  const countryOptions = computed(() => {
    const map = new Map<string, { name: string; count: number }>();
    for (const e of radarEvents.value) {
      const cc = e.venue?.country_code?.toLowerCase();
      if (!cc) continue;
      const entry = map.get(cc);
      if (entry) entry.count++;
      else map.set(cc, { name: countryName(cc, e.venue?.country || cc.toUpperCase()), count: 1 });
    }
    return [...map.entries()]
      .map(([code, { name, count }]) => ({ code, name, count }))
      .sort((a, b) => a.name.localeCompare(b.name));
  });

  const visibleEvents = computed(() => {
    let list = radarEvents.value;

    if (selectedArtistId.value) {
      list = list.filter((e) =>
        (e.artists ?? []).some((a) => a.artist_id === selectedArtistId.value),
      );
    }

    if (timeFilter.value === "week") list = list.filter((e) => withinDays(e, 7));
    else if (timeFilter.value === "month") list = list.filter((e) => withinDays(e, 31));

    if (ui.countryFilters.length)
      list = list.filter((e) => {
        const cc = e.venue?.country_code?.toLowerCase();
        return cc ? ui.countryFilters.includes(cc) : false;
      });

    if (onlyClaimed.value) list = list.filter((e) => e.status);

    list = list.filter(
      (e) => inDateRange(e, ui.startDate, ui.endDate) && matchesQuery(e, ui.eventQuery),
    );

    return sortEvents(list, ui.sort);
  });

  // The panel trigger must report what the hidden chips are doing, or a filtered list
  // looks like a broken list.
  const activeFilterCount = computed(() => {
    let n = 0;
    if (timeFilter.value !== "all") n++;
    if (onlyClaimed.value) n++;
    if (showPast.value) n++;
    if (ui.countryFilters.length) n++;
    if (groupBy.value !== "none") n++;
    if (selectedArtistId.value) n++;
    return n;
  });

  function selectArtist(id: string | null) {
    selectedArtistId.value = selectedArtistId.value === id ? null : id;
  }

  return {
    showPast,
    onlyClaimed,
    groupBy,
    timeFilter,
    selectedArtistId,
    radarEvents,
    countryOptions,
    visibleEvents,
    activeFilterCount,
    selectArtist,
  };
}

export type EventsRadar = ReturnType<typeof useEventsRadar>;
