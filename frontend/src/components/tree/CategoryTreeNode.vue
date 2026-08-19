<script setup lang="ts">
import { computed, nextTick, ref, watch } from "vue";
import type { Category } from "@/types";
import { useCategoriesStore } from "@/stores/categories";
import { useArtistsStore } from "@/stores/artists";
import { useUiStore } from "@/stores/ui";
import { useToast } from "@/composables";
import { categoryColor } from "@/utils";
import { Mono, IconButton } from "@/components/ui";
import ArtistTreeRow from "./ArtistTreeRow.vue";
import IconChevronDown from "~icons/mdi/chevron-down";
import IconChevronRight from "~icons/mdi/chevron-right";
import IconPencil from "~icons/mdi/pencil";
import IconClose from "~icons/mdi/close";

type ArtistAgg = { flags: { code: string; n: number }[]; claimed: number; upcomingListed: number };

const props = defineProps<{
  category: Category;
  selectedArtistId: string | null;
  aggregates: Map<string, ArtistAgg>;
  // Expansion is controlled by the parent tree so it can collapse/expand every node at once.
  expanded: boolean;
}>();

const emit = defineEmits<{ "select-artist": [id: string | null]; toggle: [] }>();

const categories = useCategoriesStore();
const artists = useArtistsStore();
const ui = useUiStore();
const toast = useToast();

const loadingArtists = ref(false);
const color = computed(() => categoryColor(props.category.name));

// Manual artists are history-only - hidden from the discovery tree (they live in the Diary).
const artistList = computed(() =>
  (categories.categoryArtists[props.category.id] ?? []).filter((a) => a.fetch_mode === "auto"),
);

function toggle() {
  emit("toggle");
}

// Lazy-load members the first time this node is shown expanded (covers the
// expanded-by-default case as well as a later manual expand).
watch(
  () => props.expanded,
  async (val) => {
    if (val && !categories.categoryArtists[props.category.id]) {
      loadingArtists.value = true;
      await categories.fetchCategoryArtists(props.category.id);
      loadingArtists.value = false;
    }
  },
  { immediate: true },
);

// One unit across every tree node: artist count, the size of the group - same as the
// Uncategorized node.
const countLabel = computed(() => {
  const n = props.category.artist_count;
  return `${n} ${n === 1 ? "artist" : "artists"}`;
});

// --- rename ---
const editing = ref(false);
const editName = ref("");
const nameInput = ref<HTMLInputElement | null>(null);
async function startEdit() {
  editing.value = true;
  editName.value = props.category.name;
  await nextTick();
  nameInput.value?.focus();
}
async function commitEdit() {
  const n = editName.value.trim();
  if (n && n !== props.category.name) {
    await categories.updateCategory(props.category.id, { name: n });
  }
  editing.value = false;
}

async function remove() {
  if (!confirm(`Delete category "${props.category.name}"?`)) return;
  await categories.deleteCategory(props.category.id);
  toast.success("Category deleted");
}

// --- add artist (autocomplete: reuse existing artists, avoid typo duplicates) ---
const adding = ref(false);
const newArtist = ref("");
const addInput = ref<HTMLInputElement | null>(null);

const memberIds = computed(() => new Set(artistList.value.map((a) => a.id)));

const suggestions = computed(() => {
  const q = newArtist.value.trim().toLowerCase();
  if (!q) return [];
  return artists.artists
    .filter((a) => a.name.toLowerCase().includes(q) && !memberIds.value.has(a.id))
    .slice(0, 6);
});

async function startAdd() {
  adding.value = true;
  newArtist.value = "";
  await nextTick();
  addInput.value?.focus();
}

async function addExisting(id: string, name: string) {
  await categories.addArtistToCategory(props.category.id, id);
  toast.success(`Added ${name}`);
  newArtist.value = "";
  adding.value = false;
}

async function commitAdd() {
  const name = newArtist.value.trim();
  if (!name) {
    adding.value = false;
    return;
  }
  // Reuse an exact existing artist (case-insensitive) before creating a new one.
  const existing = artists.artists.find((a) => a.name.toLowerCase() === name.toLowerCase());
  if (existing) {
    await addExisting(existing.id, existing.name);
    return;
  }
  // The discovery tree only manages auto (radar) artists; manual acts are added on the
  // Artists page / Diary side.
  const artist = await artists.createArtist({ name, fetch_mode: "auto" });
  if (artist) {
    await categories.addArtistToCategory(props.category.id, artist.id);
    toast.success(`Added ${name}`);
  }
  newArtist.value = "";
  adding.value = false;
}

