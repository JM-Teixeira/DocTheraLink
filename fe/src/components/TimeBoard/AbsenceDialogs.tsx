import { useMutation, useQuery } from "@tanstack/react-query";
import dayjs from "dayjs";
import { ArrowRight, Plane, Plus, Trash2 } from "lucide-react";
import { useEffect, useState } from "react";
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

import DateInput from "@/components/Common/DateInput";
import AbsencePlanner from "@/components/TimeBoard/AbsencePlanner";
import { apiError } from "@/components/TimeBoard/BoardDialogs";
import BoardRoutes from "@/components/TimeBoard/api";
import { addDays, today } from "@/components/TimeBoard/boardUtils";
import { AbsenceReason, AbsenceResult } from "@/components/TimeBoard/types";

import useDateFormat from "@/hooks/useDateFormat";

import * as Notification from "@/Utils/Notifications";
import query, { callApi } from "@/Utils/request/query";

const fieldClass =
  "mt-1 h-10 w-full rounded-lg border border-gray-300 px-3 text-sm shadow-sm focus:border-primary-500 focus:ring-primary-500";
const REASONS: AbsenceReason[] = ["vacation", "sick", "training", "other"];

type Physio = { id: number; first_name: string; last_name: string };
const fullName = (p?: Physio | null) =>
  p ? `${p.first_name} ${p.last_name}` : "";

