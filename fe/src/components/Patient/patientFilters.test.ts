import { describe, expect, it } from "vitest";

import {
  NO_PATIENT_FILTER,
  isFiltered,
  patientQuery,
  percent,
} from "@/components/Patient/patientFilters";

describe("patientQuery", () => {
  it("sends nothing for the empty filter", () => {
    expect(
      Object.values(patientQuery(NO_PATIENT_FILTER)).every(
        (v) => v === undefined,
      ),
    ).toBe(true);
  });

  it("maps an age band to min and max", () => {
    expect(
      patientQuery({ ...NO_PATIENT_FILTER, ageBand: "40-64" }),
    ).toMatchObject({ age_min: 40, age_max: 64 });
    expect(
      patientQuery({ ...NO_PATIENT_FILTER, ageBand: "80+" }),
    ).toMatchObject({ age_min: 80, age_max: undefined });
  });

  it("sorts care levels and passes gender and clinic", () => {
    expect(
      patientQuery({
        ...NO_PATIENT_FILTER,
        careLevels: [4, 2],
        gender: 2,
        clinic: "c1",
        search: " ana ",
      }),
    ).toMatchObject({
      care_level: [2, 4],
      gender: 2,
      facility: "c1",
      search: "ana",
    });
  });
});

describe("isFiltered", () => {
  it("detects any filter", () => {
    expect(isFiltered(NO_PATIENT_FILTER)).toBe(false);
    expect(isFiltered({ ...NO_PATIENT_FILTER, gender: 1 })).toBe(true);
    expect(isFiltered({ ...NO_PATIENT_FILTER, careLevels: [5] })).toBe(true);
  });
});

it("percent rounds and survives zero", () => {
  expect(percent(1, 3)).toBe(33);
  expect(percent(5, 0)).toBe(0);
});

describe("physio and diagnosis filters", () => {
  it("passes them and counts them as filtering", () => {
    const f = { ...NO_PATIENT_FILTER, physio: 7, diagnosis: " lombalgia " };
    expect(patientQuery(f)).toMatchObject({
      physio: 7,
      diagnosis: "lombalgia",
    });
    expect(isFiltered({ ...NO_PATIENT_FILTER, physio: 7 })).toBe(true);
    expect(isFiltered({ ...NO_PATIENT_FILTER, diagnosis: "  " })).toBe(false);
  });
});
