<script setup lang="ts">
import { computed } from "vue";
import IconLoading from "~icons/mdi/loading";

const props = withDefaults(
  defineProps<{
    tone?: "accent" | "neutral" | "danger" | "outline";
    size?: "sm" | "md";
    loading?: boolean;
    disabled?: boolean;
    type?: "button" | "submit" | "reset";
  }>(),
  { tone: "accent", size: "md", loading: false, disabled: false, type: "button" },
);

const toneCls = computed(() => {
  switch (props.tone) {
    case "neutral":
      return "bg-surface-3 text-muted border-line-2 hover:text-body";
    case "danger":
      return "bg-danger-chip text-danger border-danger-chip-border hover:brightness-125";
    case "outline":
      return "bg-transparent text-accent-text border-accent-chip-border hover:bg-accent-chip";
    default:
      return "bg-accent text-white border-accent-bright hover:brightness-110";
  }
});

const sizeCls = computed(() =>
  props.size === "sm" ? "px-2 py-1 text-label gap-1" : "px-3 py-1.5 text-xs gap-1.5",
);
</script>

<template>
  <!-- whitespace-nowrap because a button label broken across lines is never what was meant:
       it bursts the fixed-height toolbars it sits in. A label too long for the width is one to
       shorten at the call site, not to wrap here. -->
  <button
    :type="type"
    :disabled="disabled || loading"
    class="inline-flex items-center justify-center font-medium rounded-sm border transition-colors disabled:opacity-50 disabled:cursor-not-allowed font-mono whitespace-nowrap"
    :class="[toneCls, sizeCls]"
  >
    <IconLoading v-if="loading" class="animate-spin h-3.5 w-3.5" />
    <slot />
  </button>
</template>
