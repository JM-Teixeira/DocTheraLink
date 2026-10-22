import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import dayjs from "dayjs";
import {
  CircleCheck,
  Download,
  History,
  LayoutGrid,
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
import InventoryAudit from "@/components/Inventory/InventoryAudit";
import {
  HistoryDialog,
  ItemDialog,
  MovementDialog,
  SessionRulesDialog,
} from "@/components/Inventory/InventoryDialogs";
import InventoryRoutes from "@/components/Inventory/api";
import {
  STATUS_STYLES,
  actionsFor,
  groupByCategory,
  reorderCsv,
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
import query, { callApi } from "@/Utils/request/query";
import { HTTPError } from "@/Utils/request/types";
import { canManageBoard, isReadOnly } from "@/Utils/roles";

type Filter = "all" | "attention" | "out";
type Tab = "stock" | "reorder" | "audit";

const ACTION_ICONS: Partial<Record<MovementKind, typeof Plus>> = {
  use: Minus,
  open: PackageOpen,
  finish: CircleCheck,
  restock: Plus,
};

/**
 * The clinic's consumables. Single-use items are used; containers (bottles,
 * packs, rolls) are opened and later finished, so it is known who opened
 * what and since when. Coordinators also see the reorder list and the audit.
 */
export default function InventoryPage() {
  const { t } = useTranslation();
  const user = useAuthUser();
  const queryClient = useQueryClient();
  const canWrite = !isReadOnly(user);
  const manager = canManageBoard(user); // coordinators and admins

  const [tab, setTab] = useState<Tab>("stock");
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
  const [rules, setRules] = useState<StockItem | null>(null);

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
    queryClient.invalidateQueries({ queryKey: ["stock-reorder"] });
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
  const tabs: Tab[] = manager
    ? ["stock", "reorder", "audit"]
    : ["stock", "reorder"];

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
      <p className="mb-4 max-w-3xl text-sm text-gray-600">
        {t("inventory_help_kinds")}
      </p>
      <div
        className="mb-4 flex w-fit rounded-lg bg-gray-100 p-0.5"
        role="tablist"
      >
        {tabs.map((k) => (
          <button
            key={k}
            role="tab"
            aria-selected={tab === k}
            onClick={() => setTab(k)}
            className={cn(
              "rounded-md px-3 py-1.5 text-sm",
              tab === k ? "bg-white shadow-sm" : "text-gray-600",
            )}
          >
            {t(`inventory_tab__${k}`)}
          </button>
        ))}
      </div>

      {tab === "audit" && <InventoryAudit />}
      {tab === "reorder" && <ReorderList />}
      {tab === "stock" && (
        <>
          {!!summary?.open_expired && (
            <p className="mb-4 flex items-center gap-2 rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-800">
              <TriangleAlert className="size-4" aria-hidden />
              {t("open_expired_banner", { count: summary.open_expired })}
            </p>
          )}
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
            <div className="flex flex-wrap gap-2">
              {filters.map((f) => (
                <button
                  key={f.key}
                  aria-pressed={filter === f.key}
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
                        filter === f.key
                          ? "bg-white text-primary-700"
                          : "bg-gray-100",
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
                      <ItemRow
                        key={it.id}
                        item={it}
                        canWrite={canWrite}
                        manager={manager}
                        onMove={(kind) => setMoving({ item: it, kind })}
                        onHistory={() => setHistory(it)}
                        onEdit={() => setEditing(it)}
                        onRules={() => setRules(it)}
                        onRemove={() => {
                          if (
                            window.confirm(
                              t("confirm_remove_item", { name: it.name }),
                            )
                          )
                            archive.mutate(it.id);
                        }}
                      />
                    ))}
                  </ul>
                </section>
              ))}
            </div>
          )}
        </>
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
      <SessionRulesDialog
        item={rules}
        onClose={() => setRules(null)}
        onSaved={refresh}
      />
    </Page>
  );
}

