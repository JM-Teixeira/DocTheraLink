import { describe, expect, it } from "vitest";

import { TimelineEntry } from "@/components/Patient/api/routes";
import {
  countKinds,
  filterEntries,
  groupByMonth,
  monthOf,
  titleKey,
} from "@/components/Patient/timelineUtils";

const e = (over: Partial<TimelineEntry>): TimelineEntry => ({
  at: "2026-10-10T10:00:00Z",
  kind: "note",
  ...over,
});

describe("patient timeline", () => {
  const entries = [
    e({ kind: "session", status: "missed", date: "2026-10-12", hour: 9 }),
    e({ kind: "note", status: "progress" }),
    e({ kind: "referral", status: "created", at: "2026-09-02T10:00:00Z" }),
    e({ kind: "registered", at: "2026-09-01T10:00:00Z" }),
  ];

  it("filters and counts by kind", () => {
    expect(filterEntries(entries, "all")).toHaveLength(4);
    expect(filterEntries(entries, "session")).toHaveLength(1);
    expect(countKinds(entries)).toEqual({
      all: 4,
      referral: 1,
      note: 1,
      file: 0,
      session: 1,
    });
  });

  it("groups by month, sessions by their day", () => {
    expect(monthOf(entries[0])).toBe("2026-10");
    expect(
      groupByMonth(entries).map((g) => [g.month, g.entries.length]),
    ).toEqual([
      ["2026-10", 2],
      ["2026-09", 2],
    ]);
  });

  it("names each entry", () => {
    const today = "2026-10-20";
    expect(titleKey(entries[0], today)).toBe("timeline__session__missed");
    expect(titleKey(entries[2], today)).toBe("timeline__referral__created");
    expect(titleKey(entries[3], today)).toBe("timeline__registered");
    expect(titleKey(e({ kind: "file" }), today)).toBe("timeline__file");
    const booked = e({ kind: "session", status: "booked" });
    expect(titleKey({ ...booked, date: "2026-10-01" }, today)).toBe(
      "timeline__session__unrecorded",
    );
    expect(titleKey({ ...booked, date: "2026-10-25" }, today)).toBe(
      "timeline__session__booked",
    );
  });
});
