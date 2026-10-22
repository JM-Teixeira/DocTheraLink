import { describe, expect, it } from "vitest";

import {
  canChangeStock,
  roleKey,
  staffQuery,
  statusOf,
  workSteps,
} from "@/components/Staff/staffUtils";

describe("staffQuery", () => {
  const base = {
    search: "",
    role: "",
    clinic: "",
    status: "all" as const,
    limit: 20,
    offset: 0,
  };

  it("leaves out empty filters and 'all'", () => {
    expect(staffQuery(base)).toEqual({ limit: 20, offset: 0 });
  });

  it("maps every filter to its API name", () => {
    expect(
      staffQuery({
        ...base,
        search: "  inês ",
        role: "Doctor",
        clinic: "c1",
        status: "left",
        offset: 40,
      }),
    ).toEqual({
      limit: 20,
      offset: 40,
      search: "inês",
      user_type: "Doctor",
      home_facility: "c1",
      employment_status: "left",
    });
  });
});

describe("statusOf", () => {
  it("treats a missing status as active", () => {
    expect(statusOf(undefined)).toBe("active");
    expect(statusOf("")).toBe("active");
    expect(statusOf("left")).toBe("left");
    expect(statusOf("on_leave")).toBe("on_leave");
  });
});

it("roleKey prefixes the role", () => {
  expect(roleKey("Doctor")).toBe("role__Doctor");
});

describe("canChangeStock", () => {
  const gloves = (floors: number[], categories: string[]) => [
    { floors, categories },
  ];
  it("counts all-floor and all-category rules and the managers", () => {
    expect(canChangeStock("Nurse", gloves([], []), 1, "Luvas")).toBe(true);
    expect(canChangeStock("PhysioTherapist", gloves([1], []), 1, "Luvas")).toBe(
      true,
    );
    expect(canChangeStock("PhysioTherapist", gloves([2], []), 1, "Luvas")).toBe(
      false,
    );
    expect(
      canChangeStock("PhysioTherapist", gloves([], ["luvas"]), 3, "Luvas"),
    ).toBe(true);
    expect(
      canChangeStock("PhysioTherapist", gloves([1], ["Higiene"]), 1, "Luvas"),
    ).toBe(false);
    expect(canChangeStock("PhysioTherapist", [], 1, "")).toBe(false);
    expect(canChangeStock("Coordinator", [], 1, "Luvas")).toBe(true);
  });
  it("treats an empty filter as any floor or any category", () => {
    expect(
      canChangeStock(
        "PhysioTherapist",
        gloves([2], ["Higiene"]),
        null,
        "Higiene",
      ),
    ).toBe(true);
    expect(
      canChangeStock("PhysioTherapist", gloves([2], ["Higiene"]), 2, ""),
    ).toBe(true);
  });
});

describe("workSteps", () => {
  it("joined, a leave that ended, then left", () => {
    expect(
      workSteps({
        joined_on: "2026-10-01",
        events: [
          { status: "on_leave", since: "2026-10-10" },
          { status: "active", since: "2026-11-02" },
          { status: "left", since: "2026-12-31", note: "Fim de contrato" },
        ],
      }),
    ).toEqual([
      { kind: "joined", from: "2026-10-01" },
      { kind: "leave", from: "2026-10-10", until: "2026-11-02" },
      { kind: "left", from: "2026-12-31", note: "Fim de contrato" },
    ]);
  });
  it("an open leave and a return after leaving", () => {
    expect(
      workSteps({
        joined_on: "2025-01-01",
        events: [
          { status: "left", since: "2025-06-01" },
          { status: "active", since: "2026-01-05" },
          { status: "on_leave", since: "2026-10-10" },
        ],
      }).map((s) => [s.kind, s.from, s.until]),
    ).toEqual([
      ["joined", "2025-01-01", undefined],
      ["left", "2025-06-01", undefined],
      ["back", "2026-01-05", undefined],
      ["leave", "2026-10-10", undefined],
    ]);
  });
});
