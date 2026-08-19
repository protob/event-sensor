<script setup lang="ts">
import { RouterLink } from "vue-router";
import type { Artist } from "@/types";
import IconEye from "~icons/mdi/eye-outline";

defineProps<{
  name: string;
  tracked?: Artist | null;
}>();
</script>

<template>
  <component
    :is="tracked ? RouterLink : 'div'"
    :to="tracked ? `/artists/${tracked.id}` : undefined"
    class="px-2.5 py-[7px] bg-surface-2 border border-line rounded-sm text-xs flex items-center justify-between gap-1"
    :class="tracked ? 'text-accent-text hover:border-accent-chip-border' : 'text-muted'"
  >
    <span class="truncate">{{ name }}</span>
    <IconEye class="h-3.5 w-3.5 shrink-0" :class="tracked ? 'text-accent-text' : 'text-faint'" />
  </component>
</template>
