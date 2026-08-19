<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { RouterLink } from "vue-router";
import type { Event } from "@/types";
import { useEventsStore } from "@/stores/events";
import { useUiStore } from "@/stores/ui";
import { formatISO, formatEventDate, countryName, displayName } from "@/utils";
import { eventTag } from "./eventTag";
import {
  EVENT_GRID_SELECTABLE,
  EVENT_GRID_PLAIN,
  CELL_FROM_LG,
  CELL_FROM_3XL,
} from "@/composables/useRowGrid";
import { Tag, Mono, IconButton, Flag } from "@/components/ui";
import WatchAttendCell from "./WatchAttendCell.vue";
import EventLifecycleTag from "./EventLifecycleTag.vue";
import IconExternal from "~icons/mdi/open-in-new";
import IconTrash from "~icons/mdi/trash-can-outline";

const props = defineProps<{
  event: Event;
  index: number;
  selected: boolean;
  // When grouped by artist, the artist is in the group header — show the
  // venue/event in the primary cell instead of repeating the name.
  hideArtist?: boolean;
  // Multi-select (bulk) support: when checkable, a leading checkbox column appears.
  checkable?: boolean;
  checked?: boolean;
}>();

const emit = defineEmits<{
  toggle: [];
  interested: [];
  going: [];
  remove: [];
  select: [ev: MouseEvent];
}>();

const ui = useUiStore();
const events = useEventsStore();

// W/A · Artist (+optional tag) · Location · Date · actions, with a leading 26px checkbox
// gutter when the row is checkable. The templates drop columns as the pane narrows; see
// useRowGrid for the tiers and the cells they pair with.
const GRID = computed(() => (props.checkable ? EVENT_GRID_SELECTABLE : EVENT_GRID_PLAIN));

const primaryArtist = computed(() => {
  const as = props.event.artists ?? [];
  const head = as.find((a) => a.is_headliner) ?? as[0];
  return head?.artist_name ?? props.event.name;
});

// What the main cell shows: artist normally, venue/event when the artist is
// already shown in a group header.
const primaryLabel = computed(() =>
  props.hideArtist ? props.event.venue?.name || props.event.name : displayName(primaryArtist.value),
);

const tag = computed(() => eventTag(props.event));

// Lazy detail: list rows carry venue + artists but not about/lineup. Cached in the
// store, so a re-expand is free.
const detail = ref<Event | null>(null);
const detailLoading = ref(false);
async function ensureDetail() {
  if (detail.value || detailLoading.value) return;
  detailLoading.value = true;
  try {
    detail.value = await events.fetchEventDetail(props.event.id);
  } catch {
    detail.value = props.event;
  } finally {
    detailLoading.value = false;
  }
}
watch(
  () => props.selected,
  (sel) => {
    if (sel) ensureDetail();
  },
);

const view = computed(() => detail.value ?? props.event);
const lineupNames = computed(() => (view.value.lineup ?? []).map((l) => l.artist_name).join(", "));
</script>

