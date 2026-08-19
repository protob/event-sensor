<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useLocalStorage } from "@vueuse/core";
import type { Event, EventStatus } from "@/types";
import { useEventsStore } from "@/stores/events";
import { useAuthStore } from "@/stores/auth";
import { displayName } from "@/utils";
import { monthDividers } from "@/utils/eventSections";
import type { SectionRow } from "@/utils/eventSections";
import { useLibraryLenses } from "@/composables/useLibraryLenses";
import { useSelection } from "@/composables/useSelection";
import { useListKeyboard } from "@/composables/useListKeyboard";
import { Pill, Mono, CountBadge } from "@/components/ui";
import LibraryRow from "@/components/library/LibraryRow.vue";
import LibraryCard from "@/components/library/LibraryCard.vue";
import ManualEventForm from "@/components/library/ManualEventForm.vue";
import NeedsResolutionPanel from "@/components/library/NeedsResolutionPanel.vue";
import BulkActionBar from "@/components/shell/BulkActionBar.vue";
import IconList from "~icons/mdi/format-list-bulleted";
import IconGrid from "~icons/mdi/view-grid-outline";

const store = useEventsStore();
const auth = useAuthStore();

type Tab = EventStatus;
// Persisted so the chosen tab/lens survive a reload (no re-clicking "Attended" every time).
const tab = useLocalStorage<Tab>("es-library-tab", "interested");
// Two diary lenses over the same claims: a date timeline, and a by-artist history.
const lens = useLocalStorage<"date" | "artist">("es-library-lens", "date");
const view = useLocalStorage<"list" | "grid">("es-library-view", "list");

const { tabs, list, emptyMessage, artistGroups } = useLibraryLenses(() => tab.value);
const dateRows = computed<SectionRow[]>(() => monthDividers(list.value, true));

const showAdd = ref(false);
const editTarget = ref<Event | null>(null);

function openAdd() {
  editTarget.value = null;
  showAdd.value = true;
}
function openEdit(e: Event) {
  editTarget.value = e;
  showAdd.value = true;
}
function closeForm() {
  showAdd.value = false;
  editTarget.value = null;
}

async function refresh() {
  if (store.events.length === 0) await store.fetchEvents();
  if (auth.isAuthenticated) {
    await store.loadLibrary();
    await store.loadNeedsResolution();
  }
}

// After a fresh add, jump to the tab matching the new event's claim so it's immediately
// visible (adding a past show lands on Attended, not whatever tab you were on).
async function onCreated(status?: EventStatus) {
  if (status) tab.value = status;
  await refresh();
}

// --- Multi-select + bulk actions ---
const selection = useSelection(() => list.value.map((e) => e.id));
watch([list, tab], () => selection.prune());

async function bulkStatus(status: EventStatus | "") {
  await store.bulkSetStatus(selection.ids(), status === "" ? null : status);
  selection.clear();
}
async function bulkDelete() {
  await store.bulkDelete(selection.ids());
  selection.clear();
}

useListKeyboard(selection, () => list.value.length === 0);

onMounted(refresh);
</script>

