import { describe, expect, it } from "vitest";

import {
  applyTemplate,
  hasClinicalText,
  toRequest,
} from "@/components/Referral/CreateReferral";
import { catalogLang } from "@/components/Referral/referralUtils";
import { ReferralTemplate } from "@/components/Referral/types";

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

const lbp: ReferralTemplate = {
  key: "low-back-pain",
  name: "Low back pain",
  icd_code: "M54.5",
  diagnosis: "Non-specific low back pain",
  reason: "Graded exercise",
  goals: "Return to work",
  precautions: "Red flags",
  priority: "routine",
  recommended_sessions: 10,
};

describe("templates", () => {
  it("fills the clinical fields and keeps the chosen therapist", () => {
    const f = applyTemplate(
      { ...empty, therapist_id: "10", priority: "urgent", goals: "old" },
      lbp,
    );
    expect(f).toEqual({
      diagnosis: "Non-specific low back pain",
      icd_code: "M54.5",
      reason: "Graded exercise",
      goals: "Return to work",
      precautions: "Red flags",
      recommended_sessions: "10",
      priority: "routine",
      therapist_id: "10",
    });
  });

  it("asks before replacing typed clinical text only", () => {
    expect(hasClinicalText(empty)).toBe(false);
    expect(
      hasClinicalText({ ...empty, therapist_id: "3", priority: "urgent" }),
    ).toBe(false);
    expect(hasClinicalText({ ...empty, reason: "  " })).toBe(false);
    expect(hasClinicalText({ ...empty, icd_code: "M54" })).toBe(true);
  });
});

describe("catalogLang", () => {
  it("maps UI languages to catalog languages", () => {
    expect(catalogLang("pt")).toBe("pt");
    expect(catalogLang("pt-PT")).toBe("pt");
    expect(catalogLang("en")).toBe("en");
    expect(catalogLang(undefined)).toBe("en");
  });
});