/** Current and upcoming absences, with their handover, and "new absence". */
export function AbsencesDialog({
  open,
  physios,
  onClose,
  onChanged,
}: {
  open: boolean;
  physios: Physio[];
  onClose: () => void;
  onChanged: () => void;
}) {
  const { t } = useTranslation();
  const { fmt } = useDateFormat();
  const [creating, setCreating] = useState(false);
  const from = today();
  const { data: absences, refetch } = useQuery({
    queryKey: ["absences", from],
    queryFn: query(BoardRoutes.absences, {
      queryParams: { from, to: addDays(from, 180) },
    }),
    enabled: open,
  });
  const remove = useMutation({
    mutationFn: (id: string) =>
      callApi(BoardRoutes.deleteAbsence, { pathParams: { id }, silent: true }),
    onSuccess: () => {
      refetch();
      onChanged();
    },
    onError: (e: Error) => apiError(e, t("board_save_failed")),
  });

  return (
    <>
      <Dialog open={open && !creating} onOpenChange={(o) => !o && onClose()}>
        <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>{t("absences")}</DialogTitle>
            <DialogDescription>{t("absences_help")}</DialogDescription>
          </DialogHeader>
          {!absences?.length ? (
            <p className="text-sm text-gray-500">{t("no_absences")}</p>
          ) : (
            <ul className="flex flex-col gap-3">
              {absences.map((a) => (
                <li
                  key={a.id}
                  className="rounded-lg border border-gray-200 p-3 text-sm"
                >
                  <div className="flex flex-wrap items-center gap-2">
                    <Plane className="size-4 text-amber-600" aria-hidden />
                    <span className="font-medium text-gray-900">
                      {fullName(a.therapist)}
                    </span>
                    <span className="text-gray-600">
                      {t(`absence_reason__${a.reason}`)} ·{" "}
                      {dayjs(a.from).format("D MMM")} - {fmt(a.to)}
                    </span>
                    {a.cover && (
                      <span className="flex items-center gap-1 text-gray-600">
                        <ArrowRight className="size-3.5" aria-hidden />
                        {fullName(a.cover)}
                      </span>
                    )}
                    <Button
                      size="icon"
                      variant="ghost"
                      className="ml-auto"
                      aria-label={t("delete")}
                      onClick={() =>
                        window.confirm(t("confirm_delete_absence")) &&
                        remove.mutate(a.id)
                      }
                    >
                      <Trash2 className="size-4 text-rose-600" aria-hidden />
                    </Button>
                  </div>
                  {a.handover && (
                    <p className="mt-2 whitespace-pre-wrap rounded bg-gray-50 p-2 text-gray-700">
                      {a.handover}
                    </p>
                  )}
                  {!!a.notes.length && (
                    <ul className="mt-2 flex flex-col gap-1">
                      {a.notes.map((n) => (
                        <li key={n.id} className="text-gray-700">
                          <span className="font-medium">
                            {n.patient?.name}:
                          </span>{" "}
                          {n.note}
                        </li>
                      ))}
                    </ul>
                  )}
                </li>
              ))}
            </ul>
          )}
          <DialogFooter>
            <Button variant="primary" onClick={() => setCreating(true)}>
              <Plus className="mr-1 size-4" aria-hidden />
              {t("new_absence")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
      <NewAbsenceDialog
        open={open && creating}
        physios={physios}
        onClose={() => setCreating(false)}
        onDone={() => {
          setCreating(false);
          refetch();
          onChanged();
        }}
      />
    </>
  );
}

/**
 * New absence in two steps: who, when, default cover → a plan of where each
 * session goes (per patient: the cover, another physio working that hour,
 * or stays) and a handover note per patient → confirm.
 */
function NewAbsenceDialog({
  open,
  physios,
  onClose,
  onDone,
}: {
  open: boolean;
  physios: Physio[];
  onClose: () => void;
  onDone: () => void;
}) {
  const { t } = useTranslation();
  const [therapist, setTherapist] = useState(0);
  const [cover, setCover] = useState(0);
  const [from, setFrom] = useState(today());
  const [to, setTo] = useState(addDays(today(), 4));
  const [reason, setReason] = useState<AbsenceReason>("vacation");
  const [handover, setHandover] = useState("");
  const [notes, setNotes] = useState<Record<string, string>>({});
  const [preview, setPreview] = useState<AbsenceResult | null>(null);
  // Per session: who takes it (0 = stays); sessions not listed go to the cover
  const [assign, setAssign] = useState<Record<string, number>>({});
  const [replan, setReplan] = useState(0);
  useEffect(() => {
    if (open) {
      setTherapist(0);
      setCover(0);
      setHandover("");
      setNotes({});
      setAssign({});
      setPreview(null);
    }
  }, [open]);
  useEffect(() => {
    setPreview(null);
    setAssign({});
  }, [therapist, from, to]);
  useEffect(() => setPreview(null), [cover]);

  const body = (dry: boolean) => ({
    therapist_id: therapist,
    cover_id: cover || undefined,
    from,
    to,
    reason,
    handover,
    notes: Object.entries(notes)
      .filter(([, v]) => v.trim())
      .map(([patient_id, note]) => ({ patient_id, note })),
    move_patients: true,
    dry_run: dry,
    assignments: Object.entries(assign).map(
      ([appointment_id, therapist_id]) => ({
        appointment_id,
        therapist_id,
      }),
    ),
  });
  const fail = (e: Error) => apiError(e, t("board_save_failed"));
  const check = useMutation({
    mutationFn: () =>
      callApi(BoardRoutes.createAbsence, { body: body(true), silent: true }),
    onSuccess: setPreview,
    onError: fail,
  });
  const save = useMutation({
    mutationFn: () =>
      callApi(BoardRoutes.createAbsence, { body: body(false), silent: true }),
    onSuccess: (r) => {
      const moved = r.sessions.filter((s) => s.moved).length;
      Notification.Success({
        msg: t("absence_saved", { moved, total: r.sessions.length }),
      });
      onDone();
    },
    onError: fail,
  });
  // A changed destination re-checks the whole plan (room left changes)
  useEffect(() => {
    if (replan) check.mutate();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [replan]);
  const choose = (appointmentId: string, therapistId: number) => {
    if (therapistId < 0) return;
    setAssign((a) => ({ ...a, [appointmentId]: therapistId }));
    setReplan((n) => n + 1);
  };
  const valid = therapist > 0 && from && to && to >= from;
  const moved = preview?.sessions.filter((s) => s.moved).length ?? 0;

  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-5xl">
        <DialogHeader>
          <DialogTitle>{t("new_absence")}</DialogTitle>
          <DialogDescription>{t("new_absence_help")}</DialogDescription>
        </DialogHeader>
        <div className="grid gap-3 sm:grid-cols-2">
          <label className="text-sm font-medium text-gray-700">
            {t("absent_physio")}
            <select
              value={therapist}
              onChange={(e) => setTherapist(Number(e.target.value))}
              className={fieldClass}
            >
              <option value={0}>-</option>
              {physios.map((p) => (
                <option key={p.id} value={p.id}>
                  {fullName(p)}
                </option>
              ))}
            </select>
          </label>
          <label className="text-sm font-medium text-gray-700">
            {t("cover_physio")}
            <select
              value={cover}
              onChange={(e) => setCover(Number(e.target.value))}
              className={fieldClass}
            >
              <option value={0}>{t("no_cover")}</option>
              {physios
                .filter((p) => p.id !== therapist)
                .map((p) => (
                  <option key={p.id} value={p.id}>
                    {fullName(p)}
                  </option>
                ))}
            </select>
          </label>
          <label className="text-sm font-medium text-gray-700">
            {t("from")}
            <DateInput value={from} onChange={setFrom} className="mt-1" />
          </label>
          <label className="text-sm font-medium text-gray-700">
            {t("until")}
            <DateInput value={to} onChange={setTo} className="mt-1" />
          </label>
          <fieldset className="sm:col-span-2">
            <legend className="text-sm font-medium text-gray-700">
              {t("reason")}
            </legend>
            <div className="mt-1 flex flex-wrap gap-2">
              {REASONS.map((r) => (
                <button
                  key={r}
                  type="button"
                  aria-pressed={reason === r}
                  onClick={() => setReason(r)}
                  className={cn(
                    "rounded-full px-3 py-1 text-sm ring-1 ring-inset",
                    reason === r
                      ? "bg-primary-600 text-white ring-primary-600"
                      : "ring-gray-300",
                  )}
                >
                  {t(`absence_reason__${r}`)}
                </button>
              ))}
            </div>
          </fieldset>
          <label className="text-sm font-medium text-gray-700 sm:col-span-2">
            {t("handover_document")}
            <textarea
              rows={4}
              value={handover}
              onChange={(e) => setHandover(e.target.value)}
              placeholder={t("handover_placeholder")}
              className="mt-1 w-full rounded-lg border border-gray-300 p-2.5 text-sm shadow-sm focus:border-primary-500 focus:ring-primary-500"
            />
          </label>
        </div>

        {preview && (
          <div className="flex flex-col gap-3">
            <p className="text-sm font-medium text-gray-900">
              {t("absence_preview", { moved, total: preview.sessions.length })}
            </p>
            {!!preview.sessions.length && (
              <AbsencePlanner
                sessions={preview.sessions}
                physios={physios}
                absentId={therapist}
                onAssign={choose}
              />
            )}
            {!!preview.patients.length && (
              <div>
                <p className="mb-1 text-sm font-medium text-gray-700">
                  {t("handover_per_patient")}
                </p>
                <div className="flex flex-col gap-2">
                  {preview.patients.map((p) => (
                    <label key={p.id} className="text-sm text-gray-700">
                      {p.name}
                      <input
                        value={notes[p.id] ?? ""}
                        maxLength={2000}
                        onChange={(e) =>
                          setNotes((n) => ({ ...n, [p.id]: e.target.value }))
                        }
                        placeholder={t("handover_patient_placeholder")}
                        className={fieldClass}
                      />
                    </label>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}

        <DialogFooter className="gap-2">
          <Button type="button" variant="outline" onClick={onClose}>
            {t("cancel")}
          </Button>
          {preview ? (
            <Button
              type="button"
              variant="primary"
              disabled={save.isPending}
              onClick={() => save.mutate()}
            >
              {t("confirm_absence")}
            </Button>
          ) : (
            <Button
              type="button"
              variant="primary"
              disabled={!valid || check.isPending}
              onClick={() => check.mutate()}
            >
              {t("preview")}
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
