/** Filters of the clinics list (tested in facilityFilters.test.ts). */
import { FacilityModel } from "@/components/Facility/models";

import { FACILITY_TYPES } from "@/common/constants";

/** The district shown for a clinic (its district, else nothing). */
export const clinicDistrict = (f: FacilityModel) =>
  f.district_object?.name ?? "";

/** The type of a clinic as text (old records keep a numeric code). */
export const clinicType = (f: FacilityModel) =>
  typeof f.facility_type === "number"
    ? (FACILITY_TYPES.find((ft) => ft.id === f.facility_type)?.text ?? "")
    : (f.facility_type ?? "");

const fold = (s: string) =>
  s
    .normalize("NFD")
    .replace(/\p{Diacritic}/gu, "")
    .toLowerCase();

/** Clinics whose name or address contains the search (accents ignored), in the district and of the type. */
export function filterClinics(
  all: FacilityModel[],
  f: { search: string; district: string; type: string },
) {
  const q = fold(f.search.trim());
  return all.filter(
    (c) =>
      (!q || fold(`${c.name ?? ""} ${c.address ?? ""}`).includes(q)) &&
      (!f.district || clinicDistrict(c) === f.district) &&
      (!f.type || clinicType(c) === f.type),
  );
}
