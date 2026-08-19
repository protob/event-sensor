<script setup lang="ts">
defineProps<{ title: string }>();
</script>

<template>
  <div class="w-full max-w-sm">
    <div class="bg-surface border border-line rounded-lg p-6">
      <h1 class="text-center text-base font-semibold text-heading mb-5">{{ title }}</h1>
      <slot />
    </div>
  </div>
</template>