// --- per-artist fetch ---
const fetchingId = ref<string | null>(null);
async function fetchArtist(id: string) {
  fetchingId.value = id;
  await artists.fetchAndSync(id);
  fetchingId.value = null;
}

async function removeArtist(id: string, name: string) {
  if (!confirm(`Remove ${name} from ${props.category.name}?`)) return;
  await categories.removeArtistFromCategory(props.category.id, id);
}

function aggFor(id: string): ArtistAgg {
  return props.aggregates.get(id) ?? { flags: [], claimed: 0, upcomingListed: 0 };
}
</script>

<template>
  <div class="border-b border-line">
    <!-- category row -->
    <div class="group flex items-center px-2.5 py-[7px] bg-surface-2 gap-2">
      <button class="text-faint hover:text-muted shrink-0" @click="toggle">
        <IconChevronDown v-if="expanded" class="h-3.5 w-3.5" />
        <IconChevronRight v-else class="h-3.5 w-3.5" />
      </button>

      <template v-if="editing">
        <input
          ref="nameInput"
          v-model="editName"
          class="flex-1 bg-surface border border-accent-bright rounded-sm px-1 py-0.5 text-meta font-mono uppercase tracking-wide text-heading focus:outline-none"
          @keydown.enter="commitEdit"
          @keydown.esc="editing = false"
          @blur="commitEdit"
        />
      </template>
      <template v-else>
        <Mono
          size="10"
          class="font-bold tracking-wide uppercase flex-1 truncate cursor-pointer"
          :style="{ color }"
          @click="toggle"
        >
          {{ category.name }}
        </Mono>
        <!-- count hugs the right edge; in edit mode the rename/delete actions take its
             place. Fixed height (= the icon-button height) keeps the row from jumping. -->
        <div class="shrink-0 h-5 flex items-center justify-end gap-1">
          <div v-if="ui.editMode" class="flex items-center gap-1">
            <IconButton size="sm" title="Rename category" @click="startEdit">
              <IconPencil class="h-3.5 w-3.5" />
            </IconButton>
            <IconButton tone="danger" size="sm" title="Delete category" @click="remove">
              <IconClose class="h-3.5 w-3.5" />
            </IconButton>
          </div>
          <Mono v-else size="9" class="text-ghost tabular-nums">{{ countLabel }}</Mono>
        </div>
      </template>
    </div>

    <!-- expanded: artists -->
    <template v-if="expanded">
      <div v-if="loadingArtists" class="pl-[22px] py-1.5">
        <Mono size="10" class="text-faint">Loading…</Mono>
      </div>
      <ArtistTreeRow
        v-for="a in artistList"
        :key="a.id"
        :artist="a"
        :selected="selectedArtistId === a.id"
        :flags="aggFor(a.id).flags"
        :upcoming-listed="aggFor(a.id).upcomingListed"
        :fetching="fetchingId === a.id"
        @select="emit('select-artist', a.id)"
        @fetch="fetchArtist(a.id)"
        @remove="removeArtist(a.id, a.name)"
      />

      <!-- add artist row (autocomplete) -->
      <div class="pl-[22px] py-1.5 pr-2.5 border-b border-line/50">
        <template v-if="adding">
          <input
            ref="addInput"
            v-model="newArtist"
            placeholder="type to search or add…"
            class="w-full bg-surface border border-accent-bright rounded-sm px-1.5 py-0.5 text-xs text-body font-mono focus:outline-none"
            @keydown.enter="commitAdd"
            @keydown.esc="adding = false"
            @blur="adding = false"
          />
          <!-- suggestions: reuse existing artists instead of creating duplicates -->
          <div
            v-if="suggestions.length"
            class="mt-1 bg-surface-2 border border-line-2 rounded-sm overflow-hidden"
          >
            <button
              v-for="s in suggestions"
              :key="s.id"
              class="w-full text-left px-1.5 py-1 text-xs font-mono text-muted hover:bg-surface-3 hover:text-accent-text"
              @mousedown.prevent="addExisting(s.id, s.name)"
            >
              {{ s.name }}
            </button>
          </div>
          <Mono v-else-if="newArtist.trim()" size="10" class="text-faint block mt-1">
            ↵ to add “{{ newArtist.trim() }}” as new
          </Mono>
        </template>
        <button
          v-else
          class="font-mono text-meta text-faint hover:text-accent-text"
          @click="startAdd"
        >
          + add artist…
        </button>
      </div>
    </template>
  </div>
</template>
