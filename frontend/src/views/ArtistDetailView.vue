<script setup lang="ts">
import { computed, nextTick, onMounted, ref } from "vue";
import { useRoute, useRouter, RouterLink } from "vue-router";
import type { Event } from "@/types";
import { useArtistsStore } from "@/stores/artists";
import { useCategoriesStore } from "@/stores/categories";
import { useEventsStore } from "@/stores/events";
import { useAuthStore } from "@/stores/auth";
import { useToast } from "@/composables";
import { categoryColor, displayName } from "@/utils";
import { sortEvents } from "@/composables/useEventFilters";
import {
  ARTIST_EVENT_GRID,
  CELL_FROM_LG_BLOCK,
  CELL_FROM_3XL_BLOCK,
} from "@/composables/useRowGrid";
import { storeToRefs } from "pinia";
import { Tag, Mono, Pill, Flag } from "@/components/ui";
import ArtistEventRow from "@/components/events/ArtistEventRow.vue";
import ManualEventForm from "@/components/library/ManualEventForm.vue";
import IconBack from "~icons/mdi/arrow-left";
import IconFetch from "~icons/mdi/refresh";
import IconPencil from "~icons/mdi/pencil";
import IconClear from "~icons/mdi/playlist-remove";
import IconDelete from "~icons/mdi/delete";

const route = useRoute();
const router = useRouter();
const artists = useArtistsStore();
const categories = useCategoriesStore();
const events = useEventsStore();
const auth = useAuthStore();
const toast = useToast();
const { currentArtist: artist, artistEvents } = storeToRefs(artists);
const eventsLoading = ref(false);
const fetching = ref(false);
const showAdd = ref(false);

async function onManualCreated() {
  await events.fetchEvents();
  if (auth.isAuthenticated) await events.loadLibrary();
  await loadEvents();
}

// Must stay the same template ArtistEventRow uses, or the headers sit off their columns.
const GRID = ARTIST_EVENT_GRID;

const sortedEvents = computed(() => sortEvents(artistEvents.value, "date-asc"));

const countryBreakdown = computed(() => {
  const by: Record<string, number> = {};
  for (const e of artistEvents.value) {
    const cc = e.venue?.country_code;
    if (cc) by[cc] = (by[cc] ?? 0) + 1;
  }
  return Object.entries(by)
    .map(([code, n]) => ({ code, n }))
    .sort((a, b) => b.n - a.n);
});

// Per-row claim state, read reactively from the store library map so dense-toggle updates
// reflect immediately (the artistEvents list is a local snapshot).
const interestedOf = (id: string) => events.statusOf(id) != null;
const goingOf = (id: string) => events.statusOf(id) === "going";

const summary = computed(() => ({
  total: artistEvents.value.length,
  interested: artistEvents.value.filter((e) => interestedOf(e.id)).length,
  going: artistEvents.value.filter((e) => goingOf(e.id)).length,
}));

const isManual = computed(() => artist.value?.fetch_mode === "manual");
async function toggleManual() {
  if (!artist.value) return;
  await artists.setFetchMode(artist.value.id, isManual.value ? "auto" : "manual");
  toast.success(isManual.value ? "Marked manual" : "Marked auto-fetch");
}

// The owner's own category for this artist, from GET /artists/summary. Not a music genre -
// this app stores none.
const category = computed(() => {
  const id = artist.value?.id;
  return id ? (artists.summaries[id]?.categories[0]?.name ?? null) : null;
});

const initials = computed(() => (artist.value?.name || "?").slice(0, 2).toUpperCase());

async function loadEvents() {
  const id = route.params.id as string;
  eventsLoading.value = true;
  try {
    await artists.listArtistEvents(id);
  } finally {
    eventsLoading.value = false;
  }
}

