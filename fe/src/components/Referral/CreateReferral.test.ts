import { describe, expect, it } from "vitest";

import { toRequest } from "@/components/Referral/CreateReferral";

const empty = {
  diagnosis: "",
  icd_code: "",
  reason: "",
  goals: "",
  precautions: "",
  recommended_sessions: "",
  priority: "routine" as const,
  therapist_id: "",
};

describe("toRequest", () => {
  it("trims text and drops empty optional fields", () => {
    const body = toRequest("p1", {
      ...empty,
      diagnosis: "  Low back pain ",
      reason: " Core strengthening ",
      goals: "   ",
    });
    expect(body).toEqual({
      patient_id: "p1",
      diagnosis: "Low back pain",
      reason: "Core strengthening",
      priority: "routine",
      icd_code: undefined,
      goals: undefined,
      precautions: undefined,
      recommended_sessions: undefined,
      therapist_id: undefined,
    });
  });

  it("converts numeric fields", () => {
    const body = toRequest("p1", {
      ...empty,
      diagnosis: "x",
      reason: "y",
      recommended_sessions: "12",
      therapist_id: "10",
      priority: "urgent",
    });
    expect(body.recommended_sessions).toBe(12);
    expect(body.therapist_id).toBe(10);
    expect(body.priority).toBe("urgent");
  });
});
