<script setup lang="ts">
import { computed, ref } from "vue";
import { useRouter } from "vue-router";
import type { Performance } from "@/types";
import { displayName, formatShort } from "@/utils";

// Festival lineup chips: entity acts (clickable), name-only acts collapsed behind
// a "+N more" toggle. Owns the toggle so each row expands independently.
const props = defineProps<{ performances?: Performance[] }>();

const router = useRouter();

// Per-act detail suffix: " · Hotel Forum · 30 Jun" (venue and/or day), only when set.
function actMeta(p: Performance): string {
  const parts: string[] = [];
  if (p.venue_name) parts.push(p.venue_name);
  if (p.date) parts.push(formatShort(p.date));
  return parts.length ? ` · ${parts.join(" · ")}` : "";
}

const trackedActs = computed(() => (props.performances ?? []).filter((p) => p.artist_id));
const extraActs = computed(() => (props.performances ?? []).filter((p) => !p.artist_id));
const showExtras = ref(false);
</script>

<template>
  <div v-if="trackedActs.length || extraActs.length" class="flex items-center gap-1 flex-wrap">
    <span
      v-for="a in trackedActs"
      :key="a.artist_id ?? a.artist_name"
      class="px-1.5 py-px rounded-xs bg-accent-chip border border-accent-chip-border text-accent-text font-mono text-meta cursor-pointer hover:brightness-125"
      @click="router.push(`/artists/${a.artist_id}`)"
    >
      {{ displayName(a.artist_name) }}
    </span>
    <template v-if="showExtras">
      <span
        v-for="(x, i) in extraActs"
        :key="`x-${i}`"
        class="px-1.5 py-px rounded-xs bg-surface-3 border border-line-2 text-faint font-mono text-meta"
      >
        {{ displayName(x.artist_name) }}{{ actMeta(x) }}
      </span>
    </template>
    <button
      v-if="extraActs.length"
      class="text-faint hover:text-body font-mono text-meta px-1"
      :title="showExtras ? 'Hide name-only acts' : 'Show name-only acts'"
      @click="showExtras = !showExtras"
    >
      {{ showExtras ? "− hide" : `+${extraActs.length} more` }}
    </button>
  </div>
</template>
