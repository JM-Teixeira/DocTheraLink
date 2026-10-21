import { useQuery } from "@tanstack/react-query";
import dayjs from "dayjs";
import { ExternalLink, FileText, Phone, TriangleAlert } from "lucide-react";
import { navigate } from "raviger";
import { useTranslation } from "react-i18next";

import { cn } from "@/lib/utils";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

import ReferralNotes from "@/components/Referral/ReferralNotes";
import ReferralStatusBadge from "@/components/Referral/ReferralStatusBadge";
import ReferralRoutes from "@/components/Referral/api";
import { ageFromDob, fullName } from "@/components/Referral/referralUtils";
import { Referral } from "@/components/Referral/types";
import { CARE_LEVEL_STYLES } from "@/components/TimeBoard/boardUtils";

import useAuthUser from "@/hooks/useAuthUser";

import query from "@/Utils/request/query";

const OPEN = ["pending", "accepted", "in_progress"];

/** The referral to show: the open one, else the most recent. */
export function currentReferral(referrals: Referral[]) {
  return referrals.find((r) => OPEN.includes(r.status)) ?? referrals[0];
}

/**
 * Pop-up from the time board: who the patient is (age, contact, care level,
 * allergies), what the doctor asked for, and the doctor–physio thread.
 * Users who may not see referrals simply don't get that part.
 */
export default function PatientQuickView({
  patientId,
  onClose,
}: {
  patientId: string | null;
  onClose: () => void;
}) {
  const { t } = useTranslation();
  const user = useAuthUser();
  const { data: patient } = useQuery({
    queryKey: ["patient", patientId],
    queryFn: query(ReferralRoutes.getPatient, {
      pathParams: { id: patientId ?? "" },
    }),
    enabled: !!patientId,
  });
  const { data: referrals } = useQuery({
    queryKey: ["patient-referrals", patientId],
    queryFn: query(ReferralRoutes.list, {
      queryParams: { patient: patientId ?? "", limit: 5 },
      silent: true, // roles without referral access just see the patient
    }),
    enabled: !!patientId,
    retry: false,
  });
  const referral = referrals?.results.length
    ? currentReferral(referrals.results)
    : undefined;
  const age = ageFromDob(patient?.date_of_birth as string | undefined);
  const level =
    (patient as { care_level?: number } | undefined)?.care_level ?? 1;

  return (
    <Dialog open={!!patientId} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle className="flex flex-wrap items-center gap-2">
            {patient?.name ?? t("loading")}
            <span
              className={cn(
                "rounded-full px-2 py-0.5 text-xs font-medium",
                CARE_LEVEL_STYLES[level],
              )}
            >
              {t("care_level_n", { level })}
            </span>
          </DialogTitle>
          <DialogDescription className="flex flex-wrap gap-x-4 gap-y-1">
            <span>
              {age !== null ? t("age_years", { age }) : t("age_unknown")}
            </span>
            {patient?.date_of_birth && (
              <span>{dayjs(patient.date_of_birth).format("D MMM YYYY")}</span>
            )}
            {patient?.phone_number && (
              <span className="flex items-center gap-1">
                <Phone className="size-3.5" aria-hidden />
                {patient.phone_number}
              </span>
            )}
          </DialogDescription>
        </DialogHeader>

        <p className="text-xs text-gray-500">
          {t(`care_level_help__${level}`)}
        </p>
        {patient?.allergies && (
          <p className="flex items-center gap-2 rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-800">
            <TriangleAlert className="size-4 shrink-0" aria-hidden />
            {t("allergies")}: {patient.allergies}
          </p>
        )}

        {referral && (
          <section className="rounded-xl border border-gray-200 p-4">
            <div className="mb-2 flex flex-wrap items-center justify-between gap-2">
              <h3 className="flex items-center gap-2 text-sm font-semibold text-gray-900">
                <FileText className="size-4 text-primary-600" aria-hidden />
                {t("doctors_referral")}
              </h3>
              <ReferralStatusBadge status={referral.status} />
            </div>
            <dl className="grid gap-3 text-sm sm:grid-cols-2">
              <Item label={t("diagnosis")}>
                {referral.diagnosis}
                {referral.icd_code && ` (${referral.icd_code})`}
              </Item>
              <Item label={t("referring_doctor")}>
                {fullName(referral.doctor)}
              </Item>
              <Item label={t("referral_reason")} wide>
                {referral.reason}
              </Item>
              {referral.goals && (
                <Item label={t("treatment_goals")} wide>
                  {referral.goals}
                </Item>
              )}
              {referral.precautions && (
                <Item label={t("precautions")} wide>
                  <span className="text-rose-700">{referral.precautions}</span>
                </Item>
              )}
              {referral.recommended_sessions && (
                <Item label={t("recommended_sessions")}>
                  {referral.recommended_sessions}
                </Item>
              )}
            </dl>
          </section>
        )}

        {referral && <ReferralNotes referral={referral} user={user} />}

        <div className="flex flex-wrap justify-end gap-2">
          {referral && (
            <Button
              variant="outline"
              onClick={() => navigate(`/referrals/${referral.id}`)}
            >
              <ExternalLink className="mr-2 size-4" aria-hidden />
              {t("open_referral")}
            </Button>
          )}
          <Button
            variant="primary"
            onClick={() => patientId && navigate(`/patients/${patientId}`)}
          >
            <ExternalLink className="mr-2 size-4" aria-hidden />
            {t("open_patient_page")}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}

function Item({
  label,
  wide,
  children,
}: {
  label: string;
  wide?: boolean;
  children: React.ReactNode;
}) {
  return (
    <div className={wide ? "sm:col-span-2" : undefined}>
      <dt className="text-xs font-medium uppercase tracking-wide text-gray-500">
        {label}
      </dt>
      <dd className="whitespace-pre-wrap text-gray-900">{children}</dd>
    </div>
  );
}
