<script setup lang="ts">
// Saves on change; there is no draft state to keep locally.
const props = defineProps<{ note?: string | null; placeholder?: string }>();
const emit = defineEmits<{ save: [note: string] }>();
</script>

<template>
  <input
    :value="props.note ?? ''"
    type="text"
    :placeholder="placeholder ?? 'add a note…'"
    class="w-full bg-transparent border-b border-line/50 text-label text-muted font-mono py-0.5 focus:outline-none focus:border-accent-bright"
    @change="emit('save', ($event.target as HTMLInputElement).value)"
  />
</template>
