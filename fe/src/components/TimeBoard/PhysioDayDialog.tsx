import { useMutation } from "@tanstack/react-query";
import dayjs from "dayjs";
import { CalendarClock, MoveRight, Plane } from "lucide-react";
import { useState } from "react";
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

import { apiError } from "@/components/TimeBoard/BoardDialogs";
import { statusLabel } from "@/components/TimeBoard/MissDialog";
import {
  MoveSessionDialog,
  MovingSession,
} from "@/components/TimeBoard/MoveDialogs";
import BoardRoutes from "@/components/TimeBoard/api";
import {
  CARE_LEVEL_STYLES,
  DIVISION_COLORS,
  absenceOf,
  formatHour,
  hourIsFull,
  utilisation,
} from "@/components/TimeBoard/boardUtils";
import {
  Board,
  BoardTherapist,
  HourLoad,
  Shift,
} from "@/components/TimeBoard/types";

import * as Notification from "@/Utils/Notifications";
import { callApi } from "@/Utils/request/query";

/**
 * The deeper view of one physiotherapist's day: every hour on shift, where,
 * and the names of the patients - so a coordinator sees the real load.
 * Coordinators drag a patient to another hour, or use "move" for another
 * day or physio.
 */
export default function PhysioDayDialog({
  board,
  therapist,
  onClose,
  onPatient,
  onHour,
  onShift,
  onChanged,
}: {
  board: Board;
  therapist: BoardTherapist | null;
  onClose: () => void;
  onPatient: (patientId: string) => void;
  onHour: (load: HourLoad) => void;
  /** Coordinators edit a shift from its chip. */
  onShift?: (s: Shift) => void;
  /** After a session moved. */
  onChanged?: () => void;
}) {
  const { t } = useTranslation();
  const division = (id: string) => board.divisions.find((d) => d.id === id);
  const absence = therapist
    ? absenceOf(board.absences, therapist.id)
    : undefined;
  const pct = therapist ? utilisation(therapist) : 0;
  const canMove = board.permissions.manage_shifts;
  const [moving, setMoving] = useState<MovingSession | null>(null);
  const [over, setOver] = useState<number | null>(null);
  const drop = useMutation({
    mutationFn: (v: { id: string; hour: number }) =>
      callApi(BoardRoutes.moveAppointment, {
        pathParams: { id: v.id },
        body: { date: board.date, hour: v.hour, therapist_id: therapist!.id },
        silent: true,
      }),
    onSuccess: () => {
      Notification.Success({ msg: t("session_moved") });
      onChanged?.();
    },
    onError: (e: Error) => apiError(e, t("session_move_failed")),
  });

  return (
    <Dialog open={!!therapist} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>
            {therapist?.first_name} {therapist?.last_name}
          </DialogTitle>
          <DialogDescription className="capitalize">
            {dayjs(board.date).format("dddd, D MMMM YYYY")}
          </DialogDescription>
        </DialogHeader>

        {therapist && (
          <div className="grid gap-3 sm:grid-cols-3">
            <Metric label={t("patients_today")} value={therapist.patients} />
            <Metric
              label={t("credits")}
              value={`${therapist.credits}/${therapist.credit_capacity}`}
            />
            <div className="rounded-lg bg-gray-50 p-3">
              <p className="text-xs text-gray-500">{t("load")}</p>
              <p className="text-lg font-semibold text-gray-900">{pct}%</p>
              <div className="mt-1 h-1.5 rounded-full bg-gray-200">
                <div
                  className={cn(
                    "h-1.5 rounded-full",
                    pct >= 90 ? "bg-rose-500" : "bg-primary-500",
                  )}
                  style={{ width: `${pct}%` }}
                />
              </div>
            </div>
          </div>
        )}

        {!!therapist?.shifts.length && (
          <div className="flex flex-wrap items-center gap-2 text-sm">
            <span className="text-xs font-medium uppercase tracking-wide text-gray-500">
              {t("shifts")}
            </span>
            {therapist.shifts.map((s) => {
              const d = division(s.division_id);
              return (
                <button
                  key={s.id}
                  type="button"
                  disabled={!onShift}
                  onClick={() => onShift?.(s)}
                  className={cn(
                    "rounded px-2 py-0.5 text-xs",
                    d ? DIVISION_COLORS[d.color].chip : "bg-gray-100",
                    onShift && "hover:opacity-80",
                  )}
                >
                  {formatHour(s.start_hour)}-{formatHour(s.end_hour)} ·{" "}
                  {d?.name}
                </button>
              );
            })}
          </div>
        )}

        {absence && (
          <p className="flex items-center gap-2 rounded-lg bg-amber-50 px-3 py-2 text-sm text-amber-900">
            <Plane className="size-4" aria-hidden />
            {t(`absence_reason__${absence.reason}`)} ·{" "}
            {dayjs(absence.from).format("D MMM")} -{" "}
            {dayjs(absence.to).format("D MMM")}
            {absence.cover &&
              ` · ${t("covered_by", { name: `${absence.cover.first_name} ${absence.cover.last_name}` })}`}
          </p>
        )}

        {!therapist?.hours.length ? (
          <p className="text-sm text-gray-500">{t("not_working_this_day")}</p>
        ) : (
          <ol className="flex flex-col gap-2">
            {therapist.hours.map((h) => {
              const d = division(h.division_id);
              const full = hourIsFull(h);
              return (
                <li
                  key={h.hour}
                  data-hour={h.hour}
                  // Coordinators drop a patient here to move the session
                  onDragOver={(e) => {
                    if (!canMove) return;
                    e.preventDefault();
                    setOver(h.hour);
                  }}
                  onDragLeave={() => setOver(null)}
                  onDrop={(e) => {
                    if (!canMove) return;
                    e.preventDefault();
                    setOver(null);
                    const id = e.dataTransfer.getData("text/appointment");
                    const from = Number(e.dataTransfer.getData("text/hour"));
                    if (id && from !== h.hour)
                      drop.mutate({ id, hour: h.hour });
                  }}
                  className={cn(
                    "rounded-lg border p-3",
                    d ? DIVISION_COLORS[d.color].cell : "border-gray-200",
                    full && "ring-2 ring-rose-400",
                    over === h.hour && "ring-2 ring-primary-500",
                  )}
                >
                  <div className="flex flex-wrap items-center gap-2 text-sm">
                    <span className="font-mono font-semibold text-gray-800">
                      {formatHour(h.hour)}
                    </span>
                    {d && (
                      <span
                        className={cn(
                          "rounded px-1.5 text-xs",
                          DIVISION_COLORS[d.color].chip,
                        )}
                      >
                        {d.name}
                      </span>
                    )}
                    <span className="text-xs text-gray-500">
                      {t("n_of_capacity_patients", {
                        count: h.appointments.length,
                        capacity: h.capacity,
                      })}{" "}
                      · {h.credits}/{h.credit_limit} {t("credits_short")}
                    </span>
                    {board.permissions.book_patients && (
                      <Button
                        size="sm"
                        variant="ghost"
                        className="ml-auto"
                        onClick={() => onHour(h)}
                      >
                        <CalendarClock className="mr-1 size-3.5" aria-hidden />
                        {t("manage_hour")}
                      </Button>
                    )}
                  </div>
                  {h.appointments.length ? (
                    <ul className="mt-2 flex flex-wrap gap-2">
                      {h.appointments.map((a) => (
                        <li key={a.id} className="flex items-center gap-0.5">
                          <button
                            type="button"
                            draggable={canMove && a.status === "booked"}
                            onDragStart={(e) => {
                              e.dataTransfer.setData("text/appointment", a.id);
                              e.dataTransfer.setData(
                                "text/hour",
                                String(h.hour),
                              );
                              e.dataTransfer.effectAllowed = "move";
                            }}
                            onClick={() => onPatient(a.patient_id)}
                            className="flex items-center gap-1.5 rounded-full bg-white px-2.5 py-1 text-sm shadow-sm ring-1 ring-gray-200 hover:ring-primary-300"
                          >
                            <span
                              className={cn(
                                "rounded px-1 text-[11px] font-semibold",
                                CARE_LEVEL_STYLES[a.credits],
                              )}
                            >
                              {a.credits}
                            </span>
                            <span className="font-medium text-gray-900">
                              {a.patient?.name}
                            </span>
                            {a.status !== "booked" && (
                              <span className="text-xs text-gray-500">
                                · {statusLabel(t, a)}
                              </span>
                            )}
                          </button>
                          {canMove && a.status === "booked" && (
                            <Button
                              size="icon"
                              variant="ghost"
                              className="size-7"
                              aria-label={t("move_session_of", {
                                name: a.patient?.name,
                              })}
                              onClick={() =>
                                setMoving({
                                  id: a.id,
                                  patientName: a.patient?.name ?? "",
                                  date: board.date,
                                  hour: h.hour,
                                  therapistId: therapist.id,
                                })
                              }
                            >
                              <MoveRight className="size-3.5" aria-hidden />
                            </Button>
                          )}
                        </li>
                      ))}
                    </ul>
                  ) : (
                    <p className="mt-1 text-xs text-gray-500">{t("free")}</p>
                  )}
                </li>
              );
            })}
          </ol>
        )}
        {canMove && !!therapist?.hours.length && (
          <p className="text-xs text-gray-500">{t("drag_to_move_hint")}</p>
        )}
      </DialogContent>
      <MoveSessionDialog
        session={moving}
        physios={board.therapists}
        onClose={() => setMoving(null)}
        onMoved={() => {
          setMoving(null);
          onChanged?.();
        }}
      />
    </Dialog>
  );
}

function Metric({ label, value }: { label: string; value: number | string }) {
  return (
    <div className="rounded-lg bg-gray-50 p-3">
      <p className="text-xs text-gray-500">{label}</p>
      <p className="text-lg font-semibold text-gray-900">{value}</p>
    </div>
  );
}
