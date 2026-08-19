<script setup lang="ts">
import { computed, ref } from "vue";
import { DateField, Mono, formFieldClass } from "@/components/ui";
import { useSelection } from "@/composables/useSelection";

// One artist in the lineup. `entity` = true links to a catalog artist (looked up by name,
// created if missing). `entity` = false keeps it as a name-only string (no artist_id stored).
// date/venue are per-artist (festivals: which day, which venue).
export type LineupArtist = { name: string; entity: boolean; date: string; venue: string };

withDefaults(
  defineProps<{
    isFestival: boolean;
    entityNames: string[];
    eventDateRange?: { start: string; end: string };
  }>(),
  { eventDateRange: undefined },
);

const field = formFieldClass;

const model = defineModel<LineupArtist[]>({ required: true });
const artists = computed(() => model.value);
function setArtists(next: LineupArtist[]) {
  model.value = next;
}
function patch(i: number, p: Partial<LineupArtist>) {
  setArtists(artists.value.map((a, idx) => (idx === i ? { ...a, ...p } : a)));
}

function addArtist() {
  const newIdx = artists.value.length;
  setArtists([...artists.value, { name: "", entity: false, date: "", venue: "" }]);
  focused.value = newIdx;
}
function removeArtist(i: number) {
  setArtists(artists.value.filter((_, idx) => idx !== i));
  selection.clear();
  if (focused.value === i) focused.value = null;
}

// --- CSV / bulk paste ---
const showPaste = ref(false);
const pasteText = ref("");

// Accept "one name per line" OR comma/tab-separated "name, date, venue". UTF-8 preserved exactly
// (no normalization). Blank lines skipped.
function parsePaste(text: string): LineupArtist[] {
  const out: LineupArtist[] = [];
  for (const raw of text.split(/\r?\n/)) {
    const line = raw.trim();
    if (!line) continue;
    const cols = line.includes("\t") ? line.split("\t") : line.split(",");
    const name = (cols[0] ?? "").trim();
    if (!name) continue;
    out.push({
      name,
      entity: false,
      date: (cols[1] ?? "").trim(),
      venue: (cols[2] ?? "").trim(),
    });
  }
  return out;
}
function applyPaste() {
  const parsed = parsePaste(pasteText.value);
  if (parsed.length) setArtists([...artists.value, ...parsed]);
  pasteText.value = "";
  showPaste.value = false;
}

const pasteCount = computed(() => parsePaste(pasteText.value).length);

// --- Multi-select + fill-down (DAW-style) ---
// Selection over artist indices (as strings). Cleared on any structural mutation to stay honest.
const selection = useSelection(() => artists.value.map((_, i) => String(i)));
const focused = ref<number | null>(null);

const fillDay = ref("");
const fillVenue = ref("");

function selectedIdx(): number[] {
  return selection.ids().map(Number);
}
function fillDownDay() {
  if (!fillDay.value) return;
  const sel = new Set(selectedIdx());
  setArtists(artists.value.map((a, i) => (sel.has(i) ? { ...a, date: fillDay.value } : a)));
}
function fillDownVenue() {
  const sel = new Set(selectedIdx());
  setArtists(artists.value.map((a, i) => (sel.has(i) ? { ...a, venue: fillVenue.value } : a)));
}
function setEntitySelected(entity: boolean) {
  const sel = new Set(selectedIdx());
  setArtists(artists.value.map((a, i) => (sel.has(i) ? { ...a, entity } : a)));
}
function removeSelected() {
  const sel = new Set(selectedIdx());
  setArtists(artists.value.filter((_, i) => !sel.has(i)));
  selection.clear();
  focused.value = null;
}

const focusedArtist = computed(() => (focused.value != null ? artists.value[focused.value] : null));
</script>

