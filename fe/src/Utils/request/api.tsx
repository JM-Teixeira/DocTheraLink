import {
  DistrictModel,
  FacilityModel,
  FacilityRequest,
  IFacilityNotificationRequest,
  IFacilityNotificationResponse,
  IUserFacilityRequest,
  InventoryItemsModel,
  InventoryLogResponse,
  InventorySummaryResponse,
  LocalBodyModel,
  MinimumQuantityItemResponse,
  StateModel,
  WardModel,
} from "@/components/Facility/models";
import {
  DupPatientModel,
} from "@/components/Facility/models";
import { DailyRoundsModel, PatientModel } from "@/components/Patient/models";
import {
  CreateFileRequest,
  CreateFileResponse,
  FileUploadModel,
} from "@/components/Patient/models";
import {
  SkillModel,
  SkillObjectModel,
  UpdatePasswordForm,
  UserAssignedModel,
  UserModel,
} from "@/components/Users/models";

import { PaginatedResponse } from "@/Utils/request/types";

/**
 * A fake function that returns an empty object casted to type T
 * @returns Empty object as type T
 */
export function Type<T>(): T {
  return {} as T;
}

export interface JwtTokenObtainPair {
  access: string;
  refresh: string;
}

export interface LoginCredentials {
  username: string;
  password: string;
}

