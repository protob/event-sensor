<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { storeToRefs } from "pinia";
import { useEventsStore } from "@/stores/events";
import { useArtistsStore } from "@/stores/artists";
import { useCategoriesStore } from "@/stores/categories";
import { useToast } from "@/composables";
import { useSelection } from "@/composables/useSelection";
import { useListKeyboard } from "@/composables/useListKeyboard";
import { Btn, Mono } from "@/components/ui";
import BulkActionBar from "@/components/shell/BulkActionBar.vue";
import MergeArtistsModal from "@/components/artists/MergeArtistsModal.vue";
import ArtistRow from "@/components/artists/ArtistRow.vue";
import type { Artist, FetchMode } from "@/types";
import IconPlus from "~icons/mdi/plus";
import IconClear from "~icons/mdi/playlist-remove";
import IconManual from "~icons/mdi/hand-back-right-outline";

const events = useEventsStore();
const artistsStore = useArtistsStore();
const categoriesStore = useCategoriesStore();
const toast = useToast();
const { artists, loading, error } = storeToRefs(artistsStore);

const newArtistName = ref("");
const newArtistManual = ref(false);
const fetchingArtistId = ref<string | null>(null);
const clearing = ref(false);

// Per-artist stored-event count, derived from the loaded events (no extra endpoint).
const eventCounts = computed(() => {
  const m: Record<string, number> = {};
  for (const e of events.events) {
    for (const a of e.artists ?? []) m[a.artist_id] = (m[a.artist_id] ?? 0) + 1;
  }
  return m;
});

onMounted(async () => {
  await artistsStore.fetchArtists();
  if (events.events.length === 0) await events.fetchEvents();
  if (categoriesStore.categories.length === 0) await categoriesStore.fetchCategories();
});

// --- Multi-select + bulk actions ---
const selection = useSelection(() => artists.value.map((a) => a.id));
watch(artists, () => selection.prune());
useListKeyboard(selection, () => artists.value.length === 0);

const categoryOptions = computed(() =>
  categoriesStore.categories.map((c) => ({ id: c.id, name: c.name })),
);

const selectedArtists = computed(() => artists.value.filter((a) => selection.has(a.id)));
const showMerge = ref(false);

async function bulkFetchMode(mode: FetchMode) {
  await artistsStore.bulkFetchMode(selection.ids(), mode);
  selection.clear();
}
async function addCategory(categoryId: string) {
  await artistsStore.assignCategories(selection.ids(), [categoryId], undefined);
  selection.clear();
}
async function removeCategory(categoryId: string) {
  await artistsStore.assignCategories(selection.ids(), undefined, [categoryId]);
  selection.clear();
}
async function bulkDeleteArtists() {
  const ids = selection.ids();
  if (!confirm(`Delete ${ids.length} artist${ids.length === 1 ? "" : "s"}?`)) return;
  for (const id of ids) await artistsStore.deleteArtist(id);
  selection.clear();
  toast.success(`Deleted ${ids.length} artist${ids.length === 1 ? "" : "s"}`);
}
function openMerge() {
  if (selection.count.value < 2) return;
  showMerge.value = true;
}
async function doMerge(from: string[], into: string) {
  showMerge.value = false;
  await artistsStore.merge(from, into);
  selection.clear();
}

async function handleCreate() {
  const name = newArtistName.value.trim();
  if (!name) return;
  await artistsStore.createArtist({ name, fetch_mode: newArtistManual.value ? "manual" : "auto" });
  newArtistName.value = "";
  newArtistManual.value = false;
}

// Flip an artist between auto (fetched by Fetch All) and manual (on-demand only) - inline,
// no detail page needed.
async function handleToggleMode(a: Artist) {
  await artistsStore.setFetchMode(a.id, a.fetch_mode === "manual" ? "auto" : "manual");
}

async function handleFetchEvents(id: string) {
  fetchingArtistId.value = id;
  await artistsStore.fetchAndSync(id, { alsoRefreshList: true });
  fetchingArtistId.value = null;
}

async function handleClearAllEvents() {
  if (!confirm("Delete ALL un-claimed events? Artists and claimed events are kept.")) return;
  clearing.value = true;
  try {
    await artistsStore.clearAllEvents();
  } catch {
    toast.error("Failed to clear events");
  } finally {
    clearing.value = false;
  }
}

