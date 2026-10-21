/** Types of the inventory API (be/internal/modules/inventory/model.go). */

export type StockStatus = "ok" | "low" | "out";
export type MovementKind = "restock" | "use" | "adjust";

export interface StockItem {
  id: string;
  clinic_id: string;
  name: string;
  category: string;
  unit: string;
  quantity: number;
  min_quantity: number;
  status: StockStatus;
  modified_date: string;
}

export interface StockMovement {
  id: string;
  item_id: string;
  kind: MovementKind;
  delta: number;
  quantity_after: number;
  note: string;
  user?: { id: number; first_name: string; last_name: string };
  created_date: string;
}

export interface StockSummary {
  items: number;
  low: number;
  out: number;
}

export interface ItemRequest {
  name: string;
  category: string;
  unit: string;
  min_quantity: number;
  quantity?: number; // opening stock, create only
}

export interface MovementRequest {
  kind: MovementKind;
  quantity: number;
  note?: string;
}
