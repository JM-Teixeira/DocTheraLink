import { Type } from "@/Utils/request/api";
import { PaginatedResponse } from "@/Utils/request/types";

import { PatientModel } from "../models";

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
   * Get patient details
   * GET /api/v1/patient/{id}/
   */
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
