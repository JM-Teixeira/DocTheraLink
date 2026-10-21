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
  patient?: { id: string; name: string };
  referral_id?: string | null;
  status: AppointmentStatus;
  notes: string;
}

export interface HourLoad {
  hour: number;
  division_id: string;
  shift_id: string;
  capacity: number;
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
