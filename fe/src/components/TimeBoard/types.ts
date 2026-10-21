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
  hour: number;
  therapist_id: number;
  patient_id: string;
  patient?: { id: string; name: string; care_level: number };
  credits: number;
  series_id?: string | null;
  referral_id?: string | null;
  status: AppointmentStatus;
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
}

export interface DivisionRequest {
  name: string;
  color: DivisionColor;
  capacity: number;
  credit_limit?: number;
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
