import { describe, expect, it } from "vitest";

import { filterClinics } from "@/components/Facility/facilityFilters";
import { FacilityModel } from "@/components/Facility/models";

const clinics = [
  {
    id: "a",
    name: "Klyniva Demo",
    address: "Lisboa",
    district_object: { id: 1, name: "Lisboa" },
    facility_type: "Clinic",
  },
  {
    id: "b",
    name: "Climutua · Medicina Física e Reabilitação",
    address: "Rua de Cedofeita 312, Porto",
    district_object: { id: 2, name: "Porto" },
    facility_type: "Clinic",
  },
] as unknown as FacilityModel[];

describe("filterClinics", () => {
  const none = { search: "", district: "", type: "" };

  it("keeps everything without filters", () => {
    expect(filterClinics(clinics, none)).toHaveLength(2);
  });

  it("searches name and address, ignoring accents and case", () => {
    expect(
      filterClinics(clinics, { ...none, search: "fisica" }).map((c) => c.id),
    ).toEqual(["b"]);
    expect(
      filterClinics(clinics, { ...none, search: "CEDOFEITA" }).map((c) => c.id),
    ).toEqual(["b"]);
  });

  it("filters by district and type", () => {
    expect(
      filterClinics(clinics, { ...none, district: "Lisboa" }).map((c) => c.id),
    ).toEqual(["a"]);
    expect(filterClinics(clinics, { ...none, type: "Hospital" })).toHaveLength(
      0,
    );
  });
});
