import { useQuery } from "@tanstack/react-query";
import dayjs from "dayjs";
import { Bot } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";

import { cn } from "@/lib/utils";

import InventoryRoutes from "@/components/Inventory/api";
import {
  auditTotals,
  movementLabel,
} from "@/components/Inventory/inventoryUtils";
import { MovementKind } from "@/components/Inventory/types";

import query from "@/Utils/request/query";

const fieldClass =
  "h-9 rounded-lg border border-gray-300 px-2 text-sm shadow-sm focus:border-primary-500 focus:ring-primary-500";
const KINDS: MovementKind[] = ["use", "open", "finish", "restock", "adjust"];

/**
 * Audit of the clinic's stock (coordinators and admins): who used, opened,
 * restocked or counted what and when, with totals per person. Automatic use
 * by sessions on the time board is marked.
 */
export default function InventoryAudit() {
  const { t } = useTranslation();
  const [from, setFrom] = useState(
    dayjs().subtract(30, "day").format("YYYY-MM-DD"),
  );
  const [to, setTo] = useState(dayjs().format("YYYY-MM-DD"));
  const [user, setUser] = useState("");
  const [item, setItem] = useState("");
  const [kind, setKind] = useState("");

  const { data: items } = useQuery({
    queryKey: ["stock-items", "", "all"],
    queryFn: query(InventoryRoutes.list),
  });
  // Everyone who did something in the period (for the person filter)
  const { data: period } = useQuery({
    queryKey: ["stock-audit", from, to],
    queryFn: query(InventoryRoutes.audit, { queryParams: { from, to } }),
  });
  const { data: moves, isLoading } = useQuery({
    queryKey: ["stock-audit", from, to, user, item, kind],
    queryFn: query(InventoryRoutes.audit, {
      queryParams: {
        from,
        to,
        ...(user && { user }),
        ...(item && { item }),
        ...(kind && { kind }),
      },
    }),
  });
  const people = new Map(
    (period ?? [])
      .filter((m) => m.user)
      .map((m) => [m.user!.id, `${m.user!.first_name} ${m.user!.last_name}`]),
  );
  const totals = auditTotals(moves ?? []);

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-wrap items-end gap-3 rounded-xl border border-gray-200 bg-white p-3 text-sm shadow-sm">
        <label className="flex flex-col gap-1 text-xs text-gray-600">
          {t("from")}
          <input
            type="date"
            value={from}
            onChange={(e) => setFrom(e.target.value)}
            className={fieldClass}
          />
        </label>
        <label className="flex flex-col gap-1 text-xs text-gray-600">
          {t("until")}
          <input
            type="date"
            value={to}
            onChange={(e) => setTo(e.target.value)}
            className={fieldClass}
          />
        </label>
        <label className="flex flex-col gap-1 text-xs text-gray-600">
          {t("person")}
          <select
            value={user}
            onChange={(e) => setUser(e.target.value)}
            className={fieldClass}
          >
            <option value="">{t("everyone")}</option>
            {[...people].map(([id, name]) => (
              <option key={id} value={id}>
                {name}
              </option>
            ))}
          </select>
        </label>
        <label className="flex flex-col gap-1 text-xs text-gray-600">
          {t("item_name")}
          <select
            value={item}
            onChange={(e) => setItem(e.target.value)}
            className={fieldClass}
          >
            <option value="">{t("all")}</option>
            {items?.map((it) => (
              <option key={it.id} value={it.id}>
                {it.name}
              </option>
            ))}
          </select>
        </label>
        <label className="flex flex-col gap-1 text-xs text-gray-600">
          {t("action")}
          <select
            value={kind}
            onChange={(e) => setKind(e.target.value)}
            className={fieldClass}
          >
            <option value="">{t("all")}</option>
            {KINDS.map((k) => (
              <option key={k} value={k}>
                {t(`stock_kind__${k}`)}
              </option>
            ))}
          </select>
        </label>
      </div>

      {!!totals.length && (
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {totals.map((p) => (
            <section
              key={p.name}
              className="rounded-xl border border-gray-200 bg-white p-4 shadow-sm"
            >
              <h3 className="mb-2 text-base font-semibold text-gray-900">
                {p.name}
              </h3>
              <ul className="space-y-1 text-sm text-gray-700">
                {[...p.items.values()].map((it) => (
                  <li key={it.name} className="flex justify-between gap-2">
                    <span className="truncate">{it.name}</span>
                    <span className="shrink-0 text-gray-500">
                      {[
                        it.used > 0 &&
                          t("n_used", { count: it.used, unit: it.unit }),
                        it.opened > 0 && t("n_opened", { count: it.opened }),
                        it.finished > 0 &&
                          t("n_finished", { count: it.finished }),
                      ]
                        .filter(Boolean)
                        .join(" · ")}
                    </span>
                  </li>
                ))}
              </ul>
            </section>
          ))}
        </div>
      )}

      <div className="overflow-x-auto rounded-xl border border-gray-200 bg-white shadow-sm">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 text-left text-xs text-gray-500">
            <tr>
              <th className="p-3 font-medium">{t("when")}</th>
              <th className="p-3 font-medium">{t("person")}</th>
              <th className="p-3 font-medium">{t("item_name")}</th>
              <th className="p-3 font-medium">{t("action")}</th>
              <th className="p-3 text-right font-medium">{t("quantity")}</th>
              <th className="p-3 font-medium">{t("note")}</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100">
            {isLoading ? (
              <tr>
                <td colSpan={6} className="p-4 text-gray-500">
                  {t("loading")}
                </td>
              </tr>
            ) : !moves?.length ? (
              <tr>
                <td colSpan={6} className="p-4 text-gray-500">
                  {t("no_movements")}
                </td>
              </tr>
            ) : (
              moves.map((m) => (
                <tr key={m.id}>
                  <td className="whitespace-nowrap p-3 text-gray-600">
                    {dayjs(m.created_date).format("D MMM, HH:mm")}
                  </td>
                  <td className="p-3">
                    {m.user ? `${m.user.first_name} ${m.user.last_name}` : "—"}
                  </td>
                  <td className="p-3">{m.item?.name}</td>
                  <td className="p-3">
                    <span className="flex items-center gap-1.5">
                      {t(`stock_kind__${m.kind}`)}
                      {m.appointment_id && (
                        <span
                          className="flex items-center gap-1 rounded-full bg-primary-50 px-1.5 text-xs text-primary-700"
                          title={t("automatic_session_use")}
                        >
                          <Bot className="size-3" aria-hidden />
                          {t("session")}
                        </span>
                      )}
                    </span>
                  </td>
                  <td
                    className={cn(
                      "p-3 text-right font-mono",
                      m.kind === "adjust" || m.kind === "finish"
                        ? "text-gray-600"
                        : m.delta > 0
                          ? "text-emerald-700"
                          : "text-rose-700",
                    )}
                  >
                    {movementLabel(m)}
                  </td>
                  <td className="p-3 text-gray-600">
                    {m.note === "session" ? "" : m.note}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
