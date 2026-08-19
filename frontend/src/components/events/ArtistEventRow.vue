<script setup lang="ts">
import { computed } from "vue";
import { useRouter } from "vue-router";
import type { Event } from "@/types";
import { useUiStore } from "@/stores/ui";
import { formatMedium } from "@/utils";
import { eventTag } from "./eventTag";
import { ARTIST_EVENT_GRID, CELL_FROM_LG, CELL_FROM_3XL } from "@/composables/useRowGrid";
import { Tag, Mono, IconButton, Flag } from "@/components/ui";
import WatchAttendCell from "./WatchAttendCell.vue";
import EventLifecycleTag from "./EventLifecycleTag.vue";
import IconExternal from "~icons/mdi/open-in-new";
import IconTrash from "~icons/mdi/trash-can-outline";

const props = defineProps<{
  event: Event;
  index: number;
  interested?: boolean;
  going?: boolean;
}>();

const emit = defineEmits<{ interested: []; going: []; remove: [] }>();

const ui = useUiStore();
const router = useRouter();

// Shared with the header in ArtistDetailView so the two cannot drift; the tiers drop the
// tag and actions, then the location, as the content pane narrows.
const GRID = ARTIST_EVENT_GRID;
const tag = computed(() => eventTag(props.event));
</script>

<template>
  <div
    class="grid items-center px-gutter py-[9px] border-b border-line cursor-pointer hover:bg-surface-2 transition-colors"
    :class="[GRID, index % 2 === 0 ? 'bg-surface' : 'bg-surface-3']"
    @click="router.push(`/events/${event.id}`)"
  >
    <WatchAttendCell
      :interested="interested"
      :going="going"
      @interested="emit('interested')"
      @going="emit('going')"
    />
    <div class="items-center gap-1" :class="CELL_FROM_LG">
      <Tag v-if="tag" :color="tag.color">{{ tag.label }}</Tag>
      <EventLifecycleTag :event="event" />
    </div>
    <div class="min-w-0 flex flex-col justify-center pr-2">
      <span class="truncate text-xs text-body">{{ event.venue?.name ?? event.name }}</span>
      <span v-if="event.kind === 'festival'" class="truncate text-meta text-faint font-mono">{{
        event.name
      }}</span>
    </div>
    <div class="items-center gap-1.5 min-w-0" :class="CELL_FROM_3XL">
      <Flag v-if="event.venue?.country_code" :code="event.venue.country_code" :w="16" />
      <Mono size="11" class="text-muted truncate">{{ event.venue?.city ?? "—" }}</Mono>
    </div>
    <div class="text-right">
      <Mono size="11" class="text-muted whitespace-nowrap">{{
        formatMedium(event.start_date)
      }}</Mono>
    </div>
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
</template>
