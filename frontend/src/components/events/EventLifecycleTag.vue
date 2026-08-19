<script setup lang="ts">
import type { Event } from "@/types";
import { Tag } from "@/components/ui";

// Renders a world-side lifecycle marker only when it matters; ordinary upcoming `listed`
// events render nothing so rows stay calm. `past` is shown only where asked (e.g. diary).
const props = withDefaults(defineProps<{ event: Event; showPast?: boolean }>(), {
  showPast: false,
});
</script>

<template>
  <Tag
    v-if="props.event.listing_state === 'cancelled'"
    color="#ef4444"
    data-testid="event-lifecycle"
    >CANCELLED</Tag
  >
  <Tag
    v-else-if="props.event.listing_state === 'delisted'"
    color="#71717a"
    data-testid="event-lifecycle"
    >NO LONGER LISTED</Tag
  >
  <Tag
    v-else-if="props.showPast && props.event.is_past"
    color="#52525b"
    data-testid="event-lifecycle"
    >PAST</Tag
  >
</template>