async function load() {
  const id = route.params.id as string;
  await artists.fetchArtist(id);
  if (auth.isAuthenticated && Object.keys(events.libraryMap).length === 0) {
    await events.loadLibrary();
  }
  if (categories.categories.length === 0) await categories.fetchCategories();
  if (Object.keys(artists.summaries).length === 0) await artists.fetchSummaries();
  await loadEvents();
}

async function doFetch() {
  fetching.value = true;
  await artists.fetchAndSync(route.params.id as string);
  fetching.value = false;
}

// rename
const editing = ref(false);
const editName = ref("");
const nameInput = ref<HTMLInputElement | null>(null);
async function startRename() {
  editing.value = true;
  editName.value = artist.value?.name ?? "";
  await nextTick();
  nameInput.value?.focus();
}
async function commitRename() {
  const n = editName.value.trim();
  if (n && n !== artist.value?.name && artist.value) {
    await artists.updateArtist(artist.value.id, { name: n });
  }
  editing.value = false;
}

async function clearEvents() {
  if (!confirm("Clear all events for this artist?")) return;
  await artists.clearArtistEvents(route.params.id as string);
}

async function deleteArtist() {
  if (!artist.value) return;
  const ok = await artists.deleteArtistConfirmed(artist.value.id, artist.value.name);
  if (ok) router.push({ name: "artists" });
}

async function toggleInterested(e: Event) {
  await events.toggleInterested(e.id);
}
async function toggleGoing(e: Event) {
  await events.toggleGoing(e.id);
}
async function removeEvent(e: Event) {
  const ok = await events.deleteEventConfirmed(e);
  if (ok) await loadEvents();
}

onMounted(load);
</script>

