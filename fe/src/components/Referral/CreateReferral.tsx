import { useMutation, useQuery } from "@tanstack/react-query";
import { Check, Search, TriangleAlert } from "lucide-react";
import { navigate } from "raviger";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";

import { cn } from "@/lib/utils";

import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";

import Page from "@/components/Common/Page";
import ReferralRoutes from "@/components/Referral/api";
import { useTherapists } from "@/components/Referral/hooks";
import { ageFromDob, initials } from "@/components/Referral/referralUtils";
import {
  CreateReferralRequest,
  Referral,
  ReferralPriority,
} from "@/components/Referral/types";

import useDebouncedValue from "@/hooks/useDebouncedValue";

import * as Notification from "@/Utils/Notifications";
import mutate from "@/Utils/request/mutate";
import query from "@/Utils/request/query";

const inputClass =
  "h-10 w-full rounded-lg border border-gray-300 px-3 text-sm shadow-sm focus:border-primary-500 focus:ring-primary-500";

const textareaClass =
  "w-full rounded-lg border border-gray-300 p-2.5 text-sm shadow-sm focus:border-primary-500 focus:ring-primary-500";

interface Form {
  diagnosis: string;
  icd_code: string;
  reason: string;
  goals: string;
  precautions: string;
  recommended_sessions: string;
  priority: ReferralPriority;
  therapist_id: string; // "" = open to every physiotherapist
}

const EMPTY: Form = {
  diagnosis: "",
  icd_code: "",
  reason: "",
  goals: "",
  precautions: "",
  recommended_sessions: "",
  priority: "routine",
  therapist_id: "",
};

/** Builds the API body, dropping empty optional fields. */
export function toRequest(patientId: string, f: Form): CreateReferralRequest {
  const opt = (s: string) => (s.trim() ? s.trim() : undefined);
  return {
    patient_id: patientId,
    diagnosis: f.diagnosis.trim(),
    reason: f.reason.trim(),
    priority: f.priority,
    icd_code: opt(f.icd_code),
    goals: opt(f.goals),
    precautions: opt(f.precautions),
    recommended_sessions: f.recommended_sessions
      ? Number(f.recommended_sessions)
      : undefined,
    therapist_id: f.therapist_id ? Number(f.therapist_id) : undefined,
  };
}

