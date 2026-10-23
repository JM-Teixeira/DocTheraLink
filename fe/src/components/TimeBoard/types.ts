/** Types of the time-board API (be/internal/modules/schedule). */

export type DivisionColor =
  | "teal"
  | "indigo"
  | "amber"
  | "sky"
  | "rose"
  | "violet"
  | "emerald"
  | "orange";

export interface Division {
  id: string;
  name: string;
  color: DivisionColor;
  capacity: number; // patients per physiotherapist per hour
  credit_limit: number; // care-level credits per physiotherapist per hour
  floor: number; // building floor
  sort_order: number;
  active: boolean;
}

export interface Shift {
  id: string;
  date: string;
  start_hour: number;
  end_hour: number; // exclusive
  therapist_id: number;
  division_id: string;
  notes: string;
}

export type AppointmentStatus = "booked" | "attended" | "missed";

export interface Appointment {
  id: string;
  date: string; // YYYY-MM-DD
  hour: number;
  therapist_id: number;
  patient_id: string;
  patient?: { id: string; name: string; care_level: number };
  credits: number;
  series_id?: string | null;
  referral_id?: string | null;
  status: AppointmentStatus;
  /** Missed sessions: justified (ill, warned) or not, and why. */
  miss_justified?: boolean | null;
  miss_reason?: string;
  notes: string;
}

export interface HourLoad {
  hour: number;
  division_id: string;
  shift_id: string;
  capacity: number;
  credits: number;
  credit_limit: number;
  appointments: Appointment[];
}

export interface BoardTherapist {
  id: number;
  first_name: string;
  last_name: string;
  shifts: Shift[];
  hours: HourLoad[];
  patients: number;
  capacity: number;
  credits: number;
  credit_capacity: number;
}

export interface Board {
  date: string;
  open_hour: number;
  close_hour: number;
  divisions: Division[];
  therapists: BoardTherapist[];
  permissions: { manage_shifts: boolean; book_patients: boolean };
  absences: Absence[];
}

export interface DivisionRequest {
  name: string;
  color: DivisionColor;
  capacity: number;
  credit_limit?: number;
  floor?: number;
  sort_order?: number;
}

export interface ShiftRequest {
  date: string;
  start_hour: number;
  end_hour: number;
  therapist_id: number;
  division_id: string;
  notes?: string;
}

export interface AppointmentRequest {
  date: string;
  hour: number;
  therapist_id: number;
  patient_id: string;
  notes?: string;
}

/** Week / month view (GET /schedule/range/). */
export interface DayTotals {
  patients: number;
  attended: number;
  missed: number;
  capacity: number;
  credits: number;
  credit_capacity: number;
}

export interface TherapistDay {
  id: number;
  shifts: Shift[];
  patients: number;
  capacity: number;
  credits: number;
  credit_capacity: number;
}

export interface DaySummary {
  date: string;
  totals: DayTotals;
  therapists: TherapistDay[];
}

export interface RangeView {
  from: string;
  to: string;
  divisions: Division[];
  therapists: { id: number; first_name: string; last_name: string }[];
  days: DaySummary[];
  absences: Absence[];
  permissions: { manage_shifts: boolean; book_patients: boolean };
}

/** Year view (GET /schedule/year/). */
export interface MonthStats {
  month: number;
  booked: number;
  attended: number;
  missed: number;
  patients: number;
  referred: number;
  discharged: number;
}

export interface YearStats {
  year: number;
  months: MonthStats[];
}

/** Weekly treatment plan (POST /schedule/series/). */
export interface SeriesRequest {
  patient_id: string;
  therapist_id: number;
  weekdays: number[]; // 0 = Sunday … 6 = Saturday
  hour: number;
  start_date: string;
  sessions: number;
  referral_id?: string;
  dry_run?: boolean;
}

export type SkipReason =
  | "not_on_shift"
  | "full"
  | "over_credits"
  | "patient_busy";

export interface SeriesResult {
  series_id?: string;
  booked: string[];
  skipped: { date: string; reason: SkipReason }[];
}

/** A physiotherapist away, optionally covered by a colleague. */
export type AbsenceReason = "vacation" | "sick" | "training" | "other";

export interface HandoverNote {
  id: string;
  patient_id: string;
  patient?: { id: string; name: string; care_level: number };
  note: string;
}

export interface Absence {
  id: string;
  therapist_id: number;
  cover_id?: number | null;
  from: string;
  to: string;
  reason: AbsenceReason;
  handover: string;
  notes: HandoverNote[];
  therapist?: { id: number; first_name: string; last_name: string };
  cover?: { id: number; first_name: string; last_name: string } | null;
}

export interface AbsenceRequest {
  therapist_id: number;
  cover_id?: number;
  from: string;
  to: string;
  reason: AbsenceReason;
  handover: string;
  notes: { patient_id: string; note: string }[];
  move_patients: boolean;
  dry_run?: boolean;
  /** Per session: who takes it (0 = stays). Others go to the cover. */
  assignments?: { appointment_id: string; therapist_id: number }[];
}

export interface MovedSession {
  appointment_id: string;
  date: string;
  hour: number;
  patient_id: string;
  patient_name: string;
  credits: number;
  moved: boolean;
  to?: number; // who takes it
  reason?: string;
  /** Physios working at that hour, with the room they have left. */
  candidates: Candidate[];
}

export interface Candidate {
  therapist_id: number;
  places: number;
  credits: number;
  fits: boolean;
}

export interface AbsenceResult {
  absence?: Absence;
  sessions: MovedSession[];
  patients: { id: string; name: string; care_level: number }[];
}

export interface PatientHandover {
  absence: Absence;
  note: string;
}

/** How often a patient missed sessions in the window (90 days). */
export interface MissCount {
  patient_id: string;
  patient?: { id: string; name: string; care_level: number };
  unjustified: number;
  justified: number;
  last: string;
}

/** Patients with too many unjustified misses (coordinators). */
export interface MissReport {
  threshold: number;
  window_days: number;
  patients: MissCount[];
}

/** What to send when marking a session missed. */
export interface MissInput {
  justified: boolean;
  reason: string;
}

/** Moves one booked session to another slot. */
export interface MoveRequest {
  date: string;
  hour: number;
  therapist_id: number;
}

/** New weekly pattern for the rest of a plan. */
export interface PlanRescheduleRequest {
  from: string;
  weekdays: number[];
  hour: number;
  therapist_id: number;
  dry_run?: boolean;
}

export interface PlanRescheduleResult extends SeriesResult {
  replaced: number;
}

export type WaitingStatus = "waiting" | "scheduled" | "removed";
export type WaitingPeriod = "morning" | "afternoon" | "any";

/** A patient waiting for a slot (urgent first, then oldest). */
export interface WaitingEntry {
  id: string;
  patient_id: string;
  patient?: { id: string; name: string; care_level: number };
  referral_id?: string | null;
  urgent: boolean;
  weekdays: number[];
  period: WaitingPeriod;
  preferred_therapist_id?: number | null;
  note: string;
  status: WaitingStatus;
  series_id?: string | null;
  created_date: string;
  /** First free hour that suits the patient in the next 2 weeks (server) */
  slot?: {
    date: string;
    hour: number;
    therapist_id: number;
    therapist: string;
  };
}

export interface WaitingRequest {
  patient_id: string;
  urgent: boolean;
  weekdays: number[];
  period: WaitingPeriod;
  preferred_therapist_id?: number;
  note: string;
}
