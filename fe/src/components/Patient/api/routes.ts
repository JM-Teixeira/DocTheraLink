import { Type } from "@/Utils/request/api";
import { PaginatedResponse } from "@/Utils/request/types";

import { PatientModel } from "../models";

export interface PatientStats {
  total: number;
  care_levels: Record<string, number>;
  genders: Record<string, number>;
  age_bands: Record<string, number>;
  average_age: number;
  /** Registrations per month, oldest first (last 12 months). */
  monthly?: { month: string; count: number }[];
}

/** One thing that happened to a patient (be/internal/modules/timeline). */
export interface TimelineEntry {
  at: string;
  kind: "registered" | "referral" | "note" | "file" | "session";
  /** referral step, note kind or session status */
  status?: string;
  referral_id?: string;
  item_id?: string;
  text?: string;
  who?: string;
  pain?: number;
  date?: string;
  hour?: number;
  justified?: boolean;
}

export interface PatientTimeline {
  upcoming: TimelineEntry[];
  history: TimelineEntry[];
}

/** One consent record (be/internal/modules/patient/consent.go). */
export interface Consent {
  id: string;
  patient_id: string;
  kind: ConsentKind;
  granted: boolean;
  method: ConsentMethod;
  note: string;
  recorded_by?: { id: number; first_name: string; last_name: string };
  created_date: string;
}
export const CONSENT_KINDS = [
  "data_processing",
  "share_with_doctor",
  "contact_email",
  "contact_sms",
  "photos",
] as const;
export type ConsentKind = (typeof CONSENT_KINDS)[number];
export const CONSENT_METHODS = ["signed_paper", "verbal", "digital"] as const;
export type ConsentMethod = (typeof CONSENT_METHODS)[number];

export interface ConsentView {
  current: Partial<Record<ConsentKind, Consent>>;
  history: Consent[];
}

export interface ConsentRequest {
  kind: ConsentKind;
  granted: boolean;
  method: ConsentMethod;
  note: string;
}

/**
 * Patient API Routes
 * Defines API endpoints for patient-related operations
 */
const PatientRoutes = {
  /**
   * Create a new patient
   * POST /api/v1/patient/
   */
  createPatient: {
    path: "/api/v1/patient/",
    method: "POST" as const,
    TBody: Type<Partial<PatientModel>>(),
    TRes: Type<PatientModel>(),
  },

  /**
   * Search for patients (duplicate check)
   * GET /api/v1/patient/search/
   */
  searchPatient: {
    path: "/api/v1/patient/search/",
    method: "GET" as const,
    TRes: Type<PaginatedResponse<PatientModel>>(),
  },

  /**
   * Get patient list
   * GET /api/v1/patient/
   */
  listPatients: {
    path: "/api/v1/patient/",
    method: "GET" as const,
    TRes: Type<PaginatedResponse<PatientModel>>(),
  },

  /**
   * Counts by care level, gender and age band for the same filters as the list
   * GET /api/v1/patient/stats/
   */
  patientStats: {
    path: "/api/v1/patient/stats/",
    method: "GET" as const,
    TRes: Type<PatientStats>(),
  },

  /**
   * Get patient details
   * GET /api/v1/patient/{id}/
   */
  /**
   * Referrals, messages, files and sessions of a patient
   * GET /api/v1/patient/{id}/timeline/
   */
  patientTimeline: {
    path: "/api/v1/patient/{id}/timeline/",
    method: "GET" as const,
    TRes: Type<PatientTimeline>(),
  },

  /** GET/POST /api/v1/patient/{id}/consents/ */
  patientConsents: {
    path: "/api/v1/patient/{id}/consents/",
    method: "GET" as const,
    TRes: Type<ConsentView>(),
  },
  recordConsent: {
    path: "/api/v1/patient/{id}/consents/",
    method: "POST" as const,
    TBody: Type<ConsentRequest>(),
    TRes: Type<ConsentView>(),
  },

  getPatient: {
    path: "/api/v1/patient/{id}/",
    method: "GET" as const,
    TRes: Type<PatientModel>(),
  },

  /**
   * Update patient
   * PUT /api/v1/patient/{id}/
   */
  updatePatient: {
    path: "/api/v1/patient/{id}/",
    method: "PUT" as const,
    TBody: Type<Partial<PatientModel>>(),
    TRes: Type<PatientModel>(),
  },

  /**
   * Partial update patient
   * PATCH /api/v1/patient/{id}/
   */
  patchPatient: {
    path: "/api/v1/patient/{id}/",
    method: "PATCH" as const,
    TBody: Type<Partial<PatientModel>>(),
    TRes: Type<PatientModel>(),
  },
} as const;

export default PatientRoutes;
