<script setup lang="ts">
import { computed, ref, watch } from "vue";
import type { Venue } from "@/types";
import { Btn, Mono, Modal } from "@/components/ui";

// Merge confirmation: pick the surviving venue; the rest are repointed away and deleted.
const props = defineProps<{ open: boolean; venues: Venue[] }>();
const emit = defineEmits<{ close: []; confirm: [from: string[], into: string] }>();

const winner = ref<string>("");

// Default the winner to the candidate with the most events - the one most likely to be the
// "real" row rather than a stray duplicate.
watch(
  () => props.venues.map((v) => v.id).join(","),
  () => {
    const best = [...props.venues].sort((a, b) => (b.event_count ?? 0) - (a.event_count ?? 0))[0];
    winner.value = best?.id ?? "";
  },
  { immediate: true },
);

const losers = computed(() => props.venues.filter((v) => v.id !== winner.value));
const repointCount = computed(() => losers.value.reduce((n, v) => n + (v.event_count ?? 0), 0));

function label(v: Venue) {
  return [v.name, v.city, v.country_code?.toUpperCase()].filter(Boolean).join(" · ");
}

function confirm() {
  if (!winner.value || losers.value.length === 0) return;
  emit(
    "confirm",
    losers.value.map((v) => v.id),
    winner.value,
  );
}
</script>

<template>
  <Modal :open="open" title="Merge venues" @close="emit('close')">
    <div class="p-gutter flex flex-col gap-3">
      <Mono size="11" class="text-muted">
        Pick the venue to keep. The others are deleted and their events repoint to the survivor.
      </Mono>

      <div class="flex flex-col gap-1.5">
        <label
          v-for="v in venues"
          :key="v.id"
          class="flex items-center gap-2 px-2.5 py-1.5 rounded-sm border cursor-pointer transition-colors"
          :class="
            winner === v.id
              ? 'bg-accent-chip border-accent-chip-border text-accent-text'
              : 'bg-surface-2 border-line-2 text-muted hover:text-body'
          "
        >
          <input v-model="winner" type="radio" :value="v.id" class="accent-accent-bright" />
          <span class="text-prose truncate">{{ label(v) }}</span>
          <Mono size="9" class="ml-auto shrink-0 tabular-nums">{{ v.event_count ?? 0 }}</Mono>
        </label>
      </div>

      <Mono size="11" class="text-faint">
        {{ losers.length }} venue{{ losers.length === 1 ? "" : "s" }} deleted ·
        {{ repointCount }} event{{ repointCount === 1 ? "" : "s" }} repointed
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
