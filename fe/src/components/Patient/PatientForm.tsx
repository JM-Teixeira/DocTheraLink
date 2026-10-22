import { History } from "lucide-react";
import { navigate } from "raviger";
import { ReactNode, useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";

import { cn } from "@/lib/utils";

import { Button } from "@/components/ui/button";

import DateInput from "@/components/Common/DateInput";
import Loading from "@/components/Common/Loading";
import Page from "@/components/Common/Page";
import DuplicatePatientsDialog from "@/components/Patient/DuplicatePatientsDialog";
import {
  useCreatePatient,
  usePatchPatient,
  usePatient,
} from "@/components/Patient/api";
import { PatientModel } from "@/components/Patient/models";
import {
  BLOOD_GROUP_OPTIONS,
  EMPTY_PATIENT_FORM,
  GENDERS,
  PatientFormErrors,
  PatientFormState,
  PossibleDuplicate,
  RELATIONS,
  draftKey,
  duplicatesFromError,
  formToPayload,
  formatPostalCode,
  patientToForm,
  readDraft,
  sameForm,
  serverErrorsToForm,
  validatePatientForm,
  writeDraft,
} from "@/components/Patient/patientFormUtils";
import {
  CARE_LEVELS,
  CARE_LEVEL_STYLES,
} from "@/components/TimeBoard/boardUtils";

import useAuthUser from "@/hooks/useAuthUser";
import useDateFormat from "@/hooks/useDateFormat";

import * as Notification from "@/Utils/Notifications";
import { HTTPError } from "@/Utils/request/types";

const inputClass =
  "mt-1 h-10 w-full rounded-lg border border-gray-300 px-3 text-sm shadow-sm focus:border-primary-500 focus:ring-primary-500";
// Suggestions only: any insurer can be typed
const COMMON_INSURERS = [
  "ADSE",
  "SAD-GNR",
  "SAD-PSP",
  "ADM",
  "Médis",
  "Multicare",
  "AdvanceCare",
  "Allianz",
  "Fidelidade",
];

const textareaClass =
  "mt-1 w-full rounded-lg border border-gray-300 p-2.5 text-sm shadow-sm focus:border-primary-500 focus:ring-primary-500";

/** Registers a new patient, or edits one when `patientId` is given. */
export default function PatientForm({ patientId }: { patientId?: string }) {
  const { t } = useTranslation();
  const editing = !!patientId;
  const { fmtTime } = useDateFormat();
  const user = useAuthUser();
  const [form, setForm] = useState<PatientFormState>(EMPTY_PATIENT_FORM);
  const [errors, setErrors] = useState<PatientFormErrors>({});
  const [dupes, setDupes] = useState<{
    blocking: boolean;
    matches: PossibleDuplicate[];
  } | null>(null);

  // Draft auto-save: `base` is what the server has (empty for a new
  // patient); the form is saved locally while it differs from it.
  const key = draftKey(user.id ?? user.username, patientId);
  const [base, setBase] = useState<PatientFormState>(EMPTY_PATIENT_FORM);
  const [ready, setReady] = useState(false);
  const [restoredAt, setRestoredAt] = useState<string | null>(null);
  const closing = useRef(false); // saved or cancelled: stop auto-saving

  const existing = usePatient(patientId ?? "", { enabled: editing });
  useEffect(() => {
    if (ready || (editing && !existing.data)) return;
    const fromServer = existing.data
      ? patientToForm(existing.data)
      : EMPTY_PATIENT_FORM;
    const draft = readDraft(localStorage, key);
    setBase(fromServer);
    if (draft && !sameForm(draft.form, fromServer)) {
      setForm(draft.form);
      setRestoredAt(draft.saved_at);
    } else setForm(fromServer);
    setReady(true);
  }, [ready, editing, existing.data, key]);

  useEffect(() => {
    if (!ready) return;
    const timer = setTimeout(
      () => !closing.current && writeDraft(localStorage, key, form, base),
      500,
    );
    return () => clearTimeout(timer);
  }, [ready, key, form, base]);

  const dropDraft = () => writeDraft(localStorage, key, base, base);

  const create = useCreatePatient();
  const patch = usePatchPatient();
  const saving = create.isPending || patch.isPending;

  const set = (k: keyof PatientFormState) => (v: string) => {
    setForm((f) => ({ ...f, [k]: v }));
    setErrors((e) => ({ ...e, [k]: undefined }));
  };

  const onSaved = (p: PatientModel) => {
    closing.current = true;
    dropDraft();
    Notification.Success({
      msg: editing ? t("patient_updated") : t("patient_created"),
    });
    navigate(`/patients/${p.id}`);
  };
  const onError = (err: Error) => {
    const found =
      err instanceof HTTPError
        ? duplicatesFromError(err.status, err.cause)
        : null;
    if (found) {
      setDupes(found);
      return;
    }
    const fieldErrors =
      err instanceof HTTPError ? serverErrorsToForm(err.cause) : {};
    setErrors(fieldErrors);
    if (!Object.keys(fieldErrors).length)
      Notification.Error({ msg: t("patient_save_failed") });
  };

  const save = (allowDuplicate = false) => {
    const v = validatePatientForm(form);
    setErrors(v);
    if (Object.keys(v).length) return;
    const body = formToPayload(form) as Partial<PatientModel>;
    if (editing)
      patch.mutate(
        { id: patientId, data: body },
        { onSuccess: onSaved, onError },
      );
    else
      create.mutate(
        allowDuplicate ? { ...body, allow_duplicate: true } : body,
        { onSuccess: onSaved, onError },
      );
  };
  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    save();
  };

  if (editing && (existing.isLoading || !ready)) return <Loading />;

  const fieldProps = (k: keyof PatientFormState) => ({
    id: k,
    value: form[k],
    "aria-invalid": !!errors[k],
    "aria-describedby": errors[k] ? `${k}-error` : undefined,
    onChange: (
      e: React.ChangeEvent<
        HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement
      >,
    ) => set(k)(e.target.value),
  });

  return (
    <Page
      title={editing ? t("edit_patient") : t("new_patient")}
      backUrl={editing ? `/patients/${patientId}` : "/patients"}
      crumbsReplacements={
        editing ? { [patientId]: { name: existing.data?.name ?? "" } } : {}
      }
    >
      <form
        onSubmit={submit}
        noValidate
        className="mx-auto flex max-w-3xl flex-col gap-6"
      >
        {restoredAt && (
          <div
            role="status"
            className="flex flex-wrap items-center justify-between gap-2 rounded-lg bg-amber-50 p-3 text-sm text-amber-900 ring-1 ring-inset ring-amber-200"
          >
            <span className="flex items-center gap-2">
              <History className="size-4" aria-hidden />
              {t("draft_restored", { time: fmtTime(restoredAt) })}
            </span>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => {
                setForm(base);
                setErrors({});
                setRestoredAt(null);
                dropDraft();
              }}
            >
              {t("discard_draft")}
            </Button>
          </div>
        )}
        <Section title={t("identity")}>
          <Field
            label={`${t("full_name")} *`}
            id="name"
            error={errors.name}
            wide
          >
            <input
              {...fieldProps("name")}
              autoComplete="off"
              className={inputClass}
            />
          </Field>
          <Field
            label={t("date_of_birth")}
            id="date_of_birth"
            error={errors.date_of_birth}
          >
            <DateInput
              id="date_of_birth"
              value={form.date_of_birth}
              max={new Date().toISOString().slice(0, 10)}
              onChange={set("date_of_birth")}
            />
          </Field>
          <Field label={t("nationality")} id="nationality">
            <input {...fieldProps("nationality")} className={inputClass} />
          </Field>
          <Field label={t("nif")} id="nif" error={errors.nif}>
            <input
              {...fieldProps("nif")}
              inputMode="numeric"
              maxLength={11}
              placeholder={t("nif_placeholder")}
              className={inputClass}
            />
          </Field>
          <Field
            label={t("sns_number")}
            id="sns_number"
            error={errors.sns_number}
          >
            <input
              {...fieldProps("sns_number")}
              inputMode="numeric"
              maxLength={11}
              className={inputClass}
            />
          </Field>
          <fieldset className="sm:col-span-2">
            <legend className="text-sm font-medium text-gray-700">
              {t("gender")}
            </legend>
            <div className="mt-1 flex flex-wrap gap-2">
              {GENDERS.map((g) => (
                <label
                  key={g.id}
                  className={cn(
                    "cursor-pointer rounded-full px-4 py-1.5 text-sm ring-1 ring-inset transition",
                    form.gender === String(g.id)
                      ? "bg-primary-600 text-white ring-primary-600"
                      : "bg-white text-gray-700 ring-gray-300 hover:bg-gray-50",
                  )}
                >
                  <input
                    type="radio"
                    name="gender"
                    value={g.id}
                    checked={form.gender === String(g.id)}
                    onChange={() => set("gender")(String(g.id))}
                    className="sr-only"
                  />
                  {t(g.key)}
                </label>
              ))}
            </div>
          </fieldset>
          <fieldset className="sm:col-span-2">
            <legend className="text-sm font-medium text-gray-700">
              {t("care_level")}
            </legend>
            <p className="text-xs text-gray-500">{t("care_level_form_help")}</p>
            <div className="mt-2 grid gap-2 sm:grid-cols-5">
              {CARE_LEVELS.map((lvl) => (
                <label
                  key={lvl}
                  className={cn(
                    "cursor-pointer rounded-lg p-2 text-xs ring-1 ring-inset transition",
                    form.care_level === String(lvl)
                      ? "bg-primary-50 ring-2 ring-primary-500"
                      : "bg-white ring-gray-200 hover:bg-gray-50",
                  )}
                >
                  <input
                    type="radio"
                    name="care_level"
                    value={lvl}
                    checked={form.care_level === String(lvl)}
                    onChange={() => set("care_level")(String(lvl))}
                    className="sr-only"
                  />
                  <span
                    className={cn(
                      "mb-1 inline-block rounded px-1.5 font-semibold",
                      CARE_LEVEL_STYLES[lvl],
                    )}
                  >
                    {lvl}
                  </span>
                  <span className="block text-gray-600">
                    {t(`care_level_help__${lvl}`)}
                  </span>
                </label>
              ))}
            </div>
          </fieldset>
        </Section>

        <Section title={t("contact")}>
          <Field
            label={t("phone")}
            id="phone_number"
            error={errors.phone_number}
          >
            <input
              {...fieldProps("phone_number")}
              type="tel"
              placeholder={t("phone_placeholder")}
              className={inputClass}
            />
          </Field>
          <Field label={t("email")} id="email" error={errors.email}>
            <input
              {...fieldProps("email")}
              type="email"
              autoComplete="off"
              className={inputClass}
            />
          </Field>
          <Field label={t("street_address")} id="address" wide>
            <input
              {...fieldProps("address")}
              placeholder={t("street_address_placeholder")}
              className={inputClass}
            />
          </Field>
          <Field
            label={t("postal_code")}
            id="postal_code"
            error={errors.postal_code}
          >
            <input
              {...fieldProps("postal_code")}
              inputMode="numeric"
              maxLength={9}
              placeholder="0000-000"
              onBlur={() =>
                set("postal_code")(formatPostalCode(form.postal_code))
              }
              className={inputClass}
            />
          </Field>
          <Field label={t("city")} id="city">
            <input {...fieldProps("city")} className={inputClass} />
          </Field>
        </Section>

        <Section title={t("emergency_contact")}>
          <Field
            label={t("emergency_contact_name")}
            id="emergency_contact_name"
          >
            <input
              {...fieldProps("emergency_contact_name")}
              autoComplete="off"
              className={inputClass}
            />
          </Field>
          <Field
            label={t("emergency_contact_relation")}
            id="emergency_contact_relation"
          >
            <select
              {...fieldProps("emergency_contact_relation")}
              className={inputClass}
            >
              <option value="">-</option>
              {RELATIONS.map((r) => (
                <option key={r} value={r}>
                  {t(`relation__${r}`)}
                </option>
              ))}
            </select>
          </Field>
          <Field
            label={t("emergency_phone")}
            id="emergency_phone_number"
            error={errors.emergency_phone_number}
          >
            <input
              {...fieldProps("emergency_phone_number")}
              type="tel"
              className={inputClass}
            />
          </Field>
        </Section>

        <Section title={t("health_insurance")}>
          <Field label={t("insurer")} id="insurer">
            <input
              {...fieldProps("insurer")}
              list="insurers"
              placeholder={t("insurer_placeholder")}
              className={inputClass}
            />
            <datalist id="insurers">
              {COMMON_INSURERS.map((i) => (
                <option key={i} value={i} />
              ))}
            </datalist>
          </Field>
          <Field label={t("insurance_number")} id="insurance_number">
            <input {...fieldProps("insurance_number")} className={inputClass} />
          </Field>
        </Section>

        <Section title={t("clinical_background")}>
          <Field label={t("blood_group")} id="blood_group">
            <select {...fieldProps("blood_group")} className={inputClass}>
              <option value="">{t("unknown")}</option>
              {BLOOD_GROUP_OPTIONS.map((b) => (
                <option key={b} value={b}>
                  {b}
                </option>
              ))}
            </select>
          </Field>
          <Field label={t("allergies")} id="allergies">
            <input {...fieldProps("allergies")} className={inputClass} />
          </Field>
          <Field label={t("medical_history")} id="medical_history_details" wide>
            <textarea
              {...fieldProps("medical_history_details")}
              rows={3}
              className={textareaClass}
            />
          </Field>
          <Field label={t("ongoing_medication")} id="ongoing_medication" wide>
            <textarea
              {...fieldProps("ongoing_medication")}
              rows={2}
              className={textareaClass}
            />
          </Field>
        </Section>

        <div className="flex flex-wrap items-center justify-end gap-2">
          <p className="mr-auto text-xs text-gray-500">{t("draft_hint")}</p>
          <Button
            type="button"
            variant="outline"
            onClick={() => {
              closing.current = true; // cancelling throws the draft away
              dropDraft();
              navigate(editing ? `/patients/${patientId}` : "/patients");
            }}
          >
            {t("cancel")}
          </Button>
          <Button type="submit" variant="primary" disabled={saving}>
            {editing ? t("save_changes") : t("register_patient")}
          </Button>
        </div>
      </form>
      <DuplicatePatientsDialog
        open={!!dupes}
        blocking={dupes?.blocking ?? false}
        matches={dupes?.matches ?? []}
        pending={saving}
        onOpenPatient={(id) => navigate(`/patients/${id}`)}
        onConfirmDifferent={() => {
          setDupes(null);
          save(true);
        }}
        onClose={() => setDupes(null)}
      />
    </Page>
  );
}

function Section({ title, children }: { title: string; children: ReactNode }) {
  return (
    <section className="rounded-xl border border-gray-200 bg-white p-5 shadow-sm">
      <h2 className="mb-4 text-base font-semibold text-gray-900">{title}</h2>
      <div className="grid gap-4 sm:grid-cols-2">{children}</div>
    </section>
  );
}

function Field({
  label,
  id,
  error,
  wide,
  children,
}: {
  label: string;
  id: string;
  error?: string;
  wide?: boolean;
  children: ReactNode;
}) {
  const { t } = useTranslation();
  return (
    <div className={wide ? "sm:col-span-2" : undefined}>
      <label htmlFor={id} className="text-sm font-medium text-gray-700">
        {label}
      </label>
      {children}
      {error && (
        <p id={`${id}-error`} className="mt-1 text-xs text-rose-600">
          {/* client errors are translation keys; server errors are plain text */}
          {error.startsWith("error__") ? t(error) : error}
        </p>
      )}
    </div>
  );
}
