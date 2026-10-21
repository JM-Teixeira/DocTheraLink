import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import dayjs from "dayjs";
import {
  CalendarDays,
  ClipboardList,
  ShieldAlert,
  Target,
  TriangleAlert,
} from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";

import { Button } from "@/components/ui/button";

import Loading from "@/components/Common/Loading";
import Page from "@/components/Common/Page";
import ReasonDialog from "@/components/Referral/ReasonDialog";
import ReferralAttachments from "@/components/Referral/ReferralAttachments";
import ReferralNotes from "@/components/Referral/ReferralNotes";
import ReferralProgress from "@/components/Referral/ReferralProgress";
import ReferralStatusBadge from "@/components/Referral/ReferralStatusBadge";
import TherapistDialog from "@/components/Referral/TherapistDialog";
import ReferralRoutes from "@/components/Referral/api";
import {
  actionNeedsReason,
  ageFromDob,
  allowedActions,
  fullName,
  hasDoctorTitle,
} from "@/components/Referral/referralUtils";
import { Referral, ReferralAction } from "@/components/Referral/types";

import useAuthUser from "@/hooks/useAuthUser";

import * as Notification from "@/Utils/Notifications";
import { callApi } from "@/Utils/request/query";

/** Runs one workflow action against the API. */
function runAction(
  id: string,
  action: ReferralAction,
  reason?: string,
  therapistId?: number,
) {
  const pathParams = { id };
  switch (action) {
    case "reassign":
      return callApi(ReferralRoutes.reassign, {
        pathParams,
        body: { therapist_id: therapistId!, reason: reason! },
      });
    case "reopen":
      return callApi(ReferralRoutes.reopen, {
        pathParams,
        body: { therapist_id: therapistId, reason: reason! },
      });
    case "accept":
      return callApi(ReferralRoutes.accept, { pathParams, body: {} });
    case "start":
      return callApi(ReferralRoutes.start, { pathParams, body: {} });
    case "decline":
      return callApi(ReferralRoutes.decline, {
        pathParams,
        body: { reason: reason! },
      });
    case "complete":
      return callApi(ReferralRoutes.complete, {
        pathParams,
        body: { reason: reason! },
      });
    case "cancel":
      return callApi(ReferralRoutes.cancel, {
        pathParams,
        body: { reason: reason! },
      });
  }
}

const ACTION_VARIANT: Record<
  ReferralAction,
  "primary" | "outline" | "destructive"
> = {
  accept: "primary",
  start: "primary",
  complete: "primary",
  decline: "outline",
  cancel: "destructive",
  reassign: "outline",
  reopen: "primary",
};

