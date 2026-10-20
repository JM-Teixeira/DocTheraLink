/**
 * TypeScript types and interfaces for Patient Creation feature
 * Feature: Centralized Patient Creation
 */

import { BLOOD_GROUPS, GENDER_TYPES } from "@/common/constants";

// ============================================================================
// Core Patient Data Types
// ============================================================================

/**
 * Gender options for patient demographics
 */
export type Gender = (typeof GENDER_TYPES)[number]["id"];

/**
 * Blood group types
 */
export type BloodGroup = (typeof BLOOD_GROUPS)[number];

/**
 * Patient status options
 */
export enum PatientStatus {
  LIVE = "live",
  DISCHARGED = "discharged",
}

/**
 * Payment method options
 */
export enum PaymentMethod {
  INSURANCE = "insurance",
  SELF_PAY = "self_pay",
  GOVERNMENT_SCHEME = "government_scheme",
}

/**
 * Allergy severity levels
 */
export enum AllergySeverity {
  MILD = "mild",
  MODERATE = "moderate",
  SEVERE = "severe",
  LIFE_THREATENING = "life_threatening",
}

/**
 * Medical condition status
 */
export enum ConditionStatus {
  ACTIVE = "active",
  RESOLVED = "resolved",
  CHRONIC = "chronic",
}

// ============================================================================
// Address Interface
// ============================================================================

/**
 * Address information
 */
export interface Address {
  street: string;
  city: string;
  state: string;
  postalCode: string;
  country: string;
}

// ============================================================================
// Demographics Interface
// ============================================================================

/**
 * Patient demographic information
 */
export interface Demographics {
  firstName: string;
  middleName?: string;
  lastName: string;
  dateOfBirth: string; // ISO 8601 date format
  age?: number; // Calculated field
  gender: Gender;
  phone: string;
  alternatePhone?: string;
  email?: string;
  address: Address;
}

// ============================================================================
// Emergency Contact Interface
// ============================================================================

/**
 * Emergency contact information
 */
export interface EmergencyContact {
  name: string;
  phone: string;
  alternatePhone?: string;
  relationship: string;
  address?: Address;
}

// ============================================================================
// Medical History Interfaces
// ============================================================================

/**
 * Allergy information
 */
export interface Allergy {
  id: string;
  allergen: string;
  severity: AllergySeverity;
  reaction?: string;
  diagnosedDate?: string;
}

/**
 * Medication information
 */
export interface Medication {
  id: string;
  name: string;
  dosage: string;
  frequency: string;
  startDate: string;
  endDate?: string;
  prescribedBy?: string;
  notes?: string;
}

/**
 * Medical condition information
 */
export interface Condition {
  id: string;
  name: string;
  diagnosedDate: string;
  status: ConditionStatus;
  notes?: string;
}

/**
 * Complete medical history
 */
export interface MedicalHistory {
  allergies: Allergy[];
  medications: Medication[];
  conditions: Condition[];
  bloodType?: BloodGroup;
}

// ============================================================================
// Clinical Notes Interface
// ============================================================================

/**
 * Clinical notes for patient
 */
export interface ClinicalNotes {
  reasonForVisit: string;
  initialNotes?: string;
}

// ============================================================================
// Insurance Interface
// ============================================================================

/**
 * Insurance information
 */
export interface InsuranceInfo {
  provider: string;
  policyNumber: string;
  groupNumber?: string;
  validUntil?: string;
}

// ============================================================================
// Administrative Interface
// ============================================================================

/**
 * Administrative information
 */
export interface Administrative {
  facilityId: string;
  status: PatientStatus;
  admissionDate: string;
  dischargeDate?: string;
  insurance?: InsuranceInfo;
  paymentMethod: PaymentMethod;
}

// ============================================================================
// Form State Interface
// ============================================================================

/**
 * Complete patient creation form state
 */
