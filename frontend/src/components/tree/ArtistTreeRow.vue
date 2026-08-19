<script setup lang="ts">
import { computed } from "vue";
import type { Artist } from "@/types";
import { useUiStore } from "@/stores/ui";
import { displayName } from "@/utils";
import { Mono, IconButton, Flag } from "@/components/ui";
import IconFetch from "~icons/mdi/refresh";
import IconClose from "~icons/mdi/close";

const ui = useUiStore();

const props = defineProps<{
  artist: Artist;
  selected: boolean;
  flags: { code: string; n: number }[];
  upcomingListed?: number;
  fetching?: boolean;
  // Uncategorized artists have no category to be removed from — hide the button there.
  hideRemove?: boolean;
}>();

const emit = defineEmits<{ select: []; fetch: []; remove: [] }>();

const isManual = computed(() => props.artist.fetch_mode === "manual");
// Dormant: an auto artist with no upcoming listed events right now (distinct from manual).
const isDormant = computed(() => !isManual.value && (props.upcomingListed ?? 0) === 0);
</script>

<template>
  <div
    class="group flex items-center px-2.5 py-1.5 pl-8 border-b border-line/50 cursor-pointer gap-2"
    :class="selected ? 'bg-surface-sel' : 'hover:bg-surface-2'"
    @click="emit('select')"
  >
    <div class="flex-1 min-w-0">
      <div class="text-xs truncate" :class="selected ? 'text-accent-text font-bold' : 'text-muted'">
        {{ displayName(artist.name) }}
      </div>
      <div class="flex gap-1.5 mt-0.5 flex-wrap items-center">
        <span
          v-for="f in flags"
          :key="f.code"
          class="inline-flex items-center gap-0.5 cursor-pointer hover:opacity-80"
          :title="`Toggle events in ${f.code.toUpperCase()}`"
          @click.stop="ui.toggleCountry(f.code)"
        >
          <Flag :code="f.code" :w="14" />
          <Mono size="10" class="text-faint">×{{ f.n }}</Mono>
        </span>
        <Mono
          v-if="isManual"
          size="9"
          class="px-1 py-px rounded-xs bg-surface-3 border border-line-2 text-faint uppercase tracking-wide"
          title="Manual — history only, never fetched"
        >
          manual
        </Mono>
        <Mono
          v-else-if="isDormant"
          size="9"
          class="text-ghost italic"
          title="No upcoming listed events"
        >
          no upcoming
        </Mono>
      </div>
    </div>

    <span
      v-if="(upcomingListed ?? 0) > 0"
      class="shrink-0 px-1.5 py-px rounded-xs bg-surface-3 border border-line-2 text-ghost font-mono text-micro tabular-nums"
      :title="`${upcomingListed} upcoming event${upcomingListed === 1 ? '' : 's'}`"
    >
      {{ upcomingListed }}
    </span>

    <!-- remove sits inboard (reserved space, shown on hover); fetch stays the rightmost
         element so its icon aligns to the edge across every row, categorized or not. -->
    <IconButton
      v-if="!hideRemove && ui.editMode"
      tone="danger"
      size="sm"
      title="Remove artist from category"
      @click.stop="emit('remove')"
    >
      <IconClose class="h-3.5 w-3.5" />
    </IconButton>
    <IconButton
      tone="accent"
      size="sm"
      title="Fetch events from Ticketmaster"
      @click.stop="emit('fetch')"
    >
      <IconFetch class="h-3.5 w-3.5" :class="fetching ? 'animate-spin' : ''" />
    </IconButton>
  </div>
</template>
