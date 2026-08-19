<script setup lang="ts">
import type { EventStatus } from "@/types";
import { EVENT_STATUSES } from "@/utils/eventStatus";

const props = defineProps<{ status?: EventStatus | null }>();
const emit = defineEmits<{ select: [status: EventStatus] }>();
</script>

<template>
  <select
    :value="props.status ?? ''"
    class="bg-surface-2 border border-line-2 rounded-sm text-label text-muted font-mono px-1 py-0.5"
    @change="emit('select', ($event.target as HTMLSelectElement).value as EventStatus)"
  >
    <option v-for="s in EVENT_STATUSES" :key="s" :value="s">{{ s }}</option>
  </select>
</template>
