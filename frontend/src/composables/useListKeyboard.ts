import { useEventListener } from "@vueuse/core";
import type { SelectionApi } from "./useSelection";

function inEditable(t: EventTarget | null) {
  const el = t as HTMLElement | null;
  return !!el && (el.tagName === "INPUT" || el.tagName === "TEXTAREA" || el.isContentEditable);
}

// List-wide keyboard semantics: Cmd/Ctrl+A selects everything, Esc clears the
// selection. `onEscape` runs first and consumes the key when it returns true, so a
// view can close its own overlay (filter panel, drawer) on the same keypress.
export function useListKeyboard(
  selection: SelectionApi,
  isEmpty: () => boolean,
  onEscape?: () => boolean,
) {
  useEventListener(window, "keydown", (e: KeyboardEvent) => {
    if (inEditable(e.target)) return;
    if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "a") {
      if (isEmpty()) return;
      e.preventDefault();
      selection.selectAll();
    } else if (e.key === "Escape") {
      if (onEscape?.()) return;
      if (selection.count.value > 0) selection.clear();
    }
  });
}

export { inEditable };