export default function ReferralDetail({ id }: { id: string }) {
  const { t } = useTranslation();
  const user = useAuthUser();
  const queryClient = useQueryClient();
  const [reasonFor, setReasonFor] = useState<ReferralAction | null>(null);
  const [therapistFor, setTherapistFor] = useState<
    "reassign" | "reopen" | null
  >(null);

  const {
    data: referral,
    isLoading,
    isError,
  } = useQuery({
    queryKey: ["referral", id],
    queryFn: ({ signal }) =>
      callApi(ReferralRoutes.get, { pathParams: { id }, signal }),
  });

  const { mutate: act, isPending } = useMutation({
    mutationFn: (v: {
      action: ReferralAction;
      reason?: string;
      therapistId?: number;
    }) => runAction(id, v.action, v.reason, v.therapistId),
    onSuccess: (updated: Referral, v) => {
      queryClient.setQueryData(["referral", id], updated);
      queryClient.invalidateQueries({ queryKey: ["referral-notes", id] });
      queryClient.invalidateQueries({ queryKey: ["referrals"] });
      setReasonFor(null);
      setTherapistFor(null);
      Notification.Success({ msg: t(`referral_action_done__${v.action}`) });
    },
    onError: () => Notification.Error({ msg: t("referral_action_failed") }),
  });

  if (isLoading) return <Loading />;
  if (isError || !referral) {
    return (
      <Page title={t("referral")} backUrl="/referrals">
        <p className="rounded-lg bg-rose-50 p-4 text-sm text-rose-700">
          {t("referral_not_found")}
        </p>
      </Page>
    );
  }

  const actions = allowedActions(referral, user);
  const age = ageFromDob(referral.patient?.date_of_birth);
  const doctor = fullName(referral.doctor);

  return (
    <Page
      title={t("referral")}
      backUrl="/referrals"
      // Show the patient's name in the breadcrumb instead of the referral UUID
      crumbsReplacements={{ [id]: { name: referral.patient?.name ?? "" } }}
    >
      <div className="grid gap-6 lg:grid-cols-5">
        {/* Left: clinical content and workflow */}
        <div className="flex flex-col gap-6 lg:col-span-3">
          <section className="rounded-xl border border-gray-200 bg-white p-5 shadow-sm">
            <div className="flex flex-wrap items-start justify-between gap-4">
              <div>
                <h2 className="text-xl font-semibold text-gray-900">
                  {referral.patient?.name}
                </h2>
                <p className="text-sm text-gray-500">
                  {age !== null ? t("age_years", { age }) : t("age_unknown")}
                  {referral.patient?.phone_number &&
                    ` · ${referral.patient.phone_number}`}
                </p>
              </div>
              <div className="flex items-center gap-2">
                {referral.priority === "urgent" && (
                  <span className="inline-flex items-center gap-1 rounded-full bg-rose-50 px-2.5 py-0.5 text-xs font-semibold text-rose-700 ring-1 ring-inset ring-rose-200">
                    <TriangleAlert className="size-3.5" aria-hidden />
                    {t("referral_priority__urgent")}
                  </span>
                )}
                <ReferralStatusBadge status={referral.status} />
              </div>
            </div>

            <div className="mt-6">
              <ReferralProgress referral={referral} />
            </div>

            {actions.length > 0 && (
              <div className="mt-6 flex flex-wrap gap-2 border-t border-gray-100 pt-5">
                {actions.map((a) => (
                  <Button
                    key={a}
                    variant={ACTION_VARIANT[a]}
                    disabled={isPending}
                    onClick={() => {
                      if (a === "reassign" || a === "reopen")
                        setTherapistFor(a);
                      else if (actionNeedsReason(a)) setReasonFor(a);
                      else act({ action: a });
                    }}
                  >
                    {t(`referral_action__${a}`)}
                  </Button>
                ))}
              </div>
            )}
          </section>

          {/* Closing remarks */}
          {referral.outcome_summary && (
            <Callout
              tone="success"
              title={t("outcome_summary")}
              body={referral.outcome_summary}
            />
          )}
          {referral.decline_reason && (
            <Callout
              tone="danger"
              title={t("decline_reason")}
              body={referral.decline_reason}
            />
          )}
          {referral.cancel_reason && (
            <Callout
              tone="neutral"
              title={t("cancel_reason")}
              body={referral.cancel_reason}
            />
          )}

          <section className="rounded-xl border border-gray-200 bg-white p-5 shadow-sm">
            <h3 className="mb-4 text-base font-semibold text-gray-900">
              {t("clinical_details")}
            </h3>
            <dl className="grid gap-5 sm:grid-cols-2">
              <Detail icon={ClipboardList} label={t("diagnosis")}>
                {referral.diagnosis}
                {referral.icd_code && (
                  <span className="ml-2 rounded bg-gray-100 px-1.5 py-0.5 font-mono text-xs text-gray-600">
                    {referral.icd_code}
                  </span>
                )}
              </Detail>
              <Detail icon={CalendarDays} label={t("recommended_sessions")}>
                {referral.recommended_sessions ?? "—"}
              </Detail>
              <Detail icon={ClipboardList} label={t("referral_reason")} wide>
                {referral.reason}
              </Detail>
              {referral.goals && (
                <Detail icon={Target} label={t("treatment_goals")} wide>
                  {referral.goals}
                </Detail>
              )}
              {referral.precautions && (
                <Detail icon={ShieldAlert} label={t("precautions")} wide>
                  {referral.precautions}
                </Detail>
              )}
            </dl>
          </section>

          <section className="grid gap-4 rounded-xl border border-gray-200 bg-white p-5 text-sm shadow-sm sm:grid-cols-2">
            <Person
              label={t("referring_doctor")}
              name={
                hasDoctorTitle(doctor)
                  ? doctor
                  : t("doctor_name", { name: doctor })
              }
            />
            <Person
              label={t("physiotherapist")}
              name={
                referral.therapist
                  ? fullName(referral.therapist)
                  : t("referral_unassigned")
              }
            />
            <Timestamp label={t("referred_on")} value={referral.created_date} />
            {referral.accepted_at && (
              <Timestamp
                label={t("accepted_on")}
                value={referral.accepted_at}
              />
            )}
            {referral.started_at && (
              <Timestamp label={t("started_on")} value={referral.started_at} />
            )}
            {referral.completed_at && (
              <Timestamp
                label={t("completed_on")}
                value={referral.completed_at}
              />
            )}
          </section>
        </div>

        {/* Right: conversation and shared files */}
        <div className="flex flex-col gap-6 lg:col-span-2">
          <ReferralNotes referral={referral} user={user} />
          <ReferralAttachments referral={referral} user={user} />
        </div>
      </div>

      <TherapistDialog
        mode={therapistFor}
        currentTherapistId={referral.therapist_id}
        pending={isPending}
        onClose={() => setTherapistFor(null)}
        onConfirm={({ therapistId, reason }) =>
          therapistFor && act({ action: therapistFor, reason, therapistId })
        }
      />
      <ReasonDialog
        action={reasonFor}
        pending={isPending}
        onClose={() => setReasonFor(null)}
        onConfirm={(reason) => reasonFor && act({ action: reasonFor, reason })}
      />
    </Page>
  );
}