export interface PatientFormState {
  demographics: Demographics;
  emergencyContact?: EmergencyContact;
  medicalHistory: MedicalHistory;
  clinicalNotes: ClinicalNotes;
  administrative: Administrative;
}

// ============================================================================
// Validation Error Interface
// ============================================================================

/**
 * Validation error structure
 */
export interface ValidationError {
  field: string;
  message: string;
  type: "required" | "format" | "custom";
}

/**
 * Validation errors by field
 */
export type ValidationErrors = Record<string, string>;

// ============================================================================
// API Request/Response Interfaces
// ============================================================================

/**
 * Patient creation API request
 */
export interface CreatePatientRequest {
  demographics: Demographics;
  emergencyContact?: EmergencyContact;
  medicalHistory?: MedicalHistory;
  clinicalNotes: ClinicalNotes;
  administrative: Administrative;
}

/**
 * Patient creation API response
 */
export interface CreatePatientResponse {
  id: string;
  demographics: Demographics;
  emergencyContact?: EmergencyContact;
  medicalHistory: MedicalHistory;
  clinicalNotes: ClinicalNotes;
  administrative: Administrative;
  createdAt: string;
  updatedAt: string;
  createdBy: string;
}

// ============================================================================
// Draft Storage Interface
// ============================================================================

/**
 * Form draft for local storage
 */
export interface FormDraft {
  version: string;
  timestamp: string;
  data: Partial<PatientFormState>;
}


// ============================================================================
// Type Guards
// ============================================================================

/**
 * Type guard to check if a value is a valid Gender
 */
export function isGender(value: unknown): value is Gender {
  return typeof value === "number" && value >= 1 && value <= 3;
}

/**
 * Type guard to check if a value is a valid BloodGroup
 */
export function isBloodGroup(value: unknown): value is BloodGroup {
  const validBloodGroups = ["A+", "A-", "B+", "B-", "AB+", "AB-", "O+", "O-"];
  return typeof value === "string" && validBloodGroups.includes(value);
}

/**
 * Type guard to check if a value is a valid PatientStatus
 */
export function isPatientStatus(value: unknown): value is PatientStatus {
  return value === PatientStatus.LIVE || value === PatientStatus.DISCHARGED;
}

/**
 * Type guard to check if a value is a valid PaymentMethod
 */
export function isPaymentMethod(value: unknown): value is PaymentMethod {
  return (
    value === PaymentMethod.INSURANCE ||
    value === PaymentMethod.SELF_PAY ||
    value === PaymentMethod.GOVERNMENT_SCHEME
  );
}

/**
 * Type guard to check if a value is a valid AllergySeverity
 */
export function isAllergySeverity(value: unknown): value is AllergySeverity {
  return (
    value === AllergySeverity.MILD ||
    value === AllergySeverity.MODERATE ||
    value === AllergySeverity.SEVERE ||
    value === AllergySeverity.LIFE_THREATENING
  );
}

/**
 * Type guard to check if a value is a valid ConditionStatus
 */
export function isConditionStatus(value: unknown): value is ConditionStatus {
  return (
    value === ConditionStatus.ACTIVE ||
    value === ConditionStatus.RESOLVED ||
    value === ConditionStatus.CHRONIC
  );
}

/**
 * Type guard to check if a value is a valid Address
 */
export function isAddress(value: unknown): value is Address {
  if (typeof value !== "object" || value === null) return false;
  const addr = value as Record<string, unknown>;
  return (
    typeof addr.street === "string" &&
    typeof addr.city === "string" &&
    typeof addr.state === "string" &&
    typeof addr.postalCode === "string" &&
    typeof addr.country === "string"
  );
}

/**
 * Type guard to check if a value is a valid Demographics
 */
