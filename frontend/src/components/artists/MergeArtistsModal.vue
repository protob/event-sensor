<script setup lang="ts">
import { computed, ref, watch } from "vue";
import type { Artist } from "@/types";
import { useEventsStore } from "@/stores/events";
import { displayName } from "@/utils";
import { Btn, Mono, Modal } from "@/components/ui";

// Merge confirmation: pick the surviving artist (winner); the rest are merged away (deleted).
// Consequences are derived client-side from already-loaded events — no extra endpoint.
const props = defineProps<{ open: boolean; artists: Artist[] }>();
const emit = defineEmits<{ close: []; confirm: [from: string[], into: string] }>();

const events = useEventsStore();

const winner = ref<string>("");

// Default the winner to the first candidate whenever the set changes.
watch(
  () => props.artists.map((a) => a.id).join(","),
  () => {
    winner.value = props.artists[0]?.id ?? "";
  },
  { immediate: true },
);

const losers = computed(() => props.artists.filter((a) => a.id !== winner.value).map((a) => a.id));

// How many tracked performances will repoint to the winner (the losers' performance rows).
const repointCount = computed(() => {
  const loserSet = new Set(losers.value);
  let n = 0;
  for (const e of events.events) {
    for (const p of e.performances ?? []) {
      if (p.artist_id && loserSet.has(p.artist_id)) n++;
    }
  }
  return n;
});

function confirm() {
  if (!winner.value || losers.value.length === 0) return;
  emit("confirm", losers.value, winner.value);
}
</script>

<template>
  <Modal :open="open" title="Merge artists" @close="emit('close')">
    <div class="p-gutter flex flex-col gap-3">
      <Mono size="11" class="text-muted">
        Pick the artist to keep. The others are merged into it and deleted; their performances and
        category memberships repoint to the survivor.
      </Mono>

      <div class="flex flex-col gap-1.5">
        <label
          v-for="a in artists"
          :key="a.id"
          class="flex items-center gap-2 px-2.5 py-1.5 rounded-sm border cursor-pointer transition-colors"
          :class="
            winner === a.id
              ? 'bg-accent-chip border-accent-chip-border text-accent-text'
              : 'bg-surface-2 border-line-2 text-muted hover:text-body'
          "
        >
          <input v-model="winner" type="radio" :value="a.id" class="accent-accent-bright" />
          <span class="text-prose truncate">{{ displayName(a.name) }}</span>
          <Mono v-if="winner === a.id" size="9" class="ml-auto uppercase">keep</Mono>
        </label>
      </div>

      <Mono size="11" class="text-faint">
        {{ losers.length }} artist{{ losers.length === 1 ? "" : "s" }} merged away ·
        {{ repointCount }} performance{{ repointCount === 1 ? "" : "s" }} repointed
      </Mono>

      <div class="flex justify-end gap-2 pt-1">
        <Btn size="md" tone="neutral" @click="emit('close')">Cancel</Btn>
        <Btn size="md" tone="accent" :disabled="!winner || losers.length === 0" @click="confirm">
          Merge {{ losers.length }} →
        </Btn>
      </div>
    </div>
  </Modal>
</template>
