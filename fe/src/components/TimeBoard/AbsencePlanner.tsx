import dayjs from "dayjs";
import { ArrowRight, CircleCheck, CircleX } from "lucide-react";
import { useTranslation } from "react-i18next";

import { cn } from "@/lib/utils";

import {
  CARE_LEVEL_STYLES,
  formatHour,
} from "@/components/TimeBoard/boardUtils";
import { MovedSession } from "@/components/TimeBoard/types";

type Physio = { id: number; first_name: string; last_name: string };
const nameOf = (physios: Physio[], id?: number) => {
  const p = physios.find((x) => x.id === id);
  return p ? `${p.first_name} ${p.last_name}` : "";
};

/** Who receives which sessions, from the preview (moved sessions only). */
export function receivers(sessions: MovedSession[]) {
  const by = new Map<number, MovedSession[]>();
  for (const s of sessions) {
    if (s.moved && s.to) by.set(s.to, [...(by.get(s.to) ?? []), s]);
  }
  return [...by.entries()];
}

/**
 * Where each session of the absent physio goes. Per session, everyone who
 * works at that hour is offered with the room they have left (places and
 * care credits, after the other moves in this plan); "stays" keeps it with
 * the absent physio. A summary shows who receives whom.
 */
export default function AbsencePlanner({
  sessions,
  physios,
  absentId,
  onAssign,
}: {
  sessions: MovedSession[];
  physios: Physio[];
  absentId: number;
  onAssign: (appointmentId: string, therapistId: number) => void;
}) {
  const { t } = useTranslation();
  const absentName = nameOf(physios, absentId);
  const groups = receivers(sessions);

  return (
    <div className="grid gap-4 lg:grid-cols-[1fr_16rem]">
      <div className="overflow-x-auto rounded-lg border border-gray-200">
        <table className="min-w-full text-sm">
          <thead className="bg-gray-50 text-left text-xs text-gray-500">
            <tr>
              <th className="px-3 py-2 font-medium">{t("session")}</th>
              <th className="px-3 py-2 font-medium">{t("patient")}</th>
              <th className="px-3 py-2 font-medium">{t("goes_to")}</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100">
            {sessions.map((s) => {
              const current = s.moved
                ? (s.to ?? 0)
                : s.reason === "kept"
                  ? 0
                  : -1;
              return (
                <tr key={s.appointment_id}>
                  <td className="whitespace-nowrap px-3 py-2 text-gray-600">
                    <span className="capitalize">
                      {dayjs(s.date).format("ddd D MMM")}
                    </span>{" "}
                    <span className="font-mono">{formatHour(s.hour)}</span>
                  </td>
                  <td className="px-3 py-2">
                    <span className="flex items-center gap-1.5">
                      <span
                        className={cn(
                          "rounded px-1 text-[11px] font-semibold",
                          CARE_LEVEL_STYLES[s.credits] ?? "",
                        )}
                      >
                        {s.credits}
                      </span>
                      {s.patient_name}
                    </span>
                  </td>
                  <td className="px-3 py-2">
                    <div className="flex items-center gap-2">
                      {s.moved ? (
                        <CircleCheck
                          className="size-4 shrink-0 text-emerald-600"
                          aria-hidden
                        />
                      ) : (
                        <CircleX
                          className="size-4 shrink-0 text-amber-600"
                          aria-hidden
                        />
                      )}
                      <select
                        aria-label={t("goes_to_for", {
                          name: s.patient_name,
                          hour: formatHour(s.hour),
                        })}
                        value={current}
                        onChange={(e) =>
                          onAssign(s.appointment_id, Number(e.target.value))
                        }
                        className="h-9 min-w-56 rounded-lg border border-gray-300 px-2 text-sm"
                      >
                        {current === -1 && (
                          <option value={-1}>
                            {t(`skip_reason__${s.reason}`)}
                          </option>
                        )}
                        <option value={0}>
                          {t("stays_with", { name: absentName })}
                        </option>
                        {s.candidates.map((c) => (
                          <option key={c.therapist_id} value={c.therapist_id}>
                            {nameOf(physios, c.therapist_id)} ·{" "}
                            {t("n_places", { count: c.places })} ·{" "}
                            {t("n_credits_left", { count: c.credits })}
                            {!c.fits && c.therapist_id !== s.to
                              ? ` · ${t("does_not_fit")}`
                              : ""}
                          </option>
                        ))}
                      </select>
                    </div>
                    {!s.moved && s.reason && s.reason !== "kept" && (
                      <p className="mt-1 text-xs text-amber-700">
                        {t(`skip_reason__${s.reason}`)}
                      </p>
                    )}
                    {!s.candidates.length && (
                      <p className="mt-1 text-xs text-gray-500">
                        {t("nobody_on_shift")}
                      </p>
                    )}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      <aside
        className="rounded-lg border border-gray-200 bg-gray-50 p-3 text-sm"
        aria-label={t("who_receives")}
      >
        <h3 className="mb-2 font-medium text-gray-900">{t("who_receives")}</h3>
        {!groups.length ? (
          <p className="text-xs text-gray-500">{t("nobody_receives")}</p>
        ) : (
          <ul className="flex flex-col gap-3">
            {groups.map(([id, list]) => (
              <li key={id}>
                <p className="flex items-center gap-1.5 font-medium text-gray-800">
                  <ArrowRight className="size-3.5" aria-hidden />
                  {nameOf(physios, id)}
                  <span className="text-xs font-normal text-gray-500">
                    ({list.length})
                  </span>
                </p>
                <ul className="ml-5 mt-1 text-xs text-gray-600">
                  {list.map((s) => (
                    <li key={s.appointment_id}>
                      <span className="capitalize">
                        {dayjs(s.date).format("ddd D")}
                      </span>{" "}
                      {formatHour(s.hour)} · {s.patient_name}
                    </li>
                  ))}
                </ul>
              </li>
            ))}
          </ul>
        )}
      </aside>
    </div>
  );
}
