<script setup lang="ts">
import { computed, ref } from "vue";
import type { Event } from "@/types";
import type { SelectionApi } from "@/composables/useSelection";
import { useEventsStore } from "@/stores/events";
import { groupEvents, monthDividers } from "@/utils/eventSections";
import type { EventGroup } from "@/utils/eventSections";
import {
  EVENT_GRID_SELECTABLE,
  EVENT_GRID_PLAIN,
  CELL_FROM_LG_BLOCK,
  CELL_FROM_3XL_BLOCK,
} from "@/composables/useRowGrid";
import { Mono, Flag, CountBadge } from "@/components/ui";
import EventRow from "./EventRow.vue";
import IconChevron from "~icons/mdi/chevron-down";

const props = withDefaults(
  defineProps<{
    events: Event[];
    loading?: boolean;
    emptyMessage?: string;
    groupBy?: "none" | "artist" | "country";
    sort?: "date-asc" | "date-desc" | "artist";
    selection?: SelectionApi;
  }>(),
  { groupBy: "none", sort: "date-asc" },
);

const store = useEventsStore();

// Must stay the same template the rows use, or the headers sit off their columns.
const GRID = computed(() => (props.selection ? EVENT_GRID_SELECTABLE : EVENT_GRID_PLAIN));
const expandedId = ref<string | null>(null);
const collapsed = ref<Set<string>>(new Set());

function toggle(id: string) {
  expandedId.value = expandedId.value === id ? null : id;
}

function toggleGroup(key: string) {
  const next = new Set(collapsed.value);
  if (next.has(key)) next.delete(key);
  else next.add(key);
  collapsed.value = next;
}

const groups = computed<EventGroup[] | null>(() => {
  if (props.groupBy === "none") return null;
  return groupEvents(props.events, props.groupBy);
});

// --- flat list with month/year dividers ("air") ---
// Only when sorted by date and ungrouped — otherwise the group header already segments.
const dateSorted = computed(() => props.sort === "date-asc" || props.sort === "date-desc");
const flatRows = computed(() => monthDividers(props.events, dateSorted.value));

async function remove(ev: Event) {
  await store.deleteEventConfirmed(ev);
}
</script>

<template>
  <div class="flex flex-col min-h-0 flex-1">
    <!-- column headers -->
    <div
      class="grid items-center px-gutter h-[30px] border-b border-line bg-surface-2 shrink-0"
      :class="GRID"
    >
      <span v-if="selection" />
      <Mono size="9" class="text-ghost uppercase" title="Interested / Going">I/G</Mono>
      <Mono size="9" class="text-ghost uppercase">
        {{ groupBy === "artist" ? "Venue / Event" : "Artist / Event" }}
      </Mono>
      <Mono size="9" class="text-ghost uppercase" :class="CELL_FROM_3XL_BLOCK">Location</Mono>
      <Mono size="9" class="text-ghost uppercase text-right">Date</Mono>
      <span :class="CELL_FROM_LG_BLOCK" />
    </div>

    <div class="flex-1 overflow-y-auto min-h-0">
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

      <!-- flat (with month/year dividers when date-sorted) -->
      <template v-else-if="!groups">
        <template v-for="row in flatRows" :key="row.key">
          <div
            v-if="row.kind === 'divider'"
            class="sticky top-0 z-10 flex items-center gap-2 px-gutter h-[26px] bg-surface-2/95 backdrop-blur border-b border-line"
            :class="row.yearBreak ? 'border-t border-t-accent-chip-border' : ''"
          >
            <Mono size="9" class="text-faint font-bold uppercase tracking-widest">
              {{ row.label }}
            </Mono>
            <span class="flex-1 h-px bg-line" />
          </div>
          <EventRow
            v-else
            :event="row.event"
            :index="row.index"
            :selected="expandedId === row.event.id"
            :checkable="!!selection"
            :checked="selection?.has(row.event.id)"
            @toggle="toggle(row.event.id)"
            @select="selection?.onRowClick(row.event.id, $event)"
            @interested="store.toggleInterested(row.event.id)"
            @going="store.toggleGoing(row.event.id)"
            @remove="remove(row.event)"
          />
        </template>
      </template>

      <!-- grouped -->
      <template v-else>
        <div v-for="g in groups" :key="g.key">
          <button
            class="w-full flex items-center gap-2.5 px-gutter h-9 bg-accent-chip border-y border-accent-chip-border border-l-[3px] border-l-accent-bright hover:brightness-110 sticky top-0 z-10"
            @click="toggleGroup(g.key)"
          >
            <IconChevron
              class="h-4 w-4 text-accent-text transition-transform"
              :class="collapsed.has(g.key) ? '-rotate-90' : ''"
            />
            <Flag v-if="g.code" :code="g.code" :w="22" />
            <Mono size="xs" class="text-accent-text font-bold uppercase tracking-wider truncate">
              {{ g.label }}
            </Mono>
            <CountBadge class="ml-1" :n="g.events.length" />
          </button>
          <template v-if="!collapsed.has(g.key)">
            <EventRow
              v-for="(ev, i) in g.events"
              :key="ev.id"
              :event="ev"
              :index="i"
              :selected="expandedId === ev.id"
              :hide-artist="groupBy === 'artist'"
              :checkable="!!selection"
              :checked="selection?.has(ev.id)"
              @toggle="toggle(ev.id)"
              @select="selection?.onRowClick(ev.id, $event)"
              @interested="store.toggleInterested(ev.id)"
              @going="store.toggleGoing(ev.id)"
              @remove="remove(ev)"
            />
          </template>
        </div>
      </template>
    </div>
  </div>
</template>
