// Whether an event gets a kind tag, and which one: only the exceptions (festival,
// tribute / "music of X") are tagged - a "CONCERT" tag on every row is noise. Kind comes
// from the backend; the title regex is a fallback. A kept TM event is guaranteed to be
// the artist or a tribute to them (see ticketmaster.ClassifyArtistMatch). One place, so
// row, artist row and cards agree.
export interface EventTag {
  label: string;
  color: string; // hex - Tag tints bg/border/text from it
}

// Same marker set as the backend's tributeMarkerRe (kept in sync intentionally).
const TRIBUTE_RE = /\b(tribute|music of|music by|performed by|celebration of|candlelight)\b/i;

export function eventTag(event?: { kind?: string; name?: string }): EventTag | null {
  if (event?.kind === "festival") {
    return { label: "FESTIVAL", color: "#f59e0b" }; // amber - stands apart from blue accent
  }
  if (event?.kind === "tribute" || (event?.name && TRIBUTE_RE.test(event.name))) {
    return { label: "TRIBUTE", color: "#a78bfa" }; // violet - "his music, not him"
  }
  return null; // plain concert → bare row
}
