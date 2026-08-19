# Event Sensor - Frontend

Vue 3 SPA for Event Sensor. The Vite dev server runs on `:5173` and proxies `/api` to
the Go backend on `:8080`. `bun` is the package manager; the justfile recipes use it.

Vue 3 (`<script setup>`) + TypeScript, Tailwind CSS 4, Pinia, Vue Router, VueUse,
ofetch, unplugin-icons (MDI + circle-flags).

## Commands

```sh
bun install
bun run dev          # dev server
bun run build        # vue-tsc + vite build -> ../internal/spa/dist
bun run type-check   # vue-tsc
bun run lint         # eslint --fix
bun run format       # prettier
```

`VITE_API_BASE_URL` overrides the API base (default `/api`).

## Structure

```
src/
├── api/           # ofetch client + one module per resource — imported ONLY by stores
├── assets/        # main.css: theme tokens, spacing scale, touch rules
├── components/    # events/, library/, tree/, shell/, ui/, artists/, venues/, auth/
├── composables/   # stateful building blocks: selection, list keyboard, radar/lenses,
│                  #   library entry, toasts, theme
├── router/        # routes + auth guard
├── stores/        # Pinia: events, artists, categories, venues, auth, ui, settings —
│                  #   all server state, mutation and cross-store resync lives here
├── types/         # wire types mirroring the API
├── utils/         # stateless helpers: dates, country names, event sections,
│                  #   display name, category colors, status constants, api errors
└── views/         # route-level pages
```

`api/` is imported only by `stores/` — views and components read stores, never make
requests themselves.

## Tests

`bun test` runs the unit tests in `src/__tests__/` - pure functions only. Components are
covered by the browser suite in `../e2e/`, which locates elements by `data-testid`.
