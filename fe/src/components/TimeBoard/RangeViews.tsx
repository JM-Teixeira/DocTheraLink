import { useQuery } from "@tanstack/react-query";
import dayjs from "dayjs";
import { useTranslation } from "react-i18next";

import { cn } from "@/lib/utils";

import Loading from "@/components/Common/Loading";
import BoardRoutes from "@/components/TimeBoard/api";
import {
  DIVISION_COLORS,
  monthGrid,
  today,
  utilisation,
  weekDates,
  weekStart,
} from "@/components/TimeBoard/boardUtils";
import { DaySummary, Division, RangeView } from "@/components/TimeBoard/types";

import query from "@/Utils/request/query";

function useRange(from: string, to: string) {
  return useQuery({
    queryKey: ["board-range", from, to],
    queryFn: query(BoardRoutes.range, { queryParams: { from, to } }),
    refetchInterval: 120_000,
  });
}

const divisionOf = (v: RangeView, id: string): Division | undefined =>
  v.divisions.find((d) => d.id === id);

/** Load bar: share of the credits in use. */
function Bar({ credits, capacity }: { credits: number; capacity: number }) {
  const pct = utilisation({ credits, credit_capacity: capacity });
  return (
    <div
      className="h-1.5 rounded-full bg-gray-100"
      title={`${credits}/${capacity}`}
    >
      <div
        className={cn(
          "h-1.5 rounded-full",
          pct >= 100 ? "bg-rose-500" : "bg-primary-500",
        )}
        style={{ width: `${pct}%` }}
      />
    </div>
  );
}

/** Week: physios × 7 days — where each one works and how loaded. */
export function WeekView({
  date,
  onOpenDay,
  physios = [],
}: {
  date: string;
  onOpenDay: (d: string) => void;
  /** Only these physios (empty = all). */
  physios?: number[];
}) {
  const { t } = useTranslation();
  const start = weekStart(date);
  const days = weekDates(start);
  const { data: v } = useRange(days[0], days[6]);
  if (!v) return <Loading />;
  const byDate = new Map(v.days.map((d) => [d.date, d]));

  return (
    <div className="overflow-x-auto rounded-xl border border-gray-200 bg-white shadow-sm">
      <table className="w-full border-collapse text-sm">
        <thead>
          <tr className="text-xs text-gray-500">
            <th className="sticky left-0 z-10 w-48 min-w-48 border-r border-gray-200 bg-white p-3 text-left font-medium">
              {t("physiotherapist")}
            </th>
            {days.map((d) => (
              <th key={d} className="min-w-32 p-2 font-medium">
                <button
                  type="button"
                  onClick={() => onOpenDay(d)}
                  className={cn(
                    "rounded px-2 py-1 capitalize hover:bg-gray-100",
                    d === today() && "bg-primary-50 text-primary-700",
                  )}
                >
                  {dayjs(d).format("ddd D")}
                </button>
              </th>
            ))}
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-100">
          {v.therapists
            .filter((th) => !physios.length || physios.includes(th.id))
            .map((th) => (
              <tr key={th.id}>
                <th className="sticky left-0 z-10 border-r border-gray-200 bg-white p-3 text-left font-medium text-gray-900">
                  {th.first_name} {th.last_name}
                </th>
                {days.map((d) => {
                  const day = byDate
                    .get(d)
                    ?.therapists.find((x) => x.id === th.id);
                  return (
                    <td key={d} className="p-1.5 align-top">
                      {day ? (
                        <button
                          type="button"
                          onClick={() => onOpenDay(d)}
                          className="w-full rounded-md p-1.5 text-left hover:bg-gray-50"
                        >
                          <div className="flex flex-wrap gap-1">
                            {day.shifts.map((s) => {
                              const div = divisionOf(v, s.division_id);
                              return (
                                <span
                                  key={s.id}
                                  className={cn(
                                    "rounded px-1 text-[11px]",
                                    div
                                      ? DIVISION_COLORS[div.color].chip
                                      : "bg-gray-100",
                                  )}
                                >
                                  {s.start_hour}-{s.end_hour} {div?.name}
                                </span>
                              );
                            })}
                          </div>
                          <p className="mt-1 text-xs text-gray-600">
                            {t("n_patients", { count: day.patients })}
                          </p>
                          <Bar
                            credits={day.credits}
                            capacity={day.credit_capacity}
                          />
                        </button>
                      ) : (
                        <p className="p-1.5 text-xs text-gray-300">-</p>
                      )}
                    </td>
                  );
                })}
              </tr>
            ))}
          <tr className="bg-gray-50 text-xs text-gray-600">
            <th className="sticky left-0 z-10 border-r border-gray-200 bg-gray-50 p-3 text-left font-medium">
              {t("total")}
            </th>
            {days.map((d) => {
              const tot = byDate.get(d)?.totals;
              return (
                <td key={d} className="p-2">
                  {tot && tot.capacity > 0 && (
                    <>
                      {t("n_patients", { count: tot.patients })}
                      {tot.attended > 0 && (
                        <span className="text-emerald-700">
                          {" "}
                          · ✓ {tot.attended}
                        </span>
                      )}
                    </>
                  )}
                </td>
              );
            })}
          </tr>
        </tbody>
      </table>
    </div>
  );
}

