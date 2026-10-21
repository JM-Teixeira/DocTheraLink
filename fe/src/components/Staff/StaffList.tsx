import { useQuery } from "@tanstack/react-query";
import dayjs from "dayjs";
import { Plus, Search, UserRound, X } from "lucide-react";
import { navigate } from "raviger";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";

import { cn } from "@/lib/utils";

import { Button } from "@/components/ui/button";

import { FilterChips, FilterPanel } from "@/components/Common/FilterChips";
import Page from "@/components/Common/Page";
import StaffRoutes from "@/components/Staff/api";
import {
  CLINIC_ROLES,
  STATUS_STYLES,
  roleKey,
  staffQuery,
  statusOf,
} from "@/components/Staff/staffUtils";
import {
  EMPLOYMENT_STATUSES,
  EmploymentStatus,
} from "@/components/Staff/types";

import useAuthUser from "@/hooks/useAuthUser";
import useDebouncedValue from "@/hooks/useDebouncedValue";

import { initials } from "@/Utils/person";
import query from "@/Utils/request/query";
import { canManageBoard, isAdmin } from "@/Utils/roles";

const PAGE_SIZE = 20;

/**
 * The people of the clinic: one search box and chips for role and status
 * (active by default; people who left stay on record under "Left").
 */
export default function StaffList() {
  const { t } = useTranslation();
  const user = useAuthUser();
  const admin = isAdmin(user);
  const [search, setSearch] = useState("");
  const [role, setRole] = useState("");
  const [clinic, setClinic] = useState("");
  const [status, setStatus] = useState<EmploymentStatus | "all">("active");
  const [page, setPage] = useState(0);
  const term = useDebouncedValue(search);

  // Any filter change → first page
  useEffect(() => setPage(0), [term, role, clinic, status]);

  const params = staffQuery({
    search: term,
    role,
    clinic,
    status,
    limit: PAGE_SIZE,
    offset: page * PAGE_SIZE,
  });
  const { data, isLoading, isError } = useQuery({
    queryKey: ["staff", params],
    queryFn: query(StaffRoutes.list, { queryParams: params }),
  });
  const { data: clinics } = useQuery({
    queryKey: ["staff-clinics"],
    queryFn: query(StaffRoutes.clinics, { queryParams: { limit: 100 } }),
    enabled: admin,
  });
  const total = data?.count ?? 0;
  const pages = Math.max(1, Math.ceil(total / PAGE_SIZE));
  const filtered = !!(term.trim() || role || clinic || status !== "active");

  return (
    <Page
      title={t("people")}
      hideBack
      options={
        canManageBoard(user) && (
          <Button variant="primary" onClick={() => navigate("/users/add")}>
            <Plus className="mr-1.5 size-4" aria-hidden />
            {t("add_user")}
          </Button>
        )
      }
    >
      <FilterPanel>
        <div className="flex flex-wrap items-center gap-3">
          <div className="relative min-w-60 flex-1">
            <Search
              className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-gray-400"
              aria-hidden
            />
            <input
              type="search"
              aria-label={t("search_people")}
              placeholder={t("search_people_placeholder")}
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="h-10 w-full rounded-lg border border-gray-300 pl-9 pr-3 text-sm shadow-sm focus:border-primary-500 focus:ring-primary-500"
            />
          </div>
          {admin && (
            <select
              aria-label={t("clinic")}
              value={clinic}
              onChange={(e) => setClinic(e.target.value)}
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
          {filtered && (
            <button
              type="button"
              onClick={() => {
                setSearch("");
                setRole("");
                setClinic("");
                setStatus("active");
              }}
              className="flex items-center gap-1 text-xs text-gray-500 hover:text-gray-800"
            >
              <X className="size-3.5" aria-hidden />
              {t("clear_filters")}
            </button>
          )}
        </div>
        <FilterChips
          label={t("role")}
          allLabel={t("all")}
          options={CLINIC_ROLES.map((r) => ({
            value: r as string,
            label: t(roleKey(r), { defaultValue: r }),
          }))}
          value={role ? [role] : []}
          onChange={(v) => setRole(v.at(-1) ?? "")}
        />
        <FilterChips
          label={t("status")}
          allLabel={t("all")}
          options={EMPLOYMENT_STATUSES.map((s) => ({
            value: s,
            label: t(`employment__${s}`),
          }))}
          value={status === "all" ? [] : [status]}
          onChange={(v) => setStatus(v.at(-1) ?? "all")}
        />
      </FilterPanel>

      <p className="mb-2 text-sm text-gray-500" aria-live="polite">
        {t("n_people", { count: total })}
      </p>

      {isError ? (
        <p className="rounded-lg bg-rose-50 p-4 text-sm text-rose-700">
          {t("people_load_error")}
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
      ) : data?.results?.length ? (
        <>
          <div className="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-sm">
            <table className="min-w-full divide-y divide-gray-100 text-sm">
              <thead className="bg-gray-50 text-left text-xs font-medium uppercase tracking-wide text-gray-500">
                <tr>
                  <th className="px-4 py-3">{t("name")}</th>
                  <th className="px-4 py-3">{t("role")}</th>
                  {admin && (
                    <th className="hidden px-4 py-3 md:table-cell">
                      {t("clinic")}
                    </th>
                  )}
                  <th className="px-4 py-3">{t("status")}</th>
                  <th className="hidden px-4 py-3 lg:table-cell">
                    {t("last_login")}
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {data.results.map((p) => {
                  const st = statusOf(p.employment_status);
                  const name =
                    `${p.first_name} ${p.last_name}`.trim() || p.username;
                  return (
                    <tr
                      key={p.id}
                      onClick={() => navigate(`/users/${p.username}`)}
                      className={cn(
                        "cursor-pointer transition hover:bg-primary-50/50",
                        st === "left" && "text-gray-500",
                      )}
                    >
                      <td className="px-4 py-3">
                        <a
                          href={`/users/${p.username}`}
                          onClick={(e) => e.preventDefault()}
                          className="flex items-center gap-3"
                        >
                          <span className="flex size-9 shrink-0 items-center justify-center rounded-full bg-primary-100 text-xs font-semibold text-primary-700">
                            {initials(name)}
                          </span>
                          <span>
                            <span className="block font-medium text-gray-900">
                              {name}
                            </span>
                            <span className="block text-xs text-gray-500">
                              {p.username}
                            </span>
                          </span>
                        </a>
                      </td>
                      <td className="px-4 py-3 text-gray-600">
                        {t(roleKey(p.user_type), { defaultValue: p.user_type })}
                      </td>
                      {admin && (
                        <td className="hidden px-4 py-3 text-gray-600 md:table-cell">
                          {p.clinic_name || "—"}
                        </td>
                      )}
                      <td className="px-4 py-3">
                        <span
                          className={cn(
                            "rounded-full px-2 py-0.5 text-xs font-medium",
                            STATUS_STYLES[st],
                          )}
                        >
                          {t(`employment__${st}`)}
                        </span>
                        {st !== "active" && p.employment_since && (
                          <span className="ml-1 text-xs text-gray-500">
                            {dayjs(p.employment_since).format("D MMM YYYY")}
                          </span>
                        )}
                      </td>
                      <td className="hidden px-4 py-3 text-gray-500 lg:table-cell">
                        {p.last_login
                          ? dayjs(p.last_login).format("D MMM YYYY, HH:mm")
                          : t("never")}
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
            <UserRound className="size-7 text-primary-600" aria-hidden />
          </div>
          <p className="font-semibold text-gray-900">{t("no_people_found")}</p>
        </div>
      )}
    </Page>
  );
}
