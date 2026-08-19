import { computed } from "vue";
import type { Event, EventStatus } from "@/types";
import type { Ref, ComputedRef } from "vue";
import { useEventsStore } from "@/stores/events";
import { eventTag } from "@/components/events/eventTag";

// Shared behavior of the two library entry views (row and card): how the entry is
// titled, its kind badge, and the claim verbs that keep note/reason intact.
export function useLibraryEntry(event: Ref<Event> | ComputedRef<Event>) {
  const store = useEventsStore();

  const isFestival = computed(() => event.value.kind === "festival");

  // A festival is a container: titled by the FESTIVAL (event name). A concert is
  // titled by its act.
  const title = computed(() => {
    if (isFestival.value) return event.value.name;
    const as = event.value.artists ?? [];
    return (as.find((a) => a.is_headliner) ?? as[0])?.artist_name ?? event.value.name;
  });

  const kindBadge = computed(() => eventTag(event.value));

  async function setStatus(status: EventStatus) {
    await store.setStatus(event.value.id, status, { note: event.value.note ?? undefined });
  }
  async function saveNote(note: string) {
    await store.setStatus(event.value.id, event.value.status ?? "interested", { note });
  }
  async function unclaim() {
    await store.setStatus(event.value.id, null);
  }

  return { isFestival, title, kindBadge, setStatus, saveNote, unclaim };
}
