import { useQuery } from "@tanstack/react-query";
import dayjs from "dayjs";
import { Plus, Search, Users, X } from "lucide-react";
import { navigate } from "raviger";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";

import { cn } from "@/lib/utils";

import { Button } from "@/components/ui/button";

import { FilterChips, FilterPanel } from "@/components/Common/FilterChips";
import Page from "@/components/Common/Page";
import PatientMetrics from "@/components/Patient/PatientMetrics";
import { usePatientList, usePatientStats } from "@/components/Patient/api";
import {
  AGE_BANDS,
  AgeBand,
  NO_PATIENT_FILTER,
  PatientFilter,
  isFiltered,
  patientQuery,
} from "@/components/Patient/patientFilters";
import { GENDERS, genderKey } from "@/components/Patient/patientFormUtils";
import StaffRoutes from "@/components/Staff/api";
import {
  CARE_LEVELS,
  CARE_LEVEL_STYLES,
} from "@/components/TimeBoard/boardUtils";

import useAuthUser from "@/hooks/useAuthUser";
import useDebouncedValue from "@/hooks/useDebouncedValue";

import { ageFromDob, initials } from "@/Utils/person";
import query from "@/Utils/request/query";
import { canWritePatients, isAdmin } from "@/Utils/roles";

const PAGE_SIZE = 20;

/**
 * Patients visible to the current user (the backend applies clinic scoping),
 * filtered by age, care level, gender and clinic, with the metrics of
 * whatever the filters match.
 */
