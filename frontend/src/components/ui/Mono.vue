<script setup lang="ts">
// Constant JetBrains-Mono metadata wrapper.
withDefaults(
  defineProps<{
    size?: "9" | "10" | "11" | "xs" | "sm";
    dim?: boolean;
  }>(),
  { size: "xs", dim: false },
);

const sizeClass: Record<string, string> = {
  "9": "text-micro",
  "10": "text-meta",
  "11": "text-label",
  xs: "text-xs",
  sm: "text-sm",
};
</script>

<template>
  <span class="font-mono tabular-nums" :class="[sizeClass[size], dim ? 'text-faint' : '']">
    <slot />
  </span>
</template>
