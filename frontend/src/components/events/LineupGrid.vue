<script setup lang="ts">
import { computed, ref } from "vue";
import type { Artist } from "@/types";
import { Mono } from "@/components/ui";
import LineupCell from "./LineupCell.vue";

const props = defineProps<{
  names: string[];
  // Where these names come from — drives the honest header/caption.
  source: "ticketmaster" | "entity" | "manual" | "none";
  entityByName: Map<string, Artist>;
}>();

const VISIBLE = 12;
const expandAll = ref(false);

const shown = computed(() => (expandAll.value ? props.names : props.names.slice(0, VISIBLE)));
const hidden = computed(() => Math.max(0, props.names.length - VISIBLE));

const heading = computed(() => (props.source === "entity" ? "Entity artists" : "Line-up"));

function entityArtist(name: string): Artist | null {
  return props.entityByName.get(name.toLowerCase().trim()) ?? null;
}
</script>

<template>
  <div>
    <div class="flex items-baseline gap-2.5 mb-3.5">
      <Mono size="sm" class="text-body font-semibold tracking-wide uppercase">{{ heading }}</Mono>
      <Mono size="10" class="text-ghost">{{ names.length }} artists</Mono>
      <button
        v-if="hidden > 0"
        class="ml-auto font-mono text-meta text-accent-text"
        @click="expandAll = !expandAll"
      >
        {{ expandAll ? "collapse ↑" : "expand all ↓" }}
      </button>
    </div>

    <div v-if="names.length" class="grid grid-cols-2 sm:grid-cols-3 gap-0.5">
      <LineupCell v-for="(name, i) in shown" :key="i" :name="name" :tracked="entityArtist(name)" />
      <button
        v-if="!expandAll && hidden > 0"
        class="px-2.5 py-[7px] bg-surface-3 border border-accent-chip-border rounded-sm text-xs text-accent-text flex items-center gap-1.5 font-mono"
        @click="expandAll = true"
      >
        +{{ hidden }} more →
      </button>
    </div>
    <Mono v-else size="11" class="text-faint">No artists listed for this event.</Mono>

    <!-- Honest source caption. Manual events are the user's own data — no caveat shown. -->
    <div
      v-if="source === 'ticketmaster' || source === 'entity'"
      class="mt-4 flex items-center gap-2 text-meta font-mono text-faint uppercase tracking-wide"
    >
      <template v-if="source === 'ticketmaster'">
        <span>Line-up via Ticketmaster</span>
        <span class="text-ghost">·</span>
        <span>may be incomplete</span>
        <span class="text-ghost">·</span>
        <span class="text-accent-text">entity artists linked</span>
      </template>
      <template v-else>
        <span>No detailed line-up from Ticketmaster</span>
        <span class="text-ghost">·</span>
        <span>showing entity artists on this event</span>
      </template>
    </div>
  </div>
</template>
