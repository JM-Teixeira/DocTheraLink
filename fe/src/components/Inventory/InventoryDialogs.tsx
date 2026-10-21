import { useQuery } from "@tanstack/react-query";
import dayjs from "dayjs";
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
} from "@/components/Inventory/types";

import query from "@/Utils/request/query";

const fieldClass =
  "mt-1 h-10 w-full rounded-lg border border-gray-300 px-3 text-sm shadow-sm focus:border-primary-500 focus:ring-primary-500";

/** Creates (item = null) or edits an item. */
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
  const [form, setForm] = useState({
    name: "",
    category: "",
    unit: "units",
    min: "0",
    perSession: "0",
    quantity: "0",
  });
  useEffect(() => {
    if (open)
      setForm({
        name: item?.name ?? "",
        category: item?.category ?? "",
        unit: item?.unit ?? "units",
        min: String(item?.min_quantity ?? 0),
        perSession: String(item?.per_session ?? 0),
        quantity: "0",
      });
  }, [open, item]);
  const set =
    (k: keyof typeof form) => (e: React.ChangeEvent<HTMLInputElement>) =>
      setForm((f) => ({ ...f, [k]: e.target.value }));
  const valid = form.name.trim().length >= 2 && Number(form.min) >= 0;

  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="sm:max-w-md">
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
              min_quantity: Number(form.min),
              per_session: Number(form.perSession) || 0,
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

/** Records use, restock or a physical count of one item. */
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
  const problem = item ? checkMovement(kind, Number(qty), item.quantity) : null;

  return (
    <Dialog open={!!item} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="sm:max-w-sm">
        <DialogHeader>
          <DialogTitle>{t(`stock_action__${kind}`)}</DialogTitle>
          <DialogDescription>
            {item?.name} ·{" "}
            {t("in_stock_n", { count: item?.quantity ?? 0, unit: item?.unit })}
          </DialogDescription>
        </DialogHeader>
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

/** Latest movements of one item. */
export function HistoryDialog({
  item,
  onClose,
}: {
  item: StockItem | null;
  onClose: () => void;
}) {
  const { t } = useTranslation();
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
                    m.kind === "adjust"
                      ? "text-gray-600"
                      : m.delta > 0
                        ? "text-emerald-700"
                        : "text-rose-700",
                  )}
                >
                  {movementLabel(m.kind, m.delta, m.quantity_after)}
                </span>
                <div className="min-w-0 flex-1">
                  <p className="text-gray-900">
                    {t(`stock_kind__${m.kind}`)}
                    {m.note && (
                      <span className="text-gray-500"> · {m.note}</span>
                    )}
                  </p>
                  <p className="text-xs text-gray-500">
                    {m.user
                      ? `${m.user.first_name} ${m.user.last_name} · `
                      : ""}
                    {dayjs(m.created_date).format("D MMM YYYY, HH:mm")}
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
