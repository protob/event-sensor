<script setup lang="ts">
import { computed, ref } from "vue";
import { RouterLink } from "vue-router";
import { useUiStore } from "@/stores/ui";
import { useAuthStore } from "@/stores/auth";
import { useArtistsStore } from "@/stores/artists";
import { useEventsStore } from "@/stores/events";
import { useToast } from "@/composables";
import { inEditable } from "@/composables/useListKeyboard";
import { Btn, Mono, DateField } from "@/components/ui";
import ManualEventForm from "@/components/library/ManualEventForm.vue";
import { useEventListener, useMediaQuery } from "@vueuse/core";
import IconMenu from "~icons/mdi/menu";
import IconSearch from "~icons/mdi/magnify";
import IconFetch from "~icons/mdi/refresh";
import IconPencil from "~icons/mdi/pencil";
import IconPlus from "~icons/mdi/plus";

// The rail drawer state lives in AppShell — the scrim and the rail are its siblings there.
const emit = defineEmits<{ "toggle-rail": [] }>();

const ui = useUiStore();
const auth = useAuthStore();
const artists = useArtistsStore();
const events = useEventsStore();
const toast = useToast();

// Local date inputs; applied to the shared ui store on Apply.
const startLocal = ref(ui.startDate);
const endLocal = ref(ui.endDate);

function applyDates() {
  ui.startDate = startLocal.value;
  ui.endDate = endLocal.value;
}
function clearDates() {
  startLocal.value = "";
  endLocal.value = "";
  ui.clearDates();
}

// Global quick-add: open the manual event form blank from anywhere (also via the `n` shortcut).
const showAdd = ref(false);
async function onCreated() {
  await events.fetchEvents();
  if (auth.isAuthenticated) await events.loadLibrary();
}
useEventListener(window, "keydown", (e: KeyboardEvent) => {
  if (inEditable(e.target) || e.metaKey || e.ctrlKey || e.altKey) return;
  if (e.key === "n" && auth.isAuthenticated && !showAdd.value) {
    e.preventDefault();
    showAdd.value = true;
  }
});

// The hint shortens, then goes, as the box does: `placeholder` is content, not style, so
// no media query can rewrite it - useMediaQuery is a live binding to "which width am I at".
// Two queries because the three states are not a scale (full label / "search" / nothing -
// below 360px the box is narrower than any word). aria-label names the input regardless.
const roomForVerb = useMediaQuery("(min-width: 1180px)");
const roomForHint = useMediaQuery("(min-width: 360px)");
const searchHint = computed(() =>
  !roomForHint.value ? "" : roomForVerb.value ? "search events…" : "search",
);

const fetching = ref(false);

// TODO(visual): add a backend batch fetch endpoint; for now iterate tracked
// artists client-side calling POST /artists/{id}/fetch-events.
async function fetchAll() {
  if (fetching.value) return;
  fetching.value = true;
  try {
    if (artists.artists.length === 0) await artists.fetchArtists();
    // Skip manual artists client-side - don't even call the endpoint.
    const auto = artists.artists.filter((a) => a.fetch_mode === "auto");
    const manualSkipped = artists.artists.length - auto.length;
    let saved = 0;
    let european = 0;
    let failed = 0;
    for (const a of auto) {
      const res = await artists.fetchArtistEvents(a.id);
      if (res) {
        saved += res.saved_count;
        european += res.european_count;
      } else {
        failed += 1;
      }
    }
    await artists.refreshAfterFetch();
    toast.success(
      `Fetched ${auto.length} artists · ${saved} saved · ${european} in region` +
        `${manualSkipped ? ` · ${manualSkipped} manual skipped` : ""}` +
        `${failed ? ` · ${failed} failed` : ""}`,
    );
  } catch {
    toast.error("Fetch all failed");
  } finally {
    fetching.value = false;
  }
}
</script>

