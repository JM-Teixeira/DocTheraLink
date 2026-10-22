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
  emergency_contact_name: string;
  emergency_contact_relation: string; // one of RELATIONS or ""
  email: string;
  nationality: string;
  nif: string;
  sns_number: string;
  care_level: string; // "1" … "5"
  address: string; // street line
  postal_code: string; // 0000-000
  city: string;
  insurer: string;
  insurance_number: string;
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
  emergency_contact_name: "",
  emergency_contact_relation: "",
  email: "",
  nationality: "",
  nif: "",
  sns_number: "",
  care_level: "1",
  address: "",
  postal_code: "",
  city: "",
  insurer: "",
  insurance_number: "",
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
    emergency_contact_name: p.emergency_contact_name ?? "",
    emergency_contact_relation: p.emergency_contact_relation ?? "",
    email: p.email ?? "",
    nationality: p.nationality ?? "",
    nif: p.nif ?? "",
    sns_number: p.sns_number ?? "",
    care_level: String(p.care_level ?? 1),
    address: p.address ?? "",
    postal_code: p.postal_code ?? "",
    city: p.city ?? "",
    insurer: p.insurer ?? "",
    insurance_number: p.insurance_number ?? "",
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
    emergency_contact_name: opt(f.emergency_contact_name),
    emergency_contact_relation: opt(f.emergency_contact_relation),
    email: opt(f.email),
    nationality: opt(f.nationality),
    nif: opt(normalizeId(f.nif)),
    sns_number: opt(normalizeId(f.sns_number)),
    care_level: Number(f.care_level) || 1,
    address: opt(f.address),
    postal_code: opt(f.postal_code),
    city: opt(f.city),
    insurer: opt(f.insurer),
    insurance_number: opt(f.insurance_number),
    blood_group: opt(f.blood_group),
    allergies: opt(f.allergies),
    medical_history_details: opt(f.medical_history_details),
    ongoing_medication: opt(f.ongoing_medication),
  };
}

/** Removes the separators people type in identifiers ("123 456 789"). */
export const normalizeId = (s: string) => s.replace(/[\s.-]/g, "");

/**
 * Portuguese NIF: 9 digits, the last a mod-11 check digit (weights 9..2).
 * Mirrors ValidNIF in be/internal/modules/patient/identifiers.go.
 */
export function validNif(raw: string) {
  const nif = normalizeId(raw);
  if (!/^\d{9}$/.test(nif)) return false;
  let sum = 0;
  for (let i = 0; i < 8; i++) sum += Number(nif[i]) * (9 - i);
  const check = 11 - (sum % 11);
  return Number(nif[8]) === (check >= 10 ? 0 : check);
}

/** Groups a stored 9-digit identifier for display: "123 456 789". */
export const formatId = (id?: string | null) =>
  id ? id.replace(/^(\d{3})(\d{3})(\d{3})$/, "$1 $2 $3") : id;

/** SNS number (nº de utente): 9 digits; its check rule is not public. */
export const validSns = (raw: string) => /^\d{9}$/.test(normalizeId(raw));

/** Same patterns as the backend (patient/service.go). */
const PHONE_RE = /^\+?[0-9 ()\-.]{6,20}$/;
const EMAIL_RE = /^[^@\s]+@[^@\s]+\.[^@\s]+$/;
const POSTAL_RE = /^\d{4}-?\d{3}$/;

/** Who the emergency contact is; labels are relation__<id>. */
export const RELATIONS = [
  "spouse",
  "parent",
  "child",
  "sibling",
  "relative",
  "friend",
  "caregiver",
  "other",
] as const;

/** "4000123" or "4000 123" → "4000-123"; anything else is left as typed. */
export function formatPostalCode(raw: string) {
  const s = raw.replace(/\s/g, "");
  return /^\d{7}$/.test(s) ? `${s.slice(0, 4)}-${s.slice(4)}` : raw;
}

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
  if (f.nif.trim() && !validNif(f.nif)) errors.nif = "error__invalid_nif";
  if (f.sns_number.trim() && !validSns(f.sns_number))
    errors.sns_number = "error__invalid_sns";
  for (const k of ["phone_number", "emergency_phone_number"] as const) {
    if (f[k].trim() && !PHONE_RE.test(f[k].trim()))
      errors[k] = "error__invalid_phone";
  }
  if (f.email.trim() && !EMAIL_RE.test(f.email.trim()))
    errors.email = "error__invalid_email";
  if (f.postal_code.trim() && !POSTAL_RE.test(f.postal_code.replace(/\s/g, "")))
    errors.postal_code = "error__invalid_postal_code";
  return errors;
}

