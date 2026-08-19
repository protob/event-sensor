<script setup lang="ts">
withDefaults(defineProps<{ n: number; prefix?: string }>(), { prefix: "" });
</script>

<template>
  <span
    class="px-1.5 py-px rounded-xs bg-accent-strong border border-accent-bright text-accent-text font-mono text-meta tabular-nums"
  >
    {{ prefix }}{{ n }}
  </span>
</template>
