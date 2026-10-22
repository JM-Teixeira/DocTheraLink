/**
 * Types for the doctor → physiotherapist referral workflow.
 * Mirrors be/internal/modules/referral/model.go.
 */

export const REFERRAL_STATUSES = [
  "pending",
  "accepted",
  "in_progress",
  "completed",
  "declined",
  "cancelled",
] as const;

export type ReferralStatus = (typeof REFERRAL_STATUSES)[number];

export type ReferralPriority = "routine" | "urgent";

export type NoteKind = "progress" | "question" | "reply" | "system";

/** Minimal user view embedded in referral responses. */
export interface UserSummary {
  id: number;
  username: string;
  first_name: string;
  last_name: string;
  user_type: string;
}

/** Minimal patient view embedded in referral responses. */
export interface PatientSummary {
  id: string;
  name: string;
  date_of_birth?: string | null;
  phone_number?: string | null;
}

export interface Referral {
  id: string;
  patient_id: string;
  facility_id?: string | null;
  doctor_id: number;
  therapist_id?: number | null;
  diagnosis: string;
  icd_code?: string | null;
  reason: string;
  precautions?: string | null;
  goals?: string | null;
  recommended_sessions?: number | null;
  priority: ReferralPriority;
  status: ReferralStatus;
  decline_reason?: string | null;
  cancel_reason?: string | null;
  outcome_summary?: string | null;
  accepted_at?: string | null;
  started_at?: string | null;
  completed_at?: string | null;
  patient?: PatientSummary;
  doctor?: UserSummary;
  therapist?: UserSummary | null;
  created_date: string;
  modified_date: string;
  /** Questions on the thread waiting for the viewer to answer. */
  open_questions?: number;
}

/** One step of a session in the treatment plan (be/internal/modules/referral/plan.go). */
export interface PlanItem {
  activity: string;
  minutes?: number;
  dose?: string;
  equipment?: string;
  notes?: string;
}

export interface TreatmentPlan {
  id: string;
  referral_id: string;
  items: PlanItem[];
  author_id: number;
  author?: { id: number; first_name: string; last_name: string };
  created_date: string;
}

export interface PlanView {
  current: TreatmentPlan | null;
  versions: TreatmentPlan[];
  can_edit: boolean;
}

export interface ReferralNote {
  id: string;
  referral_id: string;
  author_id: number;
  kind: NoteKind;
  body: string;
  /** Replies: the question answered. */
  reply_to_id?: string | null;
  /** Progress notes: pain 0-10. */
  pain?: number | null;
  /** Questions: someone replied. */
  answered?: boolean;
  author?: UserSummary;
  created_date: string;
}

export interface CreateReferralRequest {
  patient_id: string;
  therapist_id?: number;
  diagnosis: string;
  icd_code?: string;
  reason: string;
  precautions?: string;
  goals?: string;
  recommended_sessions?: number;
  priority: ReferralPriority;
}

export interface ReasonRequest {
  reason: string;
}

export interface NoteRequest {
  kind?: Exclude<NoteKind, "system">;
  body: string;
  reply_to?: string;
  pain?: number;
}

export interface ReassignRequest {
  therapist_id: number;
  reason: string;
}

export interface ReopenRequest {
  therapist_id?: number;
  reason: string;
}

/** Number of referrals per status, from GET /referrals/counts/. */
export type ReferralCounts = Partial<Record<ReferralStatus, number>>;

/** Actions a user can take on a referral (see referralUtils.allowedActions). */
export type ReferralAction =
  | "accept"
  | "decline"
  | "start"
  | "complete"
  | "cancel"
  | "reassign"
  | "reopen";

/** One ICD-10 code from GET /catalog/icd10/ (description in the UI language). */
export interface IcdCode {
  code: string;
  description: string;
}

/** Built-in referral template from GET /catalog/referral-templates/. */
export interface ReferralTemplate {
  key: string;
  name: string;
  icd_code: string;
  diagnosis: string;
  reason: string;
  goals: string;
  precautions: string;
  priority: ReferralPriority;
  recommended_sessions: number;
}
