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

/** Roles that may change all of their clinic's stock (inventory canManage). */
const STOCK_MANAGERS = [
  "Coordinator",
  "StateAdmin",
  "DistrictAdmin",
  "LocalBodyAdmin",
];

/**
 * Whether a person may change stock of a floor and/or category ("" / null =
 * any). Mirrors Grant.Covers on the server: a rule without floors covers all
 * floors, one without categories all categories; managers change everything.
 */
export function canChangeStock(
  userType: string,
  grants: { floors: number[]; categories: string[] }[],
  floor: number | null,
  category: string,
) {
  if (STOCK_MANAGERS.includes(userType)) return true;
  const cat = category.trim().toLowerCase();
  return grants.some(
    (g) =>
      (floor === null || !g.floors?.length || g.floors.includes(floor)) &&
      (!cat ||
        !g.categories?.length ||
        g.categories.some((c) => c.trim().toLowerCase() === cat)),
  );
}

/** One step of a person's work story, as shown on their page. */
export interface WorkStep {
  kind: "joined" | "leave" | "back" | "left";
  from: string;
  until?: string; // a leave that ended
  note?: string;
}

/**
 * Turns the employment history into short steps: joined, each leave (with
 * its end when they came back), back after leaving, left.
 */
export function workSteps(h: {
  joined_on: string;
  events: { status: string; since: string; note?: string }[];
}): WorkStep[] {
  const steps: WorkStep[] = [{ kind: "joined", from: h.joined_on }];
  for (const e of h.events) {
    const last = steps[steps.length - 1];
    if (e.status === "on_leave")
      steps.push({ kind: "leave", from: e.since, note: e.note });
    else if (e.status === "left")
      steps.push({ kind: "left", from: e.since, note: e.note });
    else if (last.kind === "leave" && !last.until) last.until = e.since;
    else if (last.kind === "left") steps.push({ kind: "back", from: e.since });
  }
  return steps;
}
