/**
 * Constants for Patient Creation feature
 * Feature: Centralized Patient Creation
 */

import {
  AllergySeverity,
  ConditionStatus,
  PatientStatus,
  PaymentMethod,
} from "../types/patientCreate";

// ============================================================================
// Patient Status Options
// ============================================================================

export const PATIENT_STATUS_OPTIONS = [
  { value: PatientStatus.LIVE, label: "Live (Active)" },
  { value: PatientStatus.DISCHARGED, label: "Discharged" },
] as const;

// ============================================================================
// Payment Method Options
// ============================================================================

export const PAYMENT_METHOD_OPTIONS = [
  { value: PaymentMethod.INSURANCE, label: "Insurance" },
  { value: PaymentMethod.SELF_PAY, label: "Self Pay" },
  { value: PaymentMethod.GOVERNMENT_SCHEME, label: "Government Scheme" },
] as const;

// ============================================================================
// Allergy Severity Options
// ============================================================================

export const ALLERGY_SEVERITY_OPTIONS = [
  { value: AllergySeverity.MILD, label: "Mild" },
  { value: AllergySeverity.MODERATE, label: "Moderate" },
  { value: AllergySeverity.SEVERE, label: "Severe" },
  {
    value: AllergySeverity.LIFE_THREATENING,
    label: "Life Threatening",
  },
] as const;

// ============================================================================
// Condition Status Options
// ============================================================================

export const CONDITION_STATUS_OPTIONS = [
  { value: ConditionStatus.ACTIVE, label: "Active" },
  { value: ConditionStatus.RESOLVED, label: "Resolved" },
  { value: ConditionStatus.CHRONIC, label: "Chronic" },
] as const;

// ============================================================================
// Relationship Options
// ============================================================================

export const RELATIONSHIP_OPTIONS = [
  { value: "spouse", label: "Spouse" },
  { value: "parent", label: "Parent" },
  { value: "child", label: "Child" },
  { value: "sibling", label: "Sibling" },
  { value: "friend", label: "Friend" },
  { value: "other", label: "Other" },
] as const;

// ============================================================================
// Form Field Limits
// ============================================================================

export const FIELD_LIMITS = {
  NAME_MIN_LENGTH: 2,
  NAME_MAX_LENGTH: 50,
  REASON_FOR_VISIT_MIN_LENGTH: 10,
  REASON_FOR_VISIT_MAX_LENGTH: 1000,
  INITIAL_NOTES_MAX_LENGTH: 5000,
} as const;

// ============================================================================
// Auto-save Configuration
// ============================================================================

export const AUTO_SAVE_CONFIG = {
  INTERVAL_MS: 30000, // 30 seconds
  STORAGE_KEY: "patient_create_draft",
  VERSION: "1.0",
  MAX_AGE_MS: 24 * 60 * 60 * 1000, // 24 hours
} as const;

// ============================================================================
// Validation Patterns
// ============================================================================

export const VALIDATION_PATTERNS = {
  NAME: /^[a-zA-Z\s\-']+$/,
  PHONE:
    /^[+]?[(]?[0-9]{1,4}[)]?[-\s.]?[(]?[0-9]{1,4}[)]?[-\s.]?[0-9]{1,9}$/,
  EMAIL: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
} as const;

// ============================================================================
// Default Form Values
// ============================================================================

export const DEFAULT_FORM_VALUES = {
  demographics: {
    firstName: "",
    middleName: "",
    lastName: "",
    dateOfBirth: "",
    gender: 1,
    phone: "",
    alternatePhone: "",
    email: "",
    address: {
      street: "",
      city: "",
      state: "",
      postalCode: "",
      country: "",
    },
  },
  emergencyContact: undefined,
  medicalHistory: {
    allergies: [],
    medications: [],
    conditions: [],
    bloodType: undefined,
  },
  clinicalNotes: {
    reasonForVisit: "",
    initialNotes: "",
  },
  administrative: {
    facilityId: "",
    status: PatientStatus.LIVE,
    admissionDate: new Date().toISOString().split("T")[0],
    dischargeDate: "",
    insurance: undefined,
    paymentMethod: PaymentMethod.SELF_PAY,
  },
} as const;