<template>
  <div class="flex flex-col gap-1.5">
    <div class="flex items-center gap-2">
      <Mono size="9" class="text-muted uppercase">
        {{ isFestival ? "Lineup — artists you saw" : "Artists / lineup (optional)" }}
      </Mono>
      <span class="ml-auto flex items-center gap-3">
        <button
          class="font-mono text-label text-faint hover:text-body"
          :title="'Paste a list of names (one per line) or CSV: name, date, venue'"
          data-testid="lineup-paste"
          @click="showPaste = !showPaste"
        >
          ⎘ paste list
        </button>
        <button
          class="font-mono text-label text-accent-text hover:brightness-125"
          @click="addArtist"
        >
          + add artist
        </button>
      </span>
    </div>

    <!-- CSV / bulk paste -->
    <div
      v-if="showPaste"
      class="flex flex-col gap-1.5 p-2 rounded-sm border border-line-2 bg-surface-2/60"
    >
      <Mono size="9" class="text-faint">
        One name per line, or comma/tab-separated <b>name, date (YYYY-MM-DD), venue</b>. Appends to
        the lineup. Diacritics &amp; non-Latin scripts preserved.
      </Mono>
      <textarea
        v-model="pasteText"
        rows="5"
        placeholder="Artist 1&#10;Artist 2, 2007-07-06, Tent Stage&#10;Artist 3"
        :class="[field, 'w-full resize-y']"
        data-testid="lineup-csv"
      />
      <div class="flex justify-end gap-2">
        <button class="font-mono text-label text-faint hover:text-body" @click="showPaste = false">
          cancel
        </button>
        <button
          class="font-mono text-label text-accent-text hover:brightness-125 disabled:opacity-40"
          :disabled="!pasteText.trim()"
          data-testid="lineup-csv-apply"
          @click="applyPaste"
        >
          append {{ pasteCount }} artists →
        </button>
      </div>
    </div>

    <!-- fill-down bar (appears with a selection) -->
    <div
      v-if="selection.count.value > 0"
      class="flex items-center gap-2 flex-wrap p-1.5 rounded-sm border border-accent-chip-border bg-accent-chip/40"
    >
      <Mono size="9" class="text-accent-text uppercase tabular-nums">
        {{ selection.count.value }} selected
      </Mono>
      <template v-if="isFestival">
        <span class="flex items-center gap-1">
          <DateField v-model="fillDay" :class="[field, 'w-32']" data-testid="lineup-fill-day" />
          <button
            class="font-mono text-meta text-accent-text hover:brightness-125"
            @click="fillDownDay"
          >
            set day
          </button>
        </span>
        <span class="flex items-center gap-1">
          <input
            v-model="fillVenue"
            type="text"
            placeholder="venue"
            :class="[field, 'w-28']"
            data-testid="lineup-fill-venue"
          />
          <button
            class="font-mono text-meta text-accent-text hover:brightness-125"
            @click="fillDownVenue"
          >
            set venue
          </button>
        </span>
      </template>
      <button
        class="font-mono text-meta text-muted hover:text-body"
        @click="setEntitySelected(true)"
      >
        → entity
      </button>
      <button
        class="font-mono text-meta text-muted hover:text-body"
        @click="setEntitySelected(false)"
      >
        → name-only
      </button>
      <button class="font-mono text-meta text-muted hover:text-danger" @click="removeSelected">
        remove
      </button>
      <button
        class="ml-auto font-mono text-meta text-faint hover:text-body"
        @click="selection.clear()"
      >
        clear
      </button>
    </div>

    <!-- FESTIVAL: two-panel (left dense list, right detail of the focused artist) -->
    <div v-if="isFestival && artists.length" class="flex gap-3 min-h-[160px]">
      <div class="flex-1 min-w-0 max-h-72 overflow-y-auto rounded-sm border border-line-2">
        <div
          v-for="(artist, i) in artists"
          :key="i"
          class="flex items-center gap-2 px-2 py-1 border-b border-line/60 cursor-pointer"
          :data-testid="'lineup-act-' + i"
          :class="[
            focused === i
              ? 'bg-surface-sel'
              : selection.has(String(i))
                ? 'bg-accent-chip/30'
                : 'hover:bg-surface-2',
          ]"
          @click="focused = i"
        >
          <input
            type="checkbox"
            class="accent-accent-bright shrink-0"
            :checked="selection.has(String(i))"
            @click.stop="selection.onRowClick(String(i), $event)"
          />
          <span class="text-xs truncate flex-1" :class="artist.name ? 'text-body' : 'text-faint'">
            {{ artist.name || "(unnamed artist)" }}
          </span>
          <Mono v-if="artist.entity" size="9" class="text-accent-text shrink-0">●</Mono>
          <Mono v-if="artist.date" size="9" class="text-faint shrink-0 tabular-nums">{{
            artist.date
          }}</Mono>
          <button
            class="text-faint hover:text-danger px-0.5 shrink-0"
            title="Remove artist"
            @click.stop="removeArtist(i)"
          >
            ✕
          </button>
        </div>
      </div>

      <div
        class="w-56 shrink-0 flex flex-col gap-2 p-2 rounded-sm border border-line-2 bg-surface-2/40"
      >
        <template v-if="focusedArtist">
          <label class="flex flex-col gap-1">
            <Mono size="9" class="text-muted uppercase">Artist name</Mono>
            <input
              :value="focusedArtist.name"
              type="text"
              list="es-entity-artists"
              placeholder="artist name"
              :class="field"
              @input="patch(focused!, { name: ($event.target as HTMLInputElement).value })"
            />
          </label>
          <label
            class="flex items-center gap-1.5 font-mono text-label text-faint cursor-pointer select-none"
          >
            <input
              type="checkbox"
              class="accent-accent-bright"
              :checked="focusedArtist.entity"
              @change="patch(focused!, { entity: ($event.target as HTMLInputElement).checked })"
            />
            entity artist (in catalog)
          </label>
          <label class="flex flex-col gap-1">
            <Mono size="9" class="text-muted uppercase">Day</Mono>
            <DateField
              :model-value="focusedArtist.date"
              :class="field"
              @update:model-value="patch(focused!, { date: $event })"
            />
          </label>
          <label class="flex flex-col gap-1">
            <Mono size="9" class="text-muted uppercase">Venue</Mono>
            <input
              :value="focusedArtist.venue"
              type="text"
              placeholder="this artist's venue"
              :class="field"
              @input="patch(focused!, { venue: ($event.target as HTMLInputElement).value })"
            />
          </label>
        </template>
        <Mono v-else size="11" class="text-faint">Select an artist to edit their day / venue.</Mono>
      </div>
    </div>

    <!-- NON-FESTIVAL: compact single-column rows -->
    <template v-else-if="!isFestival">
      <div
        v-for="(artist, i) in artists"
        :key="i"
        class="flex gap-2 items-center"
        :data-testid="'lineup-act-' + i"
      >
        <input
          type="checkbox"
          class="accent-accent-bright shrink-0"
          :checked="selection.has(String(i))"
          @click="selection.onRowClick(String(i), $event)"
        />
        <input
          :value="artist.name"
          type="text"
          list="es-entity-artists"
          placeholder="artist name (type to match an entity artist)"
          :class="[field, 'flex-1 min-w-0']"
          @input="patch(i, { name: ($event.target as HTMLInputElement).value })"
        />
        <label
          class="flex items-center gap-1 font-mono text-meta text-faint shrink-0 cursor-pointer select-none"
          title="Entity artist: linked to a catalog artist by name (created if new). Off = name-only string."
        >
          <input
            type="checkbox"
            class="accent-accent-bright"
            :checked="artist.entity"
            @change="patch(i, { entity: ($event.target as HTMLInputElement).checked })"
          />
          entity
        </label>
        <button
          class="text-faint hover:text-danger px-1 shrink-0"
          title="Remove artist"
          @click="removeArtist(i)"
        >
          ✕
        </button>
      </div>
    </template>

    <datalist id="es-entity-artists">
      <option v-for="n in entityNames" :key="n" :value="n" />
    </datalist>
  </div>
</template>
