/** Filters of the patients page (tested in patientFilters.test.ts). */

/** Age bands offered as chips; the same bands the stats endpoint counts. */
export const AGE_BANDS = [
  { key: "0-17", min: 0, max: 17 },
  { key: "18-39", min: 18, max: 39 },
  { key: "40-64", min: 40, max: 64 },
  { key: "65-79", min: 65, max: 79 },
  { key: "80+", min: 80, max: undefined },
] as const;

export type AgeBand = (typeof AGE_BANDS)[number]["key"];

export interface PatientFilter {
  search: string;
  ageBand: AgeBand | "";
  careLevels: number[];
  gender: number | null;
  clinic: string;
  physio: number | null;
  diagnosis: string;
}

export const NO_PATIENT_FILTER: PatientFilter = {
  search: "",
  ageBand: "",
  careLevels: [],
  gender: null,
  clinic: "",
  physio: null,
  diagnosis: "",
};

/** Query params shared by the list and the stats (empty filters left out). */
export function patientQuery(f: PatientFilter) {
  const band = AGE_BANDS.find((b) => b.key === f.ageBand);
  return {
    search: f.search.trim() || undefined,
    age_min: band?.min,
    age_max: band?.max,
    care_level: f.careLevels.length
      ? [...f.careLevels].sort((a, b) => a - b)
      : undefined,
    gender: f.gender ?? undefined,
    facility: f.clinic || undefined,
    physio: f.physio ?? undefined,
    diagnosis: f.diagnosis.trim() || undefined,
  };
}

/** Whether anything narrows the list (search included). */
export const isFiltered = (f: PatientFilter) =>
  !!(
    f.search.trim() ||
    f.ageBand ||
    f.careLevels.length ||
    f.gender !== null ||
    f.clinic ||
    f.physio !== null ||
    f.diagnosis.trim()
  );

/** Share of n in total, 0–100. */
export const percent = (n: number, total: number) =>
  total ? Math.round((n / total) * 100) : 0;