export function isDemographics(value: unknown): value is Demographics {
  if (typeof value !== "object" || value === null) return false;
  const demo = value as Record<string, unknown>;
  return (
    typeof demo.firstName === "string" &&
    typeof demo.lastName === "string" &&
    typeof demo.dateOfBirth === "string" &&
    isGender(demo.gender) &&
    typeof demo.phone === "string" &&
    isAddress(demo.address) &&
    (demo.middleName === undefined || typeof demo.middleName === "string") &&
    (demo.age === undefined || typeof demo.age === "number") &&
    (demo.alternatePhone === undefined ||
      typeof demo.alternatePhone === "string") &&
    (demo.email === undefined || typeof demo.email === "string")
  );
}

/**
 * Type guard to check if a value is a valid EmergencyContact
 */
export function isEmergencyContact(
  value: unknown,
): value is EmergencyContact {
  if (typeof value !== "object" || value === null) return false;
  const contact = value as Record<string, unknown>;
  return (
    typeof contact.name === "string" &&
    typeof contact.phone === "string" &&
    typeof contact.relationship === "string" &&
    (contact.alternatePhone === undefined ||
      typeof contact.alternatePhone === "string") &&
    (contact.address === undefined || isAddress(contact.address))
  );
}

/**
 * Type guard to check if a value is a valid Allergy
 */
export function isAllergy(value: unknown): value is Allergy {
  if (typeof value !== "object" || value === null) return false;
  const allergy = value as Record<string, unknown>;
  return (
    typeof allergy.id === "string" &&
    typeof allergy.allergen === "string" &&
    isAllergySeverity(allergy.severity) &&
    (allergy.reaction === undefined || typeof allergy.reaction === "string") &&
    (allergy.diagnosedDate === undefined ||
      typeof allergy.diagnosedDate === "string")
  );
}

/**
 * Type guard to check if a value is a valid Medication
 */
export function isMedication(value: unknown): value is Medication {
  if (typeof value !== "object" || value === null) return false;
  const med = value as Record<string, unknown>;
  return (
    typeof med.id === "string" &&
    typeof med.name === "string" &&
    typeof med.dosage === "string" &&
    typeof med.frequency === "string" &&
    typeof med.startDate === "string" &&
    (med.endDate === undefined || typeof med.endDate === "string") &&
    (med.prescribedBy === undefined || typeof med.prescribedBy === "string") &&
    (med.notes === undefined || typeof med.notes === "string")
  );
}

/**
 * Type guard to check if a value is a valid Condition
 */
export function isCondition(value: unknown): value is Condition {
  if (typeof value !== "object" || value === null) return false;
  const condition = value as Record<string, unknown>;
  return (
    typeof condition.id === "string" &&
    typeof condition.name === "string" &&
    typeof condition.diagnosedDate === "string" &&
    isConditionStatus(condition.status) &&
    (condition.notes === undefined || typeof condition.notes === "string")
  );
}

/**
 * Type guard to check if a value is a valid MedicalHistory
 */
export function isMedicalHistory(value: unknown): value is MedicalHistory {
  if (typeof value !== "object" || value === null) return false;
  const history = value as Record<string, unknown>;
  return (
    Array.isArray(history.allergies) &&
    history.allergies.every(isAllergy) &&
    Array.isArray(history.medications) &&
    history.medications.every(isMedication) &&
    Array.isArray(history.conditions) &&
    history.conditions.every(isCondition) &&
    (history.bloodType === undefined || isBloodGroup(history.bloodType))
  );
}

/**
 * Type guard to check if a value is a valid ClinicalNotes
 */
export function isClinicalNotes(value: unknown): value is ClinicalNotes {
  if (typeof value !== "object" || value === null) return false;
  const notes = value as Record<string, unknown>;
  return (
    typeof notes.reasonForVisit === "string" &&
    (notes.initialNotes === undefined || typeof notes.initialNotes === "string")
  );
}

/**
 * Type guard to check if a value is a valid InsuranceInfo
 */
