/** Pure helpers for the home page (tested in homeUtils.test.ts). */
import { Appointment, Board, Division } from "@/components/TimeBoard/types";

/** Translation key of the greeting for an hour of the day. */
export function greetingKey(hour: number) {
  if (hour < 5 || hour >= 20) return "good_evening";
  if (hour < 13) return "good_morning";
  return "good_afternoon";
}

export interface MyHour {
  hour: number;
  division?: Division;
  appointments: Appointment[];
}

/** A physiotherapist's own hours on the board, with their patients. */
export function myDay(board: Board, userId: number): MyHour[] {
  const row = board.therapists.find((t) => t.id === userId);
  if (!row) return [];
  return row.hours.map((h) => ({
    hour: h.hour,
    division: board.divisions.find((d) => d.id === h.division_id),
    appointments: h.appointments,
  }));
}

/** The next hour (from now on) that has patients, or undefined. */
export const nextSession = (day: MyHour[], nowHour: number) =>
  day.find((h) => h.hour >= nowHour && h.appointments.length > 0);

export interface Arrival {
  hour: number;
  patientId: string;
  patient: string;
  careLevel: number;
  physio: string;
  division?: Division;
  status: Appointment["status"];
}

/** Everyone booked today, by hour then patient name (front-desk list). */
export function arrivals(board: Board): Arrival[] {
  const out: Arrival[] = [];
  for (const t of board.therapists)
    for (const h of t.hours)
      for (const a of h.appointments)
        out.push({
          hour: h.hour,
          patientId: a.patient_id,
          patient: a.patient?.name ?? "",
          careLevel: a.credits,
          physio: `${t.first_name} ${t.last_name}`,
          division: board.divisions.find((d) => d.id === h.division_id),
          status: a.status,
        });
  return out.sort(
    (a, b) => a.hour - b.hour || a.patient.localeCompare(b.patient),
  );
}

/** Today's load of the clinic: who works, how many sessions, how full. */
export function clinicLoad(board: Board) {
  const working = board.therapists.filter((t) => t.shifts.length > 0);
  const sum = (f: (t: (typeof working)[number]) => number) =>
    working.reduce((acc, t) => acc + f(t), 0);
  const credits = sum((t) => t.credits);
  const capacity = sum((t) => t.credit_capacity);
  return {
    physios: working.length,
    patients: sum((t) => t.patients),
    attended: sum((t) =>
      t.hours.reduce(
        (n, h) =>
          n + h.appointments.filter((a) => a.status === "attended").length,
        0,
      ),
    ),
    credits,
    capacity,
    percent: capacity ? Math.round((credits / capacity) * 100) : 0,
  };
}