<template>
  <div class="h-full flex flex-col bg-app min-w-0 relative">
    <!-- tabs + lens + add -->
    <!-- The pills wrap onto a second line instead of scrolling: a filter behind an edge is a
         filter nobody knows is there. min-h keeps the single-line height on a desktop. -->
    <div
      class="min-h-9 px-gutter py-1.5 flex flex-wrap items-center gap-2 border-b border-line bg-surface-2 shrink-0"
    >
      <!-- Inactive tabs render neutral (muted) so the active tab is unmistakable; an always-on
           accent background made every tab look selected at once. -->
      <Pill
        v-for="t in tabs"
        :key="t.key"
        :tone="tab === t.key ? (t.key === 'going' ? 'go' : 'accent') : 'neutral'"
        :active="tab === t.key"
        :data-testid="'library-tab-' + t.key"
        @click="tab = t.key"
      >
        {{ t.label }} ({{ t.count() }})
      </Pill>

      <span class="w-px h-4 bg-line mx-1 max-sm:hidden" />
      <Pill :active="lens === 'date'" @click="lens = 'date'">By date</Pill>
      <Pill :active="lens === 'artist'" @click="lens = 'artist'">By artist</Pill>

      <span class="w-px h-4 bg-line mx-1 max-sm:hidden" />
      <Pill :active="view === 'list'" title="List view" @click="view = 'list'"
        ><IconList class="h-3.5 w-3.5"
      /></Pill>
      <Pill :active="view === 'grid'" title="Grid view" @click="view = 'grid'"
        ><IconGrid class="h-3.5 w-3.5"
      /></Pill>

      <Pill
        tone="accent"
        :active="true"
        class="ml-auto"
        data-testid="library-add-event"
        @click="openAdd"
        >+ Add event</Pill
      >
    </div>

    <!-- "did you go?" nudge -->
    <NeedsResolutionPanel />

    <!-- body -->
    <div class="flex-1 overflow-y-auto min-h-0">
      <div
        v-if="!list.length"
        class="flex flex-col items-center justify-center h-40 px-4 text-center"
        data-testid="library-empty"
      >
        <Mono size="xs" class="text-muted">{{ emptyMessage }}</Mono>
      </div>

      <!-- BY DATE -->
      <template v-else-if="lens === 'date'">
        <!-- list view -->
        <template v-if="view === 'list'">
          <template v-for="row in dateRows" :key="row.key">
            <div
              v-if="row.kind === 'divider'"
              class="sticky top-0 z-10 flex items-center gap-2 px-gutter h-[26px] bg-surface-2/95 backdrop-blur border-b border-line"
            >
              <Mono size="9" class="text-faint font-bold uppercase tracking-widest">{{
                row.label
              }}</Mono>
              <span class="flex-1 h-px bg-line" />
            </div>
            <LibraryRow
              v-else
              :event="row.event"
              :checkable="true"
              :checked="selection.has(row.event.id)"
              @edit="openEdit"
              @select="selection.onRowClick(row.event.id, $event)"
            />
          </template>
        </template>
        <!-- grid view -->
        <div
          v-else
          class="grid gap-2.5 grid-cols-[repeat(auto-fill,minmax(min(280px,100%),1fr))] p-gutter"
        >
          <template v-for="row in dateRows" :key="row.key">
            <div v-if="row.kind === 'divider'" class="col-span-full flex items-center gap-2 pt-1">
              <Mono size="9" class="text-faint font-bold uppercase tracking-widest">{{
                row.label
              }}</Mono>
              <span class="flex-1 h-px bg-line" />
            </div>
            <LibraryCard
              v-else
              :event="row.event"
              :checkable="true"
              :checked="selection.has(row.event.id)"
              @edit="openEdit"
              @select="selection.onRowClick(row.event.id, $event)"
            />
          </template>
        </div>
      </template>

      <!-- BY ARTIST -->
      <template v-else>
        <!-- list view -->
        <template v-if="view === 'list'">
          <div v-for="g in artistGroups" :key="g.id">
            <div
              class="sticky top-0 z-10 flex items-center gap-2 px-gutter h-[26px] bg-surface-2/95 backdrop-blur border-b border-line cursor-pointer"
              @click="$router.push(`/artists/${g.id}`)"
            >
              <Mono size="10" class="text-accent-text font-bold uppercase tracking-wider truncate">
                {{ displayName(g.name) }}
              </Mono>
              <CountBadge prefix="×" :n="g.events.length" />
              <span class="flex-1 h-px bg-line" />
            </div>
            <LibraryRow
              v-for="e in g.events"
              :key="g.id + '-' + e.id"
              :event="e"
              :checkable="true"
              :checked="selection.has(e.id)"
              @edit="openEdit"
              @select="selection.onRowClick(e.id, $event)"
            />
          </div>
        </template>
        <!-- grid view -->
        <div v-else class="p-gutter space-y-4">
          <div v-for="g in artistGroups" :key="g.id">
            <div
              class="flex items-center gap-2 mb-2 cursor-pointer"
              @click="$router.push(`/artists/${g.id}`)"
            >
              <Mono size="10" class="text-accent-text font-bold uppercase tracking-wider truncate">
                {{ displayName(g.name) }}
              </Mono>
              <CountBadge prefix="×" :n="g.events.length" />
              <span class="flex-1 h-px bg-line" />
            </div>
            <div class="grid gap-2 grid-cols-[repeat(auto-fill,minmax(min(280px,100%),1fr))]">
              <LibraryCard
                v-for="e in g.events"
                :key="g.id + '-' + e.id"
                :event="e"
                :checkable="true"
                :checked="selection.has(e.id)"
                @edit="openEdit"
                @select="selection.onRowClick(e.id, $event)"
              />
            </div>
          </div>
        </div>
      </template>
    </div>

    <ManualEventForm
      :open="showAdd"
      :edit-event="editTarget"
      @close="closeForm"
      @created="onCreated"
    />

    <BulkActionBar
      :count="selection.count.value"
      context="library"
      @set-status="bulkStatus"
      @delete="bulkDelete"
      @clear="selection.clear()"
    />
  </div>
</template>
