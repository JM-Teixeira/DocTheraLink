import { useQuery } from "@tanstack/react-query";
import { CircleHelp, Inbox, Plus, Search, X } from "lucide-react";
import { navigate } from "raviger";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";

import { cn } from "@/lib/utils";

import { Button } from "@/components/ui/button";

import DateInput from "@/components/Common/DateInput";
import { chipClass } from "@/components/Common/FilterChips";
import Page from "@/components/Common/Page";
import ReferralCard from "@/components/Referral/ReferralCard";
import ReferralHelp from "@/components/Referral/ReferralHelp";
import ReferralRoutes from "@/components/Referral/api";
import {
  canCreateReferral,
  isDoctor,
  isTherapist,
  sumCounts,
} from "@/components/Referral/referralUtils";
import { ReferralPriority, ReferralStatus } from "@/components/Referral/types";
import StaffRoutes from "@/components/Staff/api";

import useAuthUser from "@/hooks/useAuthUser";
import useDebouncedValue from "@/hooks/useDebouncedValue";

import query from "@/Utils/request/query";

const PAGE_SIZE = 24;
const selectClass =
  "h-10 rounded-lg border border-gray-300 px-3 text-sm shadow-sm focus:border-primary-500 focus:ring-primary-500";

// "" means all statuses
const FILTERS: Array<ReferralStatus | ""> = [
  "",
  "pending",
  "accepted",
  "in_progress",
  "completed",
  "declined",
  "cancelled",
];

/**
 * Referral list. The backend scopes results by role: doctors see the
 * referrals they made, therapists their inbox, admins everything.
 */
