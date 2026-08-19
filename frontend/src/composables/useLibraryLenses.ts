import { computed } from "vue";
import type { Event, EventStatus } from "@/types";
import { useEventsStore } from "@/stores/events";

// The Library's two lenses over the same claims: a date timeline with month/year
// dividers, and a by-artist history ("Artist ×3"). Both read the active tab's list.
export function useLibraryLenses(tab: () => EventStatus) {
  const store = useEventsStore();

  const tabs: { key: EventStatus; label: string; count: () => number }[] = [
    { key: "interested", label: "Interested", count: () => store.interestedCount },
    { key: "going", label: "Going", count: () => store.goingCount },
    { key: "attended", label: "Attended", count: () => store.attendedCount },
    { key: "missed", label: "Missed", count: () => store.missedCount },
  ];

  // Claimed events for the active tab. Interested/Going read forward (soonest first);
  // Attended/Missed are a diary that runs into the past (most recent first).
  const list = computed(() => {
    const t = tab();
    const past = t === "attended" || t === "missed";
    return store.events
      .filter((e) => e.status === t)
      .sort((a, b) =>
        past ? b.start_date.localeCompare(a.start_date) : a.start_date.localeCompare(b.start_date),
      );
  });

  const emptyMessage = computed(
    () =>
      ({
        interested: "Nothing marked interested yet. Toggle the bookmark on any event.",
        going: "Not going to anything yet.",
        attended: "No shows logged yet — add one with + Add event.",
        missed: "Nothing marked missed.",
      })[tab()],
  );

  // BY ARTIST: group by linked entity artist. Only entity performances (artist_id set)
  // contribute - name-only acts are excluded; a multi-artist festival counts toward each
  // of its groups.
  const artistGroups = computed(() => {
    const map = new Map<string, { id: string; name: string; events: Event[] }>();
    for (const e of list.value) {
      for (const a of e.artists ?? []) {
        let g = map.get(a.artist_id);
        if (!g) map.set(a.artist_id, (g = { id: a.artist_id, name: a.artist_name, events: [] }));
        g.events.push(e);
      }
    }
    return [...map.values()]
      .map((g) => ({
        ...g,
        events: [...g.events].sort((a, b) => b.start_date.localeCompare(a.start_date)),
      }))
      .sort((a, b) => b.events.length - a.events.length || a.name.localeCompare(b.name));
  });

  return { tabs, list, emptyMessage, artistGroups };
}
