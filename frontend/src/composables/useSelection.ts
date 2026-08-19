import { ref, computed } from "vue";

// Ephemeral, list-scoped multi-selection with Linear semantics (toggle, shift-range, select-all).
// `orderedIds` returns the current visual order so shift-range follows what the user sees. Never
// persist selection — it lives only as long as the list view is mounted.
export function useSelection(orderedIds: () => string[]) {
  const selected = ref<Set<string>>(new Set());
  const anchor = ref<string | null>(null); // for shift-range

  const count = computed(() => selected.value.size);
  const has = (id: string) => selected.value.has(id);
  const ids = () => [...selected.value];

  function toggle(id: string) {
    const s = new Set(selected.value);
    if (s.has(id)) s.delete(id);
    else s.add(id);
    selected.value = s;
    anchor.value = id;
  }

  function rangeTo(id: string) {
    const order = orderedIds();
    const a = anchor.value ?? id;
    const i = order.indexOf(a);
    const j = order.indexOf(id);
    if (i < 0 || j < 0) return toggle(id);
    const [lo, hi] = i < j ? [i, j] : [j, i];
    const s = new Set(selected.value);
    for (let k = lo; k <= hi; k++) s.add(order[k]);
    selected.value = s;
  }

  function selectAll() {
    selected.value = new Set(orderedIds());
  }

  function clear() {
    selected.value = new Set();
    anchor.value = null;
  }

  // Drop ids that no longer exist (e.g. after a list refresh) so the bar count stays honest.
  function prune() {
    const live = new Set(orderedIds());
    const next = new Set([...selected.value].filter((id) => live.has(id)));
    if (next.size !== selected.value.size) selected.value = next;
  }

  // Row click: shift = range, otherwise discrete toggle (cmd/ctrl also = toggle).
  function onRowClick(id: string, ev: MouseEvent) {
    if (ev.shiftKey) rangeTo(id);
    else toggle(id);
  }

  return { selected, count, has, ids, toggle, rangeTo, selectAll, clear, prune, onRowClick };
}

export type SelectionApi = ReturnType<typeof useSelection>;
