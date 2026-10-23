import { useQuery } from "@tanstack/react-query";
import { Printer } from "lucide-react";
import { createPortal } from "react-dom";
import { useTranslation } from "react-i18next";

import { Button } from "@/components/ui/button";

import Loading from "@/components/Common/Loading";
import Page from "@/components/Common/Page";
import ReferralRoutes from "@/components/Referral/api";
import {
  ageFromDob,
  fullName,
  hasDoctorTitle,
  letterKind,
  letterReference,
} from "@/components/Referral/referralUtils";

import useDateFormat from "@/hooks/useDateFormat";

import { callApi } from "@/Utils/request/query";

/** A4 sheet with the margins printers expect. */
const PAGE_CSS =
  "@page { size: A4; margin: 16mm; } @media print { body > *:not(#section-to-print) { display: none !important; } }";
/** Letterhead fallback when the clinic cannot be loaded. */
const PRODUCT_NAME = "Klyniva";

/**
 * Printable A4 referral letter (or discharge summary once completed).
 * "Print" uses the browser dialog, which also saves it as PDF; the global
 * print copy is portalled under <body> so the app layout cannot offset it.
 */
export default function ReferralLetter({ id }: { id: string }) {
  const { t } = useTranslation();
  const { fmt: fmtDate } = useDateFormat();

  const { data: referral, isLoading } = useQuery({
    queryKey: ["referral", id],
    queryFn: ({ signal }) =>
      callApi(ReferralRoutes.get, { pathParams: { id }, signal }),
  });
  const clinicId = referral?.facility_id ?? "";
  const { data: clinic } = useQuery({
    queryKey: ["clinic", clinicId],
    queryFn: ({ signal }) =>
      callApi(ReferralRoutes.getClinic, {
        pathParams: { id: clinicId },
        signal,
      }),
    enabled: !!clinicId,
  });

  if (isLoading) return <Loading />;
  if (!referral) {
    return (
      <Page title={t("referral_letter")} backUrl="/referrals">
        <p className="rounded-lg bg-rose-50 p-4 text-sm text-rose-700">
          {t("referral_not_found")}
        </p>
      </Page>
    );
  }

  const kind = letterKind(referral);
  const doctor = fullName(referral.doctor);
  const doctorName = hasDoctorTitle(doctor)
    ? doctor
    : t("doctor_name", { name: doctor });
  const therapist = referral.therapist ? fullName(referral.therapist) : null;
  const dob = referral.patient?.date_of_birth;
  const age = ageFromDob(dob);
  const issued =
    kind === "discharge" && referral.completed_at
      ? referral.completed_at
      : referral.created_date;
  const fmt = (d?: string | null) => fmtDate(d) || "-";

  const sheet = (
    <article className="mx-auto w-full max-w-[210mm] bg-white p-10 text-[13px] leading-relaxed text-gray-900 shadow-sm ring-1 ring-gray-200 print:p-0 print:shadow-none print:ring-0">
      {/* Letterhead */}
      <header className="flex items-start justify-between border-b-2 border-primary-600 pb-4">
        <div>
          <p className="text-lg font-semibold text-primary-700">
            {clinic?.name ?? PRODUCT_NAME}
          </p>
          {clinic?.address && (
            <p className="whitespace-pre-line text-gray-600">
              {clinic.address}
              {clinic.pincode ? `, ${clinic.pincode}` : ""}
            </p>
          )}
        </div>
        <div className="text-right text-gray-600">
          <p>{fmt(issued)}</p>
          <p className="font-mono text-xs">{letterReference(referral.id)}</p>
        </div>
      </header>

      <h1 className="mt-6 text-xl font-semibold">
        {t(`letter_title__${kind}`)}
        {referral.priority === "urgent" && (
          <span className="ml-3 align-middle text-xs font-semibold uppercase tracking-wide text-rose-700">
            {t("referral_priority__urgent")}
          </span>
        )}
      </h1>

      <LetterSection title={t("patient")}>
        <dl className="grid grid-cols-3 gap-4">
          <Field label={t("name")}>{referral.patient?.name}</Field>
          <Field label={t("date_of_birth")}>
            {fmt(dob)}
            {age !== null && ` (${t("age_years", { age })})`}
          </Field>
          <Field label={t("phone_number")}>
            {referral.patient?.phone_number || "-"}
          </Field>
        </dl>
      </LetterSection>

      <LetterSection title={t("clinical_details")}>
        <dl className="grid grid-cols-3 gap-4">
          <Field label={t("diagnosis")} wide>
            {referral.diagnosis}
            {referral.icd_code && ` (ICD-10 ${referral.icd_code})`}
          </Field>
          <Field label={t("recommended_sessions")}>
            {referral.recommended_sessions ?? "-"}
          </Field>
          <Field label={t("referral_reason")} full>
            {referral.reason}
          </Field>
          {referral.goals && (
            <Field label={t("treatment_goals")} full>
              {referral.goals}
            </Field>
          )}
          {referral.precautions && (
            <Field label={t("precautions")} full>
              {referral.precautions}
            </Field>
          )}
        </dl>
      </LetterSection>

      {kind === "discharge" && (
        <LetterSection title={t("treatment_outcome")}>
          <dl className="grid grid-cols-3 gap-4">
            <Field label={t("started_on")}>{fmt(referral.started_at)}</Field>
            <Field label={t("completed_on")}>
              {fmt(referral.completed_at)}
            </Field>
            <Field label={t("physiotherapist")}>{therapist ?? "-"}</Field>
            <Field label={t("outcome_summary")} full>
              {referral.outcome_summary || "-"}
            </Field>
          </dl>
        </LetterSection>
      )}

      {/* Signatures */}
      <div className="mt-14 grid grid-cols-2 gap-12">
        <Signature label={t("referring_doctor")} name={doctorName} />
        {kind === "discharge" ? (
          <Signature label={t("physiotherapist")} name={therapist ?? ""} />
        ) : (
          therapist && (
            <Signature label={t("addressed_to")} name={therapist} unsigned />
          )
        )}
      </div>

      <footer className="mt-12 border-t border-gray-200 pt-3 text-[11px] text-gray-500">
        {t("letter_confidential")}
      </footer>
    </article>
  );

  return (
    <Page
      title={t(`letter_title__${kind}`)}
      backUrl={`/referrals/${id}`}
      crumbsReplacements={{ [id]: { name: referral.patient?.name ?? "" } }}
      options={
        <Button variant="primary" onClick={() => window.print()}>
          <Printer className="mr-2 size-4" aria-hidden />
          {t("print_or_save_pdf")}
        </Button>
      }
    >
      <style>{PAGE_CSS}</style>
      <div className="mt-4 print:hidden">{sheet}</div>
      {/* Print copy straight under <body>, outside the app layout */}
      {createPortal(
        <div id="section-to-print" className="hidden print:block print:w-full">
          {sheet}
        </div>,
        document.body,
      )}
    </Page>
  );
}

function LetterSection({
  title,
  children,
}: {
  title: string;
  children: React.ReactNode;
}) {
  return (
    <section className="mt-6 break-inside-avoid">
      <h2 className="mb-2 text-xs font-semibold uppercase tracking-wide text-primary-700">
        {title}
      </h2>
      {children}
    </section>
  );
}

function Field({
  label,
  wide,
  full,
  children,
}: {
  label: string;
  wide?: boolean;
  full?: boolean;
  children: React.ReactNode;
}) {
  return (
    <div className={full ? "col-span-3" : wide ? "col-span-2" : undefined}>
      <dt className="text-[11px] text-gray-500">{label}</dt>
      <dd className="whitespace-pre-wrap">{children}</dd>
    </div>
  );
}

function Signature({
  label,
  name,
  unsigned,
}: {
  label: string;
  name: string;
  unsigned?: boolean;
}) {
  return (
    <div>
      {!unsigned && <div className="mb-2 h-10 border-b border-gray-400" />}
      <p className="font-medium">{name}</p>
      <p className="text-[11px] text-gray-500">{label}</p>
    </div>
  );
}