<template>
  <div
    class="border-b border-line"
    :data-testid="'event-row-' + event.id"
    :class="[
      checked ? 'ring-1 ring-inset ring-accent-bright/40' : '',
      selected
        ? 'bg-surface-sel'
        : checked
          ? 'bg-accent-chip/40'
          : index % 2 === 0
            ? 'bg-surface'
            : 'bg-surface-3',
    ]"
  >
    <!-- main row -->
    <div
      class="grid items-center px-gutter py-[7px] gap-0 cursor-default hover:bg-surface-2 transition-colors"
      :class="GRID"
      @click="emit('toggle')"
    >
      <!-- selection checkbox (bulk mode) -->
      <div v-if="checkable" class="flex items-center justify-center" @click.stop>
        <input
          type="checkbox"
          class="accent-accent-bright cursor-pointer"
          :checked="checked"
          :title="'Select (Shift-click for range)'"
          data-testid="event-select"
          @click="emit('select', $event)"
        />
      </div>

      <WatchAttendCell
        :interested="!!event.status"
        :going="event.status === 'going'"
        @interested="emit('interested')"
        @going="emit('going')"
      />

      <!-- artist / event -->
      <div class="min-w-0 pr-2">
        <div class="truncate">
          <span class="text-prose" :class="selected ? 'text-heading font-semibold' : 'text-body'">
            {{ primaryLabel }}
          </span>
          <Tag v-if="tag" :color="tag.color" class="ml-1.5 align-middle">{{ tag.label }}</Tag>
          <EventLifecycleTag :event="event" class="ml-1.5 align-middle" />
        </div>
        <!-- Below @3xl the location column is gone and the row is two thirds empty space. Put
             the city back on a second line at those widths: same information, and the row gains
             the height that makes it a reasonable thing to tap. -->
        <div
          v-if="event.venue?.city || event.venue?.country_code"
          class="@3xl:hidden flex items-center gap-1 mt-0.5"
        >
          <Flag
            v-if="event.venue?.country_code"
            :code="event.venue.country_code"
            :w="14"
            class="cursor-pointer hover:opacity-80 shrink-0"
            :title="`Toggle events in ${countryName(event.venue.country_code, event.venue.country)}`"
            @click.stop="ui.toggleCountry(event.venue.country_code)"
          />
          <Mono size="10" class="text-faint truncate">{{ event.venue?.city ?? "—" }}</Mono>
        </div>
      </div>

      <!-- location (click flag → filter that country). First column to go: longest string,
           and the detail expansion still carries the venue and country. -->
      <div class="items-center gap-1.5 min-w-0" :class="CELL_FROM_3XL">
        <Flag
          v-if="event.venue?.country_code"
          :code="event.venue.country_code"
          :w="16"
          class="cursor-pointer hover:opacity-80"
          :title="`Toggle events in ${countryName(event.venue.country_code, event.venue.country)}`"
          @click.stop="ui.toggleCountry(event.venue.country_code)"
        />
        <Mono size="11" class="text-muted truncate">{{ event.venue?.city ?? "—" }}</Mono>
      </div>

      <!-- date -->
      <div class="text-right">
        <Mono size="11" class="text-muted tabular-nums">{{ formatISO(event.start_date) }}</Mono>
      </div>

      <!-- actions: delete only in edit mode (inboard); link stays at the right edge. Dropped
           on a narrow pane - the row expands on tap and the expansion holds both actions. The
           52px column cannot hold two 44px touch targets: source order decides the contested
           strip, so the link stays last and it opens the ticket page, never deletes. -->
      <div class="items-center gap-1 justify-end" :class="CELL_FROM_LG" @click.stop>
        <IconButton
          v-if="ui.editMode"
          tone="danger"
          size="sm"
          title="Delete event"
          @click="emit('remove')"
        >
          <IconTrash class="h-4 w-4" />
        </IconButton>
        <a
          v-if="event.url"
          :href="event.url"
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

    <!-- expanded inline detail -->
    <div v-if="selected" class="border-t border-line bg-surface-sel px-gutter pb-2.5 pl-10">
      <div v-if="detailLoading" class="pt-2">
        <Mono size="11" class="text-faint">Loading…</Mono>
      </div>
      <template v-else>
        <div class="flex flex-wrap gap-x-10 gap-y-2 pt-2">
          <div v-if="view.venue?.name" class="min-w-[140px]">
            <Mono size="9" class="text-ghost block mb-0.5">VENUE</Mono>
            <span class="text-label text-muted">{{ view.venue.name }}</span>
          </div>
          <div v-if="view.venue?.country_code" class="min-w-[140px]">
            <Mono size="9" class="text-ghost block mb-0.5">COUNTRY</Mono>
            <span class="text-label text-muted">
              {{ countryName(view.venue.country_code, view.venue.country) }}
            </span>
          </div>
          <div class="min-w-[140px]">
            <Mono size="9" class="text-ghost block mb-0.5">DATE</Mono>
            <span class="text-label text-muted">
              {{ formatEventDate(view.start_date, view.end_date) }}
            </span>
          </div>
          <div v-if="view.about || view.description" class="min-w-[140px] max-w-[260px]">
            <Mono size="9" class="text-ghost block mb-0.5">ABOUT</Mono>
            <span class="text-label text-muted">{{ view.about || view.description }}</span>
          </div>
          <div v-if="lineupNames" class="min-w-[140px] max-w-[280px]">
            <Mono size="9" class="text-ghost block mb-0.5">LINE-UP</Mono>
            <span class="text-label text-muted">{{ lineupNames }}</span>
          </div>
          <div v-if="view.ticket_options?.length" class="min-w-[140px] max-w-[320px]">
            <Mono size="9" class="text-ghost block mb-0.5">TICKETS</Mono>
            <div class="flex flex-wrap gap-x-3 gap-y-1">
              <a
                v-for="t in view.ticket_options"
                :key="t.url"
                :href="t.url"
                target="_blank"
                rel="noopener"
                class="text-label text-accent-text hover:underline"
              >
                {{ t.label }} ↗
              </a>
            </div>
          </div>
        </div>
        <div class="flex gap-2 mt-2.5">
          <RouterLink :to="`/events/${event.id}`" data-testid="event-open">
            <span
              class="inline-flex items-center px-3 py-1 text-label font-mono rounded-sm border bg-accent-chip border-accent-chip-border text-accent-text cursor-pointer"
            >
              Open detail →
            </span>
          </RouterLink>
          <a
            v-if="event.url"
            :href="event.url"
            target="_blank"
            rel="noopener"
            class="inline-flex items-center px-3 py-1 text-label font-mono rounded-sm border bg-accent-chip border-accent-chip-border text-accent-text"
          >
            ↗ Ticketmaster
          </a>
          <button
            class="inline-flex items-center px-3 py-1 text-label font-mono rounded-sm border bg-surface-3 border-line-2 text-muted hover:text-danger"
            @click.stop="emit('remove')"
          >
            ✕ Remove
          </button>
        </div>
      </template>
    </div>
  </div>
</template>
