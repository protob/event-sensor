<script setup lang="ts">
import { computed } from "vue";

// Filter / country pill. Slightly larger than Tag.
const props = withDefaults(
  defineProps<{
    active?: boolean;
    tone?: "accent" | "go" | "danger" | "neutral";
  }>(),
  { active: false, tone: "neutral" },
);

const cls = computed(() => {
  if (props.active) {
    switch (props.tone) {
      case "go":
        return "bg-go-chip border-go-chip-border text-go";
      case "danger":
        return "bg-danger-chip border-danger-chip-border text-danger";
      default:
        return "bg-accent-strong border-accent-bright text-accent-text";
    }
  }
  switch (props.tone) {
    case "accent":
      return "bg-accent-chip border-accent-chip-border text-accent-text";
    case "go":
      return "bg-transparent border-line-2 text-muted hover:text-go";
    case "danger":
      return "bg-transparent border-line-2 text-muted hover:text-danger";
    default:
      return "bg-transparent border-line-2 text-muted hover:text-body";
  }
});
</script>

<template>
  <span
    class="inline-flex items-center gap-1 px-2 py-0.5 text-label font-mono rounded-sm border cursor-pointer select-none whitespace-nowrap transition-colors"
    :class="cls"
  >
    <slot />
  </span>
</template>
