import type { Event } from "@/types";
import { countryName } from "./country";

// The grouping/divider pipeline shared by the events table, the events cards and the
// library's date lens. Input must already be sorted the way the consumer displays it.

export function headlinerName(e: Event): string {
  const as = e.artists ?? [];
  return (as.find((a) => a.is_headliner) ?? as[0])?.artist_name ?? e.name;
}

export interface EventGroup {
  key: string;
  label: string;
  code?: string; // country code for flag
  events: Event[];
}

export function groupEvents(events: Event[], by: "artist" | "country"): EventGroup[] {
  const map = new Map<string, EventGroup>();
  for (const e of events) {
    let key: string;
    let label: string;
    let code: string | undefined;
    if (by === "artist") {
      label = headlinerName(e);
      key = label.toLowerCase();
    } else {
      code = (e.venue?.country_code || "xx").toLowerCase();
      key = code;
      label = countryName(code, e.venue?.country || code.toUpperCase());
    }
    let g = map.get(key);
    if (!g) map.set(key, (g = { key, label, code, events: [] }));
    g.events.push(e);
  }
  // Biggest groups first - the quickest scan is the densest cluster.
  return [...map.values()].sort(
    (a, b) => b.events.length - a.events.length || a.label.localeCompare(b.label),
  );
}

export interface DateDivider {
  kind: "divider";
  key: string;
  label: string;
  yearBreak: boolean;
}

export interface SectionEvent {
  kind: "event";
  key: string;
  event: Event;
  index: number;
}

export type SectionRow = DateDivider | SectionEvent;

// Interleaves sticky month/year dividers into a date-sorted list. `index` on event
// rows is the position among events (striping), not among rows.
export function monthDividers(events: Event[], enabled = true): SectionRow[] {
  const rows: SectionRow[] = [];
  let lastYM = "";
  let lastYear = "";
  let i = 0;
  for (const e of events) {
    if (enabled) {
      const d = new Date(e.start_date);
      if (!isNaN(d.getTime())) {
        const ym = `${d.getFullYear()}-${d.getMonth()}`;
        if (ym !== lastYM) {
          const year = String(d.getFullYear());
          rows.push({
            kind: "divider",
            key: `div-${ym}`,
            label: d.toLocaleDateString("en-GB", { month: "long", year: "numeric" }).toUpperCase(),
            yearBreak: lastYear !== "" && year !== lastYear,
          });
          lastYM = ym;
          lastYear = year;
        }
      }
    }
    rows.push({ kind: "event", key: e.id, event: e, index: i++ });
  }
  return rows;
}