export default function ReferralList() {
  const { t } = useTranslation();
  const user = useAuthUser();
  const [status, setStatus] = useState<ReferralStatus | "">("");
  const [priority, setPriority] = useState<ReferralPriority | "">("");
  const [search, setSearch] = useState("");
  const term = useDebouncedValue(search.trim());
  const [doctor, setDoctor] = useState("");
  const [therapist, setTherapist] = useState("");
  const [from, setFrom] = useState("");
  const [to, setTo] = useState("");
  // Only referrals with questions waiting for me
  const [waiting, setWaiting] = useState(false);
  const [page, setPage] = useState(0);

  // Any filter change goes back to the first page
  useEffect(
    () => setPage(0),
    [status, priority, term, doctor, therapist, from, to, waiting],
  );

  // Shared by the list and the per-status counts
  const filters = {
    search: term || undefined,
    priority: priority || undefined,
    doctor: doctor || undefined,
    therapist: therapist || undefined,
    from: from || undefined,
    to: to || undefined,
    open_questions: waiting || undefined,
  };
  const anyFilter = Object.values(filters).some(Boolean);

  // A doctor already sees only their referrals, a physio only their inbox
  const pickDoctor = !isDoctor(user);
  const pickTherapist = !isTherapist(user);
  const people = (role: string, enabled: boolean) => ({
    queryKey: ["staff", "by-role", role],
    queryFn: query(StaffRoutes.list, {
      queryParams: { user_type: role, limit: 100 },
    }),
    enabled,
  });
  const { data: doctors } = useQuery(people("Doctor", pickDoctor));
  const { data: therapists } = useQuery(
    people("PhysioTherapist", pickTherapist),
  );

  const { data, isLoading, isError } = useQuery({
    queryKey: ["referrals", status, filters, page],
    queryFn: query(ReferralRoutes.list, {
      queryParams: {
        ...filters,
        status: status || undefined,
        limit: PAGE_SIZE,
        offset: page * PAGE_SIZE,
      },
    }),
  });

  const { data: counts } = useQuery({
    queryKey: ["referrals", "counts", filters],
    queryFn: query(ReferralRoutes.counts, { queryParams: filters }),
  });

  const total = data?.count ?? 0;
  const pages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  return (
    <Page
      title={isTherapist(user) ? t("referral_inbox") : t("referrals")}
      hideBack
      options={
        canCreateReferral(user) && (
          <Button variant="primary" onClick={() => navigate("/referrals/new")}>
            <Plus className="mr-1.5 size-4" aria-hidden />
            {t("new_referral")}
          </Button>
        )
      }
    >
      <ReferralHelp />
      <p className="mb-5 max-w-2xl text-sm text-gray-600">
        {isTherapist(user)
          ? t("referral_inbox_description")
          : t("referrals_description")}
      </p>

      {/* Search and priority */}
      <div className="mb-4 flex flex-wrap items-center gap-3">
        <div className="relative w-full max-w-sm">
          <Search
            className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-gray-400"
            aria-hidden
          />
          <input
            type="search"
            aria-label={t("search_patient")}
            placeholder={t("search_patient_by_name")}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="h-10 w-full rounded-lg border border-gray-300 pl-9 pr-3 text-sm shadow-sm focus:border-primary-500 focus:ring-primary-500"
          />
        </div>
        <select
          aria-label={t("priority")}
          value={priority}
          onChange={(e) => setPriority(e.target.value as ReferralPriority | "")}
          className="h-10 rounded-lg border border-gray-300 px-3 text-sm shadow-sm focus:border-primary-500 focus:ring-primary-500"
        >
          <option value="">{t("all_priorities")}</option>
          <option value="urgent">{t("referral_priority__urgent")}</option>
          <option value="routine">{t("referral_priority__routine")}</option>
        </select>
        {pickDoctor && (
          <select
            aria-label={t("referring_doctor")}
            value={doctor}
            onChange={(e) => setDoctor(e.target.value)}
            className={selectClass}
          >
            <option value="">{t("all_doctors")}</option>
            {doctors?.results.map((d) => (
              <option key={d.id} value={d.id}>
                {d.first_name} {d.last_name}
              </option>
            ))}
          </select>
        )}
        {pickTherapist && (
          <select
            aria-label={t("physiotherapist")}
            value={therapist}
            onChange={(e) => setTherapist(e.target.value)}
            className={selectClass}
          >
            <option value="">{t("all_physios")}</option>
            {therapists?.results.map((d) => (
              <option key={d.id} value={d.id}>
                {d.first_name} {d.last_name}
              </option>
            ))}
          </select>
        )}
        <label className="flex items-center gap-1.5 text-sm text-gray-600">
          {t("from")}
          <DateInput value={from} onChange={setFrom} className="w-40" />
        </label>
        <label className="flex items-center gap-1.5 text-sm text-gray-600">
          {t("until")}
          <DateInput
            value={to}
            min={from || undefined}
            onChange={setTo}
            className="w-40"
          />
        </label>
        <button
          type="button"
          aria-pressed={waiting}
          onClick={() => setWaiting((w) => !w)}
          className={chipClass(waiting)}
        >
          <CircleHelp className="size-3.5" aria-hidden />
          {t("questions_for_me")}
        </button>
        {anyFilter && (
          <button
            type="button"
            onClick={() => {
              setSearch("");
              setPriority("");
              setDoctor("");
              setTherapist("");
              setFrom("");
              setTo("");
              setWaiting(false);
            }}
            className="flex items-center gap-1 text-xs text-gray-500 hover:text-gray-800"
          >
            <X className="size-3.5" aria-hidden />
            {t("clear_filters")}
          </button>
        )}
      </div>

      {/* Status filter chips with counts */}
      <div
        role="tablist"
        aria-label={t("filter_by_status")}
        className="mb-6 flex flex-wrap gap-2"
      >
        {FILTERS.map((f) => (
          <button
            key={f || "all"}
            role="tab"
            aria-selected={status === f}
            onClick={() => setStatus(f)}
            className={cn(
              "rounded-full px-3.5 py-1.5 text-sm font-medium transition",
              status === f
                ? "bg-primary-600 text-white shadow-sm"
                : "bg-white text-gray-700 ring-1 ring-inset ring-gray-200 hover:bg-gray-50",
            )}
          >
            {f ? t(`referral_status__${f}`) : t("all")}
            {counts && (
              <span
                className={cn(
                  "ml-1.5 rounded-full px-1.5 text-xs tabular-nums",
                  status === f
                    ? "bg-white text-primary-700"
                    : "bg-gray-100 text-gray-600",
                )}
              >
                {f ? (counts[f] ?? 0) : sumCounts(counts)}
              </span>
            )}
          </button>
        ))}
      </div>

      {isLoading ? (
        <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
          {Array.from({ length: 6 }, (_, i) => (
            <div
              key={i}
              className="h-44 animate-pulse rounded-xl border border-gray-200 bg-white"
            />
          ))}
        </div>
      ) : isError ? (
        <p className="rounded-lg bg-rose-50 p-4 text-sm text-rose-700">
          {t("referrals_load_error")}
        </p>
      ) : data && data.results.length > 0 ? (
        <>
          <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
            {data.results.map((r) => (
              <ReferralCard key={r.id} referral={r} />
            ))}
          </div>
          {pages > 1 && (
            <div className="mt-6 flex items-center justify-center gap-3 text-sm">
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
        <div className="flex flex-col items-center justify-center rounded-xl border-2 border-dashed border-gray-200 bg-white px-6 py-16 text-center">
          <div className="mb-4 flex size-14 items-center justify-center rounded-full bg-primary-100">
            <Inbox className="size-7 text-primary-600" aria-hidden />
          </div>
          <p className="font-semibold text-gray-900">{t("no_referrals")}</p>
          <p className="mt-1 max-w-sm text-sm text-gray-500">
            {canCreateReferral(user)
              ? t("no_referrals_doctor_hint")
              : t("no_referrals_therapist_hint")}
          </p>
        </div>
      )}
    </Page>
  );
}
