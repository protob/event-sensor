import { expect, type APIRequestContext, type Locator, type Page } from "@playwright/test";

// The database lives for the whole run, so every spec labels the rows it creates and
// asserts on its own label rather than on an empty list.
export const uniq = () => `e2e-${Date.now().toString(36)}-${Math.floor(Math.random() * 1e4)}`;

export const daysFromNow = (n: number) =>
  new Date(Date.now() + n * 86400_000).toISOString().slice(0, 10);

export async function token(request: APIRequestContext): Promise<string> {
  const res = await request.post("/api/auth/login", {
    data: { username: "admin", password: "password" },
  });
  expect(res.ok(), "login failed").toBeTruthy();
  return (await res.json()).token;
}

// Puts the page in a logged-in state without driving the form, so only auth.spec pays
// for the login flow. The key is the one the auth store reads on boot. It is re-applied
// on every navigation, so a spec that logs out has to open its session through the form.
export async function signIn(page: Page, request: APIRequestContext) {
  const t = await token(request);
  await page.addInitScript((value) => {
    localStorage.setItem("auth_token", value);
  }, t);
}

export function authed(request: APIRequestContext, t: string) {
  const headers = { Authorization: `Bearer ${t}` };
  return {
    get: (url: string) => request.get(url, { headers }),
    post: (url: string, data: unknown) => request.post(url, { headers, data }),
    put: (url: string, data: unknown) => request.put(url, { headers, data }),
    delete: (url: string) => request.delete(url, { headers }),
  };
}

export async function createArtist(request: APIRequestContext, t: string, name: string) {
  const res = await authed(request, t).post("/api/artists", { name, fetch_mode: "manual" });
  expect(res.ok(), `create artist ${name}`).toBeTruthy();
  return (await res.json()) as { id: string; name: string };
}

export async function createEvent(
  request: APIRequestContext,
  t: string,
  body: Record<string, unknown>,
) {
  const res = await authed(request, t).post("/api/events", body);
  expect(res.ok(), "create event").toBeTruthy();
  return (await res.json()) as { id: string; name: string };
}

export async function fetchEvents(request: APIRequestContext, t: string, artistId: string) {
  const res = await authed(request, t).post(`/api/artists/${artistId}/fetch-events`, {});
  expect(res.ok(), "fetch events").toBeTruthy();
  return res.json();
}

// Points an artist at a keyword with no fixture: the stub answers with an empty page, so
// the next fetch sweeps what is un-claimed and delists what is not.
export async function silence(request: APIRequestContext, t: string, artistId: string) {
  const res = await authed(request, t).put(`/api/artists/${artistId}`, { name: `Gone ${uniq()}` });
  expect(res.ok(), "rename artist").toBeTruthy();
}

// The radar is filtered client-side, so a spec narrows it to its own rows through the
// top-bar search rather than assuming it owns the list.
export async function radarRows(page: Page, query: string): Promise<Locator> {
  await page.getByTestId("search-input").fill(query);
  return page.getByTestId("events-table").locator('[data-testid^="event-row-"]');
}

export async function rowId(row: Locator): Promise<string> {
  return (await row.getAttribute("data-testid"))!.replace(/^event-row-/, "");
}

// Kind and status are native <select>s and the contract's ids name the <option>, so the
// select is reached through the option it holds.
export function optionSelect(page: Page, testId: string): Locator {
  return page.locator("select").filter({ has: page.getByTestId(testId) });
}

export async function libraryStatuses(request: APIRequestContext, t: string) {
  const entries = await (await authed(request, t).get("/api/library")).json();
  return entries as { event_id: string; status: string }[];
}

// The manual form does not hand the id back, so the event is looked up by the unique
// name its spec gave it.
export async function eventIdByName(
  request: APIRequestContext,
  t: string,
  name: string,
): Promise<string> {
  let id = "";
  await expect
    .poll(async () => {
      const res = await authed(request, t).get(`/api/events?q=${encodeURIComponent(name)}`);
      const rows = (await res.json()) as { id: string; name: string }[];
      id = rows.find((e) => e.name === name)?.id ?? "";
      return id;
    })
    .not.toBe("");
  return id;
}
