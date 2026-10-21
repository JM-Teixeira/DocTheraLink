/** Pure helpers for the staff pages (tested in staffUtils.test.ts). */
import { EmploymentStatus } from "@/components/Staff/types";

/** The roles a clinic works with, in the order the filter shows them. */
export const CLINIC_ROLES = [
  "Doctor",
  "PhysioTherapist",
  "Coordinator",
  "Receptionist",
  "Nurse",
  "StateAdmin",
] as const;

/** Translation key of a role; unknown roles fall back to their id. */
export const roleKey = (userType: string) => `role__${userType}`;

/** Badge classes per employment status (literal, so the build keeps them). */
export const STATUS_STYLES: Record<EmploymentStatus, string> = {
  active: "bg-emerald-100 text-emerald-800",
  on_leave: "bg-amber-100 text-amber-800",
  left: "bg-gray-200 text-gray-700",
};

/** Missing status (older rows) counts as active. */
export const statusOf = (s?: string | null): EmploymentStatus =>
  s === "on_leave" || s === "left" ? s : "active";

/** Query params of the users list; "all" and empty values are left out. */
export function staffQuery(f: {
  search: string;
  role: string;
  clinic: string;
  status: EmploymentStatus | "all";
  limit: number;
  offset: number;
}) {
  const q: Record<string, string | number> = {
    limit: f.limit,
    offset: f.offset,
  };
  if (f.search.trim()) q.search = f.search.trim();
  if (f.role) q.user_type = f.role;
  if (f.clinic) q.home_facility = f.clinic;
  if (f.status !== "all") q.employment_status = f.status;
  return q;
}
