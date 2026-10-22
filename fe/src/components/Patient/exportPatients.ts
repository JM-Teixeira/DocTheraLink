/** Downloads the patients CSV (coordinators and admins; see patient/export.go). */
import careConfig from "@careConfig";

import { QueryParams } from "@/Utils/request/types";
import { getAuthorizationHeader, makeUrl } from "@/Utils/request/utils";

/** Fetches /patient/export/ with the list filters and saves the file. */
export async function downloadPatientsCsv(params: QueryParams, lang: string) {
  const headers = new Headers();
  const auth = getAuthorizationHeader();
  if (auth) headers.set("Authorization", auth);
  const res = await fetch(
    careConfig.apiUrl + makeUrl("/api/v1/patient/export/", { ...params, lang }),
    { headers },
  );
  if (!res.ok) throw new Error("patients_export_failed");
  const url = URL.createObjectURL(await res.blob());
  const link = document.createElement("a");
  link.href = url;
  link.download = exportFileName(new Date());
  link.click();
  URL.revokeObjectURL(url);
}

/** "utentes-2026-10-22.csv" (local date). */
export const exportFileName = (d: Date) =>
  `utentes-${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}.csv`;
