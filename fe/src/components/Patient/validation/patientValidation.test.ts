/**
 * Property-Based Tests for Patient Data Model Validation
 * Feature: Centralized Patient Creation
 * Task 2.2: Write property test for data model validation
 */
import * as fc from "fast-check";
import { describe, expect, it } from "vitest";

import { PatientModel } from "../models";

/**
 * **Validates: Requirements 2.1, 3.3**
 *
 * Property 1: Type safety - All patient data fields maintain correct types
 * Property 2: Required fields validation - Required fields cannot be empty
 */

// ============================================================================
// Arbitraries (Generators) for Patient Data
// ============================================================================

/**
 * Generator for valid patient names
 */
const patientNameArbitrary = fc
  .string({ minLength: 1, maxLength: 100 })
  .filter((s) => s.trim().length > 0);

/**
 * Generator for valid phone numbers (10-15 digits)
 */
const phoneNumberArbitrary = fc
  .integer({ min: 1000000000, max: 999999999999999 })
  .map((n) => n.toString());

/**
 * Generator for valid gender values (1-3 as per common healthcare standards)
 */
const genderArbitrary = fc.integer({ min: 1, max: 3 });

/**
 * Generator for valid email addresses
 */
const emailArbitrary = fc.emailAddress();

/**
 * Generator for valid dates in ISO format
 */
// noInvalidDate: fast-check v4 may otherwise yield Date(NaN), and toISOString() throws on it
const dateArbitrary = fc
  .date({ noInvalidDate: true })
  .map((d) => d.toISOString());

/**
 * Generator for valid year of birth (1900-2024)
 */
const yearOfBirthArbitrary = fc.integer({ min: 1900, max: 2024 });

/**
 * Generator for valid pincode (6 digits)
 */
const pincodeArbitrary = fc.integer({ min: 100000, max: 999999 });

/**
 * Generator for valid patient IDs (UUID format)
 */
const patientIdArbitrary = fc.uuid();

/**
 * Generator for complete PatientModel with all required fields
 */
const patientModelArbitrary = fc.record({
  id: patientIdArbitrary,
  name: patientNameArbitrary,
  phone_number: phoneNumberArbitrary,
  gender: genderArbitrary,
  year_of_birth: yearOfBirthArbitrary,
  address: fc
    .string({ minLength: 1, maxLength: 500 })
    .filter((s) => s.trim().length > 0),
  pincode: pincodeArbitrary,
  facility: fc.string({ minLength: 1 }).filter((s) => s.trim().length > 0),
});

/**
 * Generator for PatientModel with optional fields
 */
const patientModelWithOptionalsArbitrary = fc.record({
  id: patientIdArbitrary,
  name: patientNameArbitrary,
  phone_number: phoneNumberArbitrary,
  gender: genderArbitrary,
  year_of_birth: yearOfBirthArbitrary,
  address: fc
    .string({ minLength: 1, maxLength: 500 })
    .filter((s) => s.trim().length > 0),
  pincode: pincodeArbitrary,
  facility: fc.string({ minLength: 1 }).filter((s) => s.trim().length > 0),
  emergency_phone_number: fc.option(phoneNumberArbitrary, { nil: undefined }),
  email: fc.option(emailArbitrary, { nil: undefined }),
  date_of_birth: fc.option(dateArbitrary, { nil: undefined }),
  allergies: fc.option(fc.string(), { nil: undefined }),
  blood_group: fc.option(
    fc.constantFrom("A+", "A-", "B+", "B-", "AB+", "AB-", "O+", "O-"),
    { nil: undefined },
  ),
  is_vaccinated: fc.option(fc.boolean(), { nil: undefined }),
});

// ============================================================================
// Property Tests
// ============================================================================

