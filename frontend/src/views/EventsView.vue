<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useMediaQuery, useLocalStorage } from "@vueuse/core";
import { storeToRefs } from "pinia";
import { useEventsStore } from "@/stores/events";
import { useUiStore } from "@/stores/ui";
import { useAuthStore } from "@/stores/auth";
import { useEventsRadar } from "@/composables/useEventsRadar";
import { useSelection } from "@/composables/useSelection";
import { useListKeyboard } from "@/composables/useListKeyboard";
import type { EventStatus } from "@/types";
import { Pill } from "@/components/ui";
import EventTable from "@/components/events/EventTable.vue";
import EventCards from "@/components/events/EventCards.vue";
import CountryFilter from "@/components/events/CountryFilter.vue";
import CategoryTree from "@/components/tree/CategoryTree.vue";
import BulkActionBar from "@/components/shell/BulkActionBar.vue";
import IconChevron from "~icons/mdi/chevron-left";
import IconFilter from "~icons/mdi/filter-variant";
import IconList from "~icons/mdi/format-list-bulleted";
import IconCards from "~icons/mdi/view-grid-outline";

const store = useEventsStore();
const ui = useUiStore();
const auth = useAuthStore();
const radar = useEventsRadar();
const { loading } = storeToRefs(store);
const {
  showPast,
  onlyClaimed,
  groupBy,
  timeFilter,
  selectedArtistId,
  countryOptions,
  visibleEvents,
  activeFilterCount,
  selectArtist,
} = radar;

// A collapsed sidebar and a closed drawer are not the same fact, so they are not the same
// variable. The wide state is a preference and persists; the narrow one is ephemeral and
// starts closed. treeOpen asks whichever of the two the current width makes authoritative;
// the template's max-tree: utilities own what the tree looks like in each mode.
const wide = useMediaQuery("(min-width: 900px)");
const deskCollapsed = useLocalStorage("es-tree-collapsed", false);
const drawerOpen = ref(false);
const treeOpen = computed(() => (wide.value ? !deskCollapsed.value : drawerOpen.value));
function toggleTree() {
  if (wide.value) deskCollapsed.value = !deskCollapsed.value;
  else drawerOpen.value = !drawerOpen.value;
}
// Widening retires the drawer instead of leaving it armed to reappear the next time the
// window narrows. Nothing is visible at this moment, so this cannot contradict the user.
watch(wide, (w) => {
  if (w) drawerOpen.value = false;
});

// Below lg the chips collapse into a panel: a sideways scroll leaves controls behind an
// edge with nothing to say they are there.
const filtersOpen = ref(false);

// --- Multi-select + bulk actions (list/table view only) ---
// TODO(design): cards view (EventCards) has no checkbox affordance yet; selection is table-only.
const selection = useSelection(() => visibleEvents.value.map((e) => e.id));
// Drop ids that scroll out of the filtered set so the bar count stays honest.
watch(visibleEvents, () => selection.prune());

async function bulkStatus(status: EventStatus | "") {
  await store.bulkSetStatus(selection.ids(), status === "" ? null : status);
  selection.clear();
}
async function bulkDelete() {
  await store.bulkDelete(selection.ids());
  selection.clear();
}

useListKeyboard(
  selection,
  () => visibleEvents.value.length === 0,
  () => {
    if (filtersOpen.value) {
      filtersOpen.value = false;
      return true;
    }
    return false;
  },
);

onMounted(async () => {
  await store.fetchEvents();
  if (auth.isAuthenticated) await store.loadLibrary();
});
</script>