/**
 * Drafts of the patient form, so a half-filled registration survives a
 * closed tab or a lost connection. They hold patient data, so they belong to
 * one user, expire after a day and are removed on sign-out.
 */
export const DRAFT_PREFIX = "klyniva:patient-draft:";
export const DRAFT_MAX_AGE_MS = 24 * 60 * 60 * 1000;

export interface PatientDraft {
  saved_at: string; // ISO time
  form: PatientFormState;
}

/** One draft per user and per patient ("new" for a registration). */
export const draftKey = (userId: string | number, patientId?: string) =>
  `${DRAFT_PREFIX}${userId}:${patientId ?? "new"}`;

/** True when two forms hold the same values. */
export const sameForm = (a: PatientFormState, b: PatientFormState) =>
  (Object.keys(EMPTY_PATIENT_FORM) as (keyof PatientFormState)[]).every(
    (k) => a[k] === b[k],
  );

/** Reads a draft; old, broken or foreign data gives null. */
export function readDraft(
  storage: Pick<Storage, "getItem"> | undefined,
  key: string,
  now = new Date(),
): PatientDraft | null {
  try {
    const raw = storage?.getItem(key);
    if (!raw) return null;
    const d = JSON.parse(raw) as Partial<PatientDraft>;
    const at = new Date(d.saved_at ?? "");
    if (!d.form || Number.isNaN(at.getTime())) return null;
    if (now.getTime() - at.getTime() > DRAFT_MAX_AGE_MS) return null;
    // Only known fields, as strings (new fields keep their empty value)
    const form = { ...EMPTY_PATIENT_FORM };
    for (const k of Object.keys(form) as (keyof PatientFormState)[]) {
      const v = (d.form as unknown as Record<string, unknown>)[k];
      if (typeof v === "string") form[k] = v;
    }
    return { saved_at: at.toISOString(), form };
  } catch {
    return null;
  }
}

/** Saves the form; an untouched form removes the draft instead. */
export function writeDraft(
  storage: Pick<Storage, "setItem" | "removeItem"> | undefined,
  key: string,
  form: PatientFormState,
  base: PatientFormState,
  now = new Date(),
) {
  try {
    if (sameForm(form, base)) storage?.removeItem(key);
    else
      storage?.setItem(
        key,
        JSON.stringify({ saved_at: now.toISOString(), form }),
      );
  } catch {
    // storage full or blocked: the form still works, just without a draft
  }
}

/** Removes every patient draft (on sign-out). */
export function clearAllDrafts(storage: Storage | undefined) {
  try {
    if (!storage) return;
    const keys: string[] = [];
    for (let i = 0; i < storage.length; i++) {
      const k = storage.key(i);
      if (k?.startsWith(DRAFT_PREFIX)) keys.push(k);
    }
    keys.forEach((k) => storage.removeItem(k));
  } catch {
    // nothing to clear
  }
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

/** An existing patient the backend thinks may be the same person (409). */
export interface PossibleDuplicate {
  id: string;
  name: string;
  date_of_birth?: string | null;
  phone_number?: string | null;
  reason: "nif" | "sns_number" | "name_dob" | "name_phone";
  blocking: boolean;
}

/** Reads the 409 duplicate_patient answer; null for any other error. */
export function duplicatesFromError(
  status: number | undefined,
  cause: unknown,
): { blocking: boolean; matches: PossibleDuplicate[] } | null {
  const c = cause as
    | { error?: string; blocking?: boolean; duplicates?: PossibleDuplicate[] }
    | undefined;
  if (status !== 409 || c?.error !== "duplicate_patient") return null;
  return { blocking: !!c.blocking, matches: c.duplicates ?? [] };
}

/** Translation key for a gender code. */
export function genderKey(gender?: number | null) {
  return GENDERS.find((g) => g.id === gender)?.key ?? "gender__unknown";
}
