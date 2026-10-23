import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  ArrowDown,
  ArrowUp,
  ClipboardList,
  Pencil,
  Plus,
  X,
} from "lucide-react";
import { Link } from "raviger";
import { useState } from "react";
import { useTranslation } from "react-i18next";

import { Button } from "@/components/ui/button";

import ReferralRoutes from "@/components/Referral/api";
import { PlanItem, PlanView, TreatmentPlan } from "@/components/Referral/types";

import useDateFormat from "@/hooks/useDateFormat";

import * as Notification from "@/Utils/Notifications";
import query, { callApi } from "@/Utils/request/query";

const cellClass =
  "h-9 w-full rounded-md border border-gray-300 px-2 text-sm focus:border-primary-500 focus:ring-primary-500";

/** Rows being edited: every cell is a string so inputs stay controlled. */
type Row = Record<keyof PlanItem, string>;
const EMPTY_ROW: Row = {
  activity: "",
  minutes: "",
  dose: "",
  equipment: "",
  notes: "",
};

export const toRows = (items: PlanItem[]): Row[] =>
  items.map((i) => ({
    activity: i.activity,
    minutes: i.minutes ? String(i.minutes) : "",
    dose: i.dose ?? "",
    equipment: i.equipment ?? "",
    notes: i.notes ?? "",
  }));

/** Rows to API items; empty rows dropped (the server checks again). */
export const toItems = (rows: Row[]): PlanItem[] =>
  rows
    .filter((r) => r.activity.trim())
    .map((r) => ({
      activity: r.activity.trim(),
      ...(Number(r.minutes) > 0 ? { minutes: Number(r.minutes) } : {}),
      ...(r.dose.trim() ? { dose: r.dose.trim() } : {}),
      ...(r.equipment.trim() ? { equipment: r.equipment.trim() } : {}),
      ...(r.notes.trim() ? { notes: r.notes.trim() } : {}),
    }));

export const totalMinutes = (items: PlanItem[]) =>
  items.reduce((n, i) => n + (i.minutes ?? 0), 0);

/**
 * The referral's treatment plan: what is done in each session, in order, so
 * whoever covers the physio carries on the same way. Editing saves a new
 * version; earlier versions stay visible.
 */