async function handleClearArtistEvents(id: string) {
  if (!confirm("Clear all events for this artist?")) return;
  try {
    await artistsStore.clearArtistEvents(id);
  } catch {
    toast.error("Failed to clear artist events");
  }
}
</script>

<template>
  <div class="h-full flex flex-col bg-app min-w-0 relative">
    <!-- Toolbar: below sm the buttons keep icons and drop their words (the title carries the
         meaning); the name field is the only control that yields width. The form must NOT
         be min-w-0 - as a flex item it would shrink past its content and the buttons would
         overlap what follows instead of the row wrapping. -->
    <div
      class="min-h-12 px-gutter py-1.5 flex flex-wrap items-center gap-2 border-b border-line bg-surface-2 shrink-0"
    >
      <form class="flex flex-1 items-center gap-2" @submit.prevent="handleCreate">
        <input
          v-model="newArtistName"
          type="text"
          placeholder="Artist name…"
          class="flex-1 min-w-24 max-w-56 bg-surface border border-line-2 rounded-sm px-2.5 py-1.5 text-sm text-body font-mono placeholder:text-faint focus:outline-none focus:border-accent-bright"
        />
        <!-- Fetch mode: a checkbox (two-state form field, nothing reads faster); only its
             word drops below sm. Hand icon, not a refresh arrow - it marks artists you
             fetch by hand, and an arrow would say the opposite. -->
        <label
          class="shrink-0 flex items-center gap-1.5 font-mono text-label text-faint cursor-pointer select-none"
          title="Manual: excluded from Fetch All; fetched only on demand (disbanded / inactive artists you still keep)"
        >
          <input v-model="newArtistManual" type="checkbox" class="accent-accent-bright" />
          <IconManual class="h-4 w-4 sm:hidden" />
          <span class="hidden sm:inline">manual</span>
        </label>
        <Btn
          type="submit"
          tone="accent"
          size="md"
          class="shrink-0"
          title="Add artist"
          :disabled="!newArtistName.trim()"
        >
          <IconPlus class="h-3.5 w-3.5" />
          <span class="hidden sm:inline">Add Artist</span>
        </Btn>
      </form>
      <Btn
        tone="danger"
        size="md"
        class="ml-auto shrink-0"
        title="Clear all events"
        :loading="clearing"
        @click="handleClearAllEvents"
      >
        <IconClear v-if="!clearing" class="h-3.5 w-3.5" />
        <span class="hidden sm:inline">Clear All Events</span>
      </Btn>
    </div>

    <div class="flex-1 overflow-y-auto min-h-0">
      <div v-if="loading && !artists.length" class="flex items-center justify-center h-32">
        <Mono size="xs" class="text-faint">Loading artists…</Mono>
      </div>
      <p v-else-if="error" class="p-4 text-danger text-sm">{{ error }}</p>
      <div v-else-if="!artists.length" class="flex items-center justify-center h-40">
        <Mono size="xs" class="text-muted">No artists yet. Add one above to get started.</Mono>
      </div>
      <template v-else>
        <ArtistRow
          v-for="(a, i) in artists"
          :key="a.id"
          :artist="a"
          :event-count="eventCounts[a.id] ?? 0"
          :selected="selection.has(a.id)"
          :striped="i % 2 === 0"
          :fetching="fetchingArtistId === a.id"
          @select="selection.onRowClick(a.id, $event)"
          @toggle-mode="handleToggleMode(a)"
          @fetch="handleFetchEvents(a.id)"
          @clear="handleClearArtistEvents(a.id)"
          @delete="artistsStore.deleteArtistConfirmed(a.id, a.name)"
        />
      </template>
    </div>

    <BulkActionBar
      :count="selection.count.value"
      context="artists"
      :categories="categoryOptions"
      @fetch-mode="bulkFetchMode"
      @add-category="addCategory"
      @remove-category="removeCategory"
      @merge="openMerge"
      @delete="bulkDeleteArtists"
      @clear="selection.clear()"
    />

    <MergeArtistsModal
      :open="showMerge"
      :artists="selectedArtists"
      @close="showMerge = false"
      @confirm="doMerge"
    />
  </div>
</template>
