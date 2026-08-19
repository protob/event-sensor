<script setup lang="ts">
import { computed, ref, nextTick } from "vue";
import { useEventListener } from "@vueuse/core";
import { useUiStore } from "@/stores/ui";
import { Flag } from "@/components/ui";
import IconChevron from "~icons/mdi/chevron-down";
import IconCheck from "~icons/mdi/check";

// Multi-select country filter for the events list. Empty selection = whole region (all).
// Replaces the old single native <select>: a country is rarely a trip ("Benelux week" =
// be+nl+lu), so the filter is an additive set. Flag-clicks elsewhere toggle the same set.
const props = defineProps<{
  options: { code: string; name: string; count: number }[];
}>();

const ui = useUiStore();

const open = ref(false);
const query = ref("");
const trigger = ref<HTMLElement | null>(null);
const panel = ref<HTMLElement | null>(null);

// The popover is teleported to <body> and positioned with fixed coords. It has to escape
// the toolbar's `overflow-x-auto`, which otherwise clips anything spilling below the row.
const pos = ref({ top: 0, left: 0 });

function place() {
  const r = trigger.value?.getBoundingClientRect();
  if (r) pos.value = { top: r.bottom + 4, left: r.left };
}

// Search only earns its keep once the list is long; below this it's just clutter.
const showSearch = computed(() => props.options.length > 12);

const filtered = computed(() => {
  const q = query.value.trim().toLowerCase();
  if (!q) return props.options;
  return props.options.filter((o) => o.name.toLowerCase().includes(q) || o.code.includes(q));
});

const selected = computed(() => ui.countryFilters);
const isOn = (code: string) => selected.value.includes(code);

// Up to 3 flags in the trigger, then "+N"; keeps the toolbar from wrapping.
const triggerFlags = computed(() => selected.value.slice(0, 3));
const overflow = computed(() => Math.max(0, selected.value.length - 3));

async function toggleOpen() {
  open.value = !open.value;
  if (open.value) {
    query.value = "";
    await nextTick();
    place();
  }
}

// Keep the panel glued to the trigger if the page/toolbar scrolls or the window resizes.
function onReflow() {
  if (open.value) place();
}

// Close when a click lands outside the trigger and the (teleported) panel.
useEventListener(document, "mousedown", (e: MouseEvent) => {
  const t = e.target as Node;
  if (trigger.value?.contains(t) || panel.value?.contains(t)) return;
  open.value = false;
});
useEventListener(window, "resize", onReflow);
useEventListener(window, "scroll", onReflow, { capture: true });
</script>

<template>
  <div class="shrink-0">
    <button
      ref="trigger"
      type="button"
      title="Filter by country"
      class="h-[26px] flex items-center gap-1 bg-surface border rounded-sm text-label font-mono px-1.5 focus:outline-none focus:border-accent-bright"
      :class="
        selected.length
          ? 'border-accent-bright text-accent-text'
          : 'border-line-2 text-muted hover:text-body'
      "
      @click="toggleOpen"
    >
      <template v-if="selected.length">
        <Flag v-for="cc in triggerFlags" :key="cc" :code="cc" :w="16" />
        <span v-if="overflow">+{{ overflow }}</span>
      </template>
      <span v-else>All countries</span>
      <IconChevron class="h-3.5 w-3.5 opacity-70" />
    </button>

    <Teleport to="body">
      <div
        v-if="open"
        ref="panel"
        class="fixed z-[100] w-60 bg-surface border border-line-2 rounded-md shadow-lg flex flex-col"
        :style="{ top: pos.top + 'px', left: pos.left + 'px' }"
      >
        <!-- header: count + clear -->
        <div
          class="flex items-center justify-between px-2 py-1.5 border-b border-line text-label font-mono text-faint"
        >
          <span>{{ selected.length || "no" }} selected</span>
          <button
            v-if="selected.length"
            class="text-muted hover:text-danger"
            @click="ui.clearCountries()"
          >
            Clear
          </button>
        </div>

        <!-- search (only when the list is long) -->
        <div v-if="showSearch" class="p-1.5 border-b border-line">
          <input
            v-model="query"
            type="text"
            placeholder="Search countries…"
            class="w-full bg-app border border-line-2 rounded-sm text-label font-mono px-1.5 py-1 text-body placeholder:text-faint focus:outline-none focus:border-accent-bright"
          />
        </div>

        <!-- checkbox list -->
        <div class="max-h-72 overflow-y-auto py-1">
          <button
            v-for="c in filtered"
            :key="c.code"
            type="button"
            class="w-full flex items-center gap-2 pl-2 pr-3 py-1 text-xs font-mono text-left hover:bg-surface-3"
            :class="isOn(c.code) ? 'text-accent-text' : 'text-body'"
            @click="ui.toggleCountry(c.code)"
          >
            <span
              class="h-3 w-3 shrink-0 rounded-[2px] border flex items-center justify-center"
              :class="isOn(c.code) ? 'bg-accent-strong border-accent-bright' : 'border-line-2'"
            >
              <IconCheck v-if="isOn(c.code)" class="h-2.5 w-2.5 text-accent-text" />
            </span>
            <Flag :code="c.code" :w="16" />
            <span class="truncate flex-1">{{ c.name }}</span>
            <span class="text-faint tabular-nums">{{ c.count }}</span>
          </button>
          <div
            v-if="!filtered.length"
            class="px-2 py-2 text-label font-mono text-faint text-center"
          >
            No matches
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