function ItemRow({
  item: it,
  canWrite,
  manager,
  onMove,
  onHistory,
  onEdit,
  onRules,
  onRemove,
}: {
  item: StockItem;
  canWrite: boolean;
  manager: boolean;
  onMove: (kind: MovementKind) => void;
  onHistory: () => void;
  onEdit: () => void;
  onRules: () => void;
  onRemove: () => void;
}) {
  const { t } = useTranslation();
  const container = it.tracking === "container";
  const ruled = it.session_rules.filter((r) => r.quantity > 0).length;
  return (
    <li className="flex flex-wrap items-center gap-x-4 gap-y-2 px-4 py-3">
      <div className="min-w-48 flex-1">
        <p className="flex flex-wrap items-center gap-2 font-medium text-gray-900">
          {it.name}
          <span className="rounded-full bg-gray-100 px-2 py-0.5 text-xs font-normal text-gray-600">
            {t(`item_kind__${it.tracking}`)}
          </span>
          {(it.per_session > 0 || ruled > 0) && (
            <span
              className="rounded-full bg-primary-50 px-2 py-0.5 text-xs font-normal text-primary-700"
              title={t("used_per_session_help")}
            >
              {it.per_session > 0
                ? t("n_per_session", { count: it.per_session })
                : t("per_session_in_divisions", { count: ruled })}
            </span>
          )}
        </p>
        <p className="text-xs text-gray-500">
          {t("alert_at_n", { count: it.min_quantity, unit: it.unit })}
          {container && it.in_use > 0 && it.oldest_open_at && (
            <span
              className={cn(
                "ml-2",
                it.open_expired ? "font-medium text-rose-700" : "text-gray-600",
              )}
            >
              ·{" "}
              {t("n_open_since", {
                count: it.in_use,
                date: dayjs(it.oldest_open_at).format("D MMM"),
              })}
              {it.open_expired && ` · ${t("discard_now")}`}
            </span>
          )}
        </p>
      </div>
      <div className="w-48">
        <div className="flex items-baseline justify-between text-sm">
          <span className="font-semibold text-gray-900">
            {it.quantity}{" "}
            <span className="font-normal text-gray-500">
              {container ? t("sealed_unit", { unit: it.unit }) : it.unit}
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
            className={cn("h-1.5 rounded-full", STATUS_STYLES[it.status].bar)}
            style={{ width: `${stockPercent(it.quantity, it.min_quantity)}%` }}
          />
        </div>
      </div>
      <div className="flex flex-wrap items-center gap-1">
        {canWrite &&
          actionsFor(it).map((kind) => {
            const Icon = ACTION_ICONS[kind];
            const disabled =
              ((kind === "use" || kind === "open") && it.quantity <= 0) ||
              (kind === "finish" && it.in_use <= 0);
            return (
              <Button
                key={kind}
                size="sm"
                variant={kind === "adjust" ? "ghost" : "outline"}
                disabled={disabled}
                onClick={() => onMove(kind)}
              >
                {Icon && <Icon className="mr-1 size-3.5" aria-hidden />}
                {t(`stock_action__${kind}`)}
              </Button>
            );
          })}
        <Button
          size="icon"
          variant="ghost"
          aria-label={t("stock_history")}
          onClick={onHistory}
        >
          <History className="size-4" aria-hidden />
        </Button>
        {manager && !container && (
          <Button
            size="icon"
            variant="ghost"
            aria-label={t("per_session_by_division")}
            onClick={onRules}
          >
            <LayoutGrid className="size-4" aria-hidden />
          </Button>
        )}
        {canWrite && (
          <Button
            size="icon"
            variant="ghost"
            aria-label={t("edit_item")}
            onClick={onEdit}
          >
            <Pencil className="size-4" aria-hidden />
          </Button>
        )}
        {manager && (
          <Button
            size="icon"
            variant="ghost"
            aria-label={t("remove_item")}
            onClick={onRemove}
          >
            <Trash2 className="size-4 text-rose-600" aria-hidden />
          </Button>
        )}
      </div>
    </li>
  );
}

/** What to order and how much, to send to the supplier (CSV). */
function ReorderList() {
  const { t } = useTranslation();
  const { data: items, isLoading } = useQuery({
    queryKey: ["stock-reorder"],
    queryFn: query(InventoryRoutes.reorder),
  });
  const download = () => {
    const csv = reorderCsv(items ?? [], [
      t("item_name"),
      t("category"),
      t("in_stock"),
      t("to_order"),
      t("unit"),
    ]);
    const url = URL.createObjectURL(
      new Blob(["﻿" + csv], { type: "text/csv;charset=utf-8" }),
    );
    const link = document.createElement("a");
    link.href = url;
    link.download = `encomenda-${dayjs().format("YYYY-MM-DD")}.csv`;
    link.click();
    setTimeout(() => URL.revokeObjectURL(url), 1000);
  };
  return (
    <section className="rounded-xl border border-gray-200 bg-white shadow-sm">
      <header className="flex flex-wrap items-center justify-between gap-2 border-b border-gray-100 px-4 py-3">
        <p className="text-sm text-gray-600">{t("reorder_help")}</p>
        <Button
          size="sm"
          variant="outline"
          disabled={!items?.length}
          onClick={download}
        >
          <Download className="mr-1 size-4" aria-hidden />
          {t("download_csv")}
        </Button>
      </header>
      {isLoading ? (
        <p className="p-4 text-sm text-gray-500">{t("loading")}</p>
      ) : !items?.length ? (
        <p className="p-4 text-sm text-gray-500">{t("nothing_to_reorder")}</p>
      ) : (
        <table className="w-full text-sm">
          <thead className="bg-gray-50 text-left text-xs text-gray-500">
            <tr>
              <th className="p-3 font-medium">{t("item_name")}</th>
              <th className="p-3 font-medium">{t("category")}</th>
              <th className="p-3 text-right font-medium">{t("in_stock")}</th>
              <th className="p-3 text-right font-medium">{t("to_order")}</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100">
            {items.map((it) => (
              <tr key={it.id}>
                <td className="p-3 font-medium text-gray-900">{it.name}</td>
                <td className="p-3 text-gray-600">{it.category}</td>
                <td
                  className={cn(
                    "p-3 text-right",
                    it.status === "out"
                      ? "font-semibold text-rose-700"
                      : "text-amber-700",
                  )}
                >
                  {it.quantity} {it.unit}
                </td>
                <td className="p-3 text-right font-semibold text-gray-900">
                  {it.reorder_quantity} {it.unit}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </section>
  );
}
