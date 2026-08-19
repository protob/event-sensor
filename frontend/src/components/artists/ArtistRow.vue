<script setup lang="ts">
import { RouterLink } from "vue-router";
import type { Artist } from "@/types";
import { useUiStore } from "@/stores/ui";
import { displayName } from "@/utils";
import { Mono, IconButton } from "@/components/ui";
import IconFetch from "~icons/mdi/refresh";
import IconClear from "~icons/mdi/playlist-remove";
import IconDelete from "~icons/mdi/trash-can-outline";

// One artist row in the Artists list: name link, stored-event count, auto/manual
// toggle, per-artist fetch, and the edit-mode-gated clear/delete actions.
defineProps<{
  artist: Artist;
  eventCount: number;
  selected: boolean;
  striped: boolean;
  fetching: boolean;
}>();

const emit = defineEmits<{
  select: [ev: MouseEvent];
  "toggle-mode": [];
  fetch: [];
  clear: [];
  delete: [];
}>();

const ui = useUiStore();
</script>

<template>
  <div
    class="group flex items-center px-gutter py-2.5 border-b border-line gap-3"
    :class="[
      selected
        ? 'bg-accent-chip/40 ring-1 ring-inset ring-accent-bright/40'
        : striped
          ? 'bg-surface'
          : 'bg-surface-3',
    ]"
  >
    <input
      type="checkbox"
      class="accent-accent-bright cursor-pointer shrink-0"
      :checked="selected"
      title="Select (Shift-click for range)"
      @click="emit('select', $event)"
    />
    <RouterLink
      :to="`/artists/${artist.id}`"
      class="flex-1 min-w-0 truncate text-sm text-body hover:text-accent-text"
    >
      {{ displayName(artist.name) }}
    </RouterLink>
    <Mono
      size="10"
      class="text-ghost shrink-0 tabular-nums"
      :title="`${eventCount} event${eventCount === 1 ? '' : 's'} stored`"
    >
      {{ eventCount }} {{ eventCount === 1 ? "event" : "events" }}
    </Mono>
    <button
      class="shrink-0 px-1.5 py-px rounded-xs border text-meta font-mono uppercase tracking-wide transition-colors"
      :class="
        artist.fetch_mode === 'manual'
          ? 'bg-surface-3 border-line-2 text-faint hover:text-body'
          : 'border-transparent text-ghost hover:text-body hover:border-line-2'
      "
      :title="
        artist.fetch_mode === 'manual'
          ? 'Manual — excluded from Fetch All. Click to set Auto.'
          : 'Auto — fetched by Fetch All. Click to set Manual (on-demand only).'
      "
      @click="emit('toggle-mode')"
    >
      {{ artist.fetch_mode === "manual" ? "manual" : "auto" }}
    </button>
    <IconButton tone="accent" title="Fetch events" @click="emit('fetch')">
      <IconFetch class="h-4 w-4" :class="fetching ? 'animate-spin' : ''" />
    </IconButton>
    <!-- touch:ml-2 keeps the 44px hit areas clear of their left neighbour's tap zone. -->
    <IconButton
      v-if="ui.editMode"
      class="touch:ml-2"
      title="Clear this artist's events"
      @click="emit('clear')"
    >
      <IconClear class="h-4 w-4" />
    </IconButton>
    <IconButton
      v-if="ui.editMode"
      class="touch:ml-2"
      tone="danger"
      title="Delete artist"
      @click="emit('delete')"
    >
      <IconDelete class="h-4 w-4" />
    </IconButton>
  </div>
</template>