export default function TreatmentPlanCard({
  referralId,
  compact,
}: {
  referralId: string;
  /** Patient page: read-only, links to the referral to change it */
  compact?: boolean;
}) {
  const { t } = useTranslation();
  const { fmtTime } = useDateFormat();
  const queryClient = useQueryClient();
  const key = ["referral-plan", referralId];
  const { data } = useQuery({
    queryKey: key,
    queryFn: query(ReferralRoutes.plan, { pathParams: { id: referralId } }),
  });
  const [rows, setRows] = useState<Row[] | null>(null);
  const [showVersions, setShowVersions] = useState(false);
  const save = useMutation({
    mutationFn: (items: PlanItem[]) =>
      callApi(ReferralRoutes.savePlan, {
        pathParams: { id: referralId },
        body: { items },
      }),
    onSuccess: (v: PlanView) => {
      queryClient.setQueryData(key, v);
      setRows(null);
      Notification.Success({ msg: t("plan_saved") });
    },
  });

  if (!data) return null;
  const current = data.current;
  if (compact && !current) return null;
  const editing = rows !== null;
  const setCell = (i: number, k: keyof Row, v: string) =>
    setRows((rs) => rs!.map((r, j) => (j === i ? { ...r, [k]: v } : r)));
  const move = (i: number, d: -1 | 1) =>
    setRows((rs) => {
      const out = [...rs!];
      [out[i], out[i + d]] = [out[i + d], out[i]];
      return out;
    });

  return (
    <section
      aria-label={t("treatment_plan")}
      className="rounded-xl border border-gray-200 bg-white p-5 shadow-sm"
    >
      <div className="mb-3 flex items-center justify-between gap-2">
        <h3 className="flex items-center gap-2 text-base font-semibold text-gray-900">
          <ClipboardList className="size-4 text-primary-600" aria-hidden />
          {t("treatment_plan")}
        </h3>
        {!compact && data.can_edit && !editing && (
          <Button
            variant="outline"
            size="sm"
            onClick={() =>
              setRows(
                current?.items.length
                  ? toRows(current.items)
                  : [{ ...EMPTY_ROW }],
              )
            }
          >
            <Pencil className="mr-1.5 size-4" aria-hidden />
            {current ? t("edit_plan") : t("create_plan")}
          </Button>
        )}
        {compact && (
          <Link
            href={`/referrals/${referralId}`}
            className="text-sm text-primary-700 hover:underline"
          >
            {t("open_referral")}
          </Link>
        )}
      </div>

      {!editing && (
        <>
          {current ? (
            <PlanSteps plan={current} />
          ) : (
            <p className="text-sm text-gray-500">
              {data.can_edit ? t("plan_empty_can_edit") : t("plan_empty")}
            </p>
          )}
          {current && (
            <p className="mt-3 text-xs text-gray-500">
              {t("plan_updated_by", {
                who: current.author
                  ? `${current.author.first_name} ${current.author.last_name}`
                  : "-",
                when: fmtTime(current.created_date),
              })}
            </p>
          )}
          {!compact && data.versions.length > 1 && (
            <div className="mt-2">
              <Button
                variant="link"
                size="sm"
                className="h-auto p-0"
                aria-expanded={showVersions}
                onClick={() => setShowVersions((v) => !v)}
              >
                {showVersions
                  ? t("hide_plan_versions")
                  : t("show_plan_versions", {
                      count: data.versions.length - 1,
                    })}
              </Button>
              {showVersions && (
                <ol className="mt-2 space-y-3 border-l-2 border-gray-100 pl-3">
                  {data.versions.slice(1).map((v) => (
                    <li key={v.id} className="text-xs text-gray-600">
                      <p className="mb-1 font-medium text-gray-800">
                        {fmtTime(v.created_date)}
                        {v.author &&
                          ` · ${v.author.first_name} ${v.author.last_name}`}
                      </p>
                      <PlanSteps plan={v} small />
                    </li>
                  ))}
                </ol>
              )}
            </div>
          )}
        </>
      )}

      {editing && (
        <form
          onSubmit={(e) => {
            e.preventDefault();
            save.mutate(toItems(rows));
          }}
        >
          <p className="mb-2 text-xs text-gray-500">{t("plan_edit_help")}</p>
          <ol className="space-y-3">
            {rows.map((r, i) => (
              <li key={i} className="rounded-lg bg-gray-50 p-2">
                <div className="flex flex-wrap items-center gap-2">
                  <span className="w-5 text-right text-xs font-semibold text-gray-500">
                    {i + 1}.
                  </span>
                  <input
                    aria-label={`${t("plan_activity")} ${i + 1}`}
                    placeholder={t("plan_activity_placeholder")}
                    value={r.activity}
                    maxLength={120}
                    onChange={(e) => setCell(i, "activity", e.target.value)}
                    className={`${cellClass} min-w-48 flex-[3]`}
                  />
                  <input
                    aria-label={`${t("plan_minutes")} ${i + 1}`}
                    placeholder={t("plan_minutes")}
                    type="number"
                    min={1}
                    max={240}
                    value={r.minutes}
                    onChange={(e) => setCell(i, "minutes", e.target.value)}
                    className={`${cellClass} w-20`}
                  />
                  <input
                    aria-label={`${t("plan_dose")} ${i + 1}`}
                    placeholder={t("plan_dose_placeholder")}
                    value={r.dose}
                    maxLength={40}
                    onChange={(e) => setCell(i, "dose", e.target.value)}
                    className={`${cellClass} w-28`}
                  />
                  <div className="flex">
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      disabled={i === 0}
                      aria-label={`${t("move_up")} ${i + 1}`}
                      onClick={() => move(i, -1)}
                    >
                      <ArrowUp className="size-4" aria-hidden />
                    </Button>
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      disabled={i === rows.length - 1}
                      aria-label={`${t("move_down")} ${i + 1}`}
                      onClick={() => move(i, 1)}
                    >
                      <ArrowDown className="size-4" aria-hidden />
                    </Button>
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      aria-label={`${t("remove")} ${i + 1}`}
                      onClick={() =>
                        setRows((rs) => rs!.filter((_, j) => j !== i))
                      }
                    >
                      <X className="size-4" aria-hidden />
                    </Button>
                  </div>
                </div>
                <div className="mt-2 flex flex-wrap gap-2 pl-7">
                  <input
                    aria-label={`${t("plan_equipment")} ${i + 1}`}
                    placeholder={t("plan_equipment")}
                    value={r.equipment}
                    maxLength={80}
                    onChange={(e) => setCell(i, "equipment", e.target.value)}
                    className={`${cellClass} min-w-40 flex-1`}
                  />
                  <input
                    aria-label={`${t("plan_notes")} ${i + 1}`}
                    placeholder={t("plan_notes")}
                    value={r.notes}
                    maxLength={300}
                    onChange={(e) => setCell(i, "notes", e.target.value)}
                    className={`${cellClass} min-w-48 flex-[2]`}
                  />
                </div>
              </li>
            ))}
          </ol>
          <div className="mt-3 flex flex-wrap items-center justify-between gap-2">
            <Button
              type="button"
              variant="outline"
              size="sm"
              disabled={rows.length >= 30}
              onClick={() => setRows((rs) => [...rs!, { ...EMPTY_ROW }])}
            >
              <Plus className="mr-1 size-4" aria-hidden />
              {t("add_step")}
            </Button>
            <span className="text-xs text-gray-600">
              {t("plan_total_minutes", { count: totalMinutes(toItems(rows)) })}
            </span>
            <div className="flex gap-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => setRows(null)}
              >
                {t("cancel")}
              </Button>
              <Button type="submit" variant="primary" disabled={save.isPending}>
                {t("save")}
              </Button>
            </div>
          </div>
        </form>
      )}
    </section>
  );
}

/** The steps of one version, numbered, with time, dose, equipment and notes. */
function PlanSteps({ plan, small }: { plan: TreatmentPlan; small?: boolean }) {
  const { t } = useTranslation();
  return (
    <>
      <ol className={small ? "space-y-0.5" : "space-y-2"}>
        {plan.items.map((it, i) => (
          <li key={i} className="flex gap-2 text-sm">
            <span className="w-5 shrink-0 text-right font-semibold text-gray-500">
              {i + 1}.
            </span>
            <div>
              <span className="font-medium text-gray-900">{it.activity}</span>
              {[
                it.minutes && t("n_minutes", { count: it.minutes }),
                it.dose,
                it.equipment,
              ]
                .filter(Boolean)
                .map((s) => (
                  <span key={String(s)} className="text-gray-600">
                    {" "}
                    · {s}
                  </span>
                ))}
              {it.notes && !small && (
                <p className="text-xs text-gray-500">{it.notes}</p>
              )}
            </div>
          </li>
        ))}
      </ol>
      {!small && totalMinutes(plan.items) > 0 && (
        <p className="mt-2 text-xs font-medium text-gray-700">
          {t("plan_total_minutes", { count: totalMinutes(plan.items) })}
        </p>
      )}
    </>
  );
}
