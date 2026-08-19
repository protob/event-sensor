<script setup lang="ts">
import { computed, ref } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";
import { storeToRefs } from "pinia";
import { useAuthStore } from "@/stores/auth";
import { useTheme } from "@/composables";
import IconEvents from "~icons/mdi/calendar-month";
import IconSearch from "~icons/mdi/magnify";
import IconArtists from "~icons/mdi/account-music";
import IconVenues from "~icons/mdi/map-marker-outline";
import IconInterested from "~icons/mdi/bookmark-outline";
import IconSettings from "~icons/mdi/cog";
import IconMoon from "~icons/mdi/weather-night";
import IconSun from "~icons/mdi/white-balance-sunny";
import IconLogout from "~icons/mdi/logout";
import IconLogin from "~icons/mdi/login";

// `open` only reaches the drawer state below md; the overlay positioning and the transform
// come from AppShell, which owns the state. Here it drives the dialog semantics.
defineProps<{ open?: boolean }>();

const route = useRoute();
const router = useRouter();
const { isAuthenticated, user } = storeToRefs(useAuthStore());
const { logout } = useAuthStore();
const { theme, toggle: toggleTheme } = useTheme();

interface NavItem {
  name: string;
  path: string;
  icon: typeof IconEvents;
  label: string;
  auth: boolean;
}

const items: NavItem[] = [
  { name: "events", path: "/events", icon: IconEvents, label: "Events", auth: false },
  { name: "search", path: "/search", icon: IconSearch, label: "Search", auth: false },
  { name: "artists", path: "/artists", icon: IconArtists, label: "Artists", auth: true },
  { name: "venues", path: "/venues", icon: IconVenues, label: "Venues", auth: true },
  { name: "library", path: "/library", icon: IconInterested, label: "Library", auth: true },
  { name: "settings", path: "/settings", icon: IconSettings, label: "Settings", auth: true },
];

const visibleItems = computed(() => items.filter((i) => !i.auth || isAuthenticated.value));

function isActive(name: string) {
  return route.name === name;
}

const initials = computed(() => {
  const u = user.value?.username || "";
  return u.slice(0, 2).toUpperCase() || "?";
});

const popoverOpen = ref(false);

async function handleLogout() {
  popoverOpen.value = false;
  await logout();
  router.push({ name: "login" });
}
</script>

<template>
  <nav
    class="w-[calc(3.5rem+env(safe-area-inset-left))] pl-[env(safe-area-inset-left)] h-full bg-surface border-r border-line flex flex-col items-center pt-3 relative shrink-0"
    :role="open ? 'dialog' : undefined"
    :aria-modal="open ? 'true' : undefined"
    :aria-label="open ? 'Navigation' : undefined"
  >
    <!-- reserved space for a brand logo; the calendar item already navigates to /events -->
    <div class="h-5 w-5 mb-4 shrink-0" aria-hidden="true" />

    <RouterLink
      v-for="item in visibleItems"
      :key="item.name"
      :to="item.path"
      :title="item.label"
      :aria-label="item.label"
      :data-testid="'nav-' + item.name"
      class="w-full h-[46px] flex items-center justify-center border-l-2 transition-colors"
      :class="
        isActive(item.name)
          ? 'bg-accent-chip border-accent-bright text-accent-text'
          : 'border-transparent text-faint hover:text-muted'
      "
    >
      <component :is="item.icon" class="h-5 w-5" />
    </RouterLink>

    <!-- bottom: avatar / auth -->
    <div class="absolute bottom-2 left-0 right-0 flex justify-center">
      <template v-if="isAuthenticated">
        <div class="relative">
          <button
            class="h-9 w-9 rounded-full bg-accent-chip border border-accent-chip-border text-accent-text text-label font-mono font-semibold flex items-center justify-center hover:brightness-125"
            :title="user?.username"
            data-testid="user-menu"
            @click="popoverOpen = !popoverOpen"
          >
            {{ initials }}
          </button>
          <div
            v-if="popoverOpen"
            class="absolute bottom-0 left-12 z-50 w-44 bg-surface border border-line-2 rounded-md shadow-lg p-2 flex flex-col gap-1"
          >
            <div class="px-2 py-1 font-mono text-label text-muted truncate">
              {{ user?.username }}
            </div>
            <button
              class="flex items-center gap-2 px-2 py-1.5 rounded-sm text-xs text-body hover:bg-surface-3"
              @click="toggleTheme()"
            >
              <IconSun v-if="theme === 'dark'" class="h-4 w-4" />
              <IconMoon v-else class="h-4 w-4" />
              {{ theme === "dark" ? "Light mode" : "Dark mode" }}
            </button>
            <button
              class="flex items-center gap-2 px-2 py-1.5 rounded-sm text-xs text-danger hover:bg-surface-3"
              data-testid="logout"
              @click="handleLogout"
            >
              <IconLogout class="h-4 w-4" />
              Logout
            </button>
          </div>
        </div>
      </template>
      <template v-else>
        <div class="flex flex-col items-center gap-2">
          <button
            class="h-8 w-8 rounded-sm text-faint hover:text-muted flex items-center justify-center"
            :title="theme === 'dark' ? 'Light mode' : 'Dark mode'"
            @click="toggleTheme()"
          >
            <IconSun v-if="theme === 'dark'" class="h-5 w-5" />
            <IconMoon v-else class="h-5 w-5" />
          </button>
          <RouterLink
            to="/login"
            title="Login"
            class="h-8 w-8 rounded-sm text-faint hover:text-accent-text flex items-center justify-center"
          >
            <IconLogin class="h-5 w-5" />
          </RouterLink>
        </div>
      </template>
    </div>
  </nav>
</template>
