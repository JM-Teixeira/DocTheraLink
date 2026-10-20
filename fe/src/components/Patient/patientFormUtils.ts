/**
 * Pure helpers for the Klyniva patient form: converting between the API model
 * and form state, and client-side validation that mirrors the backend rules
 * (be/internal/modules/patient/service.go). The backend stays authoritative.
 */
import { PatientModel } from "@/components/Patient/models";

/** Gender codes stored by the backend (1..3). Labels are translation keys. */
export const GENDERS = [
  { id: 1, key: "gender__male" },
  { id: 2, key: "gender__female" },
  { id: 3, key: "gender__other" },
] as const;

export const BLOOD_GROUP_OPTIONS = [
  "A+",
  "A-",
  "B+",
  "B-",
  "AB+",
  "AB-",
  "O+",
  "O-",
] as const;

/** Form state: every field is a string so inputs stay controlled. */
export interface PatientFormState {
  name: string;
  gender: string; // "" | "1" | "2" | "3"
  date_of_birth: string; // YYYY-MM-DD
  phone_number: string;
  emergency_phone_number: string;
  nationality: string;
  address: string;
  blood_group: string;
  allergies: string;
  medical_history_details: string;
  ongoing_medication: string;
}

export const EMPTY_PATIENT_FORM: PatientFormState = {
  name: "",
  gender: "",
  date_of_birth: "",
  phone_number: "",
  emergency_phone_number: "",
  nationality: "",
  address: "",
  blood_group: "",
  allergies: "",
  medical_history_details: "",
  ongoing_medication: "",
};

/** Fills the form from an existing patient (edit mode). */
export function patientToForm(p: PatientModel): PatientFormState {
  return {
    name: p.name ?? "",
    gender: p.gender ? String(p.gender) : "",
    date_of_birth: p.date_of_birth ? p.date_of_birth.slice(0, 10) : "",
    phone_number: p.phone_number ?? "",
    emergency_phone_number: p.emergency_phone_number ?? "",
    nationality: p.nationality ?? "",
    address: p.address ?? "",
    blood_group: p.blood_group ?? "",
    allergies: p.allergies ?? "",
    medical_history_details: p.medical_history_details ?? "",
    ongoing_medication: p.ongoing_medication ?? "",
  };
}

/**
 * Builds the request body. Empty optional fields are sent as null so an edit
 * can clear them (the backend treats null as "clear").
 */
export function formToPayload(f: PatientFormState): Record<string, unknown> {
  const opt = (s: string) => (s.trim() ? s.trim() : null);
  return {
    name: f.name.trim(),
    gender: f.gender ? Number(f.gender) : null,
    date_of_birth: opt(f.date_of_birth),
    phone_number: opt(f.phone_number),
    emergency_phone_number: opt(f.emergency_phone_number),
    nationality: opt(f.nationality),
    address: opt(f.address),
    blood_group: opt(f.blood_group),
    allergies: opt(f.allergies),
    medical_history_details: opt(f.medical_history_details),
    ongoing_medication: opt(f.ongoing_medication),
  };
}

/** Same pattern as the backend's phone validation. */
const PHONE_RE = /^\+?[0-9 ()\-.]{6,20}$/;

/** Field → translation key of the error message. Empty object = valid. */
export type PatientFormErrors = Partial<Record<keyof PatientFormState, string>>;

export function validatePatientForm(
  f: PatientFormState,
  today = new Date(),
): PatientFormErrors {
  const errors: PatientFormErrors = {};
  if (f.name.trim().length < 2) errors.name = "error__name_min_2";
  if (f.date_of_birth) {
    const d = new Date(f.date_of_birth);
    if (Number.isNaN(d.getTime())) errors.date_of_birth = "error__invalid_date";
    else if (d > today) errors.date_of_birth = "error__date_in_future";
  }
  for (const k of ["phone_number", "emergency_phone_number"] as const) {
    if (f[k].trim() && !PHONE_RE.test(f[k].trim()))
      errors[k] = "error__invalid_phone";
  }
  return errors;
}

/**
 * Maps a backend validation error ({ details: { field: message } }) onto form
 * fields so server-side errors show next to the right input.
 */
export function serverErrorsToForm(cause: unknown): PatientFormErrors {
  const details = (cause as { details?: Record<string, string> } | undefined)
    ?.details;
  if (!details) return {};
  const out: PatientFormErrors = {};
  for (const [field, msg] of Object.entries(details)) {
    if (field in EMPTY_PATIENT_FORM) out[field as keyof PatientFormState] = msg;
  }
  return out;
}

/** Translation key for a gender code. */
export function genderKey(gender?: number | null) {
  return GENDERS.find((g) => g.id === gender)?.key ?? "gender__unknown";
}
