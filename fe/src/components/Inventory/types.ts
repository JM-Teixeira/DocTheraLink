/** Types of the inventory API (be/internal/modules/inventory/model.go). */

export type StockStatus = "ok" | "low" | "out";
export type MovementKind = "restock" | "use" | "adjust" | "open" | "finish";

/** unit = single-use (used); container = multi-use (opened, then finished). */
export type Tracking = "unit" | "container";

export interface SessionRule {
  division_id: string;
  quantity: number;
}

export interface StockItem {
  id: string;
  clinic_id: string;
  name: string;
  category: string;
  unit: string;
  quantity: number;
  min_quantity: number;
  per_session: number; // used automatically per attended session (default)
  session_rules: SessionRule[]; // per-division overrides
  tracking: Tracking;
  target_quantity: number;
  in_use: number; // containers open now
  oldest_open_at?: string | null;
  open_life_days: number;
  open_expired: boolean;
  reorder_quantity: number;
  status: StockStatus;
  modified_date: string;
}

export interface StockMovement {
  id: string;
  item_id: string;
  kind: MovementKind;
  amount: number;
  delta: number;
  quantity_after: number;
  in_use_after: number;
  note: string;
  user?: { id: number; first_name: string; last_name: string };
  item?: { id: string; name: string; unit: string; category: string };
  appointment_id?: string | null; // set for automatic session use
  created_date: string;
}

export interface StockSummary {
  items: number;
  low: number;
  out: number;
  open_expired: number;
}

export interface ItemRequest {
  name: string;
  category: string;
  unit: string;
  min_quantity: number;
  per_session: number;
  tracking: Tracking;
  target_quantity: number;
  open_life_days: number;
  quantity?: number; // opening stock, create only
}

export interface MovementRequest {
  kind: MovementKind;
  quantity: number;
  note?: string;
}
