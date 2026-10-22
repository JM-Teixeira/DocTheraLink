import { describe, expect, it } from "vitest";

import {
  filterThread,
  painSeries,
  sparkPoints,
  waitingFor,
} from "@/components/Referral/threadUtils";
import { ReferralNote } from "@/components/Referral/types";

const note = (p: Partial<ReferralNote>): ReferralNote => ({
  id: p.id ?? Math.random().toString(),
  referral_id: "r",
  author_id: 1,
  kind: "progress",
  body: "",
  created_date: "2026-10-01T10:00:00Z",
  ...p,
});

const thread = [
  note({
    id: "p1",
    kind: "progress",
    pain: 7,
    created_date: "2026-10-01T10:00:00Z",
    author_id: 10,
  }),
  note({ id: "q1", kind: "question", author_id: 1, answered: false }),
  note({ id: "s1", kind: "system" }),
  note({
    id: "p2",
    kind: "progress",
    pain: 4,
    created_date: "2026-10-08T10:00:00Z",
    author_id: 10,
  }),
  note({ id: "r1", kind: "reply", reply_to_id: "q0", author_id: 10 }),
  note({
    id: "p3",
    kind: "progress",
    created_date: "2026-10-09T10:00:00Z",
    author_id: 10,
  }),
];

describe("filterThread", () => {
  it("keeps progress notes or the questions with replies", () => {
    expect(filterThread(thread, "progress").map((n) => n.id)).toEqual([
      "p1",
      "p2",
      "p3",
    ]);
    expect(filterThread(thread, "questions").map((n) => n.id)).toEqual([
      "q1",
      "r1",
    ]);
    expect(filterThread(thread, "all")).toHaveLength(6);
  });
});

describe("painSeries", () => {
  it("uses only progress notes with a pain score, oldest first", () => {
    expect(painSeries([...thread].reverse()).map((p) => p.pain)).toEqual([
      7, 4,
    ]);
  });
});

describe("waitingFor", () => {
  it("lists unanswered questions asked by others", () => {
    expect(waitingFor(thread, 10).map((n) => n.id)).toEqual(["q1"]);
    expect(waitingFor(thread, 1)).toHaveLength(0);
  });
});

describe("sparkPoints", () => {
  it("maps 0-10 to the box, clamping out-of-range values", () => {
    expect(sparkPoints([10, 0, 5], 100, 20)).toBe("0,0 50,20 100,10");
    expect(sparkPoints([12], 100, 20)).toBe("0,0");
    expect(sparkPoints([], 100, 20)).toBe("");
  });
});
