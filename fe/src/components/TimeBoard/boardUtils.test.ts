import { describe, expect, it } from "vitest";

import {
  absenceOf,
  addDays,
  boardHours,
  coveringFor,
  filterBoard,
  floorsOf,
  formatHour,
  hourIsFull,
  hourOf,
  loadLevel,
  monthGrid,
  staffingByDivision,
  utilisation,
  weekDates,
  weekStart,
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
      credits: 0,
      credit_limit: 8,
      appointments: [],
    },
    {
      hour: 10,
      division_id: "gym",
      shift_id: "s1",
      capacity: 4,
      credits: 0,
      credit_limit: 8,
      appointments: [],
    },
    {
      hour: 14,
      division_id: "room",
      shift_id: "s2",
      capacity: 1,
      credits: 0,
      credit_limit: 8,
      appointments: [],
    },
  ],
  patients: 3,
  capacity: 9,
  credits: 0,
  credit_capacity: 24,
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
      credit_limit: 8,
      floor: 1,
      sort_order: 0,
      active: true,
    },
    {
      id: "room",
      name: "Room 1",
      color: "indigo",
      capacity: 1,
      credit_limit: 8,
      floor: 1,
      sort_order: 1,
      active: true,
    },
  ],
  therapists: [sarah],
  permissions: { manage_shifts: true, book_patients: true },
  absences: [],
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

  it("computes utilisation from credits", () => {
    expect(utilisation({ credits: 6, credit_capacity: 24 })).toBe(25);
    expect(utilisation({ credits: 0, credit_capacity: 0 })).toBe(0);
  });

  it("knows when an hour is full by patients or credits", () => {
    const h = { hour: 9, division_id: "g", shift_id: "s", appointments: [] };
    expect(hourIsFull({ ...h, capacity: 4, credits: 5, credit_limit: 8 })).toBe(
      false,
    );
    expect(hourIsFull({ ...h, capacity: 4, credits: 8, credit_limit: 8 })).toBe(
      true,
    );
    expect(hourIsFull({ ...h, capacity: 0, credits: 0, credit_limit: 8 })).toBe(
      true,
    );
  });

  it("finds weeks and builds the month calendar", () => {
    expect(weekStart("2026-10-21")).toBe("2026-10-19"); // Wednesday → Monday
    expect(weekStart("2026-10-25")).toBe("2026-10-19"); // Sunday → previous Monday
    expect(weekDates("2026-10-19")).toHaveLength(7);
    const october = monthGrid("2026-10-15");
    expect(october[0][0]).toBe("2026-09-28");
    expect(october.at(-1)?.at(-1)).toBe("2026-11-01");
    expect(october).toHaveLength(5);
  });
});

describe("board filters", () => {
  const twoFloors: Board = {
    ...board,
    divisions: [
      { ...board.divisions[0], floor: 1 },
      { ...board.divisions[1], floor: 2 },
    ],
    therapists: [
      sarah,
      { ...sarah, id: 11, first_name: "Tiago", hours: [sarah.hours[0]] },
    ],
  };

  it("lists the floors", () => {
    expect(floorsOf(twoFloors.divisions)).toEqual([1, 2]);
  });

  it("keeps only a floor's divisions, hours and physios", () => {
    const f = filterBoard(twoFloors, { floor: 2, divisions: [], physios: [] });
    expect(f.divisions.map((d) => d.id)).toEqual(["room"]);
    expect(f.therapists.map((t) => t.id)).toEqual([10]);
    expect(f.therapists[0].hours.map((h) => h.hour)).toEqual([14]);
  });

  it("filters by physio without hiding their other hours", () => {
    const f = filterBoard(twoFloors, {
      floor: null,
      divisions: [],
      physios: [11],
    });
    expect(f.therapists.map((t) => t.id)).toEqual([11]);
    expect(f.divisions).toHaveLength(2);
  });

  it("filters by division", () => {
    const f = filterBoard(twoFloors, {
      floor: null,
      divisions: ["gym"],
      physios: [],
    });
    expect(f.therapists.map((t) => t.hours.length)).toEqual([2, 1]);
  });

  it("finds absences and covers", () => {
    const abs = [
      {
        id: "a",
        therapist_id: 10,
        cover_id: 11,
        from: "",
        to: "",
        reason: "vacation" as const,
        handover: "",
        notes: [],
      },
    ];
    expect(absenceOf(abs, 10)?.cover_id).toBe(11);
    expect(absenceOf(abs, 11)).toBeUndefined();
    expect(coveringFor(abs, 11)).toHaveLength(1);
    expect(coveringFor(undefined, 11)).toEqual([]);
  });
});
