/** Pure helpers for the inventory page (tested in inventoryUtils.test.ts). */
import {
  MovementKind,
  StockItem,
  StockMovement,
  StockStatus,
} from "@/components/Inventory/types";

/** Badge and bar colours per stock status (literal classes for Tailwind). */
export const STATUS_STYLES: Record<
  StockStatus,
  { badge: string; bar: string }
> = {
  ok: {
    badge: "bg-emerald-50 text-emerald-700 ring-emerald-200",
    bar: "bg-emerald-500",
  },
  low: {
    badge: "bg-amber-50 text-amber-800 ring-amber-200",
    bar: "bg-amber-500",
  },
  out: { badge: "bg-rose-50 text-rose-700 ring-rose-200", bar: "bg-rose-500" },
};

/** Actions offered for an item: single-use items are used, containers are opened and finished. */
export function actionsFor(item: Pick<StockItem, "tracking">): MovementKind[] {
  return item.tracking === "container"
    ? ["open", "finish", "restock", "adjust"]
    : ["use", "restock", "adjust"];
}

/**
 * How full the stock bar is, 0–100. The minimum sits at 25 % of the bar, so
 * anything at or below the alert level is visibly short.
 */
export function stockPercent(quantity: number, min: number) {
  if (quantity <= 0) return 0;
  const full = Math.max(min * 4, 1);
  return Math.min(100, Math.round((quantity / full) * 100));
}

/** Items grouped by category (empty category last as "Other"), in order. */
export function groupByCategory(items: StockItem[], other: string) {
  const groups = new Map<string, StockItem[]>();
  for (const it of items) {
    const key = it.category || other;
    groups.set(key, [...(groups.get(key) ?? []), it]);
  }
  return [...groups.entries()].sort(([a], [b]) =>
    a === other ? 1 : b === other ? -1 : a.localeCompare(b),
  );
}

/** How many a movement concerned (older records only have the delta). */
const amountOf = (m: Pick<StockMovement, "amount" | "delta">) =>
  m.amount || Math.abs(m.delta);

/** Movement text for history and audit: "+20", "−3", "= 12", "✓ 1". */
export function movementLabel(
  m: Pick<StockMovement, "kind" | "amount" | "delta" | "quantity_after">,
) {
  switch (m.kind) {
    case "adjust":
      return `= ${m.quantity_after}`;
    case "finish":
      return `✓ ${amountOf(m)}`;
    case "restock":
      return `+${amountOf(m)}`;
    default:
      return `−${amountOf(m)}`;
  }
}

/**
 * Totals per person for the audit view: how much each one used, opened and
 * finished, per item ("Sarah: 12 pairs of gloves used, 2 bottles opened").
 */
export function auditTotals(moves: StockMovement[]) {
  type Line = {
    name: string;
    unit: string;
    used: number;
    opened: number;
    finished: number;
  };
  const people = new Map<number, { name: string; items: Map<string, Line> }>();
  for (const m of moves) {
    if (!m.user || !["use", "open", "finish"].includes(m.kind)) continue;
    const p = people.get(m.user.id) ?? {
      name: `${m.user.first_name} ${m.user.last_name}`,
      items: new Map(),
    };
    const it = p.items.get(m.item_id) ?? {
      name: m.item?.name ?? "",
      unit: m.item?.unit ?? "",
      used: 0,
      opened: 0,
      finished: 0,
    };
    if (m.kind === "use") it.used += amountOf(m);
    else if (m.kind === "open") it.opened += amountOf(m);
    else it.finished += amountOf(m);
    p.items.set(m.item_id, it);
    people.set(m.user.id, p);
  }
  return [...people.values()].sort((a, b) => a.name.localeCompare(b.name));
}

/**
 * Client check of a movement before sending (the API re-checks):
 * returns a translation key, or null when fine.
 */
export function checkMovement(
  kind: MovementKind,
  quantity: number,
  item: Pick<StockItem, "quantity" | "in_use">,
) {
  if (!Number.isInteger(quantity) || quantity < 0) return "stock_error__number";
  if (kind !== "adjust" && quantity < 1) return "stock_error__at_least_one";
  if ((kind === "use" || kind === "open") && quantity > item.quantity)
    return "stock_error__not_enough";
  if (kind === "finish" && quantity > item.in_use)
    return "stock_error__not_that_many_open";
  return null;
}

/** Reorder list as CSV (name;category;in stock;to order;unit) for suppliers. */
export function reorderCsv(items: StockItem[], header: string[]) {
  const cell = (v: string | number) => `"${String(v).replace(/"/g, '""')}"`;
  return [
    header,
    ...items.map((i) => [
      i.name,
      i.category,
      i.quantity,
      i.reorder_quantity,
      i.unit,
    ]),
  ]
    .map((row) => row.map(cell).join(";"))
    .join("\n");
}
