import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  History,
  Minus,
  PackageOpen,
  Pencil,
  Plus,
  Search,
  Trash2,
  TriangleAlert,
} from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";

import { cn } from "@/lib/utils";

import { Button } from "@/components/ui/button";

import Page from "@/components/Common/Page";
import {
  HistoryDialog,
  ItemDialog,
  MovementDialog,
} from "@/components/Inventory/InventoryDialogs";
import InventoryRoutes from "@/components/Inventory/api";
import {
  STATUS_STYLES,
  groupByCategory,
  stockPercent,
} from "@/components/Inventory/inventoryUtils";
import {
  ItemRequest,
  MovementKind,
  StockItem,
} from "@/components/Inventory/types";

import useAuthUser from "@/hooks/useAuthUser";
import useDebouncedValue from "@/hooks/useDebouncedValue";

import * as Notification from "@/Utils/Notifications";
import { callApi } from "@/Utils/request/query";
import query from "@/Utils/request/query";
import { HTTPError } from "@/Utils/request/types";
import { canManageBoard, isReadOnly } from "@/Utils/roles";

type Filter = "all" | "attention" | "out";

/** The clinic's consumables: stock levels, use / restock / count, alerts. */
export default function InventoryPage() {
  const { t } = useTranslation();
  const user = useAuthUser();
  const queryClient = useQueryClient();
  const canWrite = !isReadOnly(user);
  const canRemove = canManageBoard(user); // coordinators and admins

  const [search, setSearch] = useState("");
  const debounced = useDebouncedValue(search);
  const [filter, setFilter] = useState<Filter>("all");
  const [editing, setEditing] = useState<StockItem | null | undefined>(
    undefined,
  ); // null = new
  const [moving, setMoving] = useState<{
    item: StockItem;
    kind: MovementKind;
  } | null>(null);
  const [history, setHistory] = useState<StockItem | null>(null);

  const { data: summary } = useQuery({
    queryKey: ["stock-summary"],
    queryFn: query(InventoryRoutes.summary),
  });
  const { data: items, isLoading } = useQuery({
    queryKey: ["stock-items", debounced, filter],
    queryFn: query(InventoryRoutes.list, {
      queryParams: {
        search: debounced,
        ...(filter === "all" ? {} : { status: filter }),
      },
    }),
  });

  const refresh = () => {
    queryClient.invalidateQueries({ queryKey: ["stock-items"] });
    queryClient.invalidateQueries({ queryKey: ["stock-summary"] });
  };
  const failed = (e: Error) =>
    Notification.Error({
      msg:
        e instanceof HTTPError && typeof e.cause?.message === "string"
          ? e.cause.message
          : t("stock_save_failed"),
    });

  const save = useMutation({
    mutationFn: (req: ItemRequest) =>
      editing
        ? callApi(InventoryRoutes.update, {
            pathParams: { id: editing.id },
            body: req,
            silent: true,
          })
        : callApi(InventoryRoutes.create, { body: req, silent: true }),
    onSuccess: () => {
      setEditing(undefined);
      refresh();
    },
    onError: failed,
  });
  const move = useMutation({
    mutationFn: (v: { quantity: number; note: string }) =>
      callApi(InventoryRoutes.move, {
        pathParams: { id: moving!.item.id },
        body: { kind: moving!.kind, quantity: v.quantity, note: v.note },
        silent: true,
      }),
    onSuccess: (it) => {
      Notification.Success({
        msg: t("stock_updated", { name: it.name, count: it.quantity }),
      });
      setMoving(null);
      refresh();
    },
    onError: failed,
  });
  const archive = useMutation({
    mutationFn: (id: string) =>
      callApi(InventoryRoutes.archive, { pathParams: { id }, silent: true }),
    onSuccess: refresh,
    onError: failed,
  });

  const filters: { key: Filter; label: string; count?: number }[] = [
    { key: "all", label: t("all"), count: summary?.items },
    {
      key: "attention",
      label: t("needs_attention"),
      count: summary ? summary.low + summary.out : undefined,
    },
    { key: "out", label: t("out_of_stock"), count: summary?.out },
  ];

  return (
    <Page
      title={t("inventory")}
      options={
        canWrite && (
          <Button variant="primary" onClick={() => setEditing(null)}>
            <Plus className="mr-2 size-4" aria-hidden />
            {t("new_item")}
          </Button>
        )
      }
    >
      <p className="mb-4 max-w-2xl text-sm text-gray-600">
        {t("inventory_help")}
      </p>

      {/* Filters */}
      <div className="mb-4 flex flex-wrap items-center gap-3">
        <div className="relative w-full sm:w-72">
          <Search
            className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-gray-400"
            aria-hidden
          />
          <input
            aria-label={t("search_items")}
            placeholder={t("search_items")}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="h-10 w-full rounded-lg border border-gray-300 pl-9 pr-3 text-sm shadow-sm focus:border-primary-500 focus:ring-primary-500"
          />
        </div>
        <div className="flex flex-wrap gap-2" role="tablist">
          {filters.map((f) => (
            <button
              key={f.key}
              role="tab"
              aria-selected={filter === f.key}
              onClick={() => setFilter(f.key)}
              className={cn(
                "flex items-center gap-2 rounded-full px-3 py-1.5 text-sm ring-1 ring-inset transition",
                filter === f.key
                  ? "bg-primary-600 text-white ring-primary-600"
                  : "bg-white text-gray-700 ring-gray-300 hover:bg-gray-50",
              )}
            >
              {f.key !== "all" && (
                <TriangleAlert className="size-3.5" aria-hidden />
              )}
              {f.label}
              {f.count !== undefined && (
                <span
                  className={cn(
                    "rounded-full px-1.5 text-xs",
                    filter === f.key ? "bg-white/20" : "bg-gray-100",
                  )}
                >
                  {f.count}
                </span>
              )}
            </button>
          ))}
        </div>
      </div>

      {isLoading ? (
        <p className="text-sm text-gray-500">{t("loading")}</p>
      ) : !items?.length ? (
        <div className="flex flex-col items-center gap-2 rounded-xl border border-dashed border-gray-300 bg-white p-10 text-center">
          <PackageOpen className="size-8 text-gray-400" aria-hidden />
          <p className="text-sm text-gray-600">{t("no_items")}</p>
        </div>
      ) : (
        <div className="flex flex-col gap-6">
          {groupByCategory(items, t("other")).map(([category, group]) => (
            <section
              key={category}
              className="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-sm"
            >
              <h2 className="border-b border-gray-100 bg-gray-50 px-4 py-2 text-xs font-semibold uppercase tracking-wide text-gray-500">
                {category}
              </h2>
              <ul className="divide-y divide-gray-100">
                {group.map((it) => (
                  <li
                    key={it.id}
                    className="flex flex-wrap items-center gap-x-4 gap-y-2 px-4 py-3"
                  >
                    <div className="min-w-40 flex-1">
                      <p className="font-medium text-gray-900">{it.name}</p>
                      <p className="text-xs text-gray-500">
                        {t("alert_at_n", {
                          count: it.min_quantity,
                          unit: it.unit,
                        })}
                      </p>
                    </div>
                    <div className="w-48">
                      <div className="flex items-baseline justify-between text-sm">
                        <span className="font-semibold text-gray-900">
                          {it.quantity}{" "}
                          <span className="font-normal text-gray-500">
                            {it.unit}
                          </span>
                        </span>
                        <span
                          className={cn(
                            "rounded-full px-2 py-0.5 text-xs font-medium ring-1 ring-inset",
                            STATUS_STYLES[it.status].badge,
                          )}
                        >
                          {t(`stock_status__${it.status}`)}
                        </span>
                      </div>
                      <div className="mt-1 h-1.5 rounded-full bg-gray-100">
                        <div
                          className={cn(
                            "h-1.5 rounded-full",
                            STATUS_STYLES[it.status].bar,
                          )}
                          style={{
                            width: `${stockPercent(it.quantity, it.min_quantity)}%`,
                          }}
                        />
                      </div>
                    </div>
                    <div className="flex items-center gap-1">
                      {canWrite && (
                        <>
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => setMoving({ item: it, kind: "use" })}
                            disabled={it.quantity <= 0}
                          >
                            <Minus className="mr-1 size-3.5" aria-hidden />
                            {t("stock_action__use")}
                          </Button>
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() =>
                              setMoving({ item: it, kind: "restock" })
                            }
                          >
                            <Plus className="mr-1 size-3.5" aria-hidden />
                            {t("stock_action__restock")}
                          </Button>
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={() =>
                              setMoving({ item: it, kind: "adjust" })
                            }
                          >
                            {t("stock_action__adjust")}
                          </Button>
                        </>
                      )}
                      <Button
                        size="icon"
                        variant="ghost"
                        aria-label={t("stock_history")}
                        onClick={() => setHistory(it)}
                      >
                        <History className="size-4" aria-hidden />
                      </Button>
                      {canWrite && (
                        <Button
                          size="icon"
                          variant="ghost"
                          aria-label={t("edit_item")}
                          onClick={() => setEditing(it)}
                        >
                          <Pencil className="size-4" aria-hidden />
                        </Button>
                      )}
                      {canRemove && (
                        <Button
                          size="icon"
                          variant="ghost"
                          aria-label={t("remove_item")}
                          onClick={() => {
                            if (
                              window.confirm(
                                t("confirm_remove_item", { name: it.name }),
                              )
                            )
                              archive.mutate(it.id);
                          }}
                        >
                          <Trash2
                            className="size-4 text-rose-600"
                            aria-hidden
                          />
                        </Button>
                      )}
                    </div>
                  </li>
                ))}
              </ul>
            </section>
          ))}
        </div>
      )}

      <ItemDialog
        open={editing !== undefined}
        item={editing ?? null}
        pending={save.isPending}
        onClose={() => setEditing(undefined)}
        onSave={(req) => save.mutate(req)}
      />
      <MovementDialog
        item={moving?.item ?? null}
        kind={moving?.kind ?? "use"}
        pending={move.isPending}
        onClose={() => setMoving(null)}
        onSave={(quantity, note) => move.mutate({ quantity, note })}
      />
      <HistoryDialog item={history} onClose={() => setHistory(null)} />
    </Page>
  );
}
