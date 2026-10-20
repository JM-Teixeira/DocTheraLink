/**
 * Pure helpers for the referral UI. Kept free of React so they are easy to test.
 * Permission rules mirror be/internal/modules/referral/service.go — the
 * backend remains the source of truth; these only decide which buttons to show.
 */
import {
  Referral,
  ReferralAction,
  ReferralCounts,
  ReferralStatus,
  UserSummary,
} from "@/components/Referral/types";

import { RoleUser, isAdmin, isDoctor, isTherapist } from "@/Utils/roles";

/** The fields of the signed-in user the helpers need. */
export type CurrentUser = RoleUser;

export { isAdmin, isDoctor, isTherapist };

/** Doctors and admins can open referrals. */
export const canCreateReferral = (u: CurrentUser) => isDoctor(u) || isAdmin(u);

const isAssignedTherapist = (r: Referral, u: CurrentUser) =>
  isTherapist(u) && r.therapist_id === u.id;

/** Which workflow buttons to show for this user and referral. */
export function allowedActions(r: Referral, u: CurrentUser): ReferralAction[] {
  const actions: ReferralAction[] = [];
  switch (r.status) {
    case "pending":
      if (isTherapist(u) && (r.therapist_id == null || r.therapist_id === u.id))
        actions.push("accept");
      if (isAssignedTherapist(r, u)) actions.push("decline");
      break;
    case "accepted":
      if (isAssignedTherapist(r, u)) actions.push("start");
      break;
    case "in_progress":
      if (isAssignedTherapist(r, u)) actions.push("complete");
      break;
  }
  const open = ["pending", "accepted", "in_progress"].includes(r.status);
  const owner = isAdmin(u) || (isDoctor(u) && r.doctor_id === u.id);
  // Hand-over: referring doctor, admin or the assigned therapist
  if (open && (owner || isAssignedTherapist(r, u))) actions.push("reassign");
  if (open && owner) actions.push("cancel");
  if (r.status === "declined" && owner) actions.push("reopen");
  return actions;
}

/** Actions that require a free-text reason (or outcome summary). */
export const actionNeedsReason = (a: ReferralAction) =>
  a === "decline" || a === "complete" || a === "cancel";

/** Total of a per-status count map (for the "All" tab). */
export const sumCounts = (counts: ReferralCounts) =>
  Object.values(counts).reduce((a, b) => a + (b ?? 0), 0);

/** Actions that pick a physiotherapist (handled by TherapistDialog). */
export const actionNeedsTherapist = (a: ReferralAction) =>
  a === "reassign" || a === "reopen";

/** Only the referring doctor, assigned therapist and admins post on the thread. */
export function canPostNote(r: Referral, u: CurrentUser) {
  return isAdmin(u) || r.doctor_id === u.id || isAssignedTherapist(r, u);
}

/** Visual style per status: tailwind classes for badges and the accent bar. */
export const STATUS_STYLES: Record<
  ReferralStatus,
  { badge: string; dot: string }
> = {
  pending: {
    badge: "bg-amber-50 text-amber-800 ring-amber-200",
    dot: "bg-amber-500",
  },
  accepted: { badge: "bg-sky-50 text-sky-800 ring-sky-200", dot: "bg-sky-500" },
  in_progress: {
    badge: "bg-violet-50 text-violet-800 ring-violet-200",
    dot: "bg-violet-500",
  },
  completed: {
    badge: "bg-emerald-50 text-emerald-800 ring-emerald-200",
    dot: "bg-emerald-500",
  },
  declined: {
    badge: "bg-rose-50 text-rose-800 ring-rose-200",
    dot: "bg-rose-500",
  },
  cancelled: {
    badge: "bg-gray-100 text-gray-700 ring-gray-200",
    dot: "bg-gray-400",
  },
};

/** Happy-path steps shown in the progress tracker. */
export const STATUS_STEPS: ReferralStatus[] = [
  "pending",
  "accepted",
  "in_progress",
  "completed",
];

/**
 * Index of the current step in STATUS_STEPS, or -1 when the referral left the
 * happy path (declined/cancelled).
 */
export const stepIndex = (s: ReferralStatus) => STATUS_STEPS.indexOf(s);

export function fullName(u?: UserSummary | null) {
  if (!u) return "";
  return `${u.first_name} ${u.last_name}`.trim() || u.username;
}

/** True when a name already carries a doctor title ("Dr. John", "dr Ana"). */
export const hasDoctorTitle = (name: string) => /^dr\.?\s/i.test(name.trim());

// Shared person helpers, re-exported so existing imports keep working
export { ageFromDob, initials } from "@/Utils/person";