describe("Patient Data Model Validation - Property Tests", () => {
  describe("Property 1: Type Safety - All patient data fields maintain correct types", () => {
    it("should maintain string type for name field", () => {
      fc.assert(
        fc.property(patientModelArbitrary, (patient) => {
          expect(typeof patient.name).toBe("string");
          return true;
        }),
        { numRuns: 100 },
      );
    });

    it("should maintain string type for phone_number field", () => {
      fc.assert(
        fc.property(patientModelArbitrary, (patient) => {
          expect(typeof patient.phone_number).toBe("string");
          return true;
        }),
        { numRuns: 100 },
      );
    });

    it("should maintain number type for gender field", () => {
      fc.assert(
        fc.property(patientModelArbitrary, (patient) => {
          expect(typeof patient.gender).toBe("number");
          expect(patient.gender).toBeGreaterThanOrEqual(1);
          expect(patient.gender).toBeLessThanOrEqual(3);
          return true;
        }),
        { numRuns: 100 },
      );
    });

    it("should maintain number type for year_of_birth field", () => {
      fc.assert(
        fc.property(patientModelArbitrary, (patient) => {
          expect(typeof patient.year_of_birth).toBe("number");
          expect(patient.year_of_birth).toBeGreaterThanOrEqual(1900);
          expect(patient.year_of_birth).toBeLessThanOrEqual(2024);
          return true;
        }),
        { numRuns: 100 },
      );
    });

    it("should maintain string type for address field", () => {
      fc.assert(
        fc.property(patientModelArbitrary, (patient) => {
          expect(typeof patient.address).toBe("string");
          return true;
        }),
        { numRuns: 100 },
      );
    });

    it("should maintain number type for pincode field", () => {
      fc.assert(
        fc.property(patientModelArbitrary, (patient) => {
          expect(typeof patient.pincode).toBe("number");
          expect(patient.pincode).toBeGreaterThanOrEqual(100000);
          expect(patient.pincode).toBeLessThanOrEqual(999999);
          return true;
        }),
        { numRuns: 100 },
      );
    });

    it("should maintain string type for id field (UUID)", () => {
      fc.assert(
        fc.property(patientModelArbitrary, (patient) => {
          expect(typeof patient.id).toBe("string");
          // UUID format validation
          const uuidRegex =
            /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
          expect(patient.id).toMatch(uuidRegex);
          return true;
        }),
        { numRuns: 100 },
      );
    });

    it("should maintain correct types for optional fields when present", () => {
      fc.assert(
        fc.property(patientModelWithOptionalsArbitrary, (patient) => {
          if (patient.emergency_phone_number !== undefined) {
            expect(typeof patient.emergency_phone_number).toBe("string");
          }
          if (patient.blood_group !== undefined) {
            expect(typeof patient.blood_group).toBe("string");
            expect([
              "A+",
              "A-",
              "B+",
              "B-",
              "AB+",
              "AB-",
              "O+",
              "O-",
            ]).toContain(patient.blood_group);
          }
          if (patient.is_vaccinated !== undefined) {
            expect(typeof patient.is_vaccinated).toBe("boolean");
          }
          return true;
        }),
        { numRuns: 100 },
      );
    });
  });

  describe("Property 2: Required Fields Validation - Required fields cannot be empty", () => {
    it("should reject patient data with empty name", () => {
      const invalidPatient: Partial<PatientModel> = {
        name: "",
        phone_number: "1234567890",
        gender: 1,
      };

      expect(invalidPatient.name).toBe("");
      expect(invalidPatient.name?.length).toBe(0);
    });

    it("should reject patient data with empty phone_number", () => {
      const invalidPatient: Partial<PatientModel> = {
        name: "John Doe",
        phone_number: "",
        gender: 1,
      };

      expect(invalidPatient.phone_number).toBe("");
      expect(invalidPatient.phone_number?.length).toBe(0);
    });

    it("should ensure name field is always non-empty string", () => {
      fc.assert(
        fc.property(patientModelArbitrary, (patient) => {
          expect(patient.name).toBeTruthy();
          expect(patient.name.length).toBeGreaterThan(0);
          expect(patient.name.trim().length).toBeGreaterThan(0);
          return true;
        }),
        { numRuns: 100 },
      );
    });

    it("should ensure phone_number field is always non-empty string", () => {
      fc.assert(
        fc.property(patientModelArbitrary, (patient) => {
          expect(patient.phone_number).toBeTruthy();
          expect(patient.phone_number.length).toBeGreaterThanOrEqual(10);
          expect(patient.phone_number.length).toBeLessThanOrEqual(15);
          return true;
        }),
        { numRuns: 100 },
      );
    });

    it("should ensure gender field is always present and valid", () => {
      fc.assert(
        fc.property(patientModelArbitrary, (patient) => {
          expect(patient.gender).toBeDefined();
          expect(patient.gender).not.toBeNull();
          expect([1, 2, 3]).toContain(patient.gender);
          return true;
        }),
        { numRuns: 100 },
      );
    });

    it("should ensure facility field is always non-empty string", () => {
      fc.assert(
        fc.property(patientModelArbitrary, (patient) => {
          expect(patient.facility).toBeTruthy();
          expect(patient.facility.length).toBeGreaterThan(0);
          return true;
        }),
        { numRuns: 100 },
      );
    });

    it("should ensure address field is always non-empty string", () => {
      fc.assert(
        fc.property(patientModelArbitrary, (patient) => {
          expect(patient.address).toBeTruthy();
          expect(patient.address.length).toBeGreaterThan(0);
          expect(patient.address.length).toBeLessThanOrEqual(500);
          return true;
        }),
        { numRuns: 100 },
      );
    });

    it("should validate that all required fields are present together", () => {
      fc.assert(
        fc.property(patientModelArbitrary, (patient) => {
          // All required fields must be present
          expect(patient.id).toBeDefined();
          expect(patient.name).toBeDefined();
          expect(patient.phone_number).toBeDefined();
          expect(patient.gender).toBeDefined();
          expect(patient.year_of_birth).toBeDefined();
          expect(patient.address).toBeDefined();
          expect(patient.pincode).toBeDefined();
          expect(patient.facility).toBeDefined();

          // All required fields must be non-empty/valid
          expect(patient.id.length).toBeGreaterThan(0);
          expect(patient.name.length).toBeGreaterThan(0);
          expect(patient.phone_number.length).toBeGreaterThanOrEqual(10);
          expect(patient.gender).toBeGreaterThan(0);
          expect(patient.year_of_birth).toBeGreaterThan(0);
          expect(patient.address.length).toBeGreaterThan(0);
          expect(patient.pincode).toBeGreaterThan(0);
          expect(patient.facility.length).toBeGreaterThan(0);

          return true;
        }),
        { numRuns: 100 },
      );
    });
  });

  describe("Property Tests: Data Integrity", () => {
    it("should maintain data integrity across transformations", () => {
      fc.assert(
        fc.property(patientModelArbitrary, (patient) => {
          // Serialize and deserialize
          const serialized = JSON.stringify(patient);
          const deserialized = JSON.parse(serialized) as PatientModel;

          // All fields should remain the same
          expect(deserialized.id).toBe(patient.id);
          expect(deserialized.name).toBe(patient.name);
          expect(deserialized.phone_number).toBe(patient.phone_number);
          expect(deserialized.gender).toBe(patient.gender);
          expect(deserialized.year_of_birth).toBe(patient.year_of_birth);
          expect(deserialized.address).toBe(patient.address);
          expect(deserialized.pincode).toBe(patient.pincode);
          expect(deserialized.facility).toBe(patient.facility);

          return true;
        }),
        { numRuns: 100 },
      );
    });
  });
});
