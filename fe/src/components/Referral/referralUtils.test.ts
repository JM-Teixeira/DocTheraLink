import { describe, expect, it } from "vitest";

import {
  ageFromDob,
  allowedActions,
  canCreateReferral,
  canPostNote,
  fullName,
  hasDoctorTitle,
  initials,
  stepIndex,
  sumCounts,
} from "@/components/Referral/referralUtils";
import { Referral } from "@/components/Referral/types";

const doctor = { id: 1, user_type: "Doctor" };
const otherDoctor = { id: 2, user_type: "Doctor" };
const physio = { id: 10, user_type: "PhysioTherapist" };
const otherPhysio = { id: 11, user_type: "PhysioTherapist" };
const admin = { id: 99, user_type: "StateAdmin" };
const nurse = { id: 20, user_type: "Nurse" };

function referral(over: Partial<Referral> = {}): Referral {
  return {
    id: "r1",
    patient_id: "p1",
    doctor_id: doctor.id,
    therapist_id: null,
    diagnosis: "ACL",
    reason: "Rehab",
    priority: "routine",
    status: "pending",
    created_date: "2026-10-01T10:00:00Z",
    modified_date: "2026-10-01T10:00:00Z",
    ...over,
  };
}

describe("allowedActions", () => {
  it("lets any therapist accept an open referral but not decline it", () => {
    expect(allowedActions(referral(), physio)).toEqual(["accept"]);
  });

  it("only the addressed therapist acts on an addressed referral", () => {
    const r = referral({ therapist_id: physio.id });
    expect(allowedActions(r, physio)).toEqual([
      "accept",
      "decline",
      "reassign",
    ]);
    expect(allowedActions(r, otherPhysio)).toEqual([]);
  });

  it("walks the happy path for the assigned therapist", () => {
    expect(
      allowedActions(
        referral({ status: "accepted", therapist_id: physio.id }),
        physio,
      ),
    ).toEqual(["start", "reassign"]);
    expect(
      allowedActions(
        referral({ status: "in_progress", therapist_id: physio.id }),
        physio,
      ),
    ).toEqual(["complete", "reassign"]);
  });

  it("lets the referring doctor and admins hand over and cancel open referrals", () => {
    expect(allowedActions(referral(), doctor)).toEqual(["reassign", "cancel"]);
    expect(allowedActions(referral(), admin)).toEqual(["reassign", "cancel"]);
    expect(allowedActions(referral(), otherDoctor)).toEqual([]);
    expect(allowedActions(referral({ status: "completed" }), doctor)).toEqual(
      [],
    );
  });

  it("lets the referring doctor re-open a declined referral", () => {
    const r = referral({ status: "declined", therapist_id: physio.id });
    expect(allowedActions(r, doctor)).toEqual(["reopen"]);
    expect(allowedActions(r, admin)).toEqual(["reopen"]);
    expect(allowedActions(r, physio)).toEqual([]);
    expect(allowedActions(r, otherDoctor)).toEqual([]);
  });

  it("gives other roles nothing", () => {
    expect(allowedActions(referral(), nurse)).toEqual([]);
  });
});

describe("permissions", () => {
  it("only doctors and admins create referrals", () => {
    expect(canCreateReferral(doctor)).toBe(true);
    expect(canCreateReferral(admin)).toBe(true);
    expect(canCreateReferral(physio)).toBe(false);
  });

  it("only participants post notes", () => {
    const r = referral({ therapist_id: physio.id });
    expect(canPostNote(r, doctor)).toBe(true);
    expect(canPostNote(r, physio)).toBe(true);
    expect(canPostNote(r, otherPhysio)).toBe(false);
    expect(canPostNote(r, nurse)).toBe(false);
  });
});

describe("sumCounts", () => {
  it("adds up per-status counts", () => {
    expect(sumCounts({ pending: 2, accepted: 1, completed: 4 })).toBe(7);
    expect(sumCounts({})).toBe(0);
  });
});

describe("formatting helpers", () => {
  it("builds names and initials", () => {
    const u = {
      id: 1,
      username: "sj",
      first_name: "Sarah",
      last_name: "Johnson",
      user_type: "PhysioTherapist",
    };
    expect(fullName(u)).toBe("Sarah Johnson");
    expect(fullName({ ...u, first_name: "", last_name: "" })).toBe("sj");
    expect(fullName(undefined)).toBe("");
    expect(initials("Sarah  Johnson Smith")).toBe("SJ");
  });

  it("detects names that already carry a doctor title", () => {
    expect(hasDoctorTitle("Dr. John Smith")).toBe(true);
    expect(hasDoctorTitle("dr Ana")).toBe(true);
    expect(hasDoctorTitle("Drew Barry")).toBe(false);
  });

  it("computes age around the birthday", () => {
    const now = new Date("2026-10-20T12:00:00Z");
    expect(ageFromDob("1980-10-20", now)).toBe(46);
    expect(ageFromDob("1980-10-21", now)).toBe(45);
    expect(ageFromDob(null, now)).toBeNull();
    expect(ageFromDob("not a date", now)).toBeNull();
  });

  it("maps statuses onto the progress tracker", () => {
    expect(stepIndex("pending")).toBe(0);
    expect(stepIndex("completed")).toBe(3);
    expect(stepIndex("declined")).toBe(-1);
  });
});
