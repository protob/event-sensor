<script setup lang="ts">
import { computed } from "vue";

// Small uppercase chip (event kind tag).
// When `color` is a real hex, tint bg/border/text from it; otherwise use accent-chip tokens.
const props = withDefaults(
  defineProps<{
    color?: string;
  }>(),
  { color: "" },
);

const isHex = computed(() => /^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$/.test(props.color));

const style = computed(() => {
  if (!isHex.value) return {};
  return {
    backgroundColor: props.color + "18",
    borderColor: props.color + "44",
    color: props.color,
  };
});
</script>

<template>
  <span
    class="inline-flex items-center px-1.5 py-px text-micro font-mono uppercase tracking-wide rounded-xs border"
    :class="isHex ? '' : 'bg-accent-chip border-accent-chip-border text-accent-text'"
    :style="style"
  >
    <slot />
  </span>
</template>
