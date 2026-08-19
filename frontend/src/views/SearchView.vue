<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useRoute } from "vue-router";
import { storeToRefs } from "pinia";
import { useEventsStore } from "@/stores/events";
import { useAuthStore } from "@/stores/auth";
import EventTable from "@/components/events/EventTable.vue";
import ManualEventForm from "@/components/library/ManualEventForm.vue";
import { Btn } from "@/components/ui";
import IconSearch from "~icons/mdi/magnify";

const route = useRoute();
const store = useEventsStore();
const auth = useAuthStore();
const { events, loading } = storeToRefs(store);

const query = ref("");
const searched = ref(false);

// Add an event seeded from the current query.
const showAdd = ref(false);
async function onCreated() {
  if (auth.isAuthenticated) await store.loadLibrary();
  if (searched.value) await store.searchEvents(query.value.trim());
}

async function runSearch() {
  const q = query.value.trim();
  if (!q) return;
  searched.value = true;
  if (auth.isAuthenticated) await store.loadLibrary();
  await store.searchEvents(q);
}

onMounted(() => {
  const q = route.query.q as string;
  if (q) {
    query.value = q;
    runSearch();
  }
});
</script>

<template>
  <div class="h-full flex flex-col bg-app min-w-0">
    <div
      class="min-h-12 px-gutter py-1.5 flex flex-wrap items-center gap-2 border-b border-line bg-surface-2 shrink-0"
    >
      <div
        class="flex flex-1 items-center gap-2 bg-surface border border-line-2 rounded-sm px-2.5 min-w-32 max-w-md focus-within:border-accent-bright"
      >
        <IconSearch class="h-4 w-4 text-faint shrink-0" />
        <input
          v-model="query"
          type="text"
          placeholder="Search events by name, artist, or venue…"
          class="bg-transparent text-sm text-body font-mono py-2 w-full placeholder:text-faint focus:outline-none"
          @keydown.enter="runSearch"
        />
      </div>
      <!-- The label carries the query, so its width is unbounded: it wraps to its own line and
           truncates rather than pushing the field off the row. -->
      <Btn
        v-if="auth.isAuthenticated && query.trim()"
        tone="accent"
        size="sm"
        class="min-w-0 max-w-full"
        title="Create a manual event named after this query"
        @click="showAdd = true"
      >
        <span class="truncate">+ Create “{{ query.trim() }}”</span>
      </Btn>
    </div>

    <EventTable
      :events="searched ? events : []"
      :loading="loading"
      :empty-message="
        searched ? 'No results. Try a different term.' : 'Search events by name, artist, or venue.'
      "
    />

    <ManualEventForm
      :open="showAdd"
      :preset-name="query.trim()"
      @close="showAdd = false"
      @created="onCreated"
    />
  </div>
</template>
