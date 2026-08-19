const MAP: Record<string, string> = {
  ELECTRONIC: "#818cf8",
  JAZZ: "#34d399",
  ROCK: "#f87171",
};

export function categoryColor(name?: string) {
  return MAP[(name || "").toUpperCase()] ?? "var(--color-muted)";
}