<template>
  <div class="h-full flex flex-col min-w-0">
    <!-- header strip: below sm actions keep their icon and drop their word; the row wraps
         rather than clipping, so nothing is reachable only by knowing it is there. -->
    <div
      class="flex flex-wrap items-center min-h-[46px] px-gutter py-1.5 gap-2.5 border-b border-line bg-surface shrink-0"
    >
      <RouterLink
        to="/artists"
        class="flex items-center gap-1.5 font-mono text-label text-faint hover:text-muted"
      >
        <IconBack class="h-3.5 w-3.5" /> Artists
      </RouterLink>
      <Mono size="11" class="text-faint">/</Mono>
      <Mono size="11" class="text-muted truncate max-w-[30vw]">
        {{ artist ? displayName(artist.name) : "" }}
      </Mono>

      <div class="ml-auto flex flex-wrap items-center gap-2">
        <Pill
          tone="accent"
          :active="true"
          :title="
            isManual
              ? 'Manual artist — fetched only on demand (this button); Fetch All skips it'
              : 'Fetch events'
          "
          @click="doFetch"
        >
          <IconFetch class="h-3.5 w-3.5" :class="fetching ? 'animate-spin' : ''" />
          <span class="hidden sm:inline">Fetch Events</span>
        </Pill>
        <Pill
          tone="neutral"
          :title="isManual ? 'Make auto-fetch' : 'Make manual'"
          @click="toggleManual"
        >
          {{ isManual ? "↻" : "⦸" }}
          <span class="hidden sm:inline">{{ isManual ? "Make auto-fetch" : "Make manual" }}</span>
        </Pill>
        <Pill tone="accent" title="Log past show" @click="showAdd = true">
          +
          <span class="hidden sm:inline">Log past show</span>
        </Pill>
        <Pill tone="neutral" title="Rename" @click="startRename">
          <IconPencil class="h-3.5 w-3.5" />
          <span class="hidden sm:inline">Rename</span>
        </Pill>
        <Pill tone="neutral" title="Clear events" @click="clearEvents">
          <IconClear class="h-3.5 w-3.5" />
          <span class="hidden sm:inline">Clear Events</span>
        </Pill>
        <Pill tone="danger" title="Delete artist" @click="deleteArtist">
          <IconDelete class="h-3.5 w-3.5" />
          <span class="hidden sm:inline">Delete Artist</span>
        </Pill>
      </div>
    </div>

    <!-- artist header strip -->
    <div
      class="flex flex-wrap items-center gap-x-5 gap-y-3 min-h-20 px-gutter py-3 bg-surface border-b border-line shrink-0"
    >
      <div
        class="w-[52px] h-[52px] rounded-sm bg-surface-2 border border-line-2 flex items-center justify-center shrink-0 font-mono text-accent-text font-semibold"
      >
        {{ initials }}
      </div>
      <div class="min-w-0">
        <div v-if="editing">
          <input
            ref="nameInput"
            v-model="editName"
            class="bg-surface-2 border border-accent-bright rounded-sm px-2 py-1 text-lg font-bold text-heading focus:outline-none"
            @keydown.enter="commitRename"
            @keydown.esc="editing = false"
            @blur="commitRename"
          />
        </div>
        <h1 v-else class="text-xl font-bold text-heading truncate">
          {{ artist ? displayName(artist.name) : "" }}
        </h1>
        <div class="flex items-center gap-2 mt-1 flex-wrap">
          <Tag v-if="category" :color="categoryColor(category)">{{ category }}</Tag>
          <Tag v-if="isManual" color="#71717a">MANUAL</Tag>
          <Mono size="11" class="text-faint">
            {{ summary.total }} events · interested: {{ summary.interested }} · going:
            {{ summary.going }}
          </Mono>
        </div>
      </div>

      <!-- country pills -->
      <div
        class="ml-auto flex flex-wrap gap-1.5 justify-end max-w-[400px] max-sm:ml-0 max-sm:justify-start"
      >
        <span
          v-for="c in countryBreakdown"
          :key="c.code"
          class="inline-flex items-center gap-1 px-2 py-0.5 bg-surface-2 border border-line-2 rounded-sm font-mono text-label text-muted"
        >
          <Flag :code="c.code" :w="16" />
          {{ c.code.toUpperCase() }}<span class="text-ghost">×{{ c.n }}</span>
        </span>
      </div>
    </div>

    <!-- events list -->
    <div class="flex-1 min-h-0 flex flex-col bg-app">
      <div
        class="grid items-center px-gutter h-[30px] border-b border-line bg-surface-2 shrink-0"
        :class="GRID"
      >
        <Mono size="9" class="text-ghost uppercase" title="Interested / Attending">I/A</Mono>
        <Mono size="9" class="text-ghost uppercase" :class="CELL_FROM_LG_BLOCK">Type</Mono>
        <Mono size="9" class="text-ghost uppercase">Venue</Mono>
        <Mono size="9" class="text-ghost uppercase" :class="CELL_FROM_3XL_BLOCK">Location</Mono>
        <Mono size="9" class="text-ghost uppercase text-right">Date</Mono>
        <span :class="CELL_FROM_LG_BLOCK" />
      </div>
      <div class="flex-1 overflow-y-auto min-h-0">
        <div v-if="eventsLoading" class="flex items-center justify-center h-32">
          <Mono size="xs" class="text-faint">Loading events…</Mono>
        </div>
        <div
          v-else-if="sortedEvents.length === 0"
          class="flex flex-col items-center justify-center h-40 gap-1"
        >
          <Mono size="xs" class="text-muted">No events. Fetch from Ticketmaster above.</Mono>
        </div>
        <ArtistEventRow
          v-for="(e, i) in sortedEvents"
          v-else
          :key="e.id"
          :event="e"
          :index="i"
          :interested="interestedOf(e.id)"
          :going="goingOf(e.id)"
          @interested="toggleInterested(e)"
          @going="toggleGoing(e)"
          @remove="removeEvent(e)"
        />
      </div>
    </div>

    <ManualEventForm
      :open="showAdd"
      :preset-artist-id="artist?.id ?? null"
      @close="showAdd = false"
      @created="onManualCreated"
    />
  </div>
</template>
