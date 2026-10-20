import { describe, expect, it } from "vitest";

import {
  EMPTY_PATIENT_FORM,
  formToPayload,
  genderKey,
  patientToForm,
  serverErrorsToForm,
  validatePatientForm,
} from "@/components/Patient/patientFormUtils";

const today = new Date("2026-10-21T12:00:00Z");

describe("validatePatientForm", () => {
  it("requires a name of at least 2 characters", () => {
    expect(
      validatePatientForm({ ...EMPTY_PATIENT_FORM, name: " A " }, today),
    ).toEqual({
      name: "error__name_min_2",
    });
    expect(
      validatePatientForm({ ...EMPTY_PATIENT_FORM, name: "Ana" }, today),
    ).toEqual({});
  });

  it("rejects future or invalid birth dates", () => {
    const base = { ...EMPTY_PATIENT_FORM, name: "Ana" };
    expect(
      validatePatientForm({ ...base, date_of_birth: "2030-01-01" }, today)
        .date_of_birth,
    ).toBe("error__date_in_future");
    expect(
      validatePatientForm({ ...base, date_of_birth: "nope" }, today)
        .date_of_birth,
    ).toBe("error__invalid_date");
    expect(
      validatePatientForm({ ...base, date_of_birth: "1980-05-02" }, today),
    ).toEqual({});
  });

  it("validates phone numbers like the backend", () => {
    const base = { ...EMPTY_PATIENT_FORM, name: "Ana" };
    expect(
      validatePatientForm({ ...base, phone_number: "+351 912 345 678" }, today),
    ).toEqual({});
    expect(
      validatePatientForm(
        { ...base, emergency_phone_number: "call me" },
        today,
      ),
    ).toEqual({
      emergency_phone_number: "error__invalid_phone",
    });
  });
});

describe("form ↔ payload", () => {
  it("trims, converts gender and sends null for empty optionals", () => {
    const body = formToPayload({
      ...EMPTY_PATIENT_FORM,
      name: " Ana Silva ",
      gender: "2",
      allergies: "  ",
    });
    expect(body).toMatchObject({
      name: "Ana Silva",
      gender: 2,
      allergies: null,
      date_of_birth: null,
    });
  });

  it("round-trips an existing patient", () => {
    const form = patientToForm({
      id: "p1",
      name: "Rui",
      gender: 1,
      date_of_birth: "1975-03-14T00:00:00Z",
      phone_number: "+351 910 000 000",
      blood_group: "O+",
    });
    expect(form.date_of_birth).toBe("1975-03-14");
    expect(formToPayload(form)).toMatchObject({
      name: "Rui",
      gender: 1,
      date_of_birth: "1975-03-14",
      blood_group: "O+",
    });
  });
});

describe("helpers", () => {
  it("maps backend field errors onto known form fields only", () => {
    expect(
      serverErrorsToForm({
        details: { phone_number: "is not a valid phone number", id: "x" },
      }),
    ).toEqual({ phone_number: "is not a valid phone number" });
    expect(serverErrorsToForm(undefined)).toEqual({});
  });

  it("returns translation keys for gender codes", () => {
    expect(genderKey(2)).toBe("gender__female");
    expect(genderKey(undefined)).toBe("gender__unknown");
  });
});
