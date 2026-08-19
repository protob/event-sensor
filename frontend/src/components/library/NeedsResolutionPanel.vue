<script setup lang="ts">
import { storeToRefs } from "pinia";
import { useEventsStore } from "@/stores/events";
import { formatISO } from "@/utils";
import { Pill, Mono } from "@/components/ui";
import MissedReasonSelect from "./MissedReasonSelect.vue";

// "Did you go?" nudge: past `going` events awaiting an attended/missed answer.
const store = useEventsStore();
const { needsResolution } = storeToRefs(store);

async function resolve(eventId: string, status: "attended" | "missed", reason?: string) {
  await store.setStatus(eventId, status, reason ? { missed_reason: reason } : undefined);
  await store.loadNeedsResolution();
}
</script>

<template>
  <div
    v-if="needsResolution.length"
    class="shrink-0 border-b border-line bg-surface px-gutter py-2"
  >
    <Mono size="9" class="text-go font-bold uppercase tracking-wide">Did you go?</Mono>
    <div v-for="n in needsResolution" :key="n.event_id" class="flex items-center gap-2 mt-1.5">
      <Mono size="11" class="text-muted truncate flex-1">
        {{ n.name }} <span class="text-faint">· {{ formatISO(n.start_date) }}</span>
      </Mono>
      <Pill tone="go" @click="resolve(n.event_id, 'attended')">Attended</Pill>
      <MissedReasonSelect placeholder="Missed…" @select="(r) => resolve(n.event_id, 'missed', r)" />
    </div>
  </div>
</template>