export default function CreateReferral({ patientId }: { patientId?: string }) {
  const { t } = useTranslation();
  const [search, setSearch] = useState("");
  const debounced = useDebouncedValue(search);
  const [patient, setPatient] = useState<{ id: string; name: string } | null>(
    null,
  );
  const [form, setForm] = useState<Form>(EMPTY);
  const set = <K extends keyof Form>(k: K, v: Form[K]) =>
    setForm((f) => ({ ...f, [k]: v }));

  // Pre-selected patient, e.g. /referrals/new?patient=<id> from a patient page
  const { data: preset } = useQuery({
    queryKey: ["patient", patientId],
    queryFn: query(ReferralRoutes.getPatient, {
      pathParams: { id: patientId ?? "" },
    }),
    enabled: !!patientId,
  });
  useEffect(() => {
    if (preset?.id) setPatient({ id: preset.id, name: preset.name ?? "" });
  }, [preset]);

  const { data: patients, isFetching: searching } = useQuery({
    queryKey: ["referral-patient-search", debounced],
    queryFn: query(ReferralRoutes.searchPatients, {
      queryParams: { search: debounced, limit: 8 },
    }),
    enabled: !patient && debounced.trim().length >= 2,
  });

  const { data: therapists } = useTherapists();

  const { mutate: create, isPending } = useMutation({
    mutationFn: mutate(ReferralRoutes.create),
    onSuccess: (r: Referral) => {
      Notification.Success({ msg: t("referral_created") });
      navigate(`/referrals/${r.id}`);
    },
    onError: () => Notification.Error({ msg: t("referral_create_failed") }),
  });

  const valid = patient && form.diagnosis.trim() && form.reason.trim();

  return (
    <Page title={t("new_referral")} backUrl="/referrals">
      <form
        className="mx-auto flex max-w-3xl flex-col gap-6"
        onSubmit={(e) => {
          e.preventDefault();
          if (valid) create(toRequest(patient.id, form));
        }}
      >
        {/* 1. Patient */}
        <Step n={1} title={t("patient")}>
          {patient ? (
            <div className="flex items-center justify-between rounded-lg bg-primary-50 p-3 ring-1 ring-primary-200">
              <div className="flex items-center gap-3">
                <span className="flex size-9 items-center justify-center rounded-full bg-primary-600 text-sm font-semibold text-white">
                  {initials(patient.name)}
                </span>
                <span className="font-medium text-gray-900">
                  {patient.name}
                </span>
              </div>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={() => setPatient(null)}
              >
                {t("change")}
              </Button>
            </div>
          ) : (
            <div>
              <div className="relative">
                <Search
                  className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-gray-400"
                  aria-hidden
                />
                <input
                  aria-label={t("search_patient")}
                  placeholder={t("search_patient_placeholder")}
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                  className={cn(inputClass, "pl-9")}
                />
              </div>
              {debounced.trim().length >= 2 && (
                <ul className="mt-2 divide-y divide-gray-100 overflow-hidden rounded-lg border border-gray-200">
                  {searching && (
                    <li className="p-3 text-sm text-gray-500">
                      {t("loading")}
                    </li>
                  )}
                  {!searching && patients?.results.length === 0 && (
                    <li className="p-3 text-sm text-gray-500">
                      {t("no_patients_found")}
                    </li>
                  )}
                  {patients?.results.map((p) => {
                    const age = ageFromDob(
                      p.date_of_birth as string | undefined,
                    );
                    return (
                      <li key={p.id}>
                        <button
                          type="button"
                          onClick={() =>
                            setPatient({ id: p.id!, name: p.name ?? "" })
                          }
                          className="flex w-full items-center justify-between p-3 text-left text-sm hover:bg-gray-50"
                        >
                          <span className="font-medium text-gray-900">
                            {p.name}
                          </span>
                          <span className="text-gray-500">
                            {age !== null ? t("age_years", { age }) : ""}
                            {p.phone_number ? ` · ${p.phone_number}` : ""}
                          </span>
                        </button>
                      </li>
                    );
                  })}
                </ul>
              )}
            </div>
          )}
        </Step>

        {/* 2. Clinical information */}
        <Step n={2} title={t("clinical_details")}>
          <div className="grid gap-4 sm:grid-cols-3">
            <div className="sm:col-span-2">
              <Label htmlFor="diagnosis">{t("diagnosis")} *</Label>
              <input
                id="diagnosis"
                required
                maxLength={255}
                value={form.diagnosis}
                onChange={(e) => set("diagnosis", e.target.value)}
                className={cn(inputClass, "mt-1")}
              />
            </div>
            <div>
              <Label htmlFor="icd">{t("icd_code")}</Label>
              <input
                id="icd"
                maxLength={32}
                value={form.icd_code}
                placeholder="M54.5"
                onChange={(e) => set("icd_code", e.target.value)}
                className={cn(inputClass, "mt-1")}
              />
            </div>
          </div>
          <div>
            <Label htmlFor="reason">{t("referral_reason")} *</Label>
            <textarea
              id="reason"
              required
              rows={4}
              value={form.reason}
              onChange={(e) => set("reason", e.target.value)}
              placeholder={t("referral_reason_placeholder")}
              className={cn(textareaClass, "mt-1")}
            />
          </div>
          <div className="grid gap-4 sm:grid-cols-2">
            <div>
              <Label htmlFor="goals">{t("treatment_goals")}</Label>
              <textarea
                id="goals"
                rows={3}
                value={form.goals}
                onChange={(e) => set("goals", e.target.value)}
                className={cn(textareaClass, "mt-1")}
              />
            </div>
            <div>
              <Label htmlFor="precautions">{t("precautions")}</Label>
              <textarea
                id="precautions"
                rows={3}
                value={form.precautions}
                onChange={(e) => set("precautions", e.target.value)}
                className={cn(textareaClass, "mt-1")}
              />
            </div>
          </div>
        </Step>

        {/* 3. Routing */}
        <Step n={3} title={t("referral_routing")}>
          <fieldset>
            <legend className="mb-2 text-sm font-medium text-gray-700">
              {t("priority")}
            </legend>
            <div className="grid grid-cols-2 gap-3">
              {(["routine", "urgent"] as const).map((p) => (
                <label
                  key={p}
                  className={cn(
                    "flex cursor-pointer items-center gap-3 rounded-lg border p-3 text-sm transition",
                    form.priority === p
                      ? p === "urgent"
                        ? "border-rose-400 bg-rose-50 ring-1 ring-rose-400"
                        : "border-primary-400 bg-primary-50 ring-1 ring-primary-400"
                      : "border-gray-200 hover:bg-gray-50",
                  )}
                >
                  <input
                    type="radio"
                    name="priority"
                    value={p}
                    checked={form.priority === p}
                    onChange={() => set("priority", p)}
                    className="sr-only"
                  />
                  {p === "urgent" ? (
                    <TriangleAlert
                      className="size-4 text-rose-600"
                      aria-hidden
                    />
                  ) : (
                    <Check className="size-4 text-primary-600" aria-hidden />
                  )}
                  <span className="font-medium">
                    {t(`referral_priority__${p}`)}
                  </span>
                </label>
              ))}
            </div>
          </fieldset>
          <div className="grid gap-4 sm:grid-cols-2">
            <div>
              <Label htmlFor="therapist">{t("physiotherapist")}</Label>
              <select
                id="therapist"
                value={form.therapist_id}
                onChange={(e) => set("therapist_id", e.target.value)}
                className="mt-1 w-full rounded-lg border border-gray-300 p-2 text-sm shadow-sm focus:border-primary-500 focus:ring-primary-500"
              >
                <option value="">{t("any_available_physiotherapist")}</option>
                {therapists?.results.map((u) => (
                  <option key={u.id} value={u.id}>
                    {`${u.first_name} ${u.last_name}`}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <Label htmlFor="sessions">{t("recommended_sessions")}</Label>
              <input
                id="sessions"
                type="number"
                min={1}
                max={200}
                value={form.recommended_sessions}
                onChange={(e) => set("recommended_sessions", e.target.value)}
                className={cn(inputClass, "mt-1")}
              />
            </div>
          </div>
        </Step>

        <div className="flex justify-end gap-2">
          <Button
            type="button"
            variant="outline"
            onClick={() => navigate("/referrals")}
          >
            {t("cancel")}
          </Button>
          <Button
            type="submit"
            variant="primary"
            disabled={!valid || isPending}
          >
            {t("send_referral")}
          </Button>
        </div>
      </form>
    </Page>
  );
}

function Step({
  n,
  title,
  children,
}: {
  n: number;
  title: string;
  children: React.ReactNode;
}) {
  return (
    <section className="rounded-xl border border-gray-200 bg-white p-5 shadow-sm">
      <h2 className="mb-4 flex items-center gap-3 text-base font-semibold text-gray-900">
        <span className="flex size-7 items-center justify-center rounded-full bg-primary-100 text-sm text-primary-700">
          {n}
        </span>
        {title}
      </h2>
      <div className="flex flex-col gap-4">{children}</div>
    </section>
  );
}
