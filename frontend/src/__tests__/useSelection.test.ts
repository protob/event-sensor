import { test, expect } from "bun:test";
import { useSelection } from "@/composables/useSelection";

const order = ["a", "b", "c", "d"];

test("a shift-range covers the span in either direction", () => {
  const sel = useSelection(() => order);

  sel.toggle("c");
  sel.rangeTo("a");
  expect(sel.ids().sort()).toEqual(["a", "b", "c"]);

  sel.clear();
  sel.toggle("b");
  sel.rangeTo("d");
  expect(sel.ids().sort()).toEqual(["b", "c", "d"]);
});

// The anchor follows the visible order, so a re-sorted list selects what the eye sees.
test("the range follows the current order, not the click history", () => {
  let visible = [...order];
  const sel = useSelection(() => visible);

  sel.toggle("a");
  visible = ["d", "c", "b", "a"];
  sel.rangeTo("c");
  expect(sel.ids().sort()).toEqual(["a", "b", "c"]);
});

test("a range against an id that left the list degrades to a toggle", () => {
  const sel = useSelection(() => order);

  sel.toggle("a");
  sel.rangeTo("zz");
  expect(sel.ids()).toEqual(["a", "zz"]);
});

// A stale id keeps the bulk bar claiming more than it can act on.
test("prune drops ids the list no longer contains", () => {
  let visible = [...order];
  const sel = useSelection(() => visible);

  sel.selectAll();
  expect(sel.count.value).toBe(4);

  visible = ["a", "b"];
  sel.prune();
  expect(sel.ids().sort()).toEqual(["a", "b"]);
});

test("shift-click ranges and a plain click toggles", () => {
  const sel = useSelection(() => order);

  sel.onRowClick("a", { shiftKey: false } as MouseEvent);
  sel.onRowClick("c", { shiftKey: true } as MouseEvent);
  expect(sel.ids().sort()).toEqual(["a", "b", "c"]);

  sel.onRowClick("b", { shiftKey: false } as MouseEvent);
  expect(sel.ids().sort()).toEqual(["a", "c"]);
});
