<script setup lang="ts">
import { computed, nextTick, onMounted, ref } from "vue";
import { storeToRefs } from "pinia";
import { useCategoriesStore } from "@/stores/categories";
import { useArtistsStore } from "@/stores/artists";
import { useAuthStore } from "@/stores/auth";
import { useToast } from "@/composables";
import { Mono, Pill } from "@/components/ui";
import CategoryTreeNode from "./CategoryTreeNode.vue";
import ArtistTreeRow from "./ArtistTreeRow.vue";
import IconChevronDown from "~icons/mdi/chevron-down";
import IconChevronRight from "~icons/mdi/chevron-right";
import IconCollapseAll from "~icons/mdi/unfold-less-horizontal";
import IconExpandAll from "~icons/mdi/unfold-more-horizontal";

defineProps<{ selectedArtistId?: string | null }>();
const emit = defineEmits<{ "select-artist": [id: string | null] }>();

const categories = useCategoriesStore();
const artists = useArtistsStore();
const auth = useAuthStore();
const toast = useToast();
const { categories: cats, loading } = storeToRefs(categories);

type Agg = { flags: { code: string; n: number }[]; claimed: number; upcomingListed: number };

// Per-artist flags/counts from GET /artists/summary. upcomingListed powers the "dormant"
// (no upcoming) marker.
const aggregates = computed(() => {
  const out = new Map<string, Agg>();
  for (const [id, s] of Object.entries(artists.summaries)) {
    out.set(id, {
      flags: s.countries.map((c) => ({ code: c.code, n: c.count })).sort((a, b) => b.n - a.n),
      claimed: s.claimed_count,
      upcomingListed: s.upcoming_listed_count,
    });
  }
  return out;
});

function aggFor(id: string): Agg {
  return aggregates.value.get(id) ?? { flags: [], claimed: 0, upcomingListed: 0 };
}

// --- expansion (controlled) ---
// The parent tracks which nodes are *collapsed*. Categories start collapsed on mount so
// their members load lazily on first expand; brand-new categories come up expanded for
// free, and a single toggle can collapse/expand everything.
const collapsed = ref<Set<string>>(new Set());
function toggleCat(id: string) {
  const s = new Set(collapsed.value);
  if (s.has(id)) s.delete(id);
  else s.add(id);
  collapsed.value = s;
}
const anyExpanded = computed(
  () =>
    cats.value.some((c) => !collapsed.value.has(c.id)) ||
    (uncategorized.value.length > 0 && uncatExpanded.value),
);
function toggleAll() {
  if (anyExpanded.value) {
    collapsed.value = new Set(cats.value.map((c) => c.id));
    uncatExpanded.value = false;
  } else {
    collapsed.value = new Set();
    uncatExpanded.value = true;
  }
}

// --- uncategorized (virtual node) ---
// Every auto artist whose summary lists no category. The memberships come from
// /artists/summary; a category's own member list loads lazily on expand.
const uncatExpanded = ref(true);
// Manual artists are history-only and belong to the Diary, not the discovery tree.
const uncategorized = computed(() =>
  artists.artists.filter(
    (a) => a.fetch_mode === "auto" && (artists.summaries[a.id]?.categories.length ?? 0) === 0,
  ),
);

const fetchingId = ref<string | null>(null);
async function fetchUncatArtist(id: string) {
  fetchingId.value = id;
  await artists.fetchAndSync(id);
  fetchingId.value = null;
}

// --- add category ---
const adding = ref(false);
const newName = ref("");
const input = ref<HTMLInputElement | null>(null);
async function startAdd() {
  adding.value = true;
  newName.value = "";
  await nextTick();
  input.value?.focus();
}
async function commitAdd() {
  const name = newName.value.trim();
  const uid = auth.user?.id;
  if (!name || !uid) {
    adding.value = false;
    return;
  }
  const created = await categories.createCategory({ name, user_id: uid });
  if (created) toast.success(`Category "${name}" created`);
  newName.value = "";
  adding.value = false;
}

