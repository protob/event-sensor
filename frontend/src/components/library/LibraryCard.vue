<script setup lang="ts">
import { computed } from "vue";
import { useRouter } from "vue-router";
import type { Event } from "@/types";
import { formatISO, displayName } from "@/utils";
import { useLibraryEntry } from "@/composables/useLibraryEntry";
import { Mono, Flag, Tag } from "@/components/ui";
import StatusSelect from "./StatusSelect.vue";
import NoteInput from "./NoteInput.vue";
import ActChips from "./ActChips.vue";
import EventLifecycleTag from "@/components/events/EventLifecycleTag.vue";

const props = defineProps<{ event: Event; checkable?: boolean; checked?: boolean }>();
const emit = defineEmits<{ edit: [Event]; select: [ev: MouseEvent] }>();

const router = useRouter();
const { isFestival, title, kindBadge, setStatus, saveNote, unclaim } = useLibraryEntry(
  computed(() => props.event),
);
</script>

<template>
  <div
    class="flex flex-col gap-1.5 p-2.5 border border-line bg-surface hover:bg-surface-2 rounded-sm transition-colors"
    :class="checked ? 'ring-1 ring-inset ring-accent-bright/40 bg-accent-chip/30' : ''"
  >
    <div class="flex items-start gap-2">
      <input
        v-if="checkable"
        type="checkbox"
        class="accent-accent-bright cursor-pointer shrink-0 mt-0.5"
        :checked="checked"
        title="Select (Shift-click for range)"
        @click="emit('select', $event)"
      />
      <span
        class="text-prose text-body truncate cursor-pointer hover:text-accent-text flex-1 min-w-0"
        @click="router.push(`/events/${event.id}`)"
      >
        {{ displayName(title) }}
      </span>
      <Tag v-if="kindBadge" :color="kindBadge.color" class="shrink-0">{{ kindBadge.label }}</Tag>
      <EventLifecycleTag :event="event" show-past />
    </div>

    <Mono size="10" class="text-faint tabular-nums">{{ formatISO(event.start_date) }}</Mono>

    <ActChips v-if="isFestival" :performances="event.performances" />

    <div class="flex items-center gap-1.5 mt-auto">
      <Flag v-if="event.venue?.country_code" :code="event.venue.country_code" :w="14" />
      <Mono size="10" class="text-muted truncate flex-1">
        {{ event.venue?.name || event.venue?.city || "—" }}
      </Mono>
      <StatusSelect :status="event.status" @select="setStatus" />
      <button
        v-if="event.source === 'manual'"
        class="text-faint hover:text-accent-text text-label font-mono"
        title="Edit event"
        @click="emit('edit', event)"
      >
        ✎
      </button>
      <button
        class="text-faint hover:text-danger text-label font-mono"
        title="Remove from library"
        @click="unclaim"
      >
        ✕
      </button>
    </div>

    <NoteInput :note="event.note" @save="saveNote" />
  </div>
</template>
