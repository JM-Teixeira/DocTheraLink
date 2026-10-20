import { useQuery } from "@tanstack/react-query";
import { Inbox, Plus } from "lucide-react";
import { navigate } from "raviger";
import { useState } from "react";
import { useTranslation } from "react-i18next";

import { cn } from "@/lib/utils";

import { Button } from "@/components/ui/button";

import Page from "@/components/Common/Page";
import ReferralCard from "@/components/Referral/ReferralCard";
import ReferralRoutes from "@/components/Referral/api";
import {
  canCreateReferral,
  isTherapist,
} from "@/components/Referral/referralUtils";
import { ReferralStatus } from "@/components/Referral/types";

import useAuthUser from "@/hooks/useAuthUser";

import query from "@/Utils/request/query";

const PAGE_SIZE = 24;

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
  const [page, setPage] = useState(0);

  const { data, isLoading, isError } = useQuery({
    queryKey: ["referrals", status, page],
    queryFn: query(ReferralRoutes.list, {
      queryParams: {
        status: status || undefined,
        limit: PAGE_SIZE,
        offset: page * PAGE_SIZE,
      },
    }),
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
      <p className="mb-5 max-w-2xl text-sm text-gray-600">
        {isTherapist(user)
          ? t("referral_inbox_description")
          : t("referrals_description")}
      </p>

      {/* Status filter chips */}
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
            onClick={() => {
              setStatus(f);
              setPage(0);
            }}
            className={cn(
              "rounded-full px-3.5 py-1.5 text-sm font-medium transition",
              status === f
                ? "bg-primary-600 text-white shadow-sm"
                : "bg-white text-gray-700 ring-1 ring-inset ring-gray-200 hover:bg-gray-50",
            )}
          >
            {f ? t(`referral_status__${f}`) : t("all")}
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
