import { useTranslation } from "react-i18next";

import { cn } from "@/lib/utils";

import { PatientStats } from "@/components/Patient/api/routes";
import { AGE_BANDS, percent } from "@/components/Patient/patientFilters";
import { GENDERS } from "@/components/Patient/patientFormUtils";
import {
  CARE_LEVELS,
  CARE_LEVEL_STYLES,
} from "@/components/TimeBoard/boardUtils";

/** Solid bar colours per care level (the chips use the light ones). */
const LEVEL_BARS: Record<number, string> = {
  1: "bg-emerald-500",
  2: "bg-lime-500",
  3: "bg-amber-500",
  4: "bg-orange-500",
  5: "bg-rose-500",
};

/**
 * Numbers for the patients currently matched by the filters: total, average
 * age, and how they spread over care levels, age bands and gender.
 */
export default function PatientMetrics({ stats }: { stats?: PatientStats }) {
  const { t } = useTranslation();
  if (!stats)
    return (
      <div className="mb-5 h-28 animate-pulse rounded-xl border border-gray-200 bg-white" />
    );
  const total = stats.total;
  const n = (m: Record<string, number>, k: string | number) =>
    m[String(k)] ?? 0;

  return (
    <section
      aria-label={t("patient_metrics")}
      className="mb-5 grid gap-3 md:grid-cols-4"
    >
      <div className="rounded-xl border border-gray-200 bg-white p-4 shadow-sm">
        <p className="text-xs text-gray-500">{t("patients")}</p>
        <p className="text-2xl font-semibold text-gray-900">{total}</p>
        <p className="mt-1 text-xs text-gray-500">
          {t("average_age")}: {total ? Math.round(stats.average_age) : "-"}
        </p>
      </div>

      <div className="rounded-xl border border-gray-200 bg-white p-4 shadow-sm">
        <p className="mb-2 text-xs text-gray-500">{t("care_level")}</p>
        <div className="flex h-2.5 overflow-hidden rounded-full bg-gray-100">
          {CARE_LEVELS.map((l) => (
            <div
              key={l}
              className={LEVEL_BARS[l]}
              style={{ width: `${percent(n(stats.care_levels, l), total)}%` }}
            />
          ))}
        </div>
        <ul className="mt-2 flex flex-wrap gap-1.5 text-xs">
          {CARE_LEVELS.map((l) => (
            <li
              key={l}
              className={cn("rounded px-1.5", CARE_LEVEL_STYLES[l])}
              title={t("care_level_n", { level: l })}
            >
              {l}: {n(stats.care_levels, l)}
            </li>
          ))}
        </ul>
      </div>

      <div className="rounded-xl border border-gray-200 bg-white p-4 shadow-sm">
        <p className="mb-2 text-xs text-gray-500">{t("age")}</p>
        <ul className="flex flex-col gap-1 text-xs">
          {AGE_BANDS.map((b) => {
            const c = n(stats.age_bands, b.key);
            return (
              <li key={b.key} className="flex items-center gap-2">
                <span className="w-10 text-gray-600">{b.key}</span>
                <span className="h-1.5 flex-1 rounded-full bg-gray-100">
                  <span
                    className="block h-1.5 rounded-full bg-primary-500"
                    style={{ width: `${percent(c, total)}%` }}
                  />
                </span>
                <span className="w-6 text-right tabular-nums text-gray-700">
                  {c}
                </span>
              </li>
            );
          })}
        </ul>
      </div>

      <div className="rounded-xl border border-gray-200 bg-white p-4 shadow-sm">
        <p className="mb-2 text-xs text-gray-500">{t("gender")}</p>
        <ul className="flex flex-col gap-1 text-sm">
          {GENDERS.map((g) => (
            <li key={g.id} className="flex justify-between">
              <span className="text-gray-600">{t(g.key)}</span>
              <span className="tabular-nums text-gray-900">
                {n(stats.genders, g.id)}{" "}
                <span className="text-xs text-gray-500">
                  ({percent(n(stats.genders, g.id), total)}%)
                </span>
              </span>
            </li>
          ))}
        </ul>
      </div>
    </section>
  );
}
