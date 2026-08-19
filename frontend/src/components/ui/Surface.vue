<script setup lang="ts">
// Panel wrapper used for settings cards / sidebars.
defineProps<{
  header?: string;
}>();
</script>

<template>
  <section class="bg-surface border border-line rounded-md overflow-hidden">
    <header
      v-if="header || $slots.header"
      class="px-3 py-2 border-b border-line flex items-center gap-2"
    >
      <slot name="header">
        <span class="font-mono text-meta uppercase tracking-wide text-muted font-semibold">
          {{ header }}
        </span>
      </slot>
    </header>
    <div class="p-3">
      <slot />
    </div>
  </section>
</template>
