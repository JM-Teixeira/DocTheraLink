import { RouteToFacility } from "@/components/Common/RouteToFacilitySelect";
import {
  AssignedToObjectModel,
  BloodPressure,
  DailyRoundsModel,
  FacilityNameModel,
  PatientModel,
} from "@/components/Patient/models";
import { UserBareMinimum } from "@/components/Users/models";

import {
  CONSENT_PATIENT_CODE_STATUS_CHOICES,
  CONSENT_TYPE_CHOICES,
  ConsultationSuggestionValue,
  DISCHARGE_REASONS,
  PATIENT_NOTES_THREADS,
  SHIFTING_CHOICES_PEACETIME,
  UserRole,
} from "@/common/constants";

import { FeatureFlag } from "@/Utils/featureFlags";

export interface LocalBodyModel {
  id: number;
  name: string;
  body_type: number;
  localbody_code: string;
  district: number;
}

export interface DistrictModel {
  id: number;
  name: string;
  state: number;
}

export interface StateModel {
  id: number;
  name: string;
}

export interface WardModel {
  id: number;
  name: string;
  number: number;
  panchayath: string;
  local_body_id: LocalBodyModel["id"];
}

export interface FacilityModel {
  id?: string;
  name?: string;
  /** 2-4 letters; prefixes the usernames of its people (cm.joao.teixeira) */
  code?: string;
  read_cover_image_url?: string;
  facility_type?: number;
  address?: string;
  features?: number[];
  location?: {
    latitude: number;
    longitude: number;
  };
  middleware_address?: string;
  district_object?: DistrictModel;
  state_object?: StateModel;
  modified_date?: string;
  created_date?: string;
  state?: number;
  district?: number;
  pincode?: string;
  facility_flags?: FeatureFlag[];
  latitude?: string | number; // Can be string in form, number in API
  longitude?: string | number; // Can be string in form, number in API
  kasp_empanelled?: boolean;
  patient_count?: number;
  bed_count?: number;
}

export type PatientCategory =
  | "Comfort Care" // Discontinued
  | "Mild"
  | "Moderate"
  | "Critical"
  | "Actively Dying";

export interface DupPatientModel {
  id: number;
  gender: string;
  phone_number: string;
  patient_id: string;
  name: string;
  date_of_birth: string;
  year_of_birth: number;
  state_id: number;
  is_expired: boolean;
}

export interface InventoryItemsModel {
  // count?: number;
  id?: number;
  name?: string;
  default_unit?: {
    id: number;
    name: string;
  };
  allowed_units?: [
    {
      id: number;
      name: string;
    },
  ];
}

// Patient Notes Model
export interface BaseFacilityModel {
  id: string;
  name: string;
  district: number;
  state: number;
  district_object?: DistrictModel;
  state_object?: StateModel;
  facility_type: FacilityType;
  read_cover_image_url: any;
  features: any[];
  patient_count: number;
  bed_count: number;
}

export interface FacilityType {
  id: number;
  name: string;
}

export interface BaseUserModel {
  id: number;
  first_name: string;
  username: string;
  email: string;
  last_name: string;
  user_type: string;
  last_login: string;
}

export type IFacilityNotificationRequest = {
  facility: string;
  message: string;
};

export type IFacilityNotificationResponse = {
  [key: string]: string;
};

export type IUserFacilityRequest = {
  facility: string;
};

export type FacilityRequest = Omit<
  FacilityModel,
  "location" | "id" | "latitude" | "longitude"
> & {
  latitude?: number;
  longitude?: number;
};

export type InventorySummaryResponse = {
  id: string;
  item_object: {
    id: number;
    default_unit: {
      id: number;
      name: string;
    };
    allowed_units: {
      id: number;
      name: string;
    }[];
    tags: {
      id: number;
      name: string;
    }[];
    name: string;
    description: string;
    min_quantity: number;
  };
  unit_object: {
    id: number;
    name: string;
  };
  created_date: string;
  quantity: number;
  is_low: boolean;
  item: number;
};

export type MinimumQuantityItemResponse = {
  id: string;
  item_object: InventoryItemsModel;
  created_date: string;
  min_quantity: number;
  item: number;
};

export type InventoryLogResponse = InventorySummaryResponse & {
  external_id: string;
  current_stock: number;
  quantity_in_default_unit: number;
  is_incoming: boolean;
  probable_accident: boolean;
  unit: number;
  created_by: number;
};
