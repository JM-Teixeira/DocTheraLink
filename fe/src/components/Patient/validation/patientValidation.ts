/**
 * Validation schemas for Patient Creation feature
 * Uses Zod for runtime validation
 * Feature: Centralized Patient Creation
 */
import { z } from "zod";

import { BLOOD_GROUPS } from "@/common/constants";

import {
  AllergySeverity,
  ConditionStatus,
  PatientStatus,
  PaymentMethod,
} from "../types/patientCreate";

// ============================================================================
// Validation Constants
// ============================================================================

const NAME_PATTERN = /^[a-zA-Z\s\-']+$/;
const PHONE_PATTERN =
  /^[+]?[(]?[0-9]{1,4}[)]?[-\s.]?[(]?[0-9]{1,4}[)]?[-\s.]?[0-9]{1,9}$/;
const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

// ============================================================================
// Address Schema
// ============================================================================

export const addressSchema = z.object({
  street: z.string().min(1, "Street address is required"),
  city: z.string().min(1, "City is required"),
  state: z.string().min(1, "State is required"),
  postalCode: z.string().min(1, "Postal code is required"),
  country: z.string().min(1, "Country is required"),
});

// ============================================================================
// Demographics Schema
// ============================================================================

export const demographicsSchema = z
  .object({
    firstName: z
      .string()
      .min(2, "First name must be at least 2 characters")
      .max(50, "First name cannot exceed 50 characters")
      .regex(
        NAME_PATTERN,
        "First name can only contain letters, spaces, hyphens, and apostrophes",
      ),
    middleName: z
      .string()
      .max(50, "Middle name cannot exceed 50 characters")
      .regex(
        NAME_PATTERN,
        "Middle name can only contain letters, spaces, hyphens, and apostrophes",
      )
      .optional(),
    lastName: z
      .string()
      .min(2, "Last name must be at least 2 characters")
      .max(50, "Last name cannot exceed 50 characters")
      .regex(
        NAME_PATTERN,
        "Last name can only contain letters, spaces, hyphens, and apostrophes",
      ),
    dateOfBirth: z
      .string()
      .min(1, "Date of birth is required")
      .refine(
        (date) => {
          const dob = new Date(date);
          return dob <= new Date();
        },
        { message: "Date of birth cannot be in the future" },
      )
      .refine(
        (date) => {
          const dob = new Date(date);
          return dob >= new Date("1900-01-01");
        },
        { message: "Please enter a valid date of birth" },
      ),
    age: z.number().optional(),
    gender: z.number().min(1).max(3),
    phone: z
      .string()
      .min(1, "Phone number is required")
      .regex(PHONE_PATTERN, "Please enter a valid phone number"),
    alternatePhone: z
      .string()
      .regex(PHONE_PATTERN, "Please enter a valid phone number")
      .optional()
      .or(z.literal("")),
    email: z
      .string()
      .regex(EMAIL_PATTERN, "Please enter a valid email address")
      .optional()
      .or(z.literal("")),
    address: addressSchema,
  })
  .strict();

// ============================================================================
// Emergency Contact Schema
// ============================================================================

export const emergencyContactSchema = z
  .object({
    name: z
      .string()
      .min(2, "Contact name must be at least 2 characters")
      .regex(
        NAME_PATTERN,
        "Contact name can only contain letters, spaces, hyphens, and apostrophes",
      ),
    phone: z
      .string()
      .min(1, "Contact phone number is required")
      .regex(PHONE_PATTERN, "Please enter a valid phone number"),
    alternatePhone: z
      .string()
      .regex(PHONE_PATTERN, "Please enter a valid phone number")
      .optional()
      .or(z.literal("")),
    relationship: z.string().min(1, "Relationship is required"),
    address: addressSchema.optional(),
  })
  .strict()
  .optional();

// ============================================================================
// Medical History Schemas
// ============================================================================

export const allergySchema = z.object({
  id: z.string(),
  allergen: z.string().min(1, "Allergen name is required"),
  severity: z.nativeEnum(AllergySeverity),
  reaction: z.string().optional(),
  diagnosedDate: z.string().optional(),
});

export const medicationSchema = z.object({
  id: z.string(),
  name: z.string().min(1, "Medication name is required"),
  dosage: z.string().min(1, "Dosage is required"),
  frequency: z.string().min(1, "Frequency is required"),
  startDate: z.string().min(1, "Start date is required"),
  endDate: z.string().optional(),
  prescribedBy: z.string().optional(),
  notes: z.string().optional(),
});

export const conditionSchema = z.object({
  id: z.string(),
  name: z.string().min(1, "Condition name is required"),
  diagnosedDate: z.string().min(1, "Diagnosed date is required"),
  status: z.nativeEnum(ConditionStatus),
  notes: z.string().optional(),
});

export const medicalHistorySchema = z.object({
  allergies: z.array(allergySchema).default([]),
  medications: z.array(medicationSchema).default([]),
  conditions: z.array(conditionSchema).default([]),
  bloodType: z.enum(BLOOD_GROUPS as [string, ...string[]]).optional(),
});

// ============================================================================
// Clinical Notes Schema
// ============================================================================

export const clinicalNotesSchema = z.object({
  reasonForVisit: z
    .string()
    .min(10, "Please provide at least 10 characters")
    .max(1000, "Reason for visit cannot exceed 1000 characters"),
  initialNotes: z
    .string()
    .max(5000, "Initial notes cannot exceed 5000 characters")
    .optional(),
});

// ============================================================================
// Insurance Schema
// ============================================================================

export const insuranceSchema = z
  .object({
    provider: z.string().min(1, "Insurance provider is required"),
    policyNumber: z.string().min(1, "Policy number is required"),
    groupNumber: z.string().optional(),
    validUntil: z.string().optional(),
  })
  .optional();

// ============================================================================
// Administrative Schema
// ============================================================================

export const administrativeSchema = z
  .object({
    facilityId: z.string().min(1, "Facility assignment is required"),
    status: z.nativeEnum(PatientStatus),
    admissionDate: z.string().min(1, "Admission date is required"),
    dischargeDate: z.string().optional(),
    insurance: insuranceSchema,
    paymentMethod: z.nativeEnum(PaymentMethod),
  })
  .refine(
    (data) => {
      if (data.status === PatientStatus.DISCHARGED && !data.dischargeDate) {
        return false;
      }
      return true;
    },
    {
      message: "Discharge date is required when status is Discharged",
      path: ["dischargeDate"],
    },
  )
  .refine(
    (data) => {
      if (data.dischargeDate && data.admissionDate) {
        const discharge = new Date(data.dischargeDate);
        const admission = new Date(data.admissionDate);
        return discharge >= admission;
      }
      return true;
    },
    {
      message: "Discharge date cannot be before admission date",
      path: ["dischargeDate"],
    },
  );

// ============================================================================
// Complete Patient Form Schema
// ============================================================================

export const patientFormSchema = z.object({
  demographics: demographicsSchema,
  emergencyContact: emergencyContactSchema,
  medicalHistory: medicalHistorySchema,
  clinicalNotes: clinicalNotesSchema,
  administrative: administrativeSchema,
});

// ============================================================================
// Validation Helper Functions
// ============================================================================

/**
 * Validate a single field
 */
export function validateField(
  schema: z.ZodSchema,
  value: unknown,
): { isValid: boolean; error?: string } {
  const result = schema.safeParse(value);
  if (result.success) {
    return { isValid: true };
  }
  return {
    isValid: false,
    error: result.error.errors[0]?.message || "Validation failed",
  };
}

/**
 * Validate entire form
 */
export function validatePatientForm(data: unknown): {
  isValid: boolean;
  errors: Record<string, string>;
} {
  const result = patientFormSchema.safeParse(data);

  if (result.success) {
    return { isValid: true, errors: {} };
  }

  const errors: Record<string, string> = {};
  result.error.errors.forEach((err) => {
    const path = err.path.join(".");
    errors[path] = err.message;
  });

  return { isValid: false, errors };
}

/**
 * Calculate age from date of birth
 */
export function calculateAge(dateOfBirth: string): number {
  const today = new Date();
  const birthDate = new Date(dateOfBirth);
  let age = today.getFullYear() - birthDate.getFullYear();
  const monthDiff = today.getMonth() - birthDate.getMonth();

  if (
    monthDiff < 0 ||
    (monthDiff === 0 && today.getDate() < birthDate.getDate())
  ) {
    age--;
  }

  return age;
}