export function isInsuranceInfo(value: unknown): value is InsuranceInfo {
  if (typeof value !== "object" || value === null) return false;
  const insurance = value as Record<string, unknown>;
  return (
    typeof insurance.provider === "string" &&
    typeof insurance.policyNumber === "string" &&
    (insurance.groupNumber === undefined ||
      typeof insurance.groupNumber === "string") &&
    (insurance.validUntil === undefined ||
      typeof insurance.validUntil === "string")
  );
}

/**
 * Type guard to check if a value is a valid Administrative
 */
export function isAdministrative(value: unknown): value is Administrative {
  if (typeof value !== "object" || value === null) return false;
  const admin = value as Record<string, unknown>;
  return (
    typeof admin.facilityId === "string" &&
    isPatientStatus(admin.status) &&
    typeof admin.admissionDate === "string" &&
    isPaymentMethod(admin.paymentMethod) &&
    (admin.dischargeDate === undefined ||
      typeof admin.dischargeDate === "string") &&
    (admin.insurance === undefined || isInsuranceInfo(admin.insurance))
  );
}

/**
 * Type guard to check if a value is a valid PatientFormState
 */
export function isPatientFormState(value: unknown): value is PatientFormState {
  if (typeof value !== "object" || value === null) return false;
  const form = value as Record<string, unknown>;
  return (
    isDemographics(form.demographics) &&
    isMedicalHistory(form.medicalHistory) &&
    isClinicalNotes(form.clinicalNotes) &&
    isAdministrative(form.administrative) &&
    (form.emergencyContact === undefined ||
      isEmergencyContact(form.emergencyContact))
  );
}

// ============================================================================
// Validation Functions
// ============================================================================

/**
 * Validates that a string is not empty after trimming
 */
export function validateRequired(value: string | undefined): boolean {
  return value !== undefined && value.trim().length > 0;
}

/**
 * Validates email format
 */
export function validateEmail(email: string): boolean {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
}

/**
 * Validates phone number format
 * Accepts international formats with optional country code
 */
export function validatePhone(phone: string): boolean {
  const phoneRegex =
    /^[+]?[(]?[0-9]{1,4}[)]?[-\s.]?[(]?[0-9]{1,4}[)]?[-\s.]?[0-9]{1,9}$/;
  return phoneRegex.test(phone);
}

/**
 * Validates that a date is not in the future
 */
export function validateDateNotFuture(dateString: string): boolean {
  const date = new Date(dateString);
  const now = new Date();
  return date <= now;
}

/**
 * Validates that a date is in valid ISO 8601 format
 */
export function validateISODate(dateString: string): boolean {
  const date = new Date(dateString);
  return !isNaN(date.getTime()) && dateString === date.toISOString().split("T")[0];
}

/**
 * Validates name format (letters, spaces, hyphens, apostrophes only)
 */
