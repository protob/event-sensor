<script setup lang="ts">
import Btn from "./Btn.vue";

// Controlled dropdown: the parent owns which menu is open, so sibling menus in the
// same bar stay mutually exclusive. Renders a trigger button and an upward list.
const props = withDefaults(
  defineProps<{
    label: string;
    open: boolean;
    width?: string;
    maxHeight?: string;
  }>(),
  { width: "min-w-[140px]", maxHeight: "" },
);

const emit = defineEmits<{ toggle: [] }>();

// The root is a positioning wrapper; a data-testid set by the caller belongs on the
// trigger, which is the thing that gets clicked.
defineOptions({ inheritAttrs: false });
</script>

<template>
  <div class="relative">
    <Btn size="sm" tone="neutral" v-bind="$attrs" @click="emit('toggle')">{{ props.label }} ▾</Btn>
    <div
      v-if="props.open"
      class="absolute bottom-full mb-1 left-0 rounded-sm border border-line-2 bg-surface shadow-lg overflow-hidden"
      :class="[props.width, props.maxHeight]"
    >
      <slot />
    </div>
  </div>
</template>
