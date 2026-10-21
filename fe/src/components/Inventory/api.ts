import {
  ItemRequest,
  MovementRequest,
  StockItem,
  StockMovement,
  StockSummary,
} from "@/components/Inventory/types";

import { Type } from "@/Utils/request/api";

/** Inventory API (be/internal/modules/inventory/routes.go). */
const InventoryRoutes = {
  summary: {
    path: "/api/v1/inventory/summary/",
    method: "GET",
    TRes: Type<StockSummary>(),
  },
  audit: {
    path: "/api/v1/inventory/audit/",
    method: "GET",
    TRes: Type<StockMovement[]>(),
  },
  list: {
    path: "/api/v1/inventory/items/",
    method: "GET",
    TRes: Type<StockItem[]>(),
  },
  create: {
    path: "/api/v1/inventory/items/",
    method: "POST",
    TBody: Type<ItemRequest>(),
    TRes: Type<StockItem>(),
  },
  update: {
    path: "/api/v1/inventory/items/{id}/",
    method: "PUT",
    TBody: Type<ItemRequest>(),
    TRes: Type<StockItem>(),
  },
  archive: {
    path: "/api/v1/inventory/items/{id}/",
    method: "DELETE",
    TRes: Type<void>(),
  },
  move: {
    path: "/api/v1/inventory/items/{id}/movements/",
    method: "POST",
    TBody: Type<MovementRequest>(),
    TRes: Type<StockItem>(),
  },
  history: {
    path: "/api/v1/inventory/items/{id}/movements/",
    method: "GET",
    TRes: Type<StockMovement[]>(),
  },
} as const;

export default InventoryRoutes;
