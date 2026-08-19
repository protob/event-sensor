<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";
import type { Event, EventKind, EventStatus, ManualEventPayload } from "@/types";
import { useEventsStore } from "@/stores/events";
import { useArtistsStore } from "@/stores/artists";
import { useToast } from "@/composables";
import { Btn, Mono, DateField, Modal, formFieldClass } from "@/components/ui";
import LineupEditor, { type LineupArtist } from "./LineupEditor.vue";

const props = withDefaults(
  defineProps<{
    open: boolean;
    presetArtistId?: string | null;
    presetName?: string;
    editEvent?: Event | null;
  }>(),
  { presetArtistId: null, presetName: "", editEvent: null },
);
// `created` carries the claim status on a fresh add (so the Library can jump to that tab);
// it's undefined on edit, where the existing status is left alone.
const emit = defineEmits<{ close: []; created: [status?: EventStatus] }>();

const isEditing = computed(() => !!props.editEvent);

const store = useEventsStore();
const artists = useArtistsStore();
const toast = useToast();

const KINDS: EventKind[] = ["concert", "festival", "tribute", "club", "other"];

const field = formFieldClass;

// Lineup rows are owned by LineupEditor (the `LineupArtist` type is its contract).

const form = reactive({
  name: "",
  kind: "concert" as EventKind,
  date: "",
  hasTime: false,
  time: "",
  endDate: "",
  venueName: "",
  city: "",
  country: "",
  countryCode: "",
  artists: [] as LineupArtist[],
  status: "" as EventStatus | "",
});
const submitting = ref(false);

const isPastDate = computed(
  () => form.date !== "" && form.date < new Date().toISOString().slice(0, 10),
);
// A festival IS the event (the container); its artists are the lineup, each with their own day/venue.
const isFestival = computed(() => form.kind === "festival");

function artistFromPreset(): LineupArtist[] {
  if (!props.presetArtistId) return [];
  const a = artists.artists.find((x) => x.id === props.presetArtistId);
  return [{ name: a?.name ?? "", entity: true, date: "", venue: "" }];
}

function reset() {
  const e = props.editEvent;
  if (e) {
    const hasT = e.start_date.includes("T");
    Object.assign(form, {
      name: e.name,
      kind: e.kind,
      date: e.start_date.slice(0, 10),
      hasTime: hasT,
      time: hasT ? e.start_date.slice(11, 16) : "",
      endDate: e.end_date ? e.end_date.slice(0, 10) : "",
      venueName: e.venue?.name ?? "",
      city: e.venue?.city ?? "",
      country: e.venue?.country ?? "",
      countryCode: e.venue?.country_code ?? "",
      artists: (e.performances ?? []).map((p) => ({
        name: p.artist_name,
        entity: !!p.artist_id,
        date: p.date ?? "",
        venue: p.venue_name ?? "",
      })),
      status: "",
    });
    return;
  }
  Object.assign(form, {
    name: props.presetName ?? "",
    kind: "concert",
    date: "",
    hasTime: false,
    time: "",
    endDate: "",
    venueName: "",
    city: "",
    country: "",
    countryCode: "",
    artists: artistFromPreset(),
    status: "",
  });
}

const entityNames = computed(() => artists.artists.map((a) => a.name));

watch(
  () => props.open,
  (o) => {
    if (o) {
      reset();
      if (artists.artists.length === 0) artists.fetchArtists();
    }
  },
);

onMounted(() => {
  if (artists.artists.length === 0) artists.fetchArtists();
});

const effectiveStatus = computed<EventStatus>(() =>
  form.status !== "" ? form.status : isPastDate.value ? "attended" : "interested",
);

async function submit() {
  if (!form.name.trim() || !form.date) {
    toast.error("Name and date are required");
    return;
  }
  submitting.value = true;
  let startDate = form.date;
  if (form.hasTime && form.time) startDate = `${form.date}T${form.time}:00`;

  // Map lineup rows -> performances. Always auto-link by name against the existing catalog —
  // if a name matches an entity artist the artist_id is set regardless of the entity flag.
  // The entity flag only matters when there is no match: entity=true creates a new catalog
  // artist, entity=false keeps the act as a name-only string.
  const performances: NonNullable<ManualEventPayload["performances"]> = [];
  for (const a of form.artists) {
    const name = a.name.trim();
    if (!name) continue;
    let artist_id: string | undefined;
    const existing = artists.artists.find((x) => x.name.toLowerCase() === name.toLowerCase());
    if (existing) {
      artist_id = existing.id;
    } else if (a.entity) {
      artist_id = (await artists.createArtist({ name }))?.id ?? undefined;
    }
    performances.push({
      artist_id,
      artist_name: name,
      is_headliner: false,
      date: a.date.trim() || undefined,
      venue_name: a.venue.trim() || undefined,
    });
  }

  const payload: ManualEventPayload = {
    name: form.name.trim(),
    kind: form.kind,
    start_date: startDate,
    has_time: form.hasTime && !!form.time,
    end_date: form.endDate || undefined,
    performances: performances.length ? performances : undefined,

    status: effectiveStatus.value,
  };
  // Event-level location: a single venue (concert / single-site festival) or just a city
  // (urban festival — the per-act venues carry the halls). Either gives the row its city + flag.
  if (form.venueName.trim() || form.city.trim()) {
    payload.venue = {
      name: form.venueName.trim(),
      city: form.city.trim() || undefined,
      country: form.country.trim() || undefined,
      country_code: form.countryCode.trim().toLowerCase() || undefined,
    };
  }

  const ev = props.editEvent
    ? await store.updateEvent(props.editEvent.id, payload)
    : await store.createManualEvent(payload);
  submitting.value = false;
  if (ev) {
    toast.success(`${props.editEvent ? "Updated" : "Added"} “${ev.name}”`);
    emit("created", props.editEvent ? undefined : effectiveStatus.value);
    emit("close");
  }
}
</script>

