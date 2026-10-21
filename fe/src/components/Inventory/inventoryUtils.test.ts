import { describe, expect, it } from "vitest";

import {
  auditTotals,
  checkMovement,
  groupByCategory,
  movementLabel,
  stockPercent,
} from "@/components/Inventory/inventoryUtils";
import { StockItem, StockMovement } from "@/components/Inventory/types";

const item = (name: string, category: string): StockItem => ({
  id: name,
  clinic_id: "c",
  name,
  category,
  unit: "units",
  quantity: 1,
  min_quantity: 0,
  per_session: 0,
  status: "ok",
  modified_date: "",
});

describe("inventory helpers", () => {
  it("fills the stock bar relative to the minimum", () => {
    expect(stockPercent(0, 10)).toBe(0);
    expect(stockPercent(10, 10)).toBe(25);
    expect(stockPercent(40, 10)).toBe(100);
    expect(stockPercent(500, 10)).toBe(100);
    expect(stockPercent(3, 0)).toBe(100);
  });

  it("groups by category with uncategorised items last", () => {
    const groups = groupByCategory(
      [
        item("Gel", "Treatment"),
        item("Box", ""),
        item("Masks", "PPE"),
        item("Tape", "Treatment"),
      ],
      "Other",
    );
    expect(groups.map(([c, its]) => [c, its.length])).toEqual([
      ["PPE", 1],
      ["Treatment", 2],
      ["Other", 1],
    ]);
  });

  it("labels movements", () => {
    expect(movementLabel("restock", 20, 30)).toBe("+20");
    expect(movementLabel("use", -3, 27)).toBe("−3");
    expect(movementLabel("adjust", -5, 22)).toBe("= 22");
  });

  it("checks movements before sending", () => {
    expect(checkMovement("use", 2, 5)).toBeNull();
    expect(checkMovement("use", 6, 5)).toBe("stock_error__not_enough");
    expect(checkMovement("restock", 0, 5)).toBe("stock_error__at_least_one");
    expect(checkMovement("adjust", 0, 5)).toBeNull();
    expect(checkMovement("restock", 1.5, 5)).toBe("stock_error__number");
  });
});

describe("audit totals", () => {
  const sarah = { id: 10, first_name: "Sarah", last_name: "Johnson" };
  const gloves = { id: "g", name: "Gloves", unit: "pairs", category: "PPE" };
  const move = (kind: StockMovement["kind"], delta: number): StockMovement => ({
    id: Math.random().toString(),
    item_id: "g",
    kind,
    delta,
    quantity_after: 0,
    note: "",
    user: sarah,
    item: gloves,
    created_date: "",
  });

  it("sums use and openings per person and item, ignoring restocks", () => {
    const totals = auditTotals([
      move("use", -1),
      move("use", -2),
      move("open", -1),
      move("restock", 50),
    ]);
    expect(totals).toHaveLength(1);
    expect(totals[0].name).toBe("Sarah Johnson");
    expect([...totals[0].items.values()]).toEqual([
      { name: "Gloves", unit: "pairs", used: 3, opened: 1 },
    ]);
  });
});
