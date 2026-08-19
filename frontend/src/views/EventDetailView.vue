<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter, RouterLink } from "vue-router";
import { storeToRefs } from "pinia";
import type { EventStatus } from "@/types";
import { useEventsStore } from "@/stores/events";
import { useArtistsStore } from "@/stores/artists";
import { useAuthStore } from "@/stores/auth";
import { formatEventDate, countryName, categoryColor } from "@/utils";
import { Tag, Mono, Pill, Flag } from "@/components/ui";
import LineupGrid from "@/components/events/LineupGrid.vue";
import EventLifecycleTag from "@/components/events/EventLifecycleTag.vue";
import ManualEventForm from "@/components/library/ManualEventForm.vue";
import StatusSelect from "@/components/library/StatusSelect.vue";
import MissedReasonSelect from "@/components/library/MissedReasonSelect.vue";
import NoteInput from "@/components/library/NoteInput.vue";
import { eventTag } from "@/components/events/eventTag";
import IconBack from "~icons/mdi/arrow-left";
import IconInterested from "~icons/mdi/bookmark";
import IconCheck from "~icons/mdi/check-circle";

const route = useRoute();
const router = useRouter();
const store = useEventsStore();
const artists = useArtistsStore();
const auth = useAuthStore();
const { currentEvent: ev, loading, libraryMap } = storeToRefs(store);

const diaryEntry = computed(() => (ev.value ? (libraryMap.value[ev.value.id] ?? null) : null));

async function setDiaryStatus(status: EventStatus) {
  if (!ev.value) return;
  const le = diaryEntry.value;
  await store.setStatus(ev.value.id, status, {
    note: le?.note ?? undefined,
    missed_reason: le?.missed_reason ?? undefined,
  });
}
async function setMissedReason(reason: string) {
  if (!ev.value) return;
  const le = diaryEntry.value;
  await store.setStatus(ev.value.id, "missed", {
    missed_reason: reason,
    note: le?.note ?? undefined,
  });
}
async function saveDiaryNote(note: string) {
  if (!ev.value?.status) return;
  await store.setStatus(ev.value.id, ev.value.status, { note });
}

const catColor = computed(() => categoryColor(ev.value?.kind));
const kindTag = computed(() => eventTag(ev.value ?? undefined));

const heroStyle = computed(() =>
  ev.value?.image_url
    ? {
        backgroundImage: `url(${ev.value.image_url})`,
        backgroundSize: "cover",
        backgroundPosition: "center",
      }
    : {},
);

const isMultiDay = computed(() => {
  const e = ev.value;
  if (!e?.end_date) return false;
  return e.end_date.slice(0, 10) !== e.start_date.slice(0, 10);
});

const dateText = computed(() => {
  const e = ev.value;
  if (!e) return "";
  const base = formatEventDate(e.start_date, isMultiDay.value ? e.end_date : undefined);
  return isMultiDay.value ? `${base} · multi-day` : base;
});

// Where the lineup names come from: manual = the user's own data; TM attractions = may
// be incomplete; no attractions = fall back to the entity artists billed on it.
const lineup = computed(() => {
  const e = ev.value;
  if (!e) return { names: [] as string[], source: "none" as const };
  if (e.source === "manual") {
    return { names: (e.lineup ?? []).map((l) => l.artist_name), source: "manual" as const };
  }
  if (e.lineup?.length) {
    return { names: e.lineup.map((l) => l.artist_name), source: "ticketmaster" as const };
  }
  return { names: (e.artists ?? []).map((a) => a.artist_name), source: "entity" as const };
});

const entityByName = computed(() => {
  const m = new Map<string, (typeof artists.artists)[number]>();
  for (const a of artists.artists) m.set(a.name.toLowerCase().trim(), a);
  return m;
});

const showEdit = ref(false);

async function load() {
  const id = route.params.id as string;
  await store.fetchEvent(id);
  if (auth.isAuthenticated) await store.loadLibrary();
  if (artists.artists.length === 0) await artists.fetchArtists();
}

