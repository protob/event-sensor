import type { Event } from "@/types";

export function event(over: Partial<Event> = {}): Event {
  return {
    id: "e1",
    name: "Show",
    start_date: "2026-06-01",
    source: "ticketmaster",
    listing_state: "listed",
    is_past: false,
    kind: "concert",
    ...over,
  } as Event;
}
