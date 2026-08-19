<script setup lang="ts">
const model = defineModel<boolean>({ default: false });
defineProps<{ disabled?: boolean }>();
</script>

<template>
  <button
    type="button"
    role="switch"
    :aria-checked="model"
    :disabled="disabled"
    class="relative inline-flex h-5 w-[38px] items-center rounded-full border transition-colors disabled:opacity-50"
    :class="model ? 'bg-accent border-accent-bright' : 'bg-surface-3 border-line-2'"
    @click="model = !model"
  >
    <span
      class="inline-block h-3.5 w-3.5 rounded-full bg-white transition-transform"
      :class="model ? 'translate-x-[20px]' : 'translate-x-0.5'"
    />
  </button>
</template>
