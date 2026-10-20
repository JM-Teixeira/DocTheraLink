import { PatientModel } from "@/components/Patient/models";
import {
  CreateReferralRequest,
  NoteRequest,
  ReasonRequest,
  Referral,
  ReferralNote,
} from "@/components/Referral/types";
import { UserModel } from "@/components/Users/models";

import { Type } from "@/Utils/request/api";
import { PaginatedResponse } from "@/Utils/request/types";

/**
 * API routes for the referral module (be/internal/modules/referral/routes.go).
 * Use with `query()` / `mutate()` from "@/Utils/request".
 */
const ReferralRoutes = {
  list: {
    path: "/api/v1/referrals/",
    method: "GET",
    TRes: Type<PaginatedResponse<Referral>>(),
  },
  create: {
    path: "/api/v1/referrals/",
    method: "POST",
    TBody: Type<CreateReferralRequest>(),
    TRes: Type<Referral>(),
  },
  get: {
    path: "/api/v1/referrals/{id}/",
    method: "GET",
    TRes: Type<Referral>(),
  },
  accept: {
    path: "/api/v1/referrals/{id}/accept/",
    method: "POST",
    TBody: Type<Record<string, never>>(),
    TRes: Type<Referral>(),
  },
  start: {
    path: "/api/v1/referrals/{id}/start/",
    method: "POST",
    TBody: Type<Record<string, never>>(),
    TRes: Type<Referral>(),
  },
  decline: {
    path: "/api/v1/referrals/{id}/decline/",
    method: "POST",
    TBody: Type<ReasonRequest>(),
    TRes: Type<Referral>(),
  },
  complete: {
    path: "/api/v1/referrals/{id}/complete/",
    method: "POST",
    TBody: Type<ReasonRequest>(),
    TRes: Type<Referral>(),
  },
  cancel: {
    path: "/api/v1/referrals/{id}/cancel/",
    method: "POST",
    TBody: Type<ReasonRequest>(),
    TRes: Type<Referral>(),
  },
  listNotes: {
    path: "/api/v1/referrals/{id}/notes/",
    method: "GET",
    TRes: Type<ReferralNote[]>(),
  },
  addNote: {
    path: "/api/v1/referrals/{id}/notes/",
    method: "POST",
    TBody: Type<NoteRequest>(),
    TRes: Type<ReferralNote>(),
  },

  // Lookups used by the create form
  getPatient: {
    path: "/api/v1/patient/{id}/",
    method: "GET",
    TRes: Type<PatientModel>(),
  },
  searchPatients: {
    path: "/api/v1/patient/",
    method: "GET",
    TRes: Type<PaginatedResponse<PatientModel>>(),
  },
  listTherapists: {
    path: "/api/v1/users/",
    method: "GET",
    TRes: Type<PaginatedResponse<UserModel>>(),
  },
} as const;

export default ReferralRoutes;