function Detail({
  icon: Icon,
  label,
  wide,
  children,
}: {
  icon: typeof Target;
  label: string;
  wide?: boolean;
  children: React.ReactNode;
}) {
  return (
    <div className={wide ? "sm:col-span-2" : undefined}>
      <dt className="mb-1 flex items-center gap-1.5 text-xs font-medium uppercase tracking-wide text-gray-500">
        <Icon className="size-3.5" aria-hidden />
        {label}
      </dt>
      <dd className="whitespace-pre-wrap text-sm text-gray-900">{children}</dd>
    </div>
  );
}

function Person({ label, name }: { label: string; name: string }) {
  return (
    <div>
      <p className="text-xs font-medium uppercase tracking-wide text-gray-500">
        {label}
      </p>
      <p className="mt-0.5 font-medium text-gray-900">{name}</p>
    </div>
  );
}

function Timestamp({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <p className="text-xs font-medium uppercase tracking-wide text-gray-500">
        {label}
      </p>
      <time dateTime={value} className="mt-0.5 block text-gray-900">
        {dayjs(value).format("D MMM YYYY, HH:mm")}
      </time>
    </div>
  );
}

const CALLOUT_TONES = {
  success: "border-emerald-200 bg-emerald-50 text-emerald-900",
  danger: "border-rose-200 bg-rose-50 text-rose-900",
  neutral: "border-gray-200 bg-gray-50 text-gray-800",
};

function Callout({
  tone,
  title,
  body,
}: {
  tone: keyof typeof CALLOUT_TONES;
  title: string;
  body: string;
}) {
  return (
    <section className={`rounded-xl border p-4 ${CALLOUT_TONES[tone]}`}>
      <h3 className="text-sm font-semibold">{title}</h3>
      <p className="mt-1 whitespace-pre-wrap text-sm">{body}</p>
    </section>
  );
}
