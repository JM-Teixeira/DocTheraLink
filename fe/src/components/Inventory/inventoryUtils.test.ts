import { describe, expect, it } from "vitest";

import {
  actionsFor,
  auditTotals,
  checkMovement,
  groupByCategory,
  movementLabel,
  reorderCsv,
  stockPercent,
} from "@/components/Inventory/inventoryUtils";
import { StockItem, StockMovement } from "@/components/Inventory/types";

const item = (
  name: string,
  category: string,
  over: Partial<StockItem> = {},
): StockItem => ({
  id: name,
  clinic_id: "c",
  name,
  category,
  unit: "units",
  quantity: 1,
  min_quantity: 0,
  per_session: 0,
  session_rules: [],
  tracking: "unit",
  target_quantity: 0,
  in_use: 0,
  open_life_days: 0,
  open_expired: false,
  reorder_quantity: 0,
  status: "ok",
  modified_date: "",
  ...over,
});

describe("inventory helpers", () => {
  it("offers the actions of each kind of item", () => {
    expect(actionsFor({ tracking: "unit" })).toEqual([
      "use",
      "restock",
      "adjust",
    ]);
    expect(actionsFor({ tracking: "container" })).toEqual([
      "open",
      "finish",
      "restock",
      "adjust",
    ]);
  });

  it("fills the stock bar relative to the minimum", () => {
    expect(stockPercent(0, 10)).toBe(0);
    expect(stockPercent(10, 10)).toBe(25);
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
    const m = (
      kind: StockMovement["kind"],
      amount: number,
      delta: number,
      after = 0,
    ) => ({ kind, amount, delta, quantity_after: after });
    expect(movementLabel(m("restock", 20, 20))).toBe("+20");
    expect(movementLabel(m("use", 3, -3))).toBe("−3");
    expect(movementLabel(m("open", 1, -1))).toBe("−1");
    expect(movementLabel(m("finish", 1, 0))).toBe("✓ 1");
    expect(movementLabel(m("adjust", 22, -5, 22))).toBe("= 22");
    expect(movementLabel(m("use", 0, -2))).toBe("−2"); // older records
  });

  it("checks movements before sending", () => {
    const stock = { quantity: 5, in_use: 1 };
    expect(checkMovement("use", 2, stock)).toBeNull();
    expect(checkMovement("use", 6, stock)).toBe("stock_error__not_enough");
    expect(checkMovement("open", 6, stock)).toBe("stock_error__not_enough");
    expect(checkMovement("finish", 2, stock)).toBe(
      "stock_error__not_that_many_open",
    );
    expect(checkMovement("restock", 0, stock)).toBe(
      "stock_error__at_least_one",
    );
    expect(checkMovement("adjust", 0, stock)).toBeNull();
    expect(checkMovement("restock", 1.5, stock)).toBe("stock_error__number");
  });

  it("exports the reorder list as CSV", () => {
    const csv = reorderCsv(
      [
        item('Gel "frio"', "Tratamento", {
          quantity: 2,
          reorder_quantity: 10,
          unit: "frascos",
        }),
      ],
      ["Artigo", "Categoria", "Em stock", "Encomendar", "Unidade"],
    );
    expect(csv).toBe(
      '"Artigo";"Categoria";"Em stock";"Encomendar";"Unidade"\n"Gel ""frio""";"Tratamento";"2";"10";"frascos"',
    );
  });
});

describe("audit totals", () => {
  const sarah = { id: 10, first_name: "Sarah", last_name: "Johnson" };
  const move = (
    itemId: string,
    kind: StockMovement["kind"],
    amount: number,
  ): StockMovement => ({
    id: Math.random().toString(),
    item_id: itemId,
    kind,
    amount,
    delta: kind === "restock" ? amount : kind === "finish" ? 0 : -amount,
    quantity_after: 0,
    in_use_after: 0,
    note: "",
    user: sarah,
    item: { id: itemId, name: itemId, unit: "u", category: "" },
    created_date: "",
  });

  it("sums use, openings and finishes per person and item, ignoring restocks", () => {
    const totals = auditTotals([
      move("Gloves", "use", 1),
      move("Gloves", "use", 2),
      move("Gel", "open", 1),
      move("Gel", "finish", 1),
      move("Gloves", "restock", 50),
    ]);
    expect(totals).toHaveLength(1);
    expect([...totals[0].items.values()]).toEqual([
      { name: "Gloves", unit: "u", used: 3, opened: 0, finished: 0 },
      { name: "Gel", unit: "u", used: 0, opened: 1, finished: 1 },
    ]);
  });
});
