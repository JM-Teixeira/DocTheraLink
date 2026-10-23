import { afterEach, describe, expect, it } from "vitest";

import {
  getWorkingClinic,
  setWorkingClinic,
  workingClinicHeaders,
} from "@/Utils/workingClinic";

afterEach(() => localStorage.clear());

describe("working clinic (admins)", () => {
  it("is empty by default and sends nothing", () => {
    expect(getWorkingClinic()).toBe("");
    expect(workingClinicHeaders()).toEqual({});
  });
  it("is kept and sent as X-Clinic; empty clears it", () => {
    setWorkingClinic("c1");
    expect(workingClinicHeaders()).toEqual({ "X-Clinic": "c1" });
    setWorkingClinic("");
    expect(getWorkingClinic()).toBe("");
  });
});
