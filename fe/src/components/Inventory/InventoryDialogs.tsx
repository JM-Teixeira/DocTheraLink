import { useMutation, useQuery } from "@tanstack/react-query";
import { Package, PackageOpen } from "lucide-react";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";

import { cn } from "@/lib/utils";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

import InventoryRoutes from "@/components/Inventory/api";
import {
  checkMovement,
  movementLabel,
} from "@/components/Inventory/inventoryUtils";
import {
  ItemRequest,
  MovementKind,
  StockItem,
  Tracking,
} from "@/components/Inventory/types";
import BoardRoutes from "@/components/TimeBoard/api";
import { DIVISION_COLORS } from "@/components/TimeBoard/boardUtils";

import useDateFormat from "@/hooks/useDateFormat";

import * as Notification from "@/Utils/Notifications";
import query, { callApi } from "@/Utils/request/query";

const fieldClass =
  "mt-1 h-10 w-full rounded-lg border border-gray-300 px-3 text-sm shadow-sm focus:border-primary-500 focus:ring-primary-500";

type ItemForm = {
  name: string;
  category: string;
  unit: string;
  tracking: Tracking;
  min: string;
  target: string;
  perSession: string;
  openLife: string;
  quantity: string;
};

const EMPTY: ItemForm = {
  name: "",
  category: "",
  unit: "units",
  tracking: "unit",
  min: "0",
  target: "0",
  perSession: "0",
  openLife: "0",
  quantity: "0",
};

