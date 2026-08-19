// Title-cases a name for DISPLAY ONLY - the stored value is never changed, so search,
// matching and the Ticketmaster query keep using exactly what the user typed. Only
// all-lowercase words are touched; stylized casing ("cLOUDDEAD", "MF DOOM") survives.
export function displayName(name: string): string {
  return name
    .split(/(\s+)/)
    .map((w) => {
      if (!/[a-z]/.test(w)) return w;
      if (/[A-Z]/.test(w)) return w;
      return w.charAt(0).toUpperCase() + w.slice(1);
    })
    .join("");
}
