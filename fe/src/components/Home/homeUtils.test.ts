import { describe, expect, it } from "vitest";

import {
  arrivals,
  clinicLoad,
  greetingKey,
  myDay,
  nextSession,
} from "@/components/Home/homeUtils";
import { Appointment, Board } from "@/components/TimeBoard/types";

const appt = (
  id: string,
  name: string,
  status: Appointment["status"] = "booked",
  credits = 1,
): Appointment => ({
  id,
  date: "2026-10-26",
  hour: 0,
  therapist_id: 0,
  patient_id: id,
  patient: { id, name, care_level: credits },
  credits,
  status,
  notes: "",
});

const hour = (h: number, appointments: Appointment[]) => ({
  hour: h,
  division_id: "gym",
  shift_id: "s",
  capacity: 4,
  credits: appointments.reduce((n, a) => n + a.credits, 0),
  credit_limit: 8,
  appointments,
});

const board: Board = {
  date: "2026-10-21",
  open_hour: 8,
  close_hour: 20,
  divisions: [
    {
      id: "gym",
      name: "Ginásio",
      color: "teal",
      capacity: 4,
      credit_limit: 8,
      floor: 1,
      sort_order: 0,
      active: true,
    },
  ],
  therapists: [
    {
      id: 10,
      first_name: "Inês",
      last_name: "Moreira",
      shifts: [
        {
          id: "s",
          date: "",
          start_hour: 8,
          end_hour: 11,
          therapist_id: 10,
          division_id: "gym",
          notes: "",
        },
      ],
      hours: [
        hour(8, [appt("a", "Rui Lopes", "attended")]),
        hour(9, []),
        hour(10, [appt("b", "Ana Reis", "booked", 5)]),
      ],
      patients: 2,
      capacity: 12,
      credits: 6,
      credit_capacity: 24,
    },
    {
      id: 11,
      first_name: "Tiago",
      last_name: "Fonseca",
      shifts: [],
      hours: [],
      patients: 0,
      capacity: 0,
      credits: 0,
      credit_capacity: 0,
    },
  ],
  permissions: { manage_shifts: false, book_patients: false },
  absences: [],
};

describe("home helpers", () => {
  it("greets by time of day", () => {
    expect(greetingKey(8)).toBe("good_morning");
    expect(greetingKey(15)).toBe("good_afternoon");
    expect(greetingKey(21)).toBe("good_evening");
    expect(greetingKey(2)).toBe("good_evening");
  });

  it("builds a physio's day and finds the next session", () => {
    const day = myDay(board, 10);
    expect(day.map((h) => h.hour)).toEqual([8, 9, 10]);
    expect(day[0].division?.name).toBe("Ginásio");
    expect(nextSession(day, 9)?.hour).toBe(10);
    expect(nextSession(day, 11)).toBeUndefined();
    expect(myDay(board, 99)).toEqual([]);
  });

  it("lists today's arrivals by hour", () => {
    expect(
      arrivals(board).map((a) => [a.hour, a.patient, a.physio, a.careLevel]),
    ).toEqual([
      [8, "Rui Lopes", "Inês Moreira", 1],
      [10, "Ana Reis", "Inês Moreira", 5],
    ]);
  });

  it("sums the clinic's load over the physios working", () => {
    expect(clinicLoad(board)).toEqual({
      physios: 1,
      patients: 2,
      attended: 1,
      credits: 6,
      capacity: 24,
      percent: 25,
    });
  });
});
