import { describe, expect, it } from "vitest";

import {
  DRAFT_MAX_AGE_MS,
  EMPTY_PATIENT_FORM,
  clearAllDrafts,
  draftKey,
  duplicatesFromError,
  formToPayload,
  formatId,
  formatPostalCode,
  genderKey,
  patientToForm,
  readDraft,
  rowsToPayload,
  serverErrorsToForm,
  validNif,
  validSns,
  validatePatientForm,
  writeDraft,
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
      care_level: "5",
    });
    expect(body).toMatchObject({
      name: "Ana Silva",
      gender: 2,
      care_level: 5,
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
    expect(form.care_level).toBe("1"); // patients without a level count as 1
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

describe("Portuguese identifiers", () => {
  it("validates the NIF check digit", () => {
    for (const ok of ["123456789", "501964843", "100000010", "123 456 789"])
      expect(validNif(ok)).toBe(true);
    for (const bad of ["123456780", "12345678", "12345678a", ""])
      expect(validNif(bad)).toBe(false);
  });

  it("checks the SNS number format and formats for display", () => {
    expect(validSns("987.654.321")).toBe(true);
    expect(validSns("98765")).toBe(false);
    expect(formatId("123456789")).toBe("123 456 789");
    expect(formatId(null)).toBe(null);
  });

  it("flags invalid identifiers and sends them normalised", () => {
    const f = { ...EMPTY_PATIENT_FORM, name: "Ana Costa" };
    expect(validatePatientForm({ ...f, nif: "123456780" }).nif).toBe(
      "error__invalid_nif",
    );
    expect(validatePatientForm({ ...f, sns_number: "12" }).sns_number).toBe(
      "error__invalid_sns",
    );
    const body = formToPayload({ ...f, nif: "123 456 789", sns_number: "" });
    expect(body.nif).toBe("123456789");
    expect(body.sns_number).toBeNull();
  });
});

describe("duplicatesFromError", () => {
  const cause = {
    error: "duplicate_patient",
    blocking: false,
    duplicates: [
      { id: "p1", name: "Ana", reason: "name_dob", blocking: false },
    ],
  };
  it("reads the 409 answer", () => {
    expect(duplicatesFromError(409, cause)).toEqual({
      blocking: false,
      matches: cause.duplicates,
    });
  });
  it("ignores other errors", () => {
    expect(duplicatesFromError(400, cause)).toBeNull();
    expect(duplicatesFromError(409, { error: "conflict" })).toBeNull();
    expect(duplicatesFromError(409, undefined)).toBeNull();
  });
});

describe("registration extras", () => {
  it("checks e-mail and postal code, and formats the postal code", () => {
    const f = { ...EMPTY_PATIENT_FORM, name: "Ana" };
    expect(
      validatePatientForm({ ...f, email: "ana@", postal_code: "40001" }, today),
    ).toEqual({
      email: "error__invalid_email",
      postal_code: "error__invalid_postal_code",
    });
    expect(
      validatePatientForm(
        { ...f, email: "ana@ex.pt", postal_code: "4000 123" },
        today,
      ),
    ).toEqual({});
    expect(formatPostalCode("4000 123")).toBe("4000-123");
    expect(formatPostalCode("40")).toBe("40");
  });

  it("sends the new fields, empty ones as null", () => {
    const body = formToPayload({
      ...EMPTY_PATIENT_FORM,
      name: "Ana",
      city: " Porto ",
      emergency_contact_relation: "spouse",
    });
    expect(body.city).toBe("Porto");
    expect(body.emergency_contact_relation).toBe("spouse");
    expect(body.insurer).toBeNull();
  });
});

describe("form drafts", () => {
  const memory = () => {
    const m = new Map<string, string>();
    return {
      getItem: (k: string) => m.get(k) ?? null,
      setItem: (k: string, v: string) => void m.set(k, v),
      removeItem: (k: string) => void m.delete(k),
      key: (i: number) => [...m.keys()][i] ?? null,
      get length() {
        return m.size;
      },
      clear: () => m.clear(),
    } as Storage;
  };
  const key = draftKey(7);
  const filled = { ...EMPTY_PATIENT_FORM, name: "Ana", city: "Porto" };

  it("saves, restores and expires a draft", () => {
    const s = memory();
    writeDraft(s, key, filled, EMPTY_PATIENT_FORM, today);
    expect(readDraft(s, key, today)?.form.city).toBe("Porto");
    const later = new Date(today.getTime() + DRAFT_MAX_AGE_MS + 1);
    expect(readDraft(s, key, later)).toBeNull();
  });

  it("removes the draft when the form is back to its start", () => {
    const s = memory();
    writeDraft(s, key, filled, EMPTY_PATIENT_FORM, today);
    writeDraft(s, key, EMPTY_PATIENT_FORM, EMPTY_PATIENT_FORM, today);
    expect(s.length).toBe(0);
  });

  it("ignores broken data and keeps only known string fields", () => {
    const s = memory();
    s.setItem(key, "{not json");
    expect(readDraft(s, key, today)).toBeNull();
    s.setItem(
      key,
      JSON.stringify({
        saved_at: today.toISOString(),
        form: { name: "Ana", evil: "x", care_level: 3 },
      }),
    );
    const d = readDraft(s, key, today);
    expect(d?.form.name).toBe("Ana");
    expect(d?.form.care_level).toBe("1");
    expect("evil" in (d?.form ?? {})).toBe(false);
  });

  it("keeps drafts per user and patient, and clears them all on sign-out", () => {
    expect(draftKey(7, "p1")).not.toBe(draftKey(8, "p1"));
    const s = memory();
    s.setItem("other", "keep");
    writeDraft(s, draftKey(7), filled, EMPTY_PATIENT_FORM, today);
    writeDraft(s, draftKey(7, "p1"), filled, EMPTY_PATIENT_FORM, today);
    clearAllDrafts(s);
    expect(s.length).toBe(1);
    expect(s.getItem("other")).toBe("keep");
  });
});

describe("clinical lists", () => {
  it("turns rows into API items, dropping empty rows and cells", () => {
    expect(
      rowsToPayload({
        allergy_list: [
          { substance: " Penicilina ", reaction: "", severity: "severe" },
          { substance: " ", reaction: "x", severity: "" },
        ],
        medication_list: [
          { name: "Ibuprofeno", dose: "400 mg", frequency: "" },
        ],
        condition_list: [
          { name: "Asma", since: "2010" },
          { name: "HTA", since: "" },
        ],
      }),
    ).toEqual({
      allergy_list: [{ substance: "Penicilina", severity: "severe" }],
      medication_list: [{ name: "Ibuprofeno", dose: "400 mg" }],
      condition_list: [{ name: "Asma", since: 2010 }, { name: "HTA" }],
    });
  });

  it("reads lists from a patient and checks condition years", () => {
    const f = patientToForm({
      name: "Ana",
      allergy_list: [{ substance: "Látex" }],
      condition_list: [{ name: "Asma", since: 2010 }],
    });
    expect(f.allergy_list).toEqual([
      { substance: "Látex", reaction: "", severity: "" },
    ]);
    expect(f.condition_list[0].since).toBe("2010");
    expect(f.medication_list).toEqual([]);
    const bad = { ...f, condition_list: [{ name: "Asma", since: "2099" }] };
    expect(validatePatientForm(bad, today).condition_list).toBe(
      "error__invalid_year",
    );
  });

  it("restores list rows from a draft, keeping only known cells", () => {
    const s = new Map<string, string>();
    const storage = { getItem: (k: string) => s.get(k) ?? null };
    s.set(
      "k",
      JSON.stringify({
        saved_at: today.toISOString(),
        form: {
          name: "Ana",
          allergy_list: [{ substance: "Pó", evil: "x" }, "junk"],
          medication_list: "not a list",
        },
      }),
    );
    const d = readDraft(storage, "k", today);
    expect(d?.form.allergy_list).toEqual([
      { substance: "Pó", reaction: "", severity: "" },
    ]);
    expect(d?.form.medication_list).toEqual([]);
  });
});
