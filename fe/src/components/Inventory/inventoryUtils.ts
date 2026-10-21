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

/**
 * Totals per person for the audit view: how much each one used and opened,
 * per item ("Sarah: 12 pairs of gloves used, 2 boxes opened").
 */
export function auditTotals(moves: StockMovement[]) {
  const people = new Map<
    number,
    {
      name: string;
      items: Map<
        string,
        { name: string; unit: string; used: number; opened: number }
      >;
    }
  >();
  for (const m of moves) {
    if (!m.user || (m.kind !== "use" && m.kind !== "open")) continue;
    const p = people.get(m.user.id) ?? {
      name: `${m.user.first_name} ${m.user.last_name}`,
      items: new Map(),
    };
    const it = p.items.get(m.item_id) ?? {
      name: m.item?.name ?? "",
      unit: m.item?.unit ?? "",
      used: 0,
      opened: 0,
    };
    if (m.kind === "use") it.used += -m.delta;
    else it.opened += -m.delta;
    p.items.set(m.item_id, it);
    people.set(m.user.id, p);
  }
  return [...people.values()].sort((a, b) => a.name.localeCompare(b.name));
}

/** Movement text for the history list: "+20", "−3", "= 12". */
export function movementLabel(
  kind: MovementKind,
  delta: number,
  after: number,
) {
  if (kind === "adjust") return `= ${after}`;
  return delta > 0 ? `+${delta}` : `−${Math.abs(delta)}`;
}

/**
 * Client check of a movement before sending (the API re-checks):
 * returns a translation key, or null when fine.
 */
export function checkMovement(
  kind: MovementKind,
  quantity: number,
  inStock: number,
) {
  if (!Number.isInteger(quantity) || quantity < 0) return "stock_error__number";
  if (kind !== "adjust" && quantity < 1) return "stock_error__at_least_one";
  if ((kind === "use" || kind === "open") && quantity > inStock)
    return "stock_error__not_enough";
  return null;
}
