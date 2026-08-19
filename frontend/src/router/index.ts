import { createRouter, createWebHistory } from "vue-router";
import type { RouteRecordRaw } from "vue-router";
import { useAuthStore } from "@/stores/auth";

const routes: RouteRecordRaw[] = [
  {
    path: "/",
    name: "home",
    redirect: "/events",
  },
  {
    path: "/events",
    name: "events",
    component: () => import("@/views/EventsView.vue"),
  },
  {
    path: "/events/:id",
    name: "event-detail",
    component: () => import("@/views/EventDetailView.vue"),
    props: true,
  },
  {
    path: "/artists",
    name: "artists",
    component: () => import("@/views/ArtistsView.vue"),
    meta: { requiresAuth: true },
  },
  {
    path: "/artists/:id",
    name: "artist-detail",
    component: () => import("@/views/ArtistDetailView.vue"),
    props: true,
    meta: { requiresAuth: true },
  },
  {
    path: "/venues",
    name: "venues",
    component: () => import("@/views/VenuesView.vue"),
    meta: { requiresAuth: true },
  },
  {
    path: "/search",
    name: "search",
    component: () => import("@/views/SearchView.vue"),
  },
  {
    path: "/library",
    name: "library",
    component: () => import("@/views/LibraryView.vue"),
    meta: { requiresAuth: true },
  },
  {
    path: "/settings",
    name: "settings",
    component: () => import("@/views/SettingsView.vue"),
    meta: { requiresAuth: true },
  },
  {
    path: "/login",
    name: "login",
    component: () => import("@/views/LoginView.vue"),
    meta: { guestOnly: true, bare: true },
  },
  {
    path: "/reset-password",
    name: "reset-password",
    component: () => import("@/views/ResetPasswordView.vue"),
    meta: { guestOnly: true, bare: true },
  },
  {
    path: "/:pathMatch(.*)*",
    name: "not-found",
    component: () => import("@/views/NotFoundView.vue"),
  },
];

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
});

// Auth navigation guard
router.beforeEach((to, _from, next) => {
  const authStore = useAuthStore();

  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    return next({ name: "login", query: { redirect: to.fullPath } });
  }

  if (to.meta.guestOnly && authStore.isAuthenticated) {
    return next({ name: "home" });
  }

  next();
});

export default router;
