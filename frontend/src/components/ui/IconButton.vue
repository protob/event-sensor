<script setup lang="ts">
import { computed } from "vue";

// Square hover-highlight icon button (toggles, row actions).
const props = withDefaults(
  defineProps<{
    active?: boolean;
    tone?: "accent" | "go" | "danger" | "neutral";
    size?: "sm" | "md";
    title?: string;
  }>(),
  { active: false, tone: "neutral", size: "md" },
);

const toneCls = computed(() => {
  if (props.active) {
    switch (props.tone) {
      case "go":
        return "text-go";
      case "danger":
        return "text-danger";
      case "neutral":
        return "text-body";
      default:
        return "text-accent-text";
    }
  }
  switch (props.tone) {
    case "go":
      return "text-faint hover:text-go";
    case "danger":
      return "text-faint hover:text-danger";
    default:
      return "text-faint hover:text-muted";
  }
});

const sizeCls = computed(() => (props.size === "sm" ? "h-5 w-5" : "h-6 w-6"));
</script>

<template>
  <button
    type="button"
    :title="title"
    :aria-label="title"
    class="inline-flex items-center justify-center rounded-sm transition-colors hover:bg-surface-3 focus-visible:outline focus-visible:outline-1 focus-visible:outline-accent-bright"
    :class="[toneCls, sizeCls]"
  >
    <slot />
  </button>
</template>