async function remove() {
  if (!ev.value) return;
  const ok = await store.deleteEventConfirmed(ev.value);
  if (ok) router.push({ name: "events" });
}

onMounted(load);
</script>

<template>
  <div class="h-full flex flex-col min-w-0">
    <!-- header strip -->
    <div class="flex items-center h-[46px] px-4 gap-2.5 border-b border-line bg-surface shrink-0">
      <RouterLink
        to="/events"
        class="flex items-center gap-1.5 font-mono text-label text-faint hover:text-muted"
      >
        <IconBack class="h-3.5 w-3.5" /> Events
      </RouterLink>
      <Mono size="11" class="text-faint">/</Mono>
      <Mono size="11" class="text-muted truncate max-w-[40vw]">{{ ev?.name }}</Mono>

      <div v-if="ev" class="ml-auto flex items-center gap-2">
        <Pill tone="accent" :active="!!ev.status" @click="store.toggleInterested(ev.id)">
          <IconInterested class="h-3.5 w-3.5" /> Interested
        </Pill>
        <Pill
          tone="go"
          :active="ev.is_past ? ev.status === 'attended' : ev.status === 'going'"
          @click="
            ev.is_past
              ? store.setStatus(ev.id, ev.status === 'attended' ? 'interested' : 'attended')
              : store.toggleGoing(ev.id)
          "
        >
          <IconCheck class="h-3.5 w-3.5" /> {{ ev.is_past ? "Attended" : "Going" }}
        </Pill>
        <Pill v-if="ev.source === 'manual'" tone="neutral" @click="showEdit = true">✎ Edit</Pill>
        <a v-if="ev.url" :href="ev.url" target="_blank" rel="noopener">
          <Pill tone="neutral">↗ Ticketmaster</Pill>
        </a>
        <Pill tone="danger" @click="remove">✕ Remove</Pill>
      </div>
    </div>

    <!-- body -->
    <div v-if="loading && !ev" class="flex-1 flex items-center justify-center">
      <Mono size="xs" class="text-faint">Loading…</Mono>
    </div>
    <div v-else-if="!ev" class="flex-1 flex flex-col items-center justify-center gap-2">
      <Mono size="xs" class="text-muted">Event not found.</Mono>
      <RouterLink to="/events" class="font-mono text-label text-accent-text"
        >← Back to Events</RouterLink
      >
    </div>

    <div v-else class="flex-1 min-h-0 flex overflow-hidden">
      <!-- left: hero + meta -->
      <div class="w-[480px] shrink-0 border-r border-line overflow-y-auto">
        <!-- With an image: poster banner, title on a fade. Without: no empty banner - the title
             block sits in normal flow. -->
        <div
          v-if="ev.image_url"
          class="h-[220px] relative border-b border-line-2 bg-surface-2"
          :style="heroStyle"
        >
          <div
            class="absolute inset-0"
            style="background: linear-gradient(to top, var(--color-app), transparent 60%)"
          />
          <div class="absolute bottom-3.5 left-4 right-4">
            <div class="flex gap-1.5 mb-1.5 items-center">
              <!-- kindTag labels festivals/tributes; without one the plain kind shows. -->
              <Tag v-if="kindTag" :color="kindTag.color">{{ kindTag.label }}</Tag>
              <Tag v-else :color="catColor">{{ ev.kind }}</Tag>
              <EventLifecycleTag :event="ev" show-past />
            </div>
            <h1 class="text-2xl font-bold text-heading leading-tight">{{ ev.name }}</h1>
            <div v-if="ev.venue" class="mt-1 flex items-center gap-1.5">
              <Flag v-if="ev.venue.country_code" :code="ev.venue.country_code" :w="18" />
              <Mono size="11" class="text-muted">{{ ev.venue.city }}, {{ ev.venue.country }}</Mono>
            </div>
          </div>
        </div>
        <div v-else class="px-4 pt-4 pb-3.5 border-b border-line-2">
          <div class="flex gap-1.5 mb-1.5 items-center">
            <Tag v-if="kindTag" :color="kindTag.color">{{ kindTag.label }}</Tag>
            <Tag v-else :color="catColor">{{ ev.kind }}</Tag>
            <EventLifecycleTag :event="ev" show-past />
          </div>
          <h1 class="text-2xl font-bold text-heading leading-tight">{{ ev.name }}</h1>
          <div v-if="ev.venue" class="mt-1 flex items-center gap-1.5">
            <Flag v-if="ev.venue.country_code" :code="ev.venue.country_code" :w="18" />
            <Mono size="11" class="text-muted">{{ ev.venue.city }}, {{ ev.venue.country }}</Mono>
          </div>
        </div>

        <div class="px-4">
          <div class="flex gap-3 py-[9px] border-b border-line items-start">
            <Mono size="9" class="text-ghost w-[60px] shrink-0 pt-0.5 uppercase">Kind</Mono>
            <span class="text-xs text-muted">{{ ev.kind }}</span>
          </div>
          <div v-if="ev.venue?.name" class="flex gap-3 py-[9px] border-b border-line items-start">
            <Mono size="9" class="text-ghost w-[60px] shrink-0 pt-0.5 uppercase">Venue</Mono>
            <span class="text-xs text-muted">{{ ev.venue.name }}</span>
          </div>
          <div
            v-if="ev.venue?.country_code"
            class="flex gap-3 py-[9px] border-b border-line items-start"
          >
            <Mono size="9" class="text-ghost w-[60px] shrink-0 pt-0.5 uppercase">Country</Mono>
            <span class="text-xs text-muted flex items-center gap-1.5">
              {{ countryName(ev.venue.country_code, ev.venue.country) }}
              <Flag :code="ev.venue.country_code" :w="16" />
            </span>
          </div>
          <div v-if="ev.venue?.city" class="flex gap-3 py-[9px] border-b border-line items-start">
            <Mono size="9" class="text-ghost w-[60px] shrink-0 pt-0.5 uppercase">City</Mono>
            <span class="text-xs text-muted">{{ ev.venue.city }}</span>
          </div>
          <div class="flex gap-3 py-[9px] border-b border-line items-start">
            <Mono size="9" class="text-ghost w-[60px] shrink-0 pt-0.5 uppercase">Date</Mono>
            <span class="text-xs text-muted">{{ dateText }}</span>
          </div>
          <div v-if="diaryEntry" class="flex gap-3 py-[9px] border-b border-line items-start">
            <Mono size="9" class="text-ghost w-[60px] shrink-0 pt-0.5 uppercase">Diary</Mono>
            <div class="flex-1 flex flex-col gap-1.5 min-w-0">
              <div class="flex items-center gap-2 flex-wrap">
                <StatusSelect :status="diaryEntry.status" @select="setDiaryStatus" />
                <MissedReasonSelect
                  v-if="diaryEntry.status === 'missed'"
                  :reason="diaryEntry.missed_reason"
                  @select="setMissedReason"
                />
                <Mono v-if="diaryEntry.attended_date" size="10" class="text-faint">{{
                  diaryEntry.attended_date
                }}</Mono>
              </div>
              <NoteInput :note="diaryEntry.note" @save="saveDiaryNote" />
            </div>
          </div>
          <div
            v-if="ev.about || ev.description"
            class="flex gap-3 py-[9px] border-b border-line items-start"
          >
            <Mono size="9" class="text-ghost w-[60px] shrink-0 pt-0.5 uppercase">About</Mono>
            <span class="text-xs text-muted leading-relaxed">{{ ev.about || ev.description }}</span>
          </div>
        </div>
      </div>

      <!-- right: lineup -->
      <div class="flex-1 overflow-y-auto p-5 min-w-0">
        <LineupGrid :names="lineup.names" :source="lineup.source" :entity-by-name="entityByName" />
      </div>
    </div>

    <ManualEventForm
      v-if="ev"
      :open="showEdit"
      :edit-event="ev"
      @close="showEdit = false"
      @created="load"
    />
  </div>
</template>
