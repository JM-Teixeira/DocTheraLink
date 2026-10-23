import { describe, expect, it } from "vitest";

import { getDuplicateCheckParams, patientKeys } from "./index";

describe("Patient API Service", () => {
  describe("patientKeys", () => {
    it("should generate correct query keys for all patients", () => {
      expect(patientKeys.all).toEqual(["patients"]);
    });

    it("should generate correct query keys for patient lists", () => {
      expect(patientKeys.lists()).toEqual(["patients", "list"]);
    });

    it("should generate correct query keys for patient list with filters", () => {
      const filters = { facility: "facility-1", limit: 20 };
      expect(patientKeys.list(filters)).toEqual([
        "patients",
        "list",
        { facility: "facility-1", limit: 20 },
      ]);
    });

    it("should generate correct query keys for patient details", () => {
      expect(patientKeys.details()).toEqual(["patients", "detail"]);
    });

    it("should generate correct query keys for specific patient detail", () => {
      expect(patientKeys.detail("patient-123")).toEqual([
        "patients",
        "detail",
        "patient-123",
      ]);
    });

    it("should generate correct query keys for patient search", () => {
      const filters = { name: "John Doe", phone_number: "+1234567890" };
      expect(patientKeys.search(filters)).toEqual([
        "patients",
        "search",
        { name: "John Doe", phone_number: "+1234567890" },
      ]);
    });
  });

  describe("getDuplicateCheckParams", () => {
    it("should return search params with phone number", () => {
      const patientData = {
        phone_number: "+1234567890",
        name: "John Doe",
      };

      const result = getDuplicateCheckParams(patientData);

      expect(result).toEqual({
        phone_number: "+1234567890",
        name: "John Doe",
      });
    });

    it("should return search params with date of birth", () => {
      const patientData = {
        name: "Jane Smith",
        date_of_birth: "1990-01-01",
      };

      const result = getDuplicateCheckParams(patientData);

      expect(result).toEqual({
        name: "Jane Smith",
        date_of_birth: "1990-01-01",
      });
    });

    it("should use year of birth when date of birth is not available", () => {
      const patientData = {
        name: "Bob Johnson",
        year_of_birth: 1985,
      };

      const result = getDuplicateCheckParams(patientData);

      expect(result).toEqual({
        name: "Bob Johnson",
        year_of_birth: 1985,
      });
    });

    it("should prefer date of birth over year of birth", () => {
      const patientData = {
        name: "Alice Brown",
        date_of_birth: "1992-05-15",
        year_of_birth: 1992,
      };

      const result = getDuplicateCheckParams(patientData);

      expect(result).toEqual({
        name: "Alice Brown",
        date_of_birth: "1992-05-15",
      });
      expect(result).not.toHaveProperty("year_of_birth");
    });

    it("should return empty object when no searchable fields are provided", () => {
      const patientData = {
        gender: 1,
        address: "123 Main St",
      };

      const result = getDuplicateCheckParams(patientData);

      expect(result).toEqual({});
    });

    it("should handle partial data correctly", () => {
      const patientData = {
        phone_number: "+9876543210",
      };

      const result = getDuplicateCheckParams(patientData);

      expect(result).toEqual({
        phone_number: "+9876543210",
      });
    });

    it("should handle all searchable fields together", () => {
      const patientData = {
        name: "Charlie Davis",
        phone_number: "+1122334455",
        date_of_birth: "1988-12-25",
      };

      const result = getDuplicateCheckParams(patientData);

      expect(result).toEqual({
        name: "Charlie Davis",
        phone_number: "+1122334455",
        date_of_birth: "1988-12-25",
      });
    });
  });
});
