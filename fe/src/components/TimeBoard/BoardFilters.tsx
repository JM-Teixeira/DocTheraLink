import { Filter, X } from "lucide-react";
import { useTranslation } from "react-i18next";

import { cn } from "@/lib/utils";

import {
  BoardFilter,
  DIVISION_COLORS,
  NO_FILTER,
  floorsOf,
} from "@/components/TimeBoard/boardUtils";
import { Division } from "@/components/TimeBoard/types";

const STORAGE_KEY = "klyniva.board.filter";

/** The filter remembered on this browser (falls back to "everything"). */
export function loadFilter(): BoardFilter {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    return raw ? { ...NO_FILTER, ...JSON.parse(raw) } : NO_FILTER;
  } catch {
    return NO_FILTER;
  }
}

export function saveFilter(f: BoardFilter) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(f));
  } catch {
    // storage unavailable: the filter just isn't remembered
  }
}

const chip = (on: boolean) =>
  cn(
    "flex items-center gap-1.5 rounded-full px-3 py-1 text-sm ring-1 ring-inset transition",
    on
      ? "bg-primary-600 text-white ring-primary-600"
      : "bg-white text-gray-700 ring-gray-300 hover:bg-gray-50",
  );

/** Floor chips, division chips and a physio picker for the time board. */
export default function BoardFilters({
  divisions,
  physios,
  value,
  onChange,
}: {
  divisions: Division[];
  physios: { id: number; first_name: string; last_name: string }[];
  value: BoardFilter;
  onChange: (f: BoardFilter) => void;
}) {
  const { t } = useTranslation();
  const floors = floorsOf(divisions);
  const shownDivisions = divisions.filter(
    (d) => value.floor === null || (d.floor ?? 1) === value.floor,
  );
  const toggle = <T,>(list: T[], v: T) =>
    list.includes(v) ? list.filter((x) => x !== v) : [...list, v];
  const active =
    value.floor !== null ||
    value.divisions.length > 0 ||
    value.physios.length > 0;

  return (
    <div className="mb-4 flex flex-col gap-2 rounded-xl border border-gray-200 bg-white p-3 shadow-sm">
      <div className="flex flex-wrap items-center gap-2">
        <Filter className="size-4 text-gray-400" aria-hidden />
        <span className="text-xs font-medium uppercase tracking-wide text-gray-500">
          {t("floor")}
        </span>
        <button
          type="button"
          className={chip(value.floor === null)}
          onClick={() => onChange({ ...value, floor: null, divisions: [] })}
        >
          {t("all")}
        </button>
        {floors.map((fl) => (
          <button
            key={fl}
            type="button"
            className={chip(value.floor === fl)}
            onClick={() => onChange({ ...value, floor: fl, divisions: [] })}
          >
            {t("floor_n", { n: fl })}
          </button>
        ))}
        {active && (
          <button
            type="button"
            onClick={() => onChange(NO_FILTER)}
            className="ml-auto flex items-center gap-1 text-xs text-gray-500 hover:text-gray-800"
          >
            <X className="size-3.5" aria-hidden />
            {t("clear_filters")}
          </button>
        )}
      </div>
      <div className="flex flex-wrap items-center gap-2">
        <span className="text-xs font-medium uppercase tracking-wide text-gray-500">
          {t("division")}
        </span>
        {shownDivisions.map((d) => (
          <button
            key={d.id}
            type="button"
            aria-pressed={value.divisions.includes(d.id)}
            className={chip(value.divisions.includes(d.id))}
            onClick={() =>
              onChange({ ...value, divisions: toggle(value.divisions, d.id) })
            }
          >
            <span
              className={cn(
                "size-2.5 rounded-full",
                DIVISION_COLORS[d.color].dot,
              )}
            />
            {d.name}
          </button>
        ))}
      </div>
      <div className="flex flex-wrap items-center gap-2">
        <span className="text-xs font-medium uppercase tracking-wide text-gray-500">
          {t("physiotherapist")}
        </span>
        {physios.map((p) => (
          <button
            key={p.id}
            type="button"
            aria-pressed={value.physios.includes(p.id)}
            className={chip(value.physios.includes(p.id))}
            onClick={() =>
              onChange({ ...value, physios: toggle(value.physios, p.id) })
            }
          >
            {p.first_name} {p.last_name}
          </button>
        ))}
      </div>
    </div>
  );
}
