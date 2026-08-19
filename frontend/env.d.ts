/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}

// Augment vue-router RouteMeta
import "vue-router";

declare module "vue-router" {
  interface RouteMeta {
    /** Route requires an authenticated user */
    requiresAuth?: boolean;
    /** Route is only accessible to unauthenticated users (login, register) */
    guestOnly?: boolean;
    /** Route renders without app chrome (auth pages) */
    bare?: boolean;
  }
}
