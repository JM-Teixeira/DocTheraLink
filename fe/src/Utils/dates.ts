/**
 * Dates shown the way each person chose on their profile (tested in
 * dates.test.ts). Values travel as ISO "YYYY-MM-DD"; only display and typing
 * follow the preference.
 */
import dayjs from "@/Utils/dayjs";

export type DateFormat = "dmy" | "mdy" | "ymd" | "long";

export const DATE_FORMATS: DateFormat[] = ["dmy", "mdy", "ymd", "long"];

/** dayjs pattern per preference ("long" uses the language's month names). */
export const DATE_PATTERNS: Record<DateFormat, string> = {
  dmy: "DD/MM/YYYY",
  mdy: "MM/DD/YYYY",
  ymd: "YYYY-MM-DD",
  long: "D MMM YYYY",
};

/** Unknown or missing preference → day/month/year (Portugal). */
export const dateFormatOf = (pref?: string | null): DateFormat =>
  DATE_FORMATS.includes(pref as DateFormat) ? (pref as DateFormat) : "dmy";

/** "2026-10-21" → "21/10/2026" (or the chosen format); "" for no date. */
export function formatDate(
  value: string | Date | null | undefined,
  pref?: string | null,
) {
  if (!value) return "";
  const d = dayjs(value);
  return d.isValid() ? d.format(DATE_PATTERNS[dateFormatOf(pref)]) : "";
}

/** Date and time, e.g. "21/10/2026 15:56". */
export function formatDateTime(
  value: string | Date | null | undefined,
  pref?: string | null,
) {
  const date = formatDate(value, pref);
  return date ? `${date} ${dayjs(value).format("HH:mm")}` : "";
}

/** The pattern people type in (numeric formats; "long" types as d/m/y). */
export const typingPattern = (pref?: string | null) => {
  const f = dateFormatOf(pref);
  return f === "long" ? DATE_PATTERNS.dmy : DATE_PATTERNS[f];
};

/**
 * What someone typed → ISO date, or null when it isn't a real date in their
 * format (31/02 is refused, not rolled over).
 */
export function parseTypedDate(text: string, pref?: string | null) {
  const t = text.trim();
  if (!t) return null;
  const d = dayjs(t, typingPattern(pref), true);
  return d.isValid() ? d.format("YYYY-MM-DD") : null;
}
