import { useMutation, useQuery } from "@tanstack/react-query";
import dayjs from "dayjs";
import { CalendarCheck, CalendarX } from "lucide-react";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";

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
import { chipClass } from "@/components/Common/FilterChips";
import { apiError } from "@/components/TimeBoard/BoardDialogs";
import BoardRoutes from "@/components/TimeBoard/api";
import { formatHour } from "@/components/TimeBoard/boardUtils";
import { PlanRescheduleResult } from "@/components/TimeBoard/types";

import * as Notification from "@/Utils/Notifications";
import query, { callApi } from "@/Utils/request/query";

type Physio = { id: number; first_name: string; last_name: string };
const fieldClass =
  "mt-1 h-10 w-full rounded-lg border border-gray-300 px-3 text-sm shadow-sm focus:border-primary-500 focus:ring-primary-500";
const WEEK_ORDER = [1, 2, 3, 4, 5, 6, 0];

/** The session being moved, as the board shows it. */
export interface MovingSession {
  id: string;
  patientName: string;
  date: string;
  hour: number;
  therapistId: number;
}

/**
 * Moves one session to another day, hour and/or physio. The server checks
 * the same things as booking (on shift, room, care credits, patient free).
 */
export function MoveSessionDialog({
  session,
  physios,
  onClose,
  onMoved,
}: {
  session: MovingSession | null;
  physios: Physio[];
  onClose: () => void;
  onMoved: () => void;
}) {
  const { t } = useTranslation();
  const [date, setDate] = useState("");
  const [hour, setHour] = useState(9);
  const [therapist, setTherapist] = useState(0);
  useEffect(() => {
    if (!session) return;
    setDate(session.date);
    setHour(session.hour);
    setTherapist(session.therapistId);
  }, [session]);

  const move = useMutation({
    mutationFn: () =>
      callApi(BoardRoutes.moveAppointment, {
        pathParams: { id: session!.id },
        body: { date, hour, therapist_id: therapist },
        silent: true,
      }),
    onSuccess: () => {
      Notification.Success({ msg: t("session_moved") });
      onMoved();
    },
    onError: (e: Error) => apiError(e, t("session_move_failed")),
  });
  const unchanged =
    !!session &&
    date === session.date &&
    hour === session.hour &&
    therapist === session.therapistId;

  return (
    <Dialog open={!!session} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t("move_session")}</DialogTitle>
          <DialogDescription>
            {t("move_session_help", { name: session?.patientName })}
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-3 sm:grid-cols-2">
          <label className="text-sm font-medium text-gray-700 sm:col-span-2">
            {t("date")}
            <DateInput
              value={date}
              onChange={(d) => d && setDate(d)}
              className="mt-1"
            />
          </label>
          <label className="text-sm font-medium text-gray-700">
            {t("hour")}
            <select
              value={hour}
              onChange={(e) => setHour(Number(e.target.value))}
              className={fieldClass}
            >
              {Array.from({ length: 24 }, (_, h) => (
                <option key={h} value={h}>
                  {formatHour(h)}
                </option>
              ))}
            </select>
          </label>
          <label className="text-sm font-medium text-gray-700">
            {t("physiotherapist")}
            <select
              value={therapist}
              onChange={(e) => setTherapist(Number(e.target.value))}
              className={fieldClass}
            >
              {physios.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.first_name} {p.last_name}
                </option>
              ))}
            </select>
          </label>
        </div>
        <DialogFooter className="gap-2">
          <Button type="button" variant="outline" onClick={onClose}>
            {t("cancel")}
          </Button>
          <Button
            type="button"
            variant="primary"
            disabled={unchanged || move.isPending}
            onClick={() => move.mutate()}
          >
            {t("move_session")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

/**
 * Gives the rest of a weekly plan a new pattern (days, hour, physio) from a
 * date on, keeping the number of sessions still to come. Preview first.
 */
export function ReschedulePlanDialog({
  plan,
  physios,
  onClose,
  onDone,
}: {
  plan: {
    seriesId: string;
    patientName: string;
    therapistId: number;
    hour: number;
  } | null;
  physios: Physio[];
  onClose: () => void;
  onDone: () => void;
}) {
  const { t } = useTranslation();
  const [from, setFrom] = useState(dayjs().format("YYYY-MM-DD"));
  const [days, setDays] = useState<number[]>([]);
  const [hour, setHour] = useState(9);
  const [therapist, setTherapist] = useState(0);
  const [preview, setPreview] = useState<PlanRescheduleResult | null>(null);

  const { data: sessions } = useQuery({
    queryKey: ["series-sessions", plan?.seriesId],
    queryFn: query(BoardRoutes.seriesSessions, {
      pathParams: { id: plan?.seriesId ?? "" },
    }),
    enabled: !!plan,
  });
  useEffect(() => {
    if (!plan) return;
    setFrom(dayjs().format("YYYY-MM-DD"));
    setHour(plan.hour);
    setTherapist(plan.therapistId);
    setPreview(null);
  }, [plan]);
  // Start from the plan's current weekdays
  useEffect(() => {
    if (sessions?.length)
      setDays(
        [...new Set(sessions.map((s) => dayjs(s.date).day()))].sort(
          (a, b) => a - b,
        ),
      );
  }, [sessions]);
  useEffect(() => setPreview(null), [from, days, hour, therapist]);

  const body = (dry: boolean) => ({
    from,
    weekdays: days,
    hour,
    therapist_id: therapist,
    dry_run: dry,
  });
  const fail = (e: Error) => apiError(e, t("board_save_failed"));
  const check = useMutation({
    mutationFn: () =>
      callApi(BoardRoutes.reschedulePlan, {
        pathParams: { id: plan!.seriesId },
        body: body(true),
        silent: true,
      }),
    onSuccess: setPreview,
    onError: fail,
  });
  const save = useMutation({
    mutationFn: () =>
      callApi(BoardRoutes.reschedulePlan, {
        pathParams: { id: plan!.seriesId },
        body: body(false),
        silent: true,
      }),
    onSuccess: (r) => {
      Notification.Success({
        msg: t("plan_rescheduled", { count: r.booked.length }),
      });
      onDone();
    },
    onError: fail,
  });
  const toggle = (d: number) =>
    setDays((ds) => (ds.includes(d) ? ds.filter((x) => x !== d) : [...ds, d]));

  return (
    <Dialog open={!!plan} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{t("reschedule_plan")}</DialogTitle>
          <DialogDescription>
            {t("reschedule_plan_help", {
              name: plan?.patientName,
              count: sessions?.length ?? 0,
            })}
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-3">
          <label className="text-sm font-medium text-gray-700">
            {t("from")}
            <DateInput
              value={from}
              onChange={(d) => d && setFrom(d)}
              className="mt-1"
            />
          </label>
          <fieldset>
            <legend className="text-sm font-medium text-gray-700">
              {t("days_of_week")}
            </legend>
            <div className="mt-1 flex flex-wrap gap-1.5">
              {WEEK_ORDER.map((d) => (
                <button
                  key={d}
                  type="button"
                  aria-pressed={days.includes(d)}
                  className={chipClass(days.includes(d))}
                  onClick={() => toggle(d)}
                >
                  <span className="capitalize">
                    {dayjs().day(d).format("ddd")}
                  </span>
                </button>
              ))}
            </div>
          </fieldset>
          <div className="grid grid-cols-2 gap-3">
            <label className="text-sm font-medium text-gray-700">
              {t("hour")}
              <select
                value={hour}
                onChange={(e) => setHour(Number(e.target.value))}
                className={fieldClass}
              >
                {Array.from({ length: 24 }, (_, h) => (
                  <option key={h} value={h}>
                    {formatHour(h)}
                  </option>
                ))}
              </select>
            </label>
            <label className="text-sm font-medium text-gray-700">
              {t("physiotherapist")}
              <select
                value={therapist}
                onChange={(e) => setTherapist(Number(e.target.value))}
                className={fieldClass}
              >
                {physios.map((p) => (
                  <option key={p.id} value={p.id}>
                    {p.first_name} {p.last_name}
                  </option>
                ))}
              </select>
            </label>
          </div>
          {preview && (
            <div className="rounded-lg border border-gray-200 p-3 text-sm">
              <p className="mb-2 flex items-center gap-2 font-medium text-emerald-700">
                <CalendarCheck className="size-4" aria-hidden />
                {t("plan_will_move", {
                  count: preview.booked.length,
                  total: preview.replaced,
                })}
              </p>
              <div className="flex flex-wrap gap-1">
                {preview.booked.map((d) => (
                  <span
                    key={d}
                    className="rounded bg-emerald-50 px-1.5 py-0.5 text-xs capitalize text-emerald-800"
                  >
                    {dayjs(d).format("ddd D MMM")}
                  </span>
                ))}
              </div>
              {!!preview.skipped.length && (
                <p className="mt-2 flex items-center gap-2 text-xs text-amber-800">
                  <CalendarX className="size-4" aria-hidden />
                  {t("plan_skipped", { count: preview.skipped.length })}
                </p>
              )}
            </div>
          )}
        </div>
        <DialogFooter className="gap-2">
          <Button type="button" variant="outline" onClick={onClose}>
            {t("cancel")}
          </Button>
          {preview ? (
            <Button
              type="button"
              variant="primary"
              disabled={!preview.booked.length || save.isPending}
              onClick={() => save.mutate()}
            >
              {t("confirm_reschedule")}
            </Button>
          ) : (
            <Button
              type="button"
              variant="primary"
              disabled={!days.length || check.isPending}
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
