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

/** Share of the day's credits in use, 0–100 (for the per-physio bar). */
export const utilisation = (
  row: Pick<BoardTherapist, "credits" | "credit_capacity">,
) =>
  row.credit_capacity
    ? Math.round((row.credits / row.credit_capacity) * 100)
    : 0;

/**
 * Whether an hour is full: no patient place left, or not even a level-1
 * patient fits in the remaining credits.
 */
export const hourIsFull = (h: HourLoad) =>
  h.appointments.length >= h.capacity || h.credits >= h.credit_limit;

/** Monday of the date's week, as YYYY-MM-DD. */
export function weekStart(date: string) {
  const d = dayjs(date);
  const offset = (d.day() + 6) % 7; // Monday = 0
  return d.subtract(offset, "day").format("YYYY-MM-DD");
}

/** The 7 dates of the week starting at start. */
export const weekDates = (start: string) =>
  Array.from({ length: 7 }, (_, i) => addDays(start, i));

/**
 * Calendar of the date's month: weeks (Monday first) of dates; days of the
 * neighbouring months fill the first and last week.
 */
export function monthGrid(date: string) {
  const first = dayjs(date).startOf("month").format("YYYY-MM-DD");
  const last = dayjs(date).endOf("month").format("YYYY-MM-DD");
  const weeks: string[][] = [];
  for (let start = weekStart(first); start <= last; start = addDays(start, 7))
    weeks.push(weekDates(start));
  return weeks;
}

/** Care levels 1–5 (credits), with literal Tailwind classes. */
export const CARE_LEVELS = [1, 2, 3, 4, 5] as const;
export const CARE_LEVEL_STYLES: Record<number, string> = {
  1: "bg-emerald-100 text-emerald-800",
  2: "bg-lime-100 text-lime-800",
  3: "bg-amber-100 text-amber-800",
  4: "bg-orange-100 text-orange-800",
  5: "bg-rose-100 text-rose-800",
};

/** Weekday presets of a treatment plan (0 = Sunday … 6 = Saturday). */
export const PLAN_PRESETS: Record<string, number[]> = {
  once: [1],
  twice: [2, 4],
  three: [1, 3, 5],
  weekdays: [1, 2, 3, 4, 5],
};