export function validateName(name: string): boolean {
  const nameRegex = /^[a-zA-Z\s\-']+$/;
  return nameRegex.test(name) && name.trim().length >= 2;
}

/**
 * Validates postal code format (basic validation)
 */
export function validatePostalCode(postalCode: string): boolean {
  return postalCode.trim().length >= 3 && postalCode.trim().length <= 10;
}

/**
 * Calculates age from date of birth
 */
export function calculateAge(dateOfBirth: string): number {
  const dob = new Date(dateOfBirth);
  const today = new Date();
  let age = today.getFullYear() - dob.getFullYear();
  const monthDiff = today.getMonth() - dob.getMonth();

  if (monthDiff < 0 || (monthDiff === 0 && today.getDate() < dob.getDate())) {
    age--;
  }

  return age;
}

/**
 * Validates that discharge date is after admission date
 */
export function validateDischargeDateAfterAdmission(
  admissionDate: string,
  dischargeDate: string,
): boolean {
  const admission = new Date(admissionDate);
  const discharge = new Date(dischargeDate);
  return discharge >= admission;
}

/**
 * Validates Demographics data
 */
export function validateDemographics(
  demographics: Demographics,
): ValidationError[] {
  const errors: ValidationError[] = [];

  if (!validateRequired(demographics.firstName)) {
    errors.push({
      field: "demographics.firstName",
      message: "First name is required",
      type: "required",
    });
  } else if (!validateName(demographics.firstName)) {
    errors.push({
      field: "demographics.firstName",
      message:
        "First name can only contain letters, spaces, hyphens, and apostrophes",
      type: "format",
    });
  }

  if (!validateRequired(demographics.lastName)) {
    errors.push({
      field: "demographics.lastName",
      message: "Last name is required",
      type: "required",
    });
  } else if (!validateName(demographics.lastName)) {
    errors.push({
      field: "demographics.lastName",
      message:
        "Last name can only contain letters, spaces, hyphens, and apostrophes",
      type: "format",
    });
  }

  if (!validateRequired(demographics.dateOfBirth)) {
    errors.push({
      field: "demographics.dateOfBirth",
      message: "Date of birth is required",
      type: "required",
    });
  } else if (!validateISODate(demographics.dateOfBirth)) {
    errors.push({
      field: "demographics.dateOfBirth",
      message: "Please enter a valid date",
      type: "format",
    });
  } else if (!validateDateNotFuture(demographics.dateOfBirth)) {
    errors.push({
      field: "demographics.dateOfBirth",
      message: "Date of birth cannot be in the future",
      type: "custom",
    });
  }

  if (!validateRequired(demographics.phone)) {
    errors.push({
      field: "demographics.phone",
      message: "Phone number is required",
      type: "required",
    });
  } else if (!validatePhone(demographics.phone)) {
    errors.push({
      field: "demographics.phone",
      message: "Please enter a valid phone number",
      type: "format",
    });
  }

  if (demographics.alternatePhone && !validatePhone(demographics.alternatePhone)) {
    errors.push({
      field: "demographics.alternatePhone",
      message: "Please enter a valid alternate phone number",
      type: "format",
    });
  }

  if (demographics.email && !validateEmail(demographics.email)) {
    errors.push({
      field: "demographics.email",
      message: "Please enter a valid email address",
      type: "format",
    });
  }

  // Validate address
  if (!validateRequired(demographics.address.street)) {
    errors.push({
      field: "demographics.address.street",
      message: "Street address is required",
      type: "required",
    });
  }

  if (!validateRequired(demographics.address.city)) {
    errors.push({
      field: "demographics.address.city",
      message: "City is required",
      type: "required",
    });
  }

  if (!validateRequired(demographics.address.state)) {
    errors.push({
      field: "demographics.address.state",
      message: "State is required",
      type: "required",
    });
  }

  if (!validateRequired(demographics.address.postalCode)) {
    errors.push({
      field: "demographics.address.postalCode",
      message: "Postal code is required",
      type: "required",
    });
  } else if (!validatePostalCode(demographics.address.postalCode)) {
    errors.push({
      field: "demographics.address.postalCode",
      message: "Please enter a valid postal code",
      type: "format",
    });
  }

  if (!validateRequired(demographics.address.country)) {
    errors.push({
      field: "demographics.address.country",
      message: "Country is required",
      type: "required",
    });
  }

  return errors;
}

/**
 * Validates EmergencyContact data
 */
export function validateEmergencyContact(
  contact: EmergencyContact,
): ValidationError[] {
  const errors: ValidationError[] = [];

  if (!validateRequired(contact.name)) {
    errors.push({
      field: "emergencyContact.name",
      message: "Emergency contact name is required",
      type: "required",
    });
  } else if (!validateName(contact.name)) {
    errors.push({
      field: "emergencyContact.name",
      message:
        "Name can only contain letters, spaces, hyphens, and apostrophes",
      type: "format",
    });
  }

  if (!validateRequired(contact.phone)) {
    errors.push({
      field: "emergencyContact.phone",
      message: "Emergency contact phone is required",
      type: "required",
    });
  } else if (!validatePhone(contact.phone)) {
    errors.push({
      field: "emergencyContact.phone",
      message: "Please enter a valid phone number",
      type: "format",
    });
  }

  if (contact.alternatePhone && !validatePhone(contact.alternatePhone)) {
    errors.push({
      field: "emergencyContact.alternatePhone",
      message: "Please enter a valid alternate phone number",
      type: "format",
    });
  }

  if (!validateRequired(contact.relationship)) {
    errors.push({
      field: "emergencyContact.relationship",
      message: "Relationship is required",
      type: "required",
    });
  }

  return errors;
}

/**
 * Validates ClinicalNotes data
 */
export function validateClinicalNotes(
  notes: ClinicalNotes,
): ValidationError[] {
  const errors: ValidationError[] = [];

  if (!validateRequired(notes.reasonForVisit)) {
    errors.push({
      field: "clinicalNotes.reasonForVisit",
      message: "Reason for visit is required",
      type: "required",
    });
  } else if (notes.reasonForVisit.trim().length < 10) {
    errors.push({
      field: "clinicalNotes.reasonForVisit",
      message: "Please provide at least 10 characters",
      type: "custom",
    });
  } else if (notes.reasonForVisit.trim().length > 1000) {
    errors.push({
      field: "clinicalNotes.reasonForVisit",
      message: "Reason for visit cannot exceed 1000 characters",
      type: "custom",
    });
  }

  return errors;
}

/**
 * Validates Administrative data
 */
export function validateAdministrative(
  admin: Administrative,
): ValidationError[] {
  const errors: ValidationError[] = [];

  if (!validateRequired(admin.facilityId)) {
    errors.push({
      field: "administrative.facilityId",
      message: "Facility assignment is required",
      type: "required",
    });
  }

  if (!validateRequired(admin.admissionDate)) {
    errors.push({
      field: "administrative.admissionDate",
      message: "Admission date is required",
      type: "required",
    });
  } else if (!validateISODate(admin.admissionDate)) {
    errors.push({
      field: "administrative.admissionDate",
      message: "Please enter a valid date",
      type: "format",
    });
  }

  if (admin.status === PatientStatus.DISCHARGED && !admin.dischargeDate) {
    errors.push({
      field: "administrative.dischargeDate",
      message: "Discharge date is required when status is Discharged",
      type: "required",
    });
  }

  if (
    admin.dischargeDate &&
    admin.admissionDate &&
    !validateDischargeDateAfterAdmission(admin.admissionDate, admin.dischargeDate)
  ) {
    errors.push({
      field: "administrative.dischargeDate",
      message: "Discharge date cannot be before admission date",
      type: "custom",
    });
  }

  if (admin.insurance) {
    if (!validateRequired(admin.insurance.provider)) {
      errors.push({
        field: "administrative.insurance.provider",
        message: "Insurance provider is required",
        type: "required",
      });
    }

    if (!validateRequired(admin.insurance.policyNumber)) {
      errors.push({
        field: "administrative.insurance.policyNumber",
        message: "Policy number is required",
        type: "required",
      });
    }
  }

  return errors;
}

/**
 * Validates complete PatientFormState
 */
export function validatePatientFormState(
  formState: PatientFormState,
): ValidationError[] {
  const errors: ValidationError[] = [];

  // Validate demographics
  errors.push(...validateDemographics(formState.demographics));

  // Validate emergency contact if provided
  if (formState.emergencyContact) {
    errors.push(...validateEmergencyContact(formState.emergencyContact));
  }

  // Validate clinical notes
  errors.push(...validateClinicalNotes(formState.clinicalNotes));

  // Validate administrative
  errors.push(...validateAdministrative(formState.administrative));

  return errors;
}

/**
 * Converts validation errors array to ValidationErrors object
 */
export function errorsToObject(errors: ValidationError[]): ValidationErrors {
  return errors.reduce(
    (acc, error) => {
      acc[error.field] = error.message;
      return acc;
    },
    {} as ValidationErrors,
  );
}
