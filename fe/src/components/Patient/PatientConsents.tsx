import { ShieldCheck, ShieldX } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";

import { cn } from "@/lib/utils";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

import { usePatientConsents, useRecordConsent } from "@/components/Patient/api";
import {
  CONSENT_KINDS,
  CONSENT_METHODS,
  Consent,
  ConsentKind,
  ConsentMethod,
} from "@/components/Patient/api/routes";

import useDateFormat from "@/hooks/useDateFormat";

const fieldClass =
  "mt-1 h-10 w-full rounded-lg border border-gray-300 px-3 text-sm shadow-sm focus:border-primary-500 focus:ring-primary-500";

/**
 * GDPR consents of a patient: the current state of each kind, how it was
 * given and by whom, with the full history. Records are never edited: a
 * change is a new record.
 */
export default function PatientConsents({
  patientId,
  canRecord,
}: {
  patientId: string;
  canRecord: boolean;
}) {
  const { t } = useTranslation();
  const { fmt, fmtTime } = useDateFormat();
  const { data, isLoading } = usePatientConsents(patientId);
  const [editing, setEditing] = useState<ConsentKind | null>(null);
  const [showHistory, setShowHistory] = useState(false);

  if (isLoading || !data) return null;
  const who = (c: Consent) =>
    c.recorded_by
      ? `${c.recorded_by.first_name} ${c.recorded_by.last_name}`.trim()
      : "";

  return (
    <section
      aria-labelledby="consents-title"
      className="rounded-xl border border-gray-200 bg-white p-5 shadow-sm"
    >
      <h3
        id="consents-title"
        className="mb-1 text-base font-semibold text-gray-900"
      >
        {t("consents")}
      </h3>
      <p className="mb-4 text-xs text-gray-500">{t("consents_help")}</p>
      {!data.current.data_processing?.granted && (
        <p
          role="alert"
          className="mb-4 rounded-lg bg-amber-50 p-3 text-sm text-amber-900 ring-1 ring-inset ring-amber-200"
        >
          {t("consent_data_processing_missing")}
        </p>
      )}
      <ul className="divide-y divide-gray-100">
        {CONSENT_KINDS.map((kind) => {
          const c = data.current[kind];
          return (
            <li
              key={kind}
              className="flex flex-wrap items-center justify-between gap-2 py-2.5"
            >
              <div>
                <p className="text-sm font-medium text-gray-900">
                  {t(`consent__${kind}`)}
                </p>
                <p className="text-xs text-gray-500">
                  {c
                    ? [
                        fmt(c.created_date),
                        t(`consent_method__${c.method}`),
                        who(c),
                      ]
                        .filter(Boolean)
                        .join(" · ")
                    : t("consent_not_recorded")}
                </p>
              </div>
              <div className="flex items-center gap-2">
                <span
                  className={cn(
                    "inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium",
                    !c
                      ? "bg-gray-100 text-gray-600"
                      : c.granted
                        ? "bg-emerald-100 text-emerald-800"
                        : "bg-rose-100 text-rose-800",
                  )}
                >
                  {c?.granted ? (
                    <ShieldCheck className="size-3.5" aria-hidden />
                  ) : (
                    c && <ShieldX className="size-3.5" aria-hidden />
                  )}
                  {!c
                    ? t("consent_status__none")
                    : c.granted
                      ? t("consent_status__granted")
                      : t("consent_status__withdrawn")}
                </span>
                {canRecord && (
                  <Button
                    variant="outline"
                    size="sm"
                    aria-label={`${t("record_consent")}: ${t(`consent__${kind}`)}`}
                    onClick={() => setEditing(kind)}
                  >
                    {t("record_consent")}
                  </Button>
                )}
              </div>
            </li>
          );
        })}
      </ul>

      {data.history.length > 0 && (
        <div className="mt-3">
          <Button
            variant="link"
            size="sm"
            className="h-auto p-0"
            aria-expanded={showHistory}
            onClick={() => setShowHistory((v) => !v)}
          >
            {showHistory
              ? t("hide_consent_history")
              : t("show_consent_history", { count: data.history.length })}
          </Button>
          {showHistory && (
            <ol className="mt-2 space-y-1.5 text-xs text-gray-600">
              {data.history.map((c) => (
                <li key={c.id}>
                  <span className="font-medium text-gray-800">
                    {fmtTime(c.created_date)}
                  </span>{" "}
                  · {t(`consent__${c.kind}`)}:{" "}
                  {c.granted
                    ? t("consent_status__granted")
                    : t("consent_status__withdrawn")}{" "}
                  ({t(`consent_method__${c.method}`)}
                  {who(c) && `, ${who(c)}`}){c.note && ` · ${c.note}`}
                </li>
              ))}
            </ol>
          )}
        </div>
      )}

      {editing && (
        <RecordConsentDialog
          patientId={patientId}
          kind={editing}
          current={data.current[editing]}
          onClose={() => setEditing(null)}
        />
      )}
    </section>
  );
}

function RecordConsentDialog({
  patientId,
  kind,
  current,
  onClose,
}: {
  patientId: string;
  kind: ConsentKind;
  current?: Consent;
  onClose: () => void;
}) {
  const { t } = useTranslation();
  const record = useRecordConsent(patientId);
  // Default to the opposite of the current state: that is the usual change
  const [granted, setGranted] = useState(!current?.granted);
  const [method, setMethod] = useState<ConsentMethod>("signed_paper");
  const [note, setNote] = useState("");

  return (
    <Dialog open onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t(`consent__${kind}`)}</DialogTitle>
          <DialogDescription>{t("record_consent_help")}</DialogDescription>
        </DialogHeader>
        <form
          className="flex flex-col gap-4"
          onSubmit={(e) => {
            e.preventDefault();
            record.mutate(
              { kind, granted, method, note },
              { onSuccess: onClose },
            );
          }}
        >
          <fieldset>
            <legend className="text-sm font-medium text-gray-700">
              {t("consent_decision")}
            </legend>
            <div className="mt-1 flex gap-2">
              {[true, false].map((g) => (
                <label
                  key={String(g)}
                  className={cn(
                    "cursor-pointer rounded-full px-4 py-1.5 text-sm ring-1 ring-inset",
                    granted === g
                      ? "bg-primary-600 text-white ring-primary-600"
                      : "bg-white text-gray-700 ring-gray-300",
                  )}
                >
                  <input
                    type="radio"
                    name="granted"
                    className="sr-only"
                    checked={granted === g}
                    onChange={() => setGranted(g)}
                  />
                  {g ? t("consent_give") : t("consent_withdraw")}
                </label>
              ))}
            </div>
          </fieldset>
          <label className="text-sm font-medium text-gray-700">
            {t("consent_method")}
            <select
              value={method}
              onChange={(e) => setMethod(e.target.value as ConsentMethod)}
              className={fieldClass}
            >
              {CONSENT_METHODS.map((m) => (
                <option key={m} value={m}>
                  {t(`consent_method__${m}`)}
                </option>
              ))}
            </select>
          </label>
          <label className="text-sm font-medium text-gray-700">
            {t("note")}
            <input
              value={note}
              maxLength={300}
              onChange={(e) => setNote(e.target.value)}
              className={fieldClass}
            />
          </label>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={onClose}>
              {t("cancel")}
            </Button>
            <Button type="submit" variant="primary" disabled={record.isPending}>
              {t("save")}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
