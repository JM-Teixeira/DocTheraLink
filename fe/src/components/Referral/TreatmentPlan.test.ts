import { describe, expect, it } from "vitest";

import {
  toItems,
  toRows,
  totalMinutes,
} from "@/components/Referral/TreatmentPlan";

describe("treatment plan rows", () => {
  it("round-trips items, drops empty rows and empty cells", () => {
    const items = [
      { activity: "Ultrassons", minutes: 10, equipment: "Máquina 2" },
      { activity: "Codman", dose: "3x10" },
    ];
    const rows = toRows(items);
    expect(rows[1]).toEqual({
      activity: "Codman",
      minutes: "",
      dose: "3x10",
      equipment: "",
      notes: "",
    });
    expect(
      toItems([
        ...rows,
        { activity: "  ", minutes: "5", dose: "", equipment: "", notes: "" },
      ]),
    ).toEqual(items);
    expect(
      toItems([
        { activity: "x", minutes: "0", dose: " ", equipment: "", notes: "" },
      ]),
    ).toEqual([{ activity: "x" }]);
  });

  it("adds the minutes", () => {
    expect(
      totalMinutes([
        { activity: "a", minutes: 10 },
        { activity: "b" },
        { activity: "c", minutes: 5 },
      ]),
    ).toBe(15);
  });
});
