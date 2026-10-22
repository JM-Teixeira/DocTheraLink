import { useQuery } from "@tanstack/react-query";
import { Building2, Plus, Search, X } from "lucide-react";
import { navigate } from "raviger";
import { useState } from "react";
import { useTranslation } from "react-i18next";

import { Button } from "@/components/ui/button";

import ExportMenu from "@/components/Common/Export";
import { FilterChips, FilterPanel } from "@/components/Common/FilterChips";
import Page from "@/components/Common/Page";
import { FacilityCard } from "@/components/Facility/FacilityCard";
import { FacilityModel } from "@/components/Facility/models";

import useAuthUser from "@/hooks/useAuthUser";

import routes from "@/Utils/request/api";
import query from "@/Utils/request/query";
import { isAdmin } from "@/Utils/roles";

import { clinicDistrict, clinicType, filterClinics } from "./facilityFilters";

/**
 * The clinics: one search box and chips for district and type, like the
 * other lists. Staff see their own clinic; admins all of them (and add new
 * ones). There are few clinics, so filtering happens in the browser.
 */
export const FacilityList = () => {
  const { t } = useTranslation();
  const user = useAuthUser();
  const admin = isAdmin(user);
  const [search, setSearch] = useState("");
  const [district, setDistrict] = useState("");
  const [type, setType] = useState("");

  const { data, isLoading, isError } = useQuery({
    queryKey: ["clinics-all"],
    queryFn: query(routes.getPermittedFacilities, {
      queryParams: { limit: 100 },
    }),
  });
  const all: FacilityModel[] = data?.results ?? [];
  const districts = [
    ...new Set(all.map(clinicDistrict).filter(Boolean)),
  ].sort() as string[];
  const types = [
    ...new Set(all.map(clinicType).filter(Boolean)),
  ].sort() as string[];
  const shown = filterClinics(all, { search, district, type });
  const filtered = !!(search.trim() || district || type);

  return (
    <Page
      title={t("facilities")}
      breadcrumbs={false}
      hideBack
      options={
        admin && (
          <div className="flex items-center gap-2">
            <ExportMenu
              exportItems={[
                {
                  label: t("facilities"),
                  route: routes.downloadFacility,
                  filePrefix: "clinicas",
                },
              ]}
            />
            <Button
              variant="primary"
              onClick={() => navigate("/facility/create")}
            >
              <Plus className="mr-1.5 size-4" aria-hidden />
              {t("create_facility")}
            </Button>
          </div>
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
              aria-label={t("search_clinics")}
              placeholder={t("search_clinics_placeholder")}
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="h-10 w-full rounded-lg border border-gray-300 pl-9 pr-3 text-sm shadow-sm focus:border-primary-500 focus:ring-primary-500"
            />
          </div>
          {filtered && (
            <button
              type="button"
              onClick={() => {
                setSearch("");
                setDistrict("");
                setType("");
              }}
              className="flex items-center gap-1 text-xs text-gray-500 hover:text-gray-800"
            >
              <X className="size-3.5" aria-hidden />
              {t("clear_filters")}
            </button>
          )}
        </div>
        {districts.length > 1 && (
          <FilterChips
            label={t("district")}
            allLabel={t("all")}
            options={districts.map((d) => ({ value: d, label: d }))}
            value={district ? [district] : []}
            onChange={(v) => setDistrict(v.at(-1) ?? "")}
          />
        )}
        {types.length > 1 && (
          <FilterChips
            label={t("type")}
            allLabel={t("all")}
            options={types.map((d) => ({ value: d, label: d }))}
            value={type ? [type] : []}
            onChange={(v) => setType(v.at(-1) ?? "")}
          />
        )}
      </FilterPanel>

      <p className="mb-2 text-sm text-gray-500" aria-live="polite">
        {t("n_clinics", { count: shown.length })}
      </p>

      {isError ? (
        <p className="rounded-lg bg-rose-50 p-4 text-sm text-rose-700">
          {t("clinics_load_error")}
        </p>
      ) : isLoading ? (
        <div className="grid gap-4 lg:grid-cols-2">
          {Array.from({ length: 2 }, (_, i) => (
            <div
              key={i}
              className="h-28 animate-pulse rounded-xl border border-gray-200 bg-white"
            />
          ))}
        </div>
      ) : shown.length ? (
        <div className="grid gap-4 lg:grid-cols-2">
          {shown.map((f) => (
            <FacilityCard key={f.id} facility={f} />
          ))}
        </div>
      ) : (
        <div className="flex flex-col items-center rounded-xl border-2 border-dashed border-gray-200 bg-white px-6 py-16 text-center">
          <div className="mb-4 flex size-14 items-center justify-center rounded-full bg-primary-100">
            <Building2 className="size-7 text-primary-600" aria-hidden />
          </div>
          <p className="font-semibold text-gray-900">
            {filtered ? t("no_facilities") : t("no_clinics_yet")}
          </p>
        </div>
      )}
    </Page>
  );
};