/** Creates (item = null) or edits an item, including its kind. */
export function ItemDialog({
  open,
  item,
  pending,
  onClose,
  onSave,
}: {
  open: boolean;
  item: StockItem | null;
  pending: boolean;
  onClose: () => void;
  onSave: (req: ItemRequest) => void;
}) {
  const { t } = useTranslation();
  const [form, setForm] = useState<ItemForm>(EMPTY);
  useEffect(() => {
    if (!open) return;
    setForm(
      item
        ? {
            name: item.name,
            category: item.category,
            unit: item.unit,
            tracking: item.tracking,
            min: String(item.min_quantity),
            target: String(item.target_quantity),
            perSession: String(item.per_session),
            openLife: String(item.open_life_days),
            quantity: "0",
          }
        : EMPTY,
    );
  }, [open, item]);
  const set = (k: keyof ItemForm) => (e: React.ChangeEvent<HTMLInputElement>) =>
    setForm((f) => ({ ...f, [k]: e.target.value }));
  const valid = form.name.trim().length >= 2 && Number(form.min) >= 0;
  const container = form.tracking === "container";

  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{item ? t("edit_item") : t("new_item")}</DialogTitle>
        </DialogHeader>
        <form
          className="grid gap-3"
          onSubmit={(e) => {
            e.preventDefault();
            if (!valid) return;
            onSave({
              name: form.name.trim(),
              category: form.category.trim(),
              unit: form.unit.trim(),
              tracking: form.tracking,
              min_quantity: Number(form.min),
              target_quantity: Number(form.target) || 0,
              per_session: container ? 0 : Number(form.perSession) || 0,
              open_life_days: container ? Number(form.openLife) || 0 : 0,
              ...(item ? {} : { quantity: Number(form.quantity) }),
            });
          }}
        >
          <label className="text-sm font-medium text-gray-700">
            {t("item_name")} *
            <input
              value={form.name}
              onChange={set("name")}
              className={fieldClass}
              autoFocus
            />
          </label>

          <fieldset>
            <legend className="text-sm font-medium text-gray-700">
              {t("item_kind")}
            </legend>
            <div className="mt-1 grid gap-2 sm:grid-cols-2">
              {(["unit", "container"] as const).map((k) => (
                <label
                  key={k}
                  className={cn(
                    "flex cursor-pointer gap-2 rounded-lg p-3 text-sm ring-1 ring-inset",
                    form.tracking === k
                      ? "bg-primary-50 ring-2 ring-primary-500"
                      : "ring-gray-200 hover:bg-gray-50",
                  )}
                >
                  <input
                    type="radio"
                    name="tracking"
                    className="sr-only"
                    checked={form.tracking === k}
                    onChange={() => setForm((f) => ({ ...f, tracking: k }))}
                  />
                  {k === "unit" ? (
                    <Package
                      className="mt-0.5 size-4 shrink-0 text-primary-600"
                      aria-hidden
                    />
                  ) : (
                    <PackageOpen
                      className="mt-0.5 size-4 shrink-0 text-primary-600"
                      aria-hidden
                    />
                  )}
                  <span>
                    <span className="block font-medium text-gray-900">
                      {t(`item_kind__${k}`)}
                    </span>
                    <span className="block text-xs text-gray-500">
                      {t(`item_kind__${k}_help`)}
                    </span>
                  </span>
                </label>
              ))}
            </div>
          </fieldset>

          <div className="grid grid-cols-2 gap-3">
            <label className="text-sm font-medium text-gray-700">
              {t("category")}
              <input
                value={form.category}
                onChange={set("category")}
                list="stock-categories"
                className={fieldClass}
              />
              <datalist id="stock-categories">
                <option value="PPE" />
                <option value="Hygiene" />
                <option value="Treatment" />
              </datalist>
            </label>
            <label className="text-sm font-medium text-gray-700">
              {t("unit")}
              <input
                value={form.unit}
                onChange={set("unit")}
                className={fieldClass}
              />
            </label>
            <label className="text-sm font-medium text-gray-700">
              {t("alert_at_or_below")}
              <input
                type="number"
                min={0}
                value={form.min}
                onChange={set("min")}
                className={fieldClass}
              />
            </label>
            <label className="text-sm font-medium text-gray-700">
              {t("reorder_up_to")}
              <input
                type="number"
                min={0}
                value={form.target}
                onChange={set("target")}
                className={fieldClass}
              />
            </label>
            {container ? (
              <label className="col-span-2 text-sm font-medium text-gray-700">
                {t("discard_after_days")}
                <input
                  type="number"
                  min={0}
                  max={3650}
                  value={form.openLife}
                  onChange={set("openLife")}
                  className={fieldClass}
                />
                <span className="mt-1 block text-xs font-normal text-gray-500">
                  {t("discard_after_days_help")}
                </span>
              </label>
            ) : (
              <label className="col-span-2 text-sm font-medium text-gray-700">
                {t("used_per_session")}
                <input
                  type="number"
                  min={0}
                  max={100}
                  value={form.perSession}
                  onChange={set("perSession")}
                  className={fieldClass}
                />
                <span className="mt-1 block text-xs font-normal text-gray-500">
                  {t("used_per_session_help")}
                </span>
              </label>
            )}
            {!item && (
              <label className="text-sm font-medium text-gray-700">
                {t("in_stock_now")}
                <input
                  type="number"
                  min={0}
                  value={form.quantity}
                  onChange={set("quantity")}
                  className={fieldClass}
                />
              </label>
            )}
          </div>
          <DialogFooter className="mt-2 gap-2">
            <Button type="button" variant="outline" onClick={onClose}>
              {t("cancel")}
            </Button>
            <Button
              type="submit"
              variant="primary"
              disabled={!valid || pending}
            >
              {t("save")}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

/** Records use, opening, finishing, restock or a physical count of one item. */
export function MovementDialog({
  item,
  kind,
  pending,
  onClose,
  onSave,
}: {
  item: StockItem | null;
  kind: MovementKind;
  pending: boolean;
  onClose: () => void;
  onSave: (quantity: number, note: string) => void;
}) {
  const { t } = useTranslation();
  const [qty, setQty] = useState("1");
  const [note, setNote] = useState("");
  useEffect(() => {
    if (item) {
      setQty(kind === "adjust" ? String(item.quantity) : "1");
      setNote("");
    }
  }, [item, kind]);
  const problem = item ? checkMovement(kind, Number(qty), item) : null;

  return (
    <Dialog open={!!item} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="sm:max-w-sm">
        <DialogHeader>
          <DialogTitle>{t(`stock_action__${kind}`)}</DialogTitle>
          <DialogDescription>
            {item?.name} ·{" "}
            {t("in_stock_n", { count: item?.quantity ?? 0, unit: item?.unit })}
            {item?.tracking === "container" &&
              ` · ${t("n_open", { count: item.in_use })}`}
          </DialogDescription>
        </DialogHeader>
        <p className="text-xs text-gray-500">
          {t(`stock_action_help__${kind}`)}
        </p>
        <form
          className="grid gap-3"
          onSubmit={(e) => {
            e.preventDefault();
            if (!problem) onSave(Number(qty), note.trim());
          }}
        >
          <label className="text-sm font-medium text-gray-700">
            {kind === "adjust" ? t("counted_quantity") : t("quantity")}
            <input
              type="number"
              min={kind === "adjust" ? 0 : 1}
              value={qty}
              onChange={(e) => setQty(e.target.value)}
              className={fieldClass}
              autoFocus
            />
          </label>
          {problem && <p className="text-xs text-rose-600">{t(problem)}</p>}
          <label className="text-sm font-medium text-gray-700">
            {t("note_optional")}
            <input
              value={note}
              maxLength={300}
              onChange={(e) => setNote(e.target.value)}
              className={fieldClass}
            />
          </label>
          <DialogFooter className="mt-2 gap-2">
            <Button type="button" variant="outline" onClick={onClose}>
              {t("cancel")}
            </Button>
            <Button
              type="submit"
              variant="primary"
              disabled={!!problem || pending}
            >
              {t("save")}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

/** Per-division session use of a single-use item (coordinators). */
export function SessionRulesDialog({
  item,
  onClose,
  onSaved,
}: {
  item: StockItem | null;
  onClose: () => void;
  onSaved: () => void;
}) {
  const { t } = useTranslation();
  const { data: divisions } = useQuery({
    queryKey: ["divisions"],
    queryFn: query(BoardRoutes.divisions),
    enabled: !!item,
  });
  const [values, setValues] = useState<Record<string, string>>({});
  useEffect(() => {
    if (!item) return;
    setValues(
      Object.fromEntries(
        item.session_rules.map((r) => [r.division_id, String(r.quantity)]),
      ),
    );
  }, [item]);
  const save = useMutation({
    mutationFn: () =>
      callApi(InventoryRoutes.sessionRules, {
        pathParams: { id: item!.id },
        body: {
          rules: Object.entries(values)
            .filter(([, v]) => v !== "")
            .map(([division_id, v]) => ({ division_id, quantity: Number(v) })),
        },
        silent: true,
      }),
    onSuccess: () => {
      onSaved();
      onClose();
    },
    onError: () => Notification.Error({ msg: t("stock_save_failed") }),
  });

  return (
    <Dialog open={!!item} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t("per_session_by_division")}</DialogTitle>
          <DialogDescription>
            {t("per_session_by_division_help", {
              name: item?.name,
              count: item?.per_session ?? 0,
            })}
          </DialogDescription>
        </DialogHeader>
        <ul className="divide-y divide-gray-100 rounded-lg border border-gray-200">
          {divisions?.map((d) => (
            <li
              key={d.id}
              className="flex items-center gap-3 px-3 py-2 text-sm"
            >
              <span
                className={cn(
                  "size-2.5 rounded-full",
                  DIVISION_COLORS[d.color].dot,
                )}
              />
              <span className="flex-1">{d.name}</span>
              <input
                type="number"
                min={0}
                max={100}
                aria-label={d.name}
                placeholder={t("default_n", { count: item?.per_session ?? 0 })}
                value={values[d.id] ?? ""}
                onChange={(e) =>
                  setValues((v) => ({ ...v, [d.id]: e.target.value }))
                }
                className="h-9 w-28 rounded-lg border border-gray-300 px-2 text-sm"
              />
            </li>
          ))}
        </ul>
        <DialogFooter className="gap-2">
          <Button type="button" variant="outline" onClick={onClose}>
            {t("cancel")}
          </Button>
          <Button
            type="button"
            variant="primary"
            disabled={save.isPending}
            onClick={() => save.mutate()}
          >
            {t("save")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

/** Latest movements of one item. */
export function HistoryDialog({
  item,
  onClose,
}: {
  item: StockItem | null;
  onClose: () => void;
}) {
  const { t } = useTranslation();
  const { fmtTime } = useDateFormat();
  const { data: moves, isLoading } = useQuery({
    queryKey: ["stock-history", item?.id],
    queryFn: query(InventoryRoutes.history, {
      pathParams: { id: item?.id ?? "" },
    }),
    enabled: !!item,
  });
  return (
    <Dialog open={!!item} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{t("stock_history")}</DialogTitle>
          <DialogDescription>{item?.name}</DialogDescription>
        </DialogHeader>
        {isLoading ? (
          <p className="text-sm text-gray-500">{t("loading")}</p>
        ) : moves?.length ? (
          <ul className="max-h-96 divide-y divide-gray-100 overflow-auto rounded-lg border border-gray-200">
            {moves.map((m) => (
              <li
                key={m.id}
                className="flex items-center gap-3 px-3 py-2 text-sm"
              >
                <span
                  className={cn(
                    "w-14 shrink-0 text-right font-mono font-semibold",
                    m.kind === "adjust" || m.kind === "finish"
                      ? "text-gray-600"
                      : m.delta > 0
                        ? "text-emerald-700"
                        : "text-rose-700",
                  )}
                >
                  {movementLabel(m)}
                </span>
                <div className="min-w-0 flex-1">
                  <p className="text-gray-900">
                    {t(`stock_kind__${m.kind}`)}
                    {m.note && (
                      <span className="text-gray-500">
                        {" "}
                        · {m.note === "session" ? t("session") : m.note}
                      </span>
                    )}
                  </p>
                  <p className="text-xs text-gray-500">
                    {m.user
                      ? `${m.user.first_name} ${m.user.last_name} · `
                      : ""}
                    {fmtTime(m.created_date)}
                  </p>
                </div>
                <span className="text-xs text-gray-500">
                  {t("stock_after", { count: m.quantity_after })}
                </span>
              </li>
            ))}
          </ul>
        ) : (
          <p className="text-sm text-gray-500">{t("no_movements")}</p>
        )}
      </DialogContent>
    </Dialog>
  );
}