const routes = {
  // Auth Endpoints
  login: {
    path: "/api/v1/auth/login/",
    method: "POST",
    noAuth: true,
    TRes: Type<JwtTokenObtainPair>(),
    TBody: Type<LoginCredentials>(),
  },

  token_refresh: {
    path: "/api/v1/auth/token/refresh/",
    method: "POST",
    TRes: Type<JwtTokenObtainPair>(),
    TBody: Type<{ refresh: JwtTokenObtainPair["refresh"] }>(),
  },

  token_verify: {
    path: "/api/v1/auth/token/verify/",
    method: "POST",
  },

  checkResetToken: {
    path: "/api/v1/password_reset/check/",
    method: "POST",
    noAuth: true,
    TRes: Type<Record<string, never>>(),
    TBody: Type<{
      token: string;
    }>(),
  },

  resetPassword: {
    path: "/api/v1/password_reset/confirm/",
    method: "POST",
    noAuth: true,
    TRes: Type<Record<string, never>>(),
    TBody: Type<{
      password: string;
      confirm: string;
    }>(),
  },

  forgotPassword: {
    path: "/api/v1/password_reset/",
    method: "POST",
    noAuth: true,
    TRes: Type<Record<string, never>>(),
    TBody: Type<{
      username: string;
    }>(),
  },

  updatePassword: {
    path: "/api/v1/password_change/",
    method: "PUT",
    TRes: Type<Record<string, string | string[]>>(),
    TBody: Type<UpdatePasswordForm>(),
  },
  // User Endpoints
  currentUser: {
    path: "/api/v1/users/getcurrentuser/",
    TRes: Type<UserModel>(),
  },

  userList: {
    path: "/api/v1/users/",
    method: "GET",
    TRes: Type<PaginatedResponse<UserModel>>(),
  },

  userListSkill: {
    path: "/api/v1/users/{username}/skill/",
    method: "GET",
    TRes: Type<PaginatedResponse<SkillModel>>(),
  },

  userListFacility: {
    path: "/api/v1/users/{username}/get_facilities/",
    method: "GET",
    TRes: Type<PaginatedResponse<FacilityModel>>(),
  },

  addUserFacility: {
    path: "/api/v1/users/{username}/add_facility/",
    method: "PUT",
    TBody: Type<IUserFacilityRequest>(),
    TRes: Type<UserModel>(),
  },

  addUserSkill: {
    path: "/api/v1/users/{username}/skill/",
    method: "POST",
    TBody: Type<{ skill: string }>(),
    TRes: Type<SkillModel>(),
  },

  deleteUserFacility: {
    path: "/api/v1/users/{username}/delete_facility/",
    method: "DELETE",
    TBody: Type<IUserFacilityRequest>(),
    TRes: Type<Record<string, never>>(),
  },

  clearHomeFacility: {
    path: "/api/v1/users/{username}/clear_home_facility/",
    method: "DELETE",
    TRes: Type<Record<string, never>>(),
  },

  deleteUserSkill: {
    path: "/api/v1/users/{username}/skill/{id}/",
    method: "DELETE",
    TRes: Type<Record<string, never>>(),
  },

  createUser: {
    path: "/api/v1/users/",
    method: "POST",
    noAuth: true,
  },

  updateUser: {
    path: "/api/v1/users/",
    method: "PUT",
  },

  partialUpdateUser: {
    path: "/api/v1/users/{username}/",
    method: "PATCH",
    TRes: Type<UserModel>(),
    TBody: Type<Partial<UserModel>>(),
  },

  updateProfilePicture: {
    path: "/api/v1/users/{username}/profile_picture/",
    method: "PATCH",
    TRes: Type<UserModel>(),
    TBody: Type<{ profile_picture_url: string }>(),
  },

  deleteProfilePicture: {
    path: "/api/v1/users/{username}/profile_picture/",
    method: "DELETE",
    TRes: Type<UserModel>(),
  },

  deleteUser: {
    path: "/api/v1/users/{username}/",
    method: "DELETE",
    TRes: Type<Record<string, never>>(),
  },

  addUser: {
    path: "/api/v1/users/add_user/",
    method: "POST",
    TRes: Type<UserModel>(),
  },

  searchUser: {
    path: "/api/v1/users/",
  },

  getOnlineDoctors: {
    path: "/api/v1/users/?user_type=Doctor&ordering=-last_login",
  },

  // Skill Endpoints

  getAllSkills: {
    path: "/api/v1/skill/",
    method: "GET",
    TRes: Type<PaginatedResponse<SkillObjectModel>>(),
  },

  // Facility Endpoints

  getPermittedFacilities: {
    path: "/api/v1/facility/",
    TRes: Type<PaginatedResponse<FacilityModel>>(),
  },

  getAllFacilities: {
    path: "/api/v1/getallfacilities/",
    TRes: Type<PaginatedResponse<FacilityModel>>(),
  },

  createFacility: {
    path: "/api/v1/facility/",
    method: "POST",
    TRes: Type<FacilityModel>(),
    TBody: Type<FacilityRequest>(),
  },

  getPermittedFacility: {
    path: "/api/v1/facility/{id}/",
    method: "GET",
    TRes: Type<FacilityModel>(),
  },

  getAnyFacility: {
    path: "/api/v1/getallfacilities/{id}/",
    method: "GET",
    TRes: Type<FacilityModel>(),
  },

  updateFacility: {
    path: "/api/v1/facility/{id}/",
    method: "PUT",
    TRes: Type<FacilityModel>(),
    TBody: Type<FacilityRequest>(),
  },

  partialUpdateFacility: {
    path: "/api/v1/facility/{id}/",
    method: "PATCH",
    TRes: Type<FacilityModel>(),
    TBody: Type<Partial<FacilityModel>>(),
  },

  deleteFacilityCoverImage: {
    path: "/api/v1/facility/{id}/cover_image/",
    method: "DELETE",
    TRes: Type<Record<string, never>>(),
  },

  getFacilityUsers: {
    path: "/api/v1/facility/{facility_id}/get_users/",
    TRes: Type<PaginatedResponse<UserAssignedModel>>(),
  },

  // Consultation beds

  // Download Api
  deleteFacility: {
    path: "/api/v1/facility/{id}/",
    method: "DELETE",
    TRes: Type<Record<string, never>>(),
  },

  downloadFacility: {
    path: "/api/v1/facility/?csv",
    method: "GET",
    TRes: Type<string>(),
  },
  downloadPatients: {
    path: "/api/v1/patient/?csv",
    method: "GET",
    TRes: Type<string>(),
  },

  // Event Types

  // Patient

  searchPatient: {
    path: "/api/v1/patient/search/",
    TRes: Type<PaginatedResponse<DupPatientModel>>(),
  },
  patientList: {
    path: "/api/v1/patient/",
    method: "GET",
    TRes: Type<PaginatedResponse<PatientModel>>(),
  },
  addPatient: {
    path: "/api/v1/patient/",
    method: "POST",
    TRes: Type<PatientModel>(),
  },
  getPatient: {
    path: "/api/v1/patient/{id}/",
    method: "GET",
    TBody: Type<PatientModel>(),
    TRes: Type<PatientModel>(),
  },
  updatePatient: {
    path: "/api/v1/patient/{id}/",
    method: "PUT",
    TRes: Type<PatientModel>(),
  },
  patchPatient: {
    path: "/api/v1/patient/{id}/",
    method: "PATCH",
    TBody: Type<Partial<PatientModel>>(),
    TRes: Type<PatientModel>(),
  },

  // States
  statesList: {
    path: "/api/v1/state/",
    method: "GET",
    TRes: Type<PaginatedResponse<StateModel>>(),
  },

  getState: {
    path: "/api/v1/state/{id}/",
    TRes: Type<StateModel>(),
  },

  // Districts

  getDistrict: {
    path: "/api/v1/district/{id}/",
    method: "GET",
    TRes: Type<DistrictModel>(),
  },
  getDistrictByState: {
    path: "/api/v1/state/{id}/districts/",
    method: "GET",
    TRes: Type<DistrictModel[]>(),
  },
  getDistrictByName: {
    path: "/api/v1/district/",
    method: "GET",
    TRes: Type<PaginatedResponse<DistrictModel>>(),
  },
  getAllLocalBodyByDistrict: {
    path: "/api/v1/district/{id}/get_all_local_body/",
    method: "GET",
    TRes: Type<LocalBodyModel[]>(),
  },
  getLocalbodyByDistrict: {
    path: "/api/v1/district/{id}/local_bodies/",
    method: "GET",
    TRes: Type<LocalBodyModel[]>(),
  },

  // Local Body
  getLocalBody: {
    path: "/api/v1/local_body/{id}/",
    TRes: Type<LocalBodyModel>(),
  },
  getAllLocalBody: {
    path: "/api/v1/local_body/",
    method: "GET",
    TRes: Type<PaginatedResponse<LocalBodyModel>>(),
  },
  getLocalbodyByName: {
    path: "/api/v1/local_body/",
  },

  // ward
  getWard: {
    path: "/api/v1/ward/{id}/",
  },
  getWards: {
    path: "/api/v1/ward/",
  },
  getWardByLocalBody: {
    path: "/api/v1/ward/?local_body={id}",
    method: "GET",
    TRes: Type<PaginatedResponse<WardModel>>(),
  },

  //inventory
  getItems: {
    path: "/api/v1/items/",
    method: "GET",
    TRes: Type<PaginatedResponse<InventoryItemsModel>>(),
  },
  createInventory: {
    path: "/api/v1/facility/{facilityId}/inventory/",
    method: "POST",
    TRes: Type<InventoryLogResponse>(),
  },
  getInventoryLog: {
    path: "/api/v1/facility/{facilityId}/inventory/",
    method: "GET",
    TRes: Type<PaginatedResponse<InventoryLogResponse>>(),
  },
  setMinQuantity: {
    path: "/api/v1/facility/{facilityId}/min_quantity/",
    method: "POST",
    TRes: Type<MinimumQuantityItemResponse>(),
  },
  getMinQuantity: {
    path: "/api/v1/facility/{facilityId}/min_quantity/",
    method: "GET",
    TRes: Type<PaginatedResponse<InventorySummaryResponse>>(),
  },
  getMinQuantityItem: {
    path: "/api/v1/facility/{facilityId}/min_quantity/{inventoryId}/",
    method: "GET",
    TRes: Type<MinimumQuantityItemResponse>(),
  },
  updateMinQuantity: {
    path: "/api/v1/facility/{facilityId}/min_quantity/{inventoryId}/",
    method: "PATCH",
    TRes: Type<PaginatedResponse<MinimumQuantityItemResponse>>(),
  },
  getInventorySummary: {
    path: "/api/v1/facility/{facility_external_id}/inventorysummary/",
    method: "GET",
    TRes: Type<PaginatedResponse<InventorySummaryResponse>>(),
  },
  getItemName: {
    path: "/api/v1/items/",
    method: "GET",
  },
  flagInventoryItem: {
    path: "/api/v1/facility/{facility_external_id}/inventory/{external_id}/flag/",
    method: "PUT",
    TRes: Type<PaginatedResponse<InventoryLogResponse>>(),
  },
  deleteLastInventoryLog: {
    path: "/api/v1/facility/{facility_external_id}/inventory/delete_last/?item={id}",
    method: "DELETE",
    TRes: Type<Record<string, never>>(),
  },

  //Profile
  checkUsername: {
    path: "/api/v1/users/{username}/check_availability/",
    method: "GET",
    TRes: Type<Record<string, never>>(),
  },

  getUserDetails: {
    path: "/api/v1/users/{username}/",
    method: "GET",
    TRes: Type<UserModel>(),
  },
  updateUserDetails: {
    path: "/api/v1/users/",
    method: "PUT",
  },

  // FileUpload Create
  createUpload: {
    path: "/api/v1/files/",
    method: "POST",
    TBody: Type<CreateFileRequest>(),
    TRes: Type<CreateFileResponse>(),
  },
  viewUpload: {
    path: "/api/v1/files/",
    method: "GET",
    TRes: Type<PaginatedResponse<FileUploadModel>>(),
  },
  retrieveUpload: {
    path: "/api/v1/files/{id}/",
    method: "GET",
    TRes: Type<FileUploadModel>(),
  },
  editUpload: {
    path: "/api/v1/files/{id}/?file_type={fileType}&associating_id={associatingId}",
    method: "PATCH",
    TBody: Type<Partial<FileUploadModel>>(),
    TRes: Type<FileUploadModel>(),
  },

  // Assets endpoints

  // Asset transaction endpoints

  // Asset service endpoints

  // Prescription endpoints

  // Time Log Endpoints (Clock In/Out)
  timeLog: {
    clockIn: {
      path: "/api/v1/timelog/clock-in",
      method: "POST",
      TBody: Type<{ notes?: string }>(),
      TRes: Type<any>(),
    },

    clockOut: {
      path: "/api/v1/timelog/clock-out",
      method: "POST",
      TBody: Type<{ notes?: string }>(),
      TRes: Type<any>(),
    },

    getMyLogs: {
      path: "/api/v1/timelog/my-logs",
      method: "GET",
      TRes: Type<PaginatedResponse<any>>(),
    },

    getUserLogs: {
      path: "/api/v1/timelog/user/{id}",
      method: "GET",
      TRes: Type<PaginatedResponse<any>>(),
    },

    getStatus: {
      path: "/api/v1/timelog/status",
      method: "GET",
      TRes: Type<{ clocked_in: boolean; time_log: any | null }>(),
    },
  },

  // Timetable Endpoints
  timetable: {
    listSections: {
      path: "/api/v1/facility/{id}/sections/",
      method: "GET",
      TRes: Type<any[]>(),
    },

    createSection: {
      path: "/api/v1/facility/{id}/sections/",
      method: "POST",
      TBody: Type<{
        name: string;
        max_capacity?: number;
        is_active?: boolean;
        description?: string;
      }>(),
      TRes: Type<any>(),
    },

    getSection: {
      path: "/api/v1/sections/{id}",
      method: "GET",
      TRes: Type<any>(),
    },

    updateSection: {
      path: "/api/v1/sections/{id}",
      method: "PUT",
      TBody: Type<{
        name: string;
        max_capacity?: number;
        is_active?: boolean;
        description?: string;
      }>(),
      TRes: Type<any>(),
    },

    deleteSection: {
      path: "/api/v1/sections/{id}",
      method: "DELETE",
      TRes: Type<Record<string, never>>(),
    },

    getTimetable: {
      path: "/api/v1/timetable/{id}",
      method: "GET",
      TRes: Type<any[]>(),
    },

    createSlot: {
      path: "/api/v1/timetable/{id}/slot",
      method: "POST",
      TBody: Type<{
        section_id: string;
        user_id: number;
        day_of_week: number;
        start_time: string;
        end_time: string;
        is_recurring?: boolean;
        valid_from?: string;
        valid_until?: string;
        notes?: string;
      }>(),
      TRes: Type<any>(),
    },

    updateSlot: {
      path: "/api/v1/timetable/slot/{id}",
      method: "PUT",
      TBody: Type<{
        section_id: string;
        user_id: number;
        day_of_week: number;
        start_time: string;
        end_time: string;
        is_recurring?: boolean;
        valid_from?: string;
        valid_until?: string;
        notes?: string;
      }>(),
      TRes: Type<any>(),
    },

    deleteSlot: {
      path: "/api/v1/timetable/slot/{id}",
      method: "DELETE",
      TRes: Type<Record<string, never>>(),
    },
  },
} as const;

export default routes;
