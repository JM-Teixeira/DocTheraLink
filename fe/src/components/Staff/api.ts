import {
  Clinic,
  EmploymentRequest,
  Skill,
  StaffRow,
  UserSkill,
} from "@/components/Staff/types";
import { UserModel } from "@/components/Users/models";

import { DateFormat } from "@/Utils/dates";
import { Type } from "@/Utils/request/api";
import { PaginatedResponse } from "@/Utils/request/types";

/** Staff API: users list, skills, clinic and employment status. */
const StaffRoutes = {
  list: {
    path: "/api/v1/users/",
    method: "GET",
    TRes: Type<PaginatedResponse<StaffRow>>(),
  },
  clinics: {
    path: "/api/v1/facility/",
    method: "GET",
    TRes: Type<PaginatedResponse<Clinic>>(),
  },
  skills: {
    path: "/api/v1/skill/",
    method: "GET",
    TRes: Type<PaginatedResponse<Skill>>(),
  },
  userSkills: {
    path: "/api/v1/users/{username}/skill/",
    method: "GET",
    TRes: Type<PaginatedResponse<UserSkill>>(),
  },
  addSkill: {
    path: "/api/v1/users/{username}/skill/",
    method: "POST",
    TBody: Type<{ skill: string }>(),
    TRes: Type<UserSkill>(),
  },
  removeSkill: {
    path: "/api/v1/users/{username}/skill/{id}/",
    method: "DELETE",
    TRes: Type<void>(),
  },
  setClinic: {
    path: "/api/v1/users/{username}/add_facility/",
    method: "PUT",
    TBody: Type<{ facility: string }>(),
    TRes: Type<UserModel>(),
  },
  setPreferences: {
    path: "/api/v1/users/{username}/preferences/",
    method: "PATCH",
    TBody: Type<{ date_format: DateFormat }>(),
    TRes: Type<UserModel>(),
  },
  setEmployment: {
    path: "/api/v1/users/{username}/employment/",
    method: "PATCH",
    TBody: Type<EmploymentRequest>(),
    TRes: Type<UserModel>(),
  },
} as const;

export default StaffRoutes;
