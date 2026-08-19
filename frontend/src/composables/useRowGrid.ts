// Column templates for the event tables, narrowest tier first.
//
// A header and its rows must carry the identical template string or they drift apart and
// every column looks subtly misaligned - hence one shared source instead of a template
// written out at each call site.
//
// The tiers are container queries against the events pane (see the `@container` on the
// EventsView right pane and on AppShell's <main>), not viewport breakpoints. The pane's width
// depends on whether the category tree is open, which no viewport query can observe: at
// 1400px with the tree open and at 1100px with it closed the pane is the same ~1040px and
// wants the same columns.
//
// Tailwind's container scale: @lg = 32rem, @3xl = 48rem.
//
// Below the widest tier the narrow sets DROP columns rather than shrinking all of them. A
// 40px-wide truncated location is worse than no location. Order of sacrifice:
//
//   location  first - the longest string, and the flag inside it already carries the country
//   actions   next  - the whole row is tappable and the detail view holds every action
//
// The claim toggles (I/G) and the date survive to the narrowest tier, because reading what is
// on and claiming it is the entire point of the app on a phone.
//
// A shorter grid template does not remove the extra cells, it wraps them onto a
// phantom second row. Each column dropped here has a matching `hidden @lg:` / `hidden @3xl:`
// on its cell in the row and header components. When you add or remove a column, change the
// template and the cell's visibility class in the same edit or the row silently grows a
// second line.

// Events table, with the 26px bulk-selection gutter.
export const EVENT_GRID_SELECTABLE = [
  "grid-cols-[26px_44px_minmax(0,1fr)_92px]",
  "@lg:grid-cols-[26px_44px_minmax(0,1fr)_110px_52px]",
  "@3xl:grid-cols-[26px_44px_minmax(0,1fr)_150px_110px_52px]",
].join(" ");

// Events table without the selection gutter.
export const EVENT_GRID_PLAIN = [
  "grid-cols-[44px_minmax(0,1fr)_92px]",
  "@lg:grid-cols-[44px_minmax(0,1fr)_110px_52px]",
  "@3xl:grid-cols-[44px_minmax(0,1fr)_150px_110px_52px]",
].join(" ");

// Per-artist event list (ArtistEventRow + the header in ArtistDetailView). Same shape, plus a
// 90px tag column that goes at the same tier as the actions.
export const ARTIST_EVENT_GRID = [
  "grid-cols-[44px_minmax(0,1fr)_92px]",
  "@lg:grid-cols-[44px_90px_minmax(0,1fr)_100px_50px]",
  "@3xl:grid-cols-[44px_90px_minmax(0,1fr)_120px_100px_50px]",
].join(" ");

// Visibility classes, exported so a cell can never fall out of step with the tier that drops
// its column. Read them as "this cell exists from this tier up".
export const CELL_FROM_LG = "hidden @lg:flex";
export const CELL_FROM_LG_BLOCK = "hidden @lg:block";
export const CELL_FROM_3XL = "hidden @3xl:flex";
export const CELL_FROM_3XL_BLOCK = "hidden @3xl:block";
