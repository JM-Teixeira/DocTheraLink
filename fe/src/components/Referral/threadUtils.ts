/** Helpers of the referral conversation (tested in threadUtils.test.ts). */
import { ReferralNote } from "@/components/Referral/types";

export type ThreadFilter = "all" | "progress" | "questions";

/** Notes shown for a filter: progress notes, or questions with their replies. */
export function filterThread(notes: ReferralNote[], f: ThreadFilter) {
  switch (f) {
    case "progress":
      return notes.filter((n) => n.kind === "progress");
    case "questions":
      return notes.filter((n) => n.kind === "question" || n.kind === "reply");
    default:
      return notes;
  }
}

/** Pain over time from progress notes that recorded it, oldest first. */
export function painSeries(notes: ReferralNote[]) {
  return notes
    .filter((n) => n.kind === "progress" && typeof n.pain === "number")
    .map((n) => ({ date: n.created_date, pain: n.pain as number }))
    .sort((a, b) => a.date.localeCompare(b.date));
}

/** Questions others asked this user that nobody has answered yet. */
export const waitingFor = (notes: ReferralNote[], userId: number) =>
  notes.filter(
    (n) => n.kind === "question" && !n.answered && n.author_id !== userId,
  );

/** SVG polyline points for a 0-10 series in a width × height box. */
export function sparkPoints(values: number[], width: number, height: number) {
  if (!values.length) return "";
  const step = values.length > 1 ? width / (values.length - 1) : 0;
  return values
    .map(
      (v, i) =>
        `${Math.round(i * step)},${Math.round(height - (Math.min(Math.max(v, 0), 10) / 10) * height)}`,
    )
    .join(" ");
}
