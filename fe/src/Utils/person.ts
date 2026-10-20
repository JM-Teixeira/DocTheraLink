/**
 * Small helpers for displaying people (patients and staff). Pure functions so
 * they are easy to test and reuse across modules.
 */

/** Initials for avatars, e.g. "Sarah Johnson" → "SJ". */
export function initials(name: string) {
  return name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((p) => p[0]!.toUpperCase())
    .join("");
}

/** Age in whole years at `now`, or null when the birth date is unknown/invalid. */
export function ageFromDob(dob?: string | null, now = new Date()) {
  if (!dob) return null;
  const d = new Date(dob);
  if (Number.isNaN(d.getTime())) return null;
  let age = now.getFullYear() - d.getFullYear();
  const m = now.getMonth() - d.getMonth();
  if (m < 0 || (m === 0 && now.getDate() < d.getDate())) age--;
  return age;
}
