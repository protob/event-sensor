import type { Event } from "@/types";
import type { EventSort } from "@/stores/ui";
import { headlinerName } from "@/utils/eventSections";

export function matchesQuery(e: Event, q: string): boolean {
  if (!q) return true;
  const needle = q.toLowerCase();
  const hay = [
    e.name,
    e.kind,
    e.venue?.name,
    e.venue?.city,
    e.venue?.country,
    ...(e.artists?.map((a) => a.artist_name) ?? []),
  ]
    .filter(Boolean)
    .join(" ")
    .toLowerCase();
  return hay.includes(needle);
}

export function inDateRange(e: Event, start: string, end: string): boolean {
  const d = e.start_date?.slice(0, 10);
  if (!d) return true;
  if (start && d < start) return false;
  if (end && d > end) return false;
  return true;
}

function ts(e: Event): number {
  const t = new Date(e.start_date).getTime();
  return isNaN(t) ? 0 : t;
}

export function sortEvents(list: Event[], sort: EventSort): Event[] {
  const copy = [...list];
  switch (sort) {
    case "date-desc":
      return copy.sort((a, b) => ts(b) - ts(a));
    case "artist":
      return copy.sort((a, b) => headlinerName(a).localeCompare(headlinerName(b)));
    default:
      return copy.sort((a, b) => ts(a) - ts(b));
  }
}

// Days-ahead window helpers for the filter chips.
export function withinDays(e: Event, days: number): boolean {
  const t = new Date(e.start_date).getTime();
  if (isNaN(t)) return false;
  const now = Date.now();
  return t >= now - 24 * 3600 * 1000 && t <= now + days * 24 * 3600 * 1000;
}