/** Month: calendar of patients per day; click a day to open it. */
export function MonthView({
  date,
  onOpenDay,
}: {
  date: string;
  onOpenDay: (d: string) => void;
}) {
  const { t } = useTranslation();
  const weeks = monthGrid(date);
  const { data: v } = useRange(weeks[0][0], weeks.at(-1)![6]);
  if (!v) return <Loading />;
  const byDate = new Map<string, DaySummary>(v.days.map((d) => [d.date, d]));
  const month = dayjs(date).month();

  return (
    <div className="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-sm">
      <div className="grid grid-cols-7 border-b border-gray-200 bg-gray-50 text-center text-xs font-medium capitalize text-gray-500">
        {weeks[0].map((d) => (
          <div key={d} className="py-2">
            {dayjs(d).format("ddd")}
          </div>
        ))}
      </div>
      {weeks.map((w) => (
        <div
          key={w[0]}
          className="grid grid-cols-7 divide-x divide-gray-100 border-b border-gray-100 last:border-b-0"
        >
          {w.map((d) => {
            const day = byDate.get(d);
            const inMonth = dayjs(d).month() === month;
            return (
              <button
                key={d}
                type="button"
                onClick={() => onOpenDay(d)}
                className={cn(
                  "flex min-h-24 flex-col gap-1 p-2 text-left text-xs hover:bg-gray-50",
                  !inMonth && "bg-gray-50/60 text-gray-400",
                )}
              >
                <span
                  className={cn(
                    "self-start rounded-full px-1.5 text-sm font-semibold",
                    d === today() && "bg-primary-600 text-white",
                  )}
                >
                  {dayjs(d).date()}
                </span>
                {day && day.totals.capacity > 0 && (
                  <>
                    <span className="text-gray-700">
                      {t("n_patients", { count: day.totals.patients })}
                    </span>
                    {day.totals.attended > 0 && (
                      <span className="text-emerald-700">
                        ✓ {day.totals.attended}
                      </span>
                    )}
                    <span className="text-gray-500">
                      {t("n_physios", { count: day.therapists.length })}
                    </span>
                    <Bar
                      credits={day.totals.credits}
                      capacity={day.totals.credit_capacity}
                    />
                  </>
                )}
              </button>
            );
          })}
        </div>
      ))}
    </div>
  );
}

