<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useVenuesStore } from "@/stores/venues";
import { useEventsStore } from "@/stores/events";
import { countryName } from "@/utils";
import { useSelection } from "@/composables/useSelection";
import { useListKeyboard } from "@/composables/useListKeyboard";
import { Mono, Flag } from "@/components/ui";
import BulkActionBar from "@/components/shell/BulkActionBar.vue";
import MergeVenuesModal from "@/components/venues/MergeVenuesModal.vue";

const store = useVenuesStore();
const events = useEventsStore();

const query = ref("");
const showMerge = ref(false);

onMounted(() => store.fetchVenues());

// Duplicates surface first: same name + city seen more than once, most-duplicated at the
// top, then alphabetical.
const dupKey = (name: string, city?: string) =>
  `${name.trim().toLowerCase()}|${(city ?? "").trim().toLowerCase()}`;

const dupCounts = computed(() => {
  const m: Record<string, number> = {};
  for (const v of store.venues) {
    const k = dupKey(v.name, v.city);
    m[k] = (m[k] ?? 0) + 1;
  }
  return m;
});

const visible = computed(() => {
  const q = query.value.trim().toLowerCase();
  const list = q
    ? store.venues.filter(
        (v) => v.name.toLowerCase().includes(q) || (v.city ?? "").toLowerCase().includes(q),
      )
    : store.venues;
  return [...list].sort((a, b) => {
    const da = dupCounts.value[dupKey(a.name, a.city)] ?? 1;
    const db = dupCounts.value[dupKey(b.name, b.city)] ?? 1;
    if (da !== db) return db - da;
    return a.name.localeCompare(b.name);
  });
});

const isDuplicate = (v: { name: string; city?: string }) =>
  (dupCounts.value[dupKey(v.name, v.city)] ?? 1) > 1;

const selection = useSelection(() => visible.value.map((v) => v.id));
watch(visible, () => selection.prune());

const selectedVenues = computed(() => store.venues.filter((v) => selection.has(v.id)));

function openMerge() {
  if (selection.count.value < 2) return;
  showMerge.value = true;
}

async function doMerge(from: string[], into: string) {
  showMerge.value = false;
  const res = await store.merge(from, into);
  // Repointed events now carry a different venue_id; refresh so the events views agree.
  if (res) await events.fetchEvents();
  selection.clear();
}

useListKeyboard(selection, () => visible.value.length === 0);
</script>

<template>
  <div class="h-full flex flex-col bg-app min-w-0 relative">
    <div
      class="min-h-12 px-gutter py-1.5 flex flex-wrap items-center gap-2 border-b border-line bg-surface-2 shrink-0"
    >
      <!-- Grows to 224px and gives width back below that, rather than holding 224px and
           pushing the count off the row. -->
      <input
        v-model="query"
        type="text"
        placeholder="Filter venues…"
        class="flex-1 min-w-24 max-w-56 bg-surface border border-line-2 rounded-sm px-2.5 py-1.5 text-sm text-body font-mono placeholder:text-faint focus:outline-none focus:border-accent-bright"
      />
      <Mono size="10" class="text-faint shrink-0 whitespace-nowrap">
        {{ visible.length }} venue{{ visible.length === 1 ? "" : "s" }}
      </Mono>
    </div>

    <div class="flex-1 overflow-y-auto min-h-0">
      <div
        v-if="store.loading && !store.venues.length"
        class="flex items-center justify-center h-32"
      >
        <Mono size="xs" class="text-faint">Loading venues…</Mono>
      </div>
      <p v-else-if="store.error" class="p-4 text-danger text-sm">{{ store.error }}</p>
      <div v-else-if="!visible.length" class="flex items-center justify-center h-40">
        <Mono size="xs" class="text-muted">No venues yet.</Mono>
      </div>
      <template v-else>
        <div
          v-for="(v, i) in visible"
          :key="v.id"
          class="group flex items-center px-gutter py-2.5 border-b border-line gap-3"
          :class="[
            selection.has(v.id)
              ? 'bg-accent-chip/40 ring-1 ring-inset ring-accent-bright/40'
              : i % 2 === 0
                ? 'bg-surface'
                : 'bg-surface-3',
          ]"
        >
          <input
            type="checkbox"
            class="accent-accent-bright cursor-pointer shrink-0"
            :checked="selection.has(v.id)"
            title="Select (Shift-click for range)"
            @click="selection.onRowClick(v.id, $event)"
          />
          <Flag v-if="v.country_code" :code="v.country_code" class="shrink-0" />
          <span class="flex-1 min-w-0 truncate text-sm text-body">{{ v.name }}</span>
          <Mono
            v-if="isDuplicate(v)"
            size="9"
            class="shrink-0 px-1.5 py-px rounded-xs bg-surface-3 border border-line-2 text-faint uppercase"
            title="Another venue shares this name and city"
          >
            dup
          </Mono>
          <Mono size="10" class="text-ghost shrink-0 truncate max-w-[180px]">
            {{ v.city || countryName(v.country_code ?? "") }}
          </Mono>
          <Mono size="10" class="text-ghost shrink-0 tabular-nums w-16 text-right">
            {{ v.event_count ?? 0 }} ev
          </Mono>
        </div>
      </template>
    </div>

    <BulkActionBar
      :count="selection.count.value"
      context="venues"
      @merge="openMerge"
      @clear="selection.clear()"
    />

    <MergeVenuesModal
      :open="showMerge"
      :venues="selectedVenues"
      @close="showMerge = false"
      @confirm="doMerge"
    />
  </div>
</template>
