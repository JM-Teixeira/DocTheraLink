import { FacilityModel } from "@/components/Facility/models";
import { PatientModel } from "@/components/Patient/models";
import { Attachment } from "@/components/Referral/attachments";
import {
  CreateReferralRequest,
  IcdCode,
  NoteRequest,
  ReasonRequest,
  ReassignRequest,
  Referral,
  ReferralCounts,
  ReferralNote,
  ReferralTemplate,
  ReopenRequest,
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
  counts: {
    path: "/api/v1/referrals/counts/",
    method: "GET",
    TRes: Type<ReferralCounts>(),
  },
  reassign: {
    path: "/api/v1/referrals/{id}/reassign/",
    method: "POST",
    TBody: Type<ReassignRequest>(),
    TRes: Type<Referral>(),
  },
  reopen: {
    path: "/api/v1/referrals/{id}/reopen/",
    method: "POST",
    TBody: Type<ReopenRequest>(),
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

  listAttachments: {
    path: "/api/v1/referrals/{id}/attachments/",
    method: "GET",
    TRes: Type<Attachment[]>(),
  },
  deleteAttachment: {
    path: "/api/v1/attachments/{id}/",
    method: "DELETE",
    TRes: Type<void>(),
  },

  // Lookups used by the create form
  icd10: {
    path: "/api/v1/catalog/icd10/",
    method: "GET",
    TRes: Type<IcdCode[]>(),
  },
  templates: {
    path: "/api/v1/catalog/referral-templates/",
    method: "GET",
    TRes: Type<ReferralTemplate[]>(),
  },
  getClinic: {
    path: "/api/v1/facility/{id}/",
    method: "GET",
    TRes: Type<FacilityModel>(),
  },
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
