import { defineConfig } from "@playwright/test";

// 8099 is the e2e port and nothing else: 8080 is `just dev`, 5173 is Vite.
const PORT = process.env.ES_E2E_PORT ?? "8099";
const baseURL = `http://127.0.0.1:${PORT}`;

export default defineConfig({
  testDir: ".",
  timeout: 30_000,
  // A retry turns a race into a pass and hides it. One worker, because the suite shares
  // one SQLite file and one single-user account.
  retries: 0,
  workers: 1,
  reporter: "line",
  use: {
    baseURL,
    trace: "retain-on-failure",
  },
  webServer: {
    command: "./start.sh",
    // Public on a loopback bind, and it answers only once the migrations have run.
    url: `${baseURL}/api/events`,
    reuseExistingServer: false,
    // bun install + vite build + go build from a cold cache.
    timeout: 300_000,
  },
});
