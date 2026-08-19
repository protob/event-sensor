<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { useEventListener } from "@vueuse/core";
import IconRail from "./IconRail.vue";
import TopBar from "./TopBar.vue";
import StatusBar from "./StatusBar.vue";

const route = useRoute();

const BARE = new Set(["login", "reset-password"]);
const bare = computed(() => route.meta.bare === true || BARE.has(String(route.name)));

// Below md the rail is an overlay drawer. The state lives here rather than in IconRail
// because the scrim and the TopBar trigger both need it, and both are siblings of the rail.
const railOpen = ref(false);

// Any navigation closes the drawer — otherwise it stays over the page you just navigated to.
watch(
  () => route.fullPath,
  () => (railOpen.value = false),
);

useEventListener(window, "keydown", (e: KeyboardEvent) => {
  if (e.key === "Escape" && railOpen.value) railOpen.value = false;
});
</script>

<template>
  <!-- Bare layout: auth pages, full-screen centered, no chrome. -->
  <main v-if="bare" class="h-dvh w-full overflow-auto bg-app flex items-center justify-center p-4">
    <slot />
  </main>

  <!-- Full app chrome: rail · topbar · main · statusbar -->
  <div v-else class="h-dvh w-full overflow-hidden bg-app flex">
    <!-- Below md the rail is an overlay drawer; from md up it is a static column. `invisible`
         while closed is not cosmetic: it takes the off-screen links out of the tab order and
         the accessibility tree, which `-translate-x-full` alone does not. -->
    <IconRail
      :open="railOpen"
      class="max-md:fixed max-md:inset-y-0 max-md:left-0 max-md:z-40 max-md:transition-[transform,visibility]"
      :class="
        railOpen
          ? 'max-md:translate-x-0 max-md:visible'
          : 'max-md:-translate-x-full max-md:invisible'
      "
    />

    <!-- Scrim, only while the drawer is open. -->
    <div
      v-if="railOpen"
      class="fixed inset-0 z-30 bg-black/50 md:hidden"
      @click="railOpen = false"
    />

    <div class="flex-1 min-w-0 flex flex-col">
      <TopBar @toggle-rail="railOpen = !railOpen" />
      <!-- @container makes the content pane the measuring root for the tables inside it;
           overflow-x-auto is the honest last resort when a pane cannot narrow further. -->
      <main class="@container flex-1 min-h-0 overflow-hidden overflow-x-auto">
        <slot />
      </main>
      <StatusBar />
    </div>
  </div>
</template>
