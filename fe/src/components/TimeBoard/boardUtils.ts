/** Pure helpers for the time board (tested in boardUtils.test.ts). */
import dayjs from "dayjs";

import {
  Board,
  BoardTherapist,
  DivisionColor,
  HourLoad,
} from "@/components/TimeBoard/types";

/** Literal Tailwind classes per division colour (so the build keeps them). */
export const DIVISION_COLORS: Record<
  DivisionColor,
  { cell: string; chip: string; dot: string }
> = {
  teal: {
    cell: "bg-teal-50 border-teal-200",
    chip: "bg-teal-100 text-teal-800",
    dot: "bg-teal-500",
  },
  indigo: {
    cell: "bg-indigo-50 border-indigo-200",
    chip: "bg-indigo-100 text-indigo-800",
    dot: "bg-indigo-500",
  },
  amber: {
    cell: "bg-amber-50 border-amber-200",
    chip: "bg-amber-100 text-amber-800",
    dot: "bg-amber-500",
  },
  sky: {
    cell: "bg-sky-50 border-sky-200",
    chip: "bg-sky-100 text-sky-800",
    dot: "bg-sky-500",
  },
  rose: {
    cell: "bg-rose-50 border-rose-200",
    chip: "bg-rose-100 text-rose-800",
    dot: "bg-rose-500",
  },
  violet: {
    cell: "bg-violet-50 border-violet-200",
    chip: "bg-violet-100 text-violet-800",
    dot: "bg-violet-500",
  },
  emerald: {
    cell: "bg-emerald-50 border-emerald-200",
    chip: "bg-emerald-100 text-emerald-800",
    dot: "bg-emerald-500",
  },
  orange: {
    cell: "bg-orange-50 border-orange-200",
    chip: "bg-orange-100 text-orange-800",
    dot: "bg-orange-500",
  },
};

export const COLOR_NAMES = Object.keys(DIVISION_COLORS) as DivisionColor[];

/** Hours shown as columns: open_hour … close_hour − 1. */
export const boardHours = (b: Pick<Board, "open_hour" | "close_hour">) =>
  Array.from(
    { length: Math.max(b.close_hour - b.open_hour, 0) },
    (_, i) => b.open_hour + i,
  );

/** "9" → "09:00". */
export const formatHour = (h: number) => `${String(h).padStart(2, "0")}:00`;

/** The physiotherapist's load at an hour, or undefined when off shift. */
export const hourOf = (
  row: BoardTherapist,
  hour: number,
): HourLoad | undefined => row.hours.find((h) => h.hour === hour);

export type LoadLevel = "free" | "some" | "full";

/** free = nobody yet, some = room left, full = capacity reached. */
export function loadLevel(booked: number, capacity: number): LoadLevel {
  if (booked >= capacity) return "full";
  return booked === 0 ? "free" : "some";
}

/** Date n days away, as YYYY-MM-DD. */
export const addDays = (date: string, n: number) =>
  dayjs(date).add(n, "day").format("YYYY-MM-DD");

export const today = () => dayjs().format("YYYY-MM-DD");

/**
 * The "by division" view: for each division and hour, which
 * physiotherapists are there and how many patients they have.
 */
export function staffingByDivision(board: Board) {
  const out = new Map<
    string,
    Map<number, { therapist: BoardTherapist; load: HourLoad }[]>
  >();
  for (const d of board.divisions) out.set(d.id, new Map());
  for (const t of board.therapists) {
    for (const load of t.hours) {
      const byHour = out.get(load.division_id) ?? new Map();
      byHour.set(load.hour, [
        ...(byHour.get(load.hour) ?? []),
        { therapist: t, load },
      ]);
      out.set(load.division_id, byHour);
    }
  }
  return out;
}

/** Share of the day's capacity in use, 0–100 (for the per-physio bar). */
export const utilisation = (
  row: Pick<BoardTherapist, "patients" | "capacity">,
) => (row.capacity ? Math.round((row.patients / row.capacity) * 100) : 0);
