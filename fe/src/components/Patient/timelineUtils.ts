/**
 * Pure helpers for the patient timeline (be/internal/modules/timeline):
 * filtering by kind, grouping by month and the label of each entry.
 */
import { TimelineEntry } from "@/components/Patient/api/routes";

export type TimelineFilter = "all" | "referral" | "note" | "file" | "session";
export const TIMELINE_FILTERS: TimelineFilter[] = [
  "all",
  "referral",
  "note",
  "file",
  "session",
];

export const filterEntries = (es: TimelineEntry[], f: TimelineFilter) =>
  f === "all" ? es : es.filter((e) => e.kind === f);

/** How many entries of each kind (for the filter chips). */
export function countKinds(es: TimelineEntry[]) {
  const out: Record<TimelineFilter, number> = {
    all: es.length,
    referral: 0,
    note: 0,
    file: 0,
    session: 0,
  };
  for (const e of es) if (e.kind in out) out[e.kind as TimelineFilter]++;
  return out;
}

/** "YYYY-MM" of an entry: the session day, else the local date of `at`. */
export function monthOf(e: TimelineEntry) {
  if (e.date) return e.date.slice(0, 7);
  const d = new Date(e.at);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}`;
}

/** Consecutive entries of the same month together (input already sorted). */
export function groupByMonth(es: TimelineEntry[]) {
  const groups: { month: string; entries: TimelineEntry[] }[] = [];
  for (const e of es) {
    const m = monthOf(e);
    const last = groups[groups.length - 1];
    if (last?.month === m) last.entries.push(e);
    else groups.push({ month: m, entries: [e] });
  }
  return groups;
}

/** Translation key of an entry's title. A past session never marked is "unrecorded". */
export function titleKey(e: TimelineEntry, today: string) {
  if (e.kind === "session" && e.status === "booked")
    return (e.date ?? "") < today
      ? "timeline__session__unrecorded"
      : "timeline__session__booked";
  if (e.kind === "registered" || e.kind === "file")
    return `timeline__${e.kind}`;
  return `timeline__${e.kind}__${e.status}`;
}