<template>
  <!-- Nothing here scrolls sideways: everything fits, drops out once it stops earning its
         place, or - the search box - yields its own width. overflow-x-auto is only the floor
         below 320px, past the body's min-width yield limit. -->
  <header
    class="h-[46px] bg-surface border-b border-line flex items-center px-gutter gap-1.5 sm:gap-2.5 shrink-0 overflow-x-auto"
  >
    <!-- Drawer trigger; from md up the rail is a static column and needs no trigger. -->
    <button
      class="md:hidden shrink-0 text-faint hover:text-body p-1 -ml-1"
      title="Menu"
      aria-label="Open navigation"
      @click="emit('toggle-rail')"
    >
      <IconMenu class="h-5 w-5" />
    </button>

    <!-- Below sm the rail already identifies the app; the brand is the first thing to go. -->
    <RouterLink
      to="/events"
      class="hidden sm:flex shrink-0 items-center gap-1.5 text-muted hover:text-body"
    >
      <Mono size="xs" class="text-muted">Event Sensor</Mono>
      <Mono size="9" class="text-ghost">v3</Mono>
    </RouterLink>

    <!-- Global date range, first of the right-side controls - it carries the ml-auto that
         divides the row (a spacer would cost a gap on either side: 12px a 320px bar cannot
         spare). Below md it is hidden and the sort takes the margin. -->
    <div class="ml-auto hidden md:flex items-center gap-1.5 shrink-0">
      <DateField
        v-model="startLocal"
        title="Start date (YYYY-MM-DD)"
        class="w-[7.5rem] bg-surface-2 border border-line-2 rounded-sm text-xs text-muted font-mono px-1.5 py-1 placeholder:text-faint focus:outline-none focus:border-accent-bright"
      />
      <Mono size="xs" class="text-faint">→</Mono>
      <DateField
        v-model="endLocal"
        title="End date (YYYY-MM-DD)"
        class="w-[7.5rem] bg-surface-2 border border-line-2 rounded-sm text-xs text-muted font-mono px-1.5 py-1 placeholder:text-faint focus:outline-none focus:border-accent-bright"
      />
      <Btn tone="accent" size="sm" @click="applyDates">Apply</Btn>
      <Btn tone="neutral" size="sm" @click="clearDates">Clear</Btn>
    </div>

    <!-- sort -->
    <select
      v-model="ui.sort"
      title="Sort events"
      class="max-md:ml-auto shrink-0 bg-surface-2 border border-line-2 rounded-sm text-xs text-muted font-mono px-1.5 py-1 focus:outline-none focus:border-accent-bright"
    >
      <option value="date-asc">Date ↑</option>
      <option value="date-desc">Date ↓</option>
      <option value="artist">Artist</option>
    </select>

    <!-- search events. The only shrinkable control in the row, deliberately: when the bar is
         tight this is what gives, because a narrow search box still works and a clipped button
         does not. min-w keeps it wide enough to show a word of what was typed. -->
    <div
      class="min-w-14 flex items-center gap-1.5 bg-surface-2 border border-line-2 rounded-sm px-2 w-24 sm:w-28 md:w-40"
    >
      <IconSearch class="h-3.5 w-3.5 text-faint shrink-0" />
      <input
        v-model="ui.eventQuery"
        type="text"
        :placeholder="searchHint"
        aria-label="Search events"
        class="bg-transparent text-xs text-body font-mono py-1 w-full placeholder:text-faint focus:outline-none"
      />
    </div>

    <template v-if="auth.isAuthenticated">
      <!-- global quick-add (press `n`) -->
      <button
        class="shrink-0 inline-flex items-center gap-1.5 px-2.5 py-1.5 text-xs font-mono rounded-sm border border-accent-bright bg-accent-strong text-accent-text hover:brightness-110"
        title="Add event (n)"
        @click="showAdd = true"
      >
        <IconPlus class="h-3.5 w-3.5" />
        <span class="hidden sm:inline">Add</span>
      </button>
      <!-- edit mode: reveals rename/delete/remove controls across the app -->
      <button
        class="shrink-0 inline-flex items-center gap-1.5 px-2.5 py-1.5 text-xs font-mono rounded-sm border transition-colors"
        :class="
          ui.editMode
            ? 'border-accent-bright bg-accent-strong text-accent-text'
            : 'border-line-2 text-muted hover:text-body'
        "
        :title="ui.editMode ? 'Editing — click to lock' : 'Edit mode (show delete/rename controls)'"
        @click="ui.toggleEditMode()"
      >
        <IconPencil class="h-3.5 w-3.5" />
        <span class="hidden sm:inline">Edit</span>
      </button>
      <button
        class="shrink-0 inline-flex items-center gap-1.5 px-2.5 py-1.5 text-xs font-mono rounded-sm border border-danger-chip-border bg-danger-chip text-danger hover:brightness-125 disabled:opacity-50"
        :disabled="fetching"
        title="Fetch events for every tracked artist"
        @click="fetchAll"
      >
        <IconFetch class="h-3.5 w-3.5" :class="fetching ? 'animate-spin' : ''" />
        <span class="hidden sm:inline">Fetch All</span>
      </button>
    </template>
    <template v-else>
      <RouterLink to="/login">
        <Btn tone="outline" size="sm">Login</Btn>
      </RouterLink>
    </template>

    <ManualEventForm :open="showAdd" @close="showAdd = false" @created="onCreated" />
  </header>
</template>
