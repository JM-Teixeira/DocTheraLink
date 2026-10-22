import dayjs from "dayjs";
import relativeTime from "dayjs/plugin/relativeTime";
import {
  ArrowRight,
  CircleHelp,
  Stethoscope,
  TriangleAlert,
} from "lucide-react";
import { Link } from "raviger";
import { useTranslation } from "react-i18next";

import { cn } from "@/lib/utils";

import ReferralStatusBadge from "@/components/Referral/ReferralStatusBadge";
import {
  STATUS_STYLES,
  ageFromDob,
  fullName,
  hasDoctorTitle,
  initials,
} from "@/components/Referral/referralUtils";
import { Referral } from "@/components/Referral/types";

dayjs.extend(relativeTime);

/** Summary card for one referral in the list. */
export default function ReferralCard({ referral }: { referral: Referral }) {
  const { t } = useTranslation();
  const patientName = referral.patient?.name ?? t("unknown_patient");
  const age = ageFromDob(referral.patient?.date_of_birth);
  const urgent = referral.priority === "urgent";
  const doctor = fullName(referral.doctor);

  return (
    <Link
      href={`/referrals/${referral.id}`}
      className="group relative flex overflow-hidden rounded-xl border border-gray-200 bg-white shadow-sm transition hover:-translate-y-0.5 hover:border-primary-300 hover:shadow-md focus:outline-none focus-visible:ring-2 focus-visible:ring-primary-500"
    >
      {/* Status accent bar */}
      <span
        className={cn("w-1.5 shrink-0", STATUS_STYLES[referral.status].dot)}
        aria-hidden
      />

      <div className="flex min-w-0 flex-1 flex-col gap-3 p-4">
        <div className="flex items-start justify-between gap-3">
          <div className="flex min-w-0 items-center gap-3">
            <div className="flex size-10 shrink-0 items-center justify-center rounded-full bg-primary-100 text-sm font-semibold text-primary-700">
              {initials(patientName)}
            </div>
            <div className="min-w-0">
              <p className="truncate font-semibold text-gray-900">
                {patientName}
              </p>
              <p className="text-xs text-gray-500">
                {age !== null ? t("age_years", { age }) : t("age_unknown")}
              </p>
            </div>
          </div>
          <div className="flex shrink-0 flex-col items-end gap-1.5">
            <ReferralStatusBadge status={referral.status} />
            {!!referral.open_questions && (
              <span className="inline-flex items-center gap-1 rounded-full bg-amber-50 px-2 py-0.5 text-xs font-semibold text-amber-800">
                <CircleHelp className="size-3.5" aria-hidden />
                {t("n_questions_waiting", { count: referral.open_questions })}
              </span>
            )}
            {urgent && (
              <span className="inline-flex items-center gap-1 text-xs font-semibold text-rose-600">
                <TriangleAlert className="size-3.5" aria-hidden />
                {t("referral_priority__urgent")}
              </span>
            )}
          </div>
        </div>

        <div>
          <p className="flex items-center gap-1.5 text-sm font-medium text-gray-800">
            <Stethoscope className="size-4 text-primary-600" aria-hidden />
            <span className="truncate">{referral.diagnosis}</span>
          </p>
          <p className="mt-1 line-clamp-2 text-sm text-gray-600">
            {referral.reason}
          </p>
        </div>

        <div className="flex items-center justify-between gap-2 border-t border-gray-100 pt-3 text-xs text-gray-500">
          <span className="flex min-w-0 items-center gap-1.5">
            <span className="truncate">
              {hasDoctorTitle(doctor)
                ? doctor
                : t("doctor_name", { name: doctor })}
            </span>
            <ArrowRight className="size-3.5 shrink-0" aria-hidden />
            <span className="truncate">
              {referral.therapist
                ? fullName(referral.therapist)
                : t("referral_unassigned")}
            </span>
          </span>
          <time
            dateTime={referral.created_date}
            title={dayjs(referral.created_date).format("LLL")}
            className="shrink-0"
          >
            {dayjs(referral.created_date).fromNow()}
          </time>
        </div>
      </div>
    </Link>
  );
}
