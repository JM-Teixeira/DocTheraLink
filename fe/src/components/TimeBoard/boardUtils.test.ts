import { describe, expect, it } from "vitest";

import {
  addDays,
  boardHours,
  formatHour,
  hourOf,
  loadLevel,
  staffingByDivision,
  utilisation,
} from "@/components/TimeBoard/boardUtils";
import { Board, BoardTherapist } from "@/components/TimeBoard/types";

const sarah: BoardTherapist = {
  id: 10,
  first_name: "Sarah",
  last_name: "Johnson",
  shifts: [],
  hours: [
    {
      hour: 9,
      division_id: "gym",
      shift_id: "s1",
      capacity: 4,
      appointments: [],
    },
    {
      hour: 10,
      division_id: "gym",
      shift_id: "s1",
      capacity: 4,
      appointments: [],
    },
    {
      hour: 14,
      division_id: "room",
      shift_id: "s2",
      capacity: 1,
      appointments: [],
    },
  ],
  patients: 3,
  capacity: 9,
};
const board: Board = {
  date: "2026-10-26",
  open_hour: 8,
  close_hour: 12,
  divisions: [
    {
      id: "gym",
      name: "Gym",
      color: "teal",
      capacity: 4,
      sort_order: 0,
      active: true,
    },
    {
      id: "room",
      name: "Room 1",
      color: "indigo",
      capacity: 1,
      sort_order: 1,
      active: true,
    },
  ],
  therapists: [sarah],
  permissions: { manage_shifts: true, book_patients: true },
};

describe("time board helpers", () => {
  it("lists the hour columns and formats them", () => {
    expect(boardHours(board)).toEqual([8, 9, 10, 11]);
    expect(formatHour(9)).toBe("09:00");
    expect(boardHours({ open_hour: 10, close_hour: 9 })).toEqual([]);
  });

  it("finds a physio's hour", () => {
    expect(hourOf(sarah, 10)?.division_id).toBe("gym");
    expect(hourOf(sarah, 12)).toBeUndefined();
  });

  it("classifies the load", () => {
    expect(loadLevel(0, 3)).toBe("free");
    expect(loadLevel(2, 3)).toBe("some");
    expect(loadLevel(3, 3)).toBe("full");
    expect(loadLevel(5, 3)).toBe("full");
  });

  it("moves between days", () => {
    expect(addDays("2026-10-30", 3)).toBe("2026-11-02");
    expect(addDays("2026-10-26", -1)).toBe("2026-10-25");
  });

  it("groups physios per division and hour", () => {
    const by = staffingByDivision(board);
    expect(
      by
        .get("gym")
        ?.get(9)
        ?.map((x) => x.therapist.first_name),
    ).toEqual(["Sarah"]);
    expect(by.get("room")?.get(14)?.length).toBe(1);
    expect(by.get("room")?.get(9)).toBeUndefined();
  });

  it("computes utilisation", () => {
    expect(utilisation(sarah)).toBe(33);
    expect(utilisation({ patients: 0, capacity: 0 })).toBe(0);
  });
});