<template>
  <Modal
    :open="open"
    :title="isEditing ? 'Edit event' : 'Add event'"
    width="max-w-2xl"
    @close="emit('close')"
  >
    <div class="p-gutter flex flex-col gap-3">
      <label class="flex flex-col gap-1">
        <Mono size="9" class="text-muted uppercase">{{
          isFestival ? "Festival name *" : "Name *"
        }}</Mono>
        <input
          v-model="form.name"
          type="text"
          :placeholder="isFestival ? 'e.g. Open’er Festival 2007' : 'Event name'"
          :class="field"
          data-testid="manual-name"
        />
        <Mono v-if="isFestival" size="9" class="text-faint">
          The festival is the event — add the artists you saw (with day / venue) below.
        </Mono>
      </label>

      <div class="flex gap-3">
        <label class="flex flex-col gap-1 flex-1">
          <Mono size="9" class="text-muted uppercase">Kind</Mono>
          <select v-model="form.kind" :class="field">
            <option
              v-for="k in KINDS"
              :key="k"
              :value="k"
              :data-testid="k === 'festival' ? 'manual-kind-festival' : undefined"
            >
              {{ k }}
            </option>
          </select>
        </label>
        <label class="flex flex-col gap-1 flex-1">
          <Mono size="9" class="text-muted uppercase">{{
            isFestival ? "Start date * (YYYY-MM-DD)" : "Date * (YYYY-MM-DD)"
          }}</Mono>
          <DateField v-model="form.date" :class="field" data-testid="manual-start-date" />
        </label>
      </div>

      <div class="flex gap-3 items-end">
        <label class="flex items-center gap-1.5 font-mono text-label text-faint cursor-pointer">
          <input v-model="form.hasTime" type="checkbox" class="accent-accent-bright" /> has time
        </label>
        <label v-if="form.hasTime" class="flex flex-col gap-1">
          <Mono size="9" class="text-muted uppercase">Time (24h HH:MM)</Mono>
          <input
            v-model="form.time"
            type="text"
            inputmode="numeric"
            placeholder="HH:MM"
            maxlength="5"
            :class="field"
          />
        </label>
        <label class="flex flex-col gap-1 flex-1">
          <Mono size="9" class="text-muted uppercase">{{
            isFestival ? "End date (last day, YYYY-MM-DD)" : "End date (optional, YYYY-MM-DD)"
          }}</Mono>
          <DateField v-model="form.endDate" :class="field" data-testid="manual-end-date" />
        </label>
      </div>

      <div class="grid grid-cols-2 gap-3">
        <label class="flex flex-col gap-1">
          <Mono size="9" class="text-muted uppercase">{{
            isFestival ? "Main venue (optional)" : "Venue"
          }}</Mono>
          <input
            v-model="form.venueName"
            type="text"
            :placeholder="isFestival ? 'Venue name (optional)' : 'Venue name'"
            :class="field"
            data-testid="manual-venue-name"
          />
        </label>
        <label class="flex flex-col gap-1">
          <Mono size="9" class="text-muted uppercase">City</Mono>
          <input v-model="form.city" type="text" :class="field" data-testid="manual-venue-city" />
        </label>
        <label class="flex flex-col gap-1">
          <Mono size="9" class="text-muted uppercase">Country</Mono>
          <input
            v-model="form.country"
            type="text"
            :class="field"
            data-testid="manual-venue-country"
          />
        </label>
        <label class="flex flex-col gap-1">
          <Mono size="9" class="text-muted uppercase">Country code</Mono>
          <input
            v-model="form.countryCode"
            type="text"
            maxlength="2"
            placeholder="de"
            :class="field"
          />
        </label>
      </div>

      <!-- Lineup editor: CSV paste, multi-select fill-down, per-artist day/venue (festival),
             entity/name-only toggle. Festivals get the two-panel layout. -->
      <LineupEditor
        v-model="form.artists"
        :is-festival="isFestival"
        :entity-names="entityNames"
        :event-date-range="{ start: form.date, end: form.endDate }"
      />

      <!-- Initial claim is only relevant on create; editing leaves status to the diary. -->
      <label v-if="!isEditing" class="flex flex-col gap-1">
        <Mono size="9" class="text-muted uppercase">
          Initial status (default: {{ isPastDate ? "attended" : "interested" }})
        </Mono>
        <select v-model="form.status" :class="field">
          <option value="">Auto ({{ isPastDate ? "attended" : "interested" }})</option>
          <option value="interested" data-testid="manual-status-interested">interested</option>
          <option value="going" data-testid="manual-status-going">going</option>
          <option value="attended" data-testid="manual-status-attended">attended</option>
          <option value="missed" data-testid="manual-status-missed">missed</option>
        </select>
      </label>
    </div>

    <div class="flex justify-end gap-2 px-4 h-12 items-center border-t border-line">
      <Btn tone="neutral" size="sm" @click="emit('close')">Cancel</Btn>
      <Btn tone="accent" size="sm" :loading="submitting" data-testid="manual-save" @click="submit">
        {{ isEditing ? "Save changes" : "Add event" }}
      </Btn>
    </div>
  </Modal>
</template>