onMounted(async () => {
  if (categories.categories.length === 0) await categories.fetchCategories();
  // Needed for the add-artist autocomplete (reuse existing artists).
  if (artists.artists.length === 0) artists.fetchArtists();
  // One call for every artist's categories, countries and counts.
  await artists.fetchSummaries();
  // Categories open collapsed: each node fetches its own members on first expand, so the
  // landing page costs a single request. Brand-new categories still open expanded (they
  // are created after this snapshot and never enter the set).
  collapsed.value = new Set(categories.categories.map((c) => c.id));
});
</script>

<template>
  <aside class="h-full bg-surface flex flex-col">
    <!-- header -->
    <div class="shrink-0 px-2.5 py-1.5 border-b border-line flex items-center gap-2">
      <Mono size="10" class="text-faint font-bold tracking-wide uppercase">Categories</Mono>
      <button
        v-if="cats.length"
        class="ml-auto text-faint hover:text-muted shrink-0"
        :title="anyExpanded ? 'Collapse all' : 'Expand all'"
        @click="toggleAll"
      >
        <IconCollapseAll v-if="anyExpanded" class="h-4 w-4" />
        <IconExpandAll v-else class="h-4 w-4" />
      </button>
      <span :class="cats.length ? '' : 'ml-auto'">
        <Pill tone="accent" :active="true" @click="startAdd">+ New</Pill>
      </span>
    </div>

    <!-- tree -->
    <div class="flex-1 overflow-y-auto">
      <div v-if="loading && cats.length === 0" class="px-2.5 py-2">
        <Mono size="10" class="text-faint">Loading…</Mono>
      </div>
      <div v-else-if="cats.length === 0" class="px-2.5 py-3">
        <Mono size="10" class="text-muted">No categories yet. Add one below.</Mono>
      </div>
      <CategoryTreeNode
        v-for="c in cats"
        :key="c.id"
        :category="c"
        :selected-artist-id="selectedArtistId ?? null"
        :aggregates="aggregates"
        :expanded="!collapsed.has(c.id)"
        @toggle="toggleCat(c.id)"
        @select-artist="emit('select-artist', $event)"
      />

      <!-- uncategorized (virtual node): artists in no category -->
      <div v-if="uncategorized.length" class="border-b border-line">
        <div class="flex items-center px-2.5 py-[7px] bg-surface-2 gap-2">
          <button
            class="text-faint hover:text-muted shrink-0"
            @click="uncatExpanded = !uncatExpanded"
          >
            <IconChevronDown v-if="uncatExpanded" class="h-3.5 w-3.5" />
            <IconChevronRight v-else class="h-3.5 w-3.5" />
          </button>
          <Mono
            size="10"
            class="font-bold tracking-wide uppercase flex-1 truncate text-faint cursor-pointer"
            @click="uncatExpanded = !uncatExpanded"
          >
            Uncategorized
          </Mono>
          <Mono size="9" class="text-ghost shrink-0 tabular-nums">
            {{ uncategorized.length }} {{ uncategorized.length === 1 ? "artist" : "artists" }}
          </Mono>
        </div>
        <template v-if="uncatExpanded">
          <ArtistTreeRow
            v-for="a in uncategorized"
            :key="a.id"
            :artist="a"
            :selected="selectedArtistId === a.id"
            :flags="aggFor(a.id).flags"
            :upcoming-listed="aggFor(a.id).upcomingListed"
            :fetching="fetchingId === a.id"
            hide-remove
            @select="emit('select-artist', a.id)"
            @fetch="fetchUncatArtist(a.id)"
          />
        </template>
      </div>
    </div>

    <!-- add category footer -->
    <div class="shrink-0 p-2 border-t border-line">
      <input
        v-if="adding"
        ref="input"
        v-model="newName"
        placeholder="category name…"
        class="w-full bg-surface-2 border border-accent-bright rounded-sm px-2 py-1 text-xs font-mono text-body focus:outline-none"
        @keydown.enter="commitAdd"
        @keydown.esc="adding = false"
        @blur="commitAdd"
      />
      <button
        v-else
        class="w-full bg-surface-2 border border-line-2 rounded-sm px-2 py-1 text-meta font-mono text-faint hover:text-muted text-center"
        @click="startAdd"
      >
        + Add Category
      </button>
    </div>
  </aside>
</template>
