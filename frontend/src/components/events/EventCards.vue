<script setup lang="ts">
import { computed } from "vue";
import { useRouter } from "vue-router";
import type { Event } from "@/types";
import { useEventsStore } from "@/stores/events";
import { useUiStore } from "@/stores/ui";
import { countryName, formatISO, displayName } from "@/utils";
import { groupEvents, monthDividers, headlinerName } from "@/utils/eventSections";
import { Mono, Flag, Tag, IconButton, CountBadge } from "@/components/ui";
import { eventTag } from "./eventTag";
import WatchAttendCell from "./WatchAttendCell.vue";
import EventLifecycleTag from "./EventLifecycleTag.vue";
import IconExternal from "~icons/mdi/open-in-new";
import IconTrash from "~icons/mdi/trash-can-outline";

const props = withDefaults(
  defineProps<{
    events: Event[];
    loading?: boolean;
    emptyMessage?: string;
    groupBy?: "none" | "artist" | "country";
    sort?: "date-asc" | "date-desc" | "artist";
  }>(),
  { groupBy: "none", sort: "date-asc" },
);

const store = useEventsStore();
const ui = useUiStore();
const router = useRouter();

// A flat stream of "blocks": full-width headers/dividers + event cards. Rendered into a
// single responsive grid where headers span all columns (col-span-full).
type Block =
  | { kind: "header"; key: string; label: string; code?: string; count?: number }
  | { kind: "divider"; key: string; label: string; yearBreak: boolean }
  | { kind: "event"; key: string; event: Event };

const dateSorted = computed(() => props.sort === "date-asc" || props.sort === "date-desc");

const blocks = computed<Block[]>(() => {
  if (props.groupBy !== "none") {
    const out: Block[] = [];
    for (const g of groupEvents(props.events, props.groupBy)) {
      out.push({
        kind: "header",
        key: `h-${g.key}`,
        label: g.label,
        code: g.code,
        count: g.events.length,
      });
      for (const e of g.events) out.push({ kind: "event", key: e.id, event: e });
    }
    return out;
  }

  return monthDividers(props.events, dateSorted.value).map((r) =>
    r.kind === "divider"
      ? { kind: "divider", key: r.key, label: r.label, yearBreak: r.yearBreak }
      : { kind: "event", key: r.key, event: r.event },
  );
});

async function remove(ev: Event) {
  await store.deleteEventConfirmed(ev);
}
</script>

<template>
  <div class="flex-1 overflow-y-auto min-h-0 px-gutter py-3">
    <div v-if="loading" class="flex items-center justify-center h-32">
      <Mono size="xs" class="text-faint">Loading events…</Mono>
    </div>
    <div
      v-else-if="events.length === 0"
      class="flex flex-col items-center justify-center h-40 gap-1 text-center px-4"
    >
      <Mono size="xs" class="text-muted">
        {{ emptyMessage || "No events. Add an artist and fetch from Ticketmaster." }}
      </Mono>
    </div>

    <div v-else class="grid gap-2.5 grid-cols-[repeat(auto-fill,minmax(min(240px,100%),1fr))]">
      <template v-for="b in blocks" :key="b.key">
        <!-- group header -->
        <div v-if="b.kind === 'header'" class="col-span-full flex items-center gap-2 pt-1">
          <Flag v-if="b.code" :code="b.code" :w="22" />
          <Mono size="xs" class="text-accent-text font-bold uppercase tracking-wider truncate">
            {{ b.label }}
          </Mono>
          <CountBadge :n="b.count ?? 0" />
          <span class="flex-1 h-px bg-line" />
        </div>

        <!-- month/year divider -->
        <div
          v-else-if="b.kind === 'divider'"
          class="col-span-full flex items-center gap-2 pt-1"
          :class="b.yearBreak ? 'mt-1 border-t-2 border-t-accent-bright pt-2' : ''"
        >
          <Mono size="9" class="text-faint font-bold uppercase tracking-widest">{{ b.label }}</Mono>
          <span class="flex-1 h-px bg-line" />
        </div>

        <!-- event card -->
        <div
          v-else
          class="group flex flex-col gap-1.5 p-2.5 rounded-sm border border-line bg-surface hover:bg-surface-2 hover:border-line-2 transition-colors cursor-pointer"
          @click="router.push(`/events/${b.event.id}`)"
        >
          <div class="flex items-start justify-between gap-2">
            <WatchAttendCell
              :interested="!!b.event.status"
              :going="b.event.status === 'going'"
              @interested="store.toggleInterested(b.event.id)"
              @going="store.toggleGoing(b.event.id)"
            />
            <Mono size="11" class="text-muted tabular-nums shrink-0 mt-1">
              {{ formatISO(b.event.start_date) }}
            </Mono>
          </div>

          <div class="min-w-0">
            <div class="flex items-center gap-1.5 min-w-0">
              <span class="text-prose text-heading font-medium truncate">
                {{
                  groupBy === "artist"
                    ? b.event.venue?.name || b.event.name
                    : displayName(headlinerName(b.event))
                }}
              </span>
              <Tag v-if="eventTag(b.event)" :color="eventTag(b.event)!.color" class="shrink-0">
                {{ eventTag(b.event)!.label }}
              </Tag>
              <EventLifecycleTag :event="b.event" class="shrink-0" />
            </div>
            <div v-if="groupBy !== 'artist'" class="text-label text-muted truncate mt-0.5">
              {{ b.event.venue?.name ?? "—" }}
            </div>
          </div>

          <div class="flex items-center justify-between gap-2 mt-auto pt-1">
            <div class="flex items-center gap-1.5 min-w-0">
              <Flag
                v-if="b.event.venue?.country_code"
                :code="b.event.venue.country_code"
                :w="16"
                class="cursor-pointer hover:opacity-80"
                :title="`Toggle events in ${countryName(b.event.venue.country_code, b.event.venue.country)}`"
                @click.stop="ui.toggleCountry(b.event.venue.country_code)"
              />
              <Mono size="11" class="text-muted truncate">{{ b.event.venue?.city ?? "—" }}</Mono>
            </div>
            <!-- delete only in edit mode (inboard); the link stays the rightmost element
                 so its icon aligns to the card edge in every card. -->
            <div class="flex items-center gap-1 shrink-0" @click.stop>
              <IconButton
                v-if="ui.editMode"
                tone="danger"
                size="sm"
                title="Delete event"
                @click="remove(b.event)"
              >
                <IconTrash class="h-4 w-4" />
              </IconButton>
              <a
                v-if="b.event.url"
                :href="b.event.url"
                target="_blank"
                rel="noopener"
                class="inline-flex items-center"
              >
                <IconButton tone="accent" size="sm" title="Open on Ticketmaster">
                  <IconExternal class="h-4 w-4" />
                </IconButton>
              </a>
            </div>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>