/** Year: sessions held per month, and patients referred and discharged. */
export function YearView({
  year,
  onOpenMonth,
}: {
  year: number;
  onOpenMonth: (date: string) => void;
}) {
  const { t } = useTranslation();
  const { data } = useQuery({
    queryKey: ["board-year", year],
    queryFn: query(BoardRoutes.year, { queryParams: { year } }),
  });
  if (!data) return <Loading />;
  const sum = (k: "attended" | "missed" | "discharged" | "referred") =>
    data.months.reduce((a, m) => a + m[k], 0);
  const peak = Math.max(
    1,
    ...data.months.map((m) => m.attended + m.missed + m.booked),
  );

  return (
    <div className="flex flex-col gap-4">
      <div className="grid gap-3 sm:grid-cols-4">
        {(
          [
            ["sessions_attended", sum("attended"), "text-emerald-700"],
            ["sessions_missed", sum("missed"), "text-rose-700"],
            ["patients_referred", sum("referred"), "text-gray-900"],
            ["patients_discharged", sum("discharged"), "text-primary-700"],
          ] as const
        ).map(([key, value, color]) => (
          <div
            key={key}
            className="rounded-xl border border-gray-200 bg-white p-4 shadow-sm"
          >
            <p className="text-xs font-medium uppercase tracking-wide text-gray-500">
              {t(key)}
            </p>
            <p className={cn("mt-1 text-2xl font-semibold", color)}>{value}</p>
          </div>
        ))}
      </div>
      <div className="overflow-x-auto rounded-xl border border-gray-200 bg-white shadow-sm">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 text-xs text-gray-500">
            <tr>
              <th className="p-3 text-left font-medium">{t("month")}</th>
              <th className="w-1/2 p-3 text-left font-medium">
                {t("sessions")}
              </th>
              <th className="p-3 text-right font-medium">
                {t("patients_seen")}
              </th>
              <th className="p-3 text-right font-medium">
                {t("patients_referred")}
              </th>
              <th className="p-3 text-right font-medium">
                {t("patients_discharged")}
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100">
            {data.months.map((m) => {
              const total = m.attended + m.missed + m.booked;
              return (
                <tr key={m.month} className="hover:bg-gray-50">
                  <td className="p-3">
                    <button
                      type="button"
                      onClick={() =>
                        onOpenMonth(
                          dayjs(new Date(year, m.month - 1, 1)).format(
                            "YYYY-MM-DD",
                          ),
                        )
                      }
                      className="font-medium capitalize text-primary-700 hover:underline"
                    >
                      {dayjs(new Date(year, m.month - 1, 1)).format("MMMM")}
                    </button>
                  </td>
                  <td className="p-3">
                    <div className="flex items-center gap-2">
                      <div
                        className="flex h-3 flex-1 overflow-hidden rounded-full bg-gray-100"
                        title={`${m.attended} / ${m.missed} / ${m.booked}`}
                      >
                        <div
                          className="bg-emerald-500"
                          style={{ width: `${(m.attended / peak) * 100}%` }}
                        />
                        <div
                          className="bg-rose-400"
                          style={{ width: `${(m.missed / peak) * 100}%` }}
                        />
                        <div
                          className="bg-gray-300"
                          style={{ width: `${(m.booked / peak) * 100}%` }}
                        />
                      </div>
                      <span className="w-10 text-right text-xs text-gray-600">
                        {total}
                      </span>
                    </div>
                  </td>
                  <td className="p-3 text-right">{m.patients}</td>
                  <td className="p-3 text-right">{m.referred}</td>
                  <td className="p-3 text-right font-medium text-primary-700">
                    {m.discharged}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
        <div className="flex gap-4 border-t border-gray-100 p-3 text-xs text-gray-600">
          <span className="flex items-center gap-1.5">
            <span className="size-2.5 rounded-full bg-emerald-500" />{" "}
            {t("appointment_status__attended")}
          </span>
          <span className="flex items-center gap-1.5">
            <span className="size-2.5 rounded-full bg-rose-400" />{" "}
            {t("appointment_status__missed")}
          </span>
          <span className="flex items-center gap-1.5">
            <span className="size-2.5 rounded-full bg-gray-300" />{" "}
            {t("appointment_status__booked")}
          </span>
        </div>
      </div>
    </div>
  );
}
