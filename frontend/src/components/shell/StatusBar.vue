<script setup lang="ts">
import { computed } from "vue";
import { useEventsStore } from "@/stores/events";
import { useUiStore } from "@/stores/ui";
import { formatRelativeDate } from "@/utils";

const events = useEventsStore();
const ui = useUiStore();

const total = computed(() => events.events.length);
const interested = computed(() => events.interestedCount);
const going = computed(() => events.goingCount);
const attended = computed(() => events.attendedCount);

const sortLabel = computed(
  () => ({ "date-asc": "DATE ASC", "date-desc": "DATE DESC", artist: "ARTIST" })[ui.sort],
);

const lastFetch = computed(() => (ui.lastFetch ? formatRelativeDate(ui.lastFetch) : "—"));
</script>

<template>
  <!-- box-content keeps the h-7 text row at 28px and puts the home-indicator inset below it,
       rather than eating into the row. env() is 0px on hardware without a notch. -->
  <footer
    class="h-7 bg-surface border-t border-line flex items-center px-gutter gap-3 sm:gap-5 text-micro font-mono text-ghost overflow-hidden whitespace-nowrap shrink-0 pb-[env(safe-area-inset-bottom)] box-content"
  >
    <span class="text-muted">
      {{ total }} events
      <span class="text-ghost">·</span>
      <span class="text-accent-text">{{ interested }} interested</span>
      <span class="text-ghost">·</span>
      <span class="text-go">{{ going }} going</span>
      <span class="text-ghost">·</span>
      <span class="text-muted">{{ attended }} attended</span>
    </span>
    <span class="text-faint hidden sm:inline">|</span>
    <span class="hidden sm:inline">sorted by {{ sortLabel }}</span>
    <span class="mr-auto" />
    <span class="hidden md:inline">Last fetch: {{ lastFetch }}</span>
  </footer>
</template>
