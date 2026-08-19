const NAMES: Record<string, string> = {
  gb: "Great Britain",
  uk: "United Kingdom",
  de: "Germany",
  fr: "France",
  at: "Austria",
  pl: "Poland",
  nl: "Netherlands",
  be: "Belgium",
  es: "Spain",
  it: "Italy",
  cz: "Czech Republic",
  ch: "Switzerland",
  se: "Sweden",
  dk: "Denmark",
  no: "Norway",
  fi: "Finland",
  ie: "Ireland",
  pt: "Portugal",
  hu: "Hungary",
  gr: "Greece",
  ro: "Romania",
  sk: "Slovakia",
  si: "Slovenia",
  hr: "Croatia",
  lu: "Luxembourg",
  bg: "Bulgaria",
  cy: "Cyprus",
  ee: "Estonia",
  lv: "Latvia",
  lt: "Lithuania",
  mt: "Malta",
  is: "Iceland",
  li: "Liechtenstein",
  ad: "Andorra",
  mc: "Monaco",
  sm: "San Marino",
  va: "Vatican City",
  al: "Albania",
  ba: "Bosnia & Herzegovina",
  xk: "Kosovo",
  me: "Montenegro",
  mk: "North Macedonia",
  rs: "Serbia",
  tr: "Turkey",
};

// ─── Travel-region presets ───
// The region is "countries worth the trip for a show", not strict geographic Europe.
//
// These presets feed the Settings region picker, which saves the chosen codes to the
// `region.codes` user setting. That setting is what the BACKEND reads at fetch time to
// decide which events to store (see ticketmaster.DefaultRegionCodes, the same list on the
// Go side). The Events page's country chips are a separate, transient view filter over
// what is already stored - the two are deliberately not connected.

// Mirror of the backend's ticketmaster.DefaultRegionCodes. Keep both in sync.
const DEFAULT_REGION_CODES = [
  // EU members
  "at",
  "be",
  "bg",
  "hr",
  "cy",
  "cz",
  "dk",
  "ee",
  "fi",
  "fr",
  "de",
  "gr",
  "hu",
  "ie",
  "it",
  "lv",
  "lt",
  "lu",
  "mt",
  "nl",
  "pl",
  "pt",
  "ro",
  "sk",
  "si",
  "es",
  "se",
  // EFTA / Schengen non-EU
  "is",
  "li",
  "no",
  "ch",
  // UK + microstates
  "gb",
  "ad",
  "mc",
  "sm",
  // Western Balkans
  "al",
  "ba",
  "xk",
  "me",
  "mk",
  "rs",
  // Turkey
  "tr",
];

export interface RegionPreset {
  id: string;
  label: string;
  codes: string[];
}

export const REGION_PRESETS: RegionPreset[] = [
  { id: "europe-default", label: "Europe + Turkey", codes: DEFAULT_REGION_CODES },
];

// Rectangular flags (flagcdn) read far better than circle flags at small sizes.
export function flagUrl(cc: string): string {
  return `https://flagcdn.com/${(cc || "xx").toLowerCase()}.svg`;
}

export function countryName(cc: string, fallback = ""): string {
  return NAMES[(cc || "").toLowerCase()] ?? (fallback || (cc || "").toUpperCase());
}
