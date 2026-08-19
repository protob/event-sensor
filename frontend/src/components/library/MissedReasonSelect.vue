<script setup lang="ts">
import { MISSED_REASONS } from "@/utils/eventStatus";

const props = withDefaults(defineProps<{ reason?: string | null; placeholder?: string }>(), {
  placeholder: "reason…",
});
const emit = defineEmits<{ select: [reason: string] }>();
</script>

<template>
  <select
    :value="props.reason ?? ''"
    class="bg-surface-2 border border-line-2 rounded-sm text-label text-muted font-mono px-1 py-0.5"
    @change="emit('select', ($event.target as HTMLSelectElement).value)"
  >
    <option value="">{{ props.placeholder }}</option>
    <option v-for="r in MISSED_REASONS" :key="r" :value="r">{{ r }}</option>
  </select>
</template>