export default function PatientList() {
  const { t } = useTranslation();
  const user = useAuthUser();
  const admin = isAdmin(user);
  const [filter, setFilter] = useState<PatientFilter>(NO_PATIENT_FILTER);
  const [page, setPage] = useState(0);
  const term = useDebouncedValue(filter.search.trim());
  const set = (patch: Partial<PatientFilter>) =>
    setFilter((f) => ({ ...f, ...patch }));

  const params = patientQuery({ ...filter, search: term });
  const paramsKey = JSON.stringify(params);
  // Any filter change → back to the first page
  useEffect(() => setPage(0), [paramsKey]);

  const { data, isLoading, isError } = usePatientList({
    ...params,
    limit: PAGE_SIZE,
    offset: page * PAGE_SIZE,
  });
  const { data: stats } = usePatientStats(params);
  const { data: clinics } = useQuery({
    queryKey: ["staff-clinics"],
    queryFn: query(StaffRoutes.clinics, { queryParams: { limit: 100 } }),
    enabled: admin,
  });
  const total = data?.count ?? 0;
  const pages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  return (
    <Page
      title={t("patients")}
      hideBack
      options={
        canWritePatients(user) && (
          <Button variant="primary" onClick={() => navigate("/patients/new")}>
            <Plus className="mr-1.5 size-4" aria-hidden />
            {t("new_patient")}
          </Button>
        )
      }
    >
      <FilterPanel>
        <div className="flex flex-wrap items-center gap-3">
          <div className="relative min-w-60 max-w-md flex-1">
            <Search
              className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-gray-400"
              aria-hidden
            />
            <input
              type="search"
              aria-label={t("search_patient")}
              placeholder={t("search_patient_by_name")}
              value={filter.search}
              onChange={(e) => set({ search: e.target.value })}
              className="h-10 w-full rounded-lg border border-gray-300 pl-9 pr-3 text-sm shadow-sm focus:border-primary-500 focus:ring-primary-500"
            />
          </div>
          {admin && (
            <select
              aria-label={t("clinic")}
              value={filter.clinic}
              onChange={(e) => set({ clinic: e.target.value })}
              className="h-10 rounded-lg border border-gray-300 px-3 text-sm shadow-sm focus:border-primary-500 focus:ring-primary-500"
            >
              <option value="">{t("all_clinics")}</option>
              {clinics?.results.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name}
                </option>
              ))}
            </select>
          )}
          {isFiltered(filter) && (
            <button
              type="button"
              onClick={() => setFilter(NO_PATIENT_FILTER)}
              className="flex items-center gap-1 text-xs text-gray-500 hover:text-gray-800"
            >
              <X className="size-3.5" aria-hidden />
              {t("clear_filters")}
            </button>
          )}
        </div>
        <FilterChips
          label={t("care_level")}
          allLabel={t("all")}
          options={CARE_LEVELS.map((l) => ({
            value: l as number,
            label: t("care_level_n", { level: l }),
          }))}
          value={filter.careLevels}
          onChange={(careLevels) => set({ careLevels })}
        />
        <FilterChips
          label={t("age")}
          allLabel={t("all")}
          options={AGE_BANDS.map((b) => ({
            value: b.key as AgeBand,
            label: t("age_band", { band: b.key }),
          }))}
          value={filter.ageBand ? [filter.ageBand] : []}
          onChange={(v) => set({ ageBand: v.at(-1) ?? "" })}
        />
        <FilterChips
          label={t("gender")}
          allLabel={t("all")}
          options={GENDERS.map((g) => ({
            value: g.id as number,
            label: t(g.key),
          }))}
          value={filter.gender === null ? [] : [filter.gender]}
          onChange={(v) => set({ gender: v.at(-1) ?? null })}
        />
      </FilterPanel>

      <PatientMetrics stats={stats} />

      {isError ? (
        <p className="rounded-lg bg-rose-50 p-4 text-sm text-rose-700">
          {t("patients_load_error")}
        </p>
      ) : isLoading ? (
        <div className="space-y-2">
          {Array.from({ length: 6 }, (_, i) => (
            <div
              key={i}
              className="h-14 animate-pulse rounded-lg border border-gray-200 bg-white"
            />
          ))}
        </div>
      ) : data && data.results.length > 0 ? (
        <>
          <div className="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-sm">
            <table className="min-w-full divide-y divide-gray-100 text-sm">
              <thead className="bg-gray-50 text-left text-xs font-medium uppercase tracking-wide text-gray-500">
                <tr>
                  <th className="px-4 py-3">{t("name")}</th>
                  <th className="hidden px-4 py-3 sm:table-cell">{t("age")}</th>
                  <th className="px-4 py-3">{t("care_level")}</th>
                  <th className="hidden px-4 py-3 md:table-cell">
                    {t("gender")}
                  </th>
                  <th className="hidden px-4 py-3 md:table-cell">
                    {t("phone")}
                  </th>
                  <th className="hidden px-4 py-3 lg:table-cell">
                    {t("registered_on")}
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {data.results.map((p) => {
                  const age = ageFromDob(p.date_of_birth);
                  return (
                    <tr
                      key={p.id}
                      onClick={() => navigate(`/patients/${p.id}`)}
                      className="cursor-pointer transition hover:bg-primary-50/50"
                    >
                      <td className="px-4 py-3">
                        <a
                          href={`/patients/${p.id}`}
                          onClick={(e) => e.preventDefault()}
                          className="flex items-center gap-3 font-medium text-gray-900"
                        >
                          <span className="flex size-9 shrink-0 items-center justify-center rounded-full bg-primary-100 text-xs font-semibold text-primary-700">
                            {initials(p.name ?? "")}
                          </span>
                          {p.name}
                        </a>
                      </td>
                      <td className="hidden px-4 py-3 text-gray-600 sm:table-cell">
                        {age ?? "—"}
                      </td>
                      <td className="px-4 py-3">
                        {p.care_level ? (
                          <span
                            className={cn(
                              "rounded px-1.5 py-0.5 text-xs font-semibold",
                              CARE_LEVEL_STYLES[p.care_level],
                            )}
                          >
                            {p.care_level}
                          </span>
                        ) : (
                          "—"
                        )}
                      </td>
                      <td className="hidden px-4 py-3 text-gray-600 md:table-cell">
                        {t(genderKey(p.gender))}
                      </td>
                      <td className="hidden px-4 py-3 text-gray-600 md:table-cell">
                        {p.phone_number || "—"}
                      </td>
                      <td className="hidden px-4 py-3 text-gray-500 lg:table-cell">
                        {p.created_date
                          ? dayjs(p.created_date).format("D MMM YYYY")
                          : "—"}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
          {pages > 1 && (
            <div className="mt-5 flex items-center justify-center gap-3 text-sm">
              <Button
                variant="outline"
                size="sm"
                disabled={page === 0}
                onClick={() => setPage((p) => p - 1)}
              >
                {t("previous")}
              </Button>
              <span className="text-gray-600">
                {page + 1} / {pages}
              </span>
              <Button
                variant="outline"
                size="sm"
                disabled={page + 1 >= pages}
                onClick={() => setPage((p) => p + 1)}
              >
                {t("next")}
              </Button>
            </div>
          )}
        </>
      ) : (
        <div className="flex flex-col items-center rounded-xl border-2 border-dashed border-gray-200 bg-white px-6 py-16 text-center">
          <div className="mb-4 flex size-14 items-center justify-center rounded-full bg-primary-100">
            <Users className="size-7 text-primary-600" aria-hidden />
          </div>
          <p className="font-semibold text-gray-900">
            {isFiltered(filter) ? t("no_patients_found") : t("no_patients_yet")}
          </p>
          {!isFiltered(filter) && canWritePatients(user) && (
            <p className="mt-1 text-sm text-gray-500">
              {t("no_patients_hint")}
            </p>
          )}
        </div>
      )}
    </Page>
  );
}
