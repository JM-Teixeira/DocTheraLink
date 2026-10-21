import { useQuery } from "@tanstack/react-query";
import dayjs from "dayjs";
import { Pencil, Phone, Stethoscope } from "lucide-react";
import { navigate } from "raviger";
import { ReactNode } from "react";
import { useTranslation } from "react-i18next";

import { Button } from "@/components/ui/button";

import Loading from "@/components/Common/Loading";
import Page from "@/components/Common/Page";
import { usePatient } from "@/components/Patient/api";
import { formatId, genderKey } from "@/components/Patient/patientFormUtils";
import ReferralCard from "@/components/Referral/ReferralCard";
import ReferralRoutes from "@/components/Referral/api";
import { canCreateReferral } from "@/components/Referral/referralUtils";

import useAuthUser from "@/hooks/useAuthUser";

import { ageFromDob, initials } from "@/Utils/person";
import query from "@/Utils/request/query";
import { canSeeReferrals, canWritePatients } from "@/Utils/roles";

export default function PatientDetail({ id }: { id: string }) {
  const { t } = useTranslation();
  const user = useAuthUser();
  const { data: patient, isLoading, isError } = usePatient(id);

  // Referral history for this patient (only roles allowed to see referrals)
  const referrals = useQuery({
    queryKey: ["referrals", "patient", id],
    queryFn: query(ReferralRoutes.list, {
      queryParams: { patient: id, limit: 50 },
    }),
    enabled: canSeeReferrals(user),
  });

  if (isLoading) return <Loading />;
  if (isError || !patient) {
    return (
      <Page title={t("patient")} backUrl="/patients">
        <p className="rounded-lg bg-rose-50 p-4 text-sm text-rose-700">
          {t("patient_not_found")}
        </p>
      </Page>
    );
  }

  const age = ageFromDob(patient.date_of_birth);
  const subtitle = [
    age !== null ? t("age_years", { age }) : null,
    patient.gender ? t(genderKey(patient.gender)) : null,
    patient.date_of_birth
      ? dayjs(patient.date_of_birth).format("D MMM YYYY")
      : null,
  ]
    .filter(Boolean)
    .join(" · ");

  return (
    <Page
      title={t("patient")}
      backUrl="/patients"
      crumbsReplacements={{ [id]: { name: patient.name ?? "" } }}
    >
      <div className="flex flex-col gap-6">
        {/* Header */}
        <section className="flex flex-wrap items-center justify-between gap-4 rounded-xl border border-gray-200 bg-white p-5 shadow-sm">
          <div className="flex items-center gap-4">
            <span className="flex size-14 items-center justify-center rounded-full bg-primary-100 text-lg font-semibold text-primary-700">
              {initials(patient.name ?? "")}
            </span>
            <div>
              <h2 className="text-xl font-semibold text-gray-900">
                {patient.name}
              </h2>
              <p className="text-sm text-gray-500">
                {subtitle || t("age_unknown")}
              </p>
              {patient.phone_number && (
                <a
                  href={`tel:${patient.phone_number}`}
                  className="mt-1 inline-flex items-center gap-1.5 text-sm text-primary-700 hover:underline"
                >
                  <Phone className="size-3.5" aria-hidden />
                  {patient.phone_number}
                </a>
              )}
            </div>
          </div>
          <div className="flex flex-wrap gap-2">
            {canWritePatients(user) && (
              <Button
                variant="outline"
                onClick={() => navigate(`/patients/${id}/edit`)}
              >
                <Pencil className="mr-1.5 size-4" aria-hidden />
                {t("edit")}
              </Button>
            )}
            {canCreateReferral(user) && (
              <Button
                variant="primary"
                onClick={() => navigate(`/referrals/new?patient=${id}`)}
              >
                <Stethoscope className="mr-1.5 size-4" aria-hidden />
                {t("refer_to_physiotherapy")}
              </Button>
            )}
          </div>
        </section>

        <div className="grid gap-6 lg:grid-cols-2">
          <Card title={t("contact")}>
            <Item label={t("phone")} value={patient.phone_number} />
            <Item
              label={t("emergency_phone")}
              value={patient.emergency_phone_number}
            />
            <Item label={t("address")} value={patient.address} wide />
            <Item label={t("nationality")} value={patient.nationality} />
            <Item label={t("nif")} value={formatId(patient.nif)} />
            <Item
              label={t("sns_number")}
              value={formatId(patient.sns_number)}
            />
          </Card>
          <Card title={t("clinical_background")}>
            <Item
              label={t("care_level")}
              value={`${patient.care_level ?? 1} · ${t(`care_level_help__${patient.care_level ?? 1}`)}`}
            />
            <Item label={t("blood_group")} value={patient.blood_group} />
            <Item label={t("allergies")} value={patient.allergies} />
            <Item
              label={t("medical_history")}
              value={patient.medical_history_details}
              wide
            />
            <Item
              label={t("ongoing_medication")}
              value={patient.ongoing_medication}
              wide
            />
          </Card>
        </div>

        {canSeeReferrals(user) && (
          <section>
            <h3 className="mb-3 text-base font-semibold text-gray-900">
              {t("referrals")}
            </h3>
            {referrals.data?.results.length ? (
              <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
                {referrals.data.results.map((r) => (
                  <ReferralCard key={r.id} referral={r} />
                ))}
              </div>
            ) : (
              <p className="rounded-xl border-2 border-dashed border-gray-200 bg-white p-6 text-center text-sm text-gray-500">
                {referrals.isLoading
                  ? t("loading")
                  : t("no_referrals_for_patient")}
              </p>
            )}
          </section>
        )}
      </div>
    </Page>
  );
}

function Card({ title, children }: { title: string; children: ReactNode }) {
  return (
    <section className="rounded-xl border border-gray-200 bg-white p-5 shadow-sm">
      <h3 className="mb-4 text-base font-semibold text-gray-900">{title}</h3>
      <dl className="grid gap-4 sm:grid-cols-2">{children}</dl>
    </section>
  );
}

function Item({
  label,
  value,
  wide,
}: {
  label: string;
  value?: string | null;
  wide?: boolean;
}) {
  return (
    <div className={wide ? "sm:col-span-2" : undefined}>
      <dt className="text-xs font-medium uppercase tracking-wide text-gray-500">
        {label}
      </dt>
      <dd className="mt-0.5 whitespace-pre-wrap text-sm text-gray-900">
        {value || "—"}
      </dd>
    </div>
  );
}
