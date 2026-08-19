<script setup lang="ts">
import { RouterLink } from "vue-router";
</script>

<template>
  <div class="h-full flex flex-col items-center justify-center gap-3 bg-app">
    <span class="font-mono text-5xl font-bold text-faint tabular-nums">404</span>
    <p class="text-muted text-sm">Page not found</p>
    <RouterLink to="/events" class="font-mono text-label text-accent-text"
      >← Back to Events</RouterLink
    >
  </div>
</template>
