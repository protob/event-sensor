<script setup lang="ts">
import { IconButton } from "@/components/ui";
import IconInterested from "~icons/mdi/bookmark";
import IconInterestedOutline from "~icons/mdi/bookmark-outline";
import IconCheck from "~icons/mdi/check-circle";
import IconCheckOutline from "~icons/mdi/check-circle-outline";

// Two dense toggles on the new library model:
//  - bookmark filled  = claimed at all (status != null); click toggles interested <-> unclaimed
//  - check filled     = status === 'going';            click toggles going <-> interested
defineProps<{
  interested?: boolean;
  going?: boolean;
}>();

const emit = defineEmits<{ interested: []; going: [] }>();
</script>

<template>
  <div class="flex items-center gap-0.5 shrink-0" @click.stop>
    <IconButton
      :active="interested"
      tone="accent"
      size="sm"
      :title="interested ? 'Interested' : 'Not interested'"
      data-testid="claim-interested"
      @click="emit('interested')"
    >
      <IconInterested v-if="interested" class="h-4 w-4" />
      <IconInterestedOutline v-else class="h-4 w-4" />
    </IconButton>
    <IconButton
      :active="going"
      tone="go"
      size="sm"
      :title="going ? 'Going' : 'Not going'"
      data-testid="claim-going"
      @click="emit('going')"
    >
      <IconCheck v-if="going" class="h-4 w-4" />
      <IconCheckOutline v-else class="h-4 w-4" />
    </IconButton>
  </div>
</template>
