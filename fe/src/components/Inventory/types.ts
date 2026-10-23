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
  floor?: number | null; // where it is kept (none = whole clinic)
  can_change: boolean; // the viewer may use/open/finish/restock/count/edit it
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
  floor?: number | null;
  clinic_id?: string; // admins
}

export interface MovementRequest {
  kind: MovementKind;
  quantity: number;
  note?: string;
}

/** A person may change the items on these floors and in these categories (empty = all). */
export interface StockGrant {
  id: string;
  clinic_id: string;
  user_id: number;
  floors: number[];
  categories: string[];
}

export interface GrantRule {
  floors: number[];
  categories: string[];
}

/** What the viewer may change: everything (managers) or what their grants cover. */
export interface StockAccess {
  manage: boolean;
  grants: StockGrant[];
}
