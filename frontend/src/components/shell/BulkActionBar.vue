<script setup lang="ts">
import { ref } from "vue";
import { onClickOutside } from "@vueuse/core";
import type { EventStatus } from "@/types";
import { Btn, DropdownMenu } from "@/components/ui";

// A contextual bulk-action bar that slides up from the viewport bottom while a selection is
// active. Shows only the actions valid for the current list context and emits intents the
// parent view handles (calling the store bulk actions with selection.ids()).
withDefaults(
  defineProps<{
    count: number;
    context: "events" | "artists" | "library" | "category" | "venues";
    categories?: { id: string; name: string }[];
  }>(),
  { categories: () => [] },
);

const emit = defineEmits<{
  "set-status": [status: EventStatus | ""];
  delete: [];
  "fetch-mode": [mode: "auto" | "manual"];
  merge: [];
  "add-category": [categoryId: string];
  "remove-category": [categoryId: string];
  clear: [];
}>();

// One open menu at a time across the whole bar.
type Menu = "status" | "fetch" | "add-cat" | "remove-cat" | null;
const open = ref<Menu>(null);
const root = ref<HTMLElement | null>(null);
onClickOutside(root, () => (open.value = null));

function toggle(menu: Menu) {
  open.value = open.value === menu ? null : menu;
}

const STATUSES: { value: EventStatus | ""; label: string }[] = [
  { value: "interested", label: "Interested" },
  { value: "going", label: "Going" },
  { value: "attended", label: "Attended" },
  { value: "missed", label: "Missed" },
  { value: "", label: "Unclaim" },
];

const FETCH_MODES: { value: "auto" | "manual"; label: string }[] = [
  { value: "auto", label: "Auto" },
  { value: "manual", label: "Manual" },
];

function pick(fn: () => void) {
  fn();
  open.value = null;
}

const itemCls =
  "w-full text-left px-3 py-1.5 text-label text-muted hover:bg-surface-2 hover:text-body";
</script>

<template>
  <Transition
    enter-active-class="transition duration-150 ease-out"
    enter-from-class="translate-y-full opacity-0"
    enter-to-class="translate-y-0 opacity-100"
    leave-active-class="transition duration-100 ease-in"
    leave-from-class="translate-y-0 opacity-100"
    leave-to-class="translate-y-full opacity-0"
  >
    <div
      v-if="count > 0"
      ref="root"
      class="absolute bottom-3 left-1/2 -translate-x-1/2 z-30 flex items-center gap-3 px-3 py-2 rounded-md border border-line-2 bg-surface shadow-xl shadow-black/30 font-mono"
      data-testid="bulk-bar"
    >
      <span
        class="text-xs text-heading font-semibold tabular-nums whitespace-nowrap"
        data-testid="bulk-count"
      >
        {{ count }} selected
      </span>
      <span class="w-px h-5 bg-line" />

      <!-- events / library actions -->
      <template v-if="context === 'events' || context === 'library'">
        <DropdownMenu label="Set status" :open="open === 'status'" @toggle="toggle('status')">
          <button
            v-for="s in STATUSES"
            :key="s.value || 'unclaim'"
            :class="itemCls"
            :data-testid="'bulk-status-' + (s.value || 'unclaim')"
            @click="pick(() => emit('set-status', s.value))"
          >
            {{ s.label }}
          </button>
        </DropdownMenu>
        <Btn size="sm" tone="danger" data-testid="bulk-delete" @click="emit('delete')">Delete</Btn>
      </template>

      <!-- artists actions -->
      <template v-else-if="context === 'artists'">
        <DropdownMenu
          label="Fetch-mode"
          width="min-w-[120px]"
          :open="open === 'fetch'"
          @toggle="toggle('fetch')"
        >
          <button
            v-for="m in FETCH_MODES"
            :key="m.value"
            :class="itemCls"
            @click="pick(() => emit('fetch-mode', m.value))"
          >
            {{ m.label }}
          </button>
        </DropdownMenu>

        <DropdownMenu
          v-if="categories.length"
          label="Add to category"
          width="min-w-[160px]"
          max-height="max-h-64 overflow-y-auto"
          :open="open === 'add-cat'"
          @toggle="toggle('add-cat')"
        >
          <button
            v-for="c in categories"
            :key="c.id"
            :class="[itemCls, 'truncate']"
            @click="pick(() => emit('add-category', c.id))"
          >
            {{ c.name }}
          </button>
        </DropdownMenu>

        <DropdownMenu
          v-if="categories.length"
          label="Remove from category"
          width="min-w-[160px]"
          max-height="max-h-64 overflow-y-auto"
          :open="open === 'remove-cat'"
          @toggle="toggle('remove-cat')"
        >
          <button
            v-for="c in categories"
            :key="c.id"
            :class="[itemCls, 'truncate']"
            @click="pick(() => emit('remove-category', c.id))"
          >
            {{ c.name }}
          </button>
        </DropdownMenu>

        <Btn size="sm" tone="outline" :disabled="count < 2" @click="emit('merge')">Merge</Btn>
        <Btn size="sm" tone="danger" @click="emit('delete')">Delete</Btn>
      </template>

      <!-- venues actions -->
      <template v-else-if="context === 'venues'">
        <Btn size="sm" tone="outline" :disabled="count < 2" @click="emit('merge')">Merge</Btn>
      </template>

      <span class="w-px h-5 bg-line" />
      <button
        class="text-meta text-faint hover:text-muted whitespace-nowrap"
        title="Clear selection (Esc)"
        data-testid="bulk-clear"
        @click="emit('clear')"
      >
        Esc to clear
      </button>
    </div>
  </Transition>
</template>
