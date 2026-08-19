<script setup lang="ts">
import { useEventListener } from "@vueuse/core";
import Mono from "./Mono.vue";

// Scrim + panel chrome shared by the app's modals. Stays mounted at the call site;
// the `open` prop gates rendering and the Escape binding.
const props = withDefaults(
  defineProps<{
    open: boolean;
    title: string;
    width?: string;
  }>(),
  { width: "max-w-md" },
);

const emit = defineEmits<{ close: [] }>();

useEventListener(window, "keydown", (e: KeyboardEvent) => {
  if (e.key === "Escape" && props.open) emit("close");
});
</script>

<template>
  <div
    v-if="props.open"
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"
    @click.self="emit('close')"
  >
    <div
      class="w-full bg-surface border border-line-2 rounded-md shadow-xl max-h-[90dvh] overflow-y-auto"
      :class="props.width"
    >
      <div class="flex items-center justify-between px-4 h-11 border-b border-line">
        <Mono size="xs" class="text-heading font-bold uppercase tracking-wide">
          {{ props.title }}
        </Mono>
        <button class="text-faint hover:text-body" @click="emit('close')">✕</button>
      </div>
      <slot />
    </div>
  </div>
</template>