<template>
  <!-- relative so the tree can become an absolutely-positioned overlay below 900px -->
  <div class="h-full flex min-w-0 relative">
    <!-- left pane: category tree -->
    <div
      v-if="treeOpen"
      class="w-[clamp(200px,22vw,300px)] shrink-0 border-r border-line overflow-hidden max-tree:absolute max-tree:inset-y-0 max-tree:left-0 max-tree:z-50 max-tree:w-[280px] max-tree:shadow-xl max-tree:bg-surface"
    >
      <CategoryTree :selected-artist-id="selectedArtistId" @select-artist="selectArtist" />
    </div>

    <!-- As a drawer the tree covers the chevron that closes it, so it needs a scrim. Bound
         to drawerOpen, not treeOpen: only the drawer mode has one. Layers in this pane,
         low to high: sticky month dividers 10, filter panel 20, scrim 40, tree 50. -->
    <div v-if="drawerOpen" class="absolute inset-0 z-40 bg-black/50" @click="drawerOpen = false" />

    <!-- right pane: events. @container so rows size against this pane - its width depends on
         whether the tree is open, which no viewport breakpoint can see. -->
    <div class="@container flex-1 min-w-0 flex flex-col bg-app relative">
      <!-- Filter row: tree toggle and view switch stay in the row at every width; chips are
           inline from lg and live in a panel below it - not a sideways scroll, whose controls
           hide past an edge that gives no sign they exist. -->
      <div class="relative z-20 shrink-0 border-b border-line bg-surface-2">
        <div class="h-9 px-gutter flex items-center gap-2">
          <button
            class="text-faint hover:text-muted shrink-0"
            :title="treeOpen ? 'Hide categories' : 'Show categories'"
            @click="toggleTree"
          >
            <IconChevron class="h-4 w-4" :class="treeOpen ? '' : 'rotate-180'" />
          </button>

          <!-- Panel trigger. Lit whenever a filter is on, so a short list never looks broken. -->
          <button
            class="lg:hidden shrink-0 inline-flex items-center gap-1.5 px-2 py-1 rounded-sm border text-label font-mono transition-colors"
            :class="
              filtersOpen || activeFilterCount
                ? 'border-accent-bright bg-accent-strong text-accent-text'
                : 'border-line-2 text-muted hover:text-body'
            "
            :aria-expanded="filtersOpen"
            title="Filters"
            @click="filtersOpen = !filtersOpen"
          >
            <IconFilter class="h-3.5 w-3.5" />
            Filters
            <span
              v-if="activeFilterCount"
              class="inline-flex items-center justify-center min-w-4 h-4 px-1 rounded-full bg-accent-bright text-app text-micro tabular-nums"
            >
              {{ activeFilterCount }}
            </span>
          </button>

          <!-- The chips. One copy: below lg the same element becomes the panel, so there is no
               second set of controls to keep in step with this one. -->
          <div
            class="flex items-center gap-2 min-w-0 max-lg:absolute max-lg:top-full max-lg:inset-x-0 max-lg:flex-wrap max-lg:gap-y-2 max-lg:p-3 max-lg:bg-surface-2 max-lg:border-b max-lg:border-line max-lg:shadow-lg"
            :class="filtersOpen ? '' : 'max-lg:hidden'"
          >
            <Pill :active="timeFilter === 'all'" @click="timeFilter = 'all'">All</Pill>
            <Pill :active="timeFilter === 'week'" @click="timeFilter = 'week'">This Week</Pill>
            <Pill :active="timeFilter === 'month'" @click="timeFilter = 'month'">This Month</Pill>
            <span class="w-px h-4 bg-line shrink-0 max-lg:hidden" />
            <Pill tone="accent" :active="onlyClaimed" @click="onlyClaimed = !onlyClaimed">
              Interested
            </Pill>
            <Pill tone="neutral" :active="showPast" @click="showPast = !showPast"> Show past </Pill>

            <span class="w-px h-4 bg-line shrink-0 max-lg:hidden" />
            <!-- multi-select country filter (empty = whole region) -->
            <CountryFilter :options="countryOptions" />
            <select
              v-model="groupBy"
              title="Group events"
              class="bg-surface border border-line-2 rounded-sm text-label text-muted font-mono px-1.5 py-1 focus:outline-none focus:border-accent-bright shrink-0"
            >
              <option value="none">Group: None</option>
              <option value="artist">Group: Artist</option>
              <option value="country">Group: Country</option>
            </select>

            <span
              v-if="selectedArtistId"
              class="shrink-0 inline-flex items-center gap-1 text-label font-mono text-accent-text cursor-pointer"
              @click="selectArtist(null)"
            >
              artist filter ✕
            </span>
          </div>

          <!-- list / cards view toggle: outside the chip container so it keeps its place in the
               row at every width instead of travelling into the panel. -->
          <div
            class="ml-auto shrink-0 flex items-center border border-line-2 rounded-sm overflow-hidden"
          >
            <button
              class="px-1.5 py-1 flex items-center"
              :class="
                ui.eventView === 'list'
                  ? 'bg-accent-strong text-accent-text'
                  : 'text-faint hover:text-muted'
              "
              title="List view"
              @click="ui.setEventView('list')"
            >
              <IconList class="h-4 w-4" />
            </button>
            <button
              class="px-1.5 py-1 flex items-center border-l border-line-2"
              :class="
                ui.eventView === 'cards'
                  ? 'bg-accent-strong text-accent-text'
                  : 'text-faint hover:text-muted'
              "
              title="Cards view"
              @click="ui.setEventView('cards')"
            >
              <IconCards class="h-4 w-4" />
            </button>
          </div>
        </div>

        <!-- Tapping the list closes the panel; without this it stays open over the rows. -->
        <div
          v-if="filtersOpen"
          class="lg:hidden fixed inset-0 -z-10"
          @click="filtersOpen = false"
        />
      </div>

      <EventCards
        v-if="ui.eventView === 'cards'"
        :events="visibleEvents"
        :loading="loading"
        :group-by="groupBy"
        :sort="ui.sort"
      />
      <EventTable
        v-else
        :events="visibleEvents"
        :loading="loading"
        :group-by="groupBy"
        :sort="ui.sort"
        :selection="selection"
      />

      <BulkActionBar
        :count="selection.count.value"
        context="events"
        @set-status="bulkStatus"
        @delete="bulkDelete"
        @clear="selection.clear()"
      />
    </div>
  </div>
</template>
