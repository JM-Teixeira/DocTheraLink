import { useMutation, useQuery } from "@tanstack/react-query";
import dayjs from "dayjs";
import { CalendarCheck, CalendarX, Search } from "lucide-react";
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
import PatientIds from "@/components/Patient/PatientIds";
import { apiError } from "@/components/TimeBoard/BoardDialogs";
import { PatientRef } from "@/components/TimeBoard/HourDialog";
import BoardRoutes from "@/components/TimeBoard/api";
import {
  CARE_LEVEL_STYLES,
  PLAN_PRESETS,
  formatHour,
  weekStart,
} from "@/components/TimeBoard/boardUtils";
import { SeriesResult } from "@/components/TimeBoard/types";

import useDateFormat from "@/hooks/useDateFormat";
import useDebouncedValue from "@/hooks/useDebouncedValue";

import * as Notification from "@/Utils/Notifications";
import query, { callApi } from "@/Utils/request/query";

const fieldClass =
  "mt-1 h-10 w-full rounded-lg border border-gray-300 px-3 text-sm shadow-sm focus:border-primary-500 focus:ring-primary-500";

/** Weekdays in display order (Monday first); values are 0 = Sunday … 6. */
const WEEK_ORDER = [1, 2, 3, 4, 5, 6, 0];

export interface PlanPreset {
  patient: PatientRef | null;
  therapistId: number;
  hour: number;
  startDate: string;
  /** Preferred weekdays (e.g. from the waiting list); default the start date weekday. */
  weekdays?: number[];
}

/**
 * Weekly treatment plan: a patient with a physio on chosen weekdays at an
 * hour, from a date, for N sessions. "Preview" shows which dates fit and
 * which are skipped (physio off, hour full, too many credits, patient busy)
 * before anything is booked.
 */
export function SeriesDialog({
  preset,
  therapists,
  onClose,
  onDone,
}: {
  preset: PlanPreset | null;
  therapists: { id: number; first_name: string; last_name: string }[];
  onClose: () => void;
  onDone: (r: SeriesResult) => void;
}) {
  const { t } = useTranslation();
  const [patient, setPatient] = useState<PatientRef | null>(null);
  const [search, setSearch] = useState("");
  const debounced = useDebouncedValue(search);
  const [therapist, setTherapist] = useState(0);
  const [hour, setHour] = useState(10);
  const [days, setDays] = useState<number[]>(PLAN_PRESETS.three);
  const [start, setStart] = useState("");
  const [sessions, setSessions] = useState(10);
  const [preview, setPreview] = useState<SeriesResult | null>(null);

  useEffect(() => {
    if (!preset) return;
    setPatient(preset.patient);
    setTherapist(preset.therapistId);
    setHour(preset.hour);
    setStart(preset.startDate);
    setDays(
      preset.weekdays?.length
        ? preset.weekdays
        : [dayjs(preset.startDate).day()],
    );
    setSessions(10);
    setSearch("");
    setPreview(null);
  }, [preset]);
  // Any change invalidates the preview
  useEffect(
    () => setPreview(null),
    [patient, therapist, hour, days, start, sessions],
  );

  const { data: found } = useQuery({
    queryKey: ["plan-patient-search", debounced],
    queryFn: query(BoardRoutes.searchPatients, {
      queryParams: { search: debounced, limit: 6 },
    }),
    enabled: !!preset && !patient && debounced.trim().length >= 2,
  });

  const body = (dryRun: boolean) => ({
    patient_id: patient!.id,
    therapist_id: therapist,
    weekdays: days,
    hour,
    start_date: start,
    sessions,
    dry_run: dryRun,
  });
  const fail = (e: Error) => apiError(e, t("board_save_failed"));
  const check = useMutation({
    mutationFn: () =>
      callApi(BoardRoutes.series, { body: body(true), silent: true }),
    onSuccess: setPreview,
    onError: fail,
  });
  const save = useMutation({
    mutationFn: () =>
      callApi(BoardRoutes.series, { body: body(false), silent: true }),
    onSuccess: (r) => {
      Notification.Success({
        msg: t("plan_booked", { count: r.booked.length }),
      });
      onDone(r);
    },
    onError: fail,
  });
  const valid =
    !!patient && therapist > 0 && days.length > 0 && !!start && sessions >= 1;
  const toggle = (d: number) =>
    setDays((ds) => (ds.includes(d) ? ds.filter((x) => x !== d) : [...ds, d]));

  return (
    <Dialog open={!!preset} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{t("weekly_plan")}</DialogTitle>
          <DialogDescription>{t("weekly_plan_help")}</DialogDescription>
        </DialogHeader>

        <div className="grid gap-3">
          {/* Patient */}
          {patient ? (
            <div className="flex items-center justify-between rounded-lg bg-primary-50 px-3 py-2 ring-1 ring-primary-200">
              <span className="flex items-center gap-2 font-medium text-gray-900">
                {patient.name}
                <span
                  className={cn(
                    "rounded px-1.5 text-xs font-semibold",
                    CARE_LEVEL_STYLES[patient.care_level ?? 1],
                  )}
                >
                  {patient.care_level ?? 1}
                </span>
              </span>
              <Button
                size="sm"
                variant="ghost"
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
                  aria-label={t("patient")}
                  placeholder={t("assign_patient_placeholder")}
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                  className={cn(fieldClass, "mt-0 pl-9")}
                />
              </div>
              {!!found?.results.length && (
                <ul className="mt-2 divide-y divide-gray-100 rounded-lg border border-gray-200">
                  {found.results.map((p) => (
                    <li key={p.id}>
                      <button
                        type="button"
                        onClick={() =>
                          setPatient({
                            id: p.id!,
                            name: p.name ?? "",
                            care_level: p.care_level ?? 1,
                          })
                        }
                        className="w-full px-3 py-2 text-left text-sm hover:bg-gray-50"
                      >
                        <span className="block font-medium">{p.name}</span>
                        <PatientIds nif={p.nif} sns={p.sns_number} />
                      </button>
                    </li>
                  ))}
                </ul>
              )}
            </div>
          )}

          <div className="grid grid-cols-2 gap-3">
            <label className="text-sm font-medium text-gray-700">
              {t("physiotherapist")}
              <select
                value={therapist}
                onChange={(e) => setTherapist(Number(e.target.value))}
                className={fieldClass}
              >
                <option value={0}>-</option>
                {therapists.map((th) => (
                  <option key={th.id} value={th.id}>
                    {th.first_name} {th.last_name}
                  </option>
                ))}
              </select>
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
          </div>

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
                  onClick={() => toggle(d)}
                  className={cn(
                    "w-12 rounded-md py-1.5 text-sm capitalize ring-1 ring-inset",
                    days.includes(d)
                      ? "bg-primary-600 text-white ring-primary-600"
                      : "bg-white text-gray-700 ring-gray-300",
                  )}
                >
                  {dayjs().day(d).format("ddd")}
                </button>
              ))}
            </div>
            <div className="mt-2 flex flex-wrap gap-1.5 text-xs">
              {Object.entries(PLAN_PRESETS).map(([key, value]) => (
                <button
                  key={key}
                  type="button"
                  onClick={() => setDays(value)}
                  className="rounded-full bg-gray-100 px-2.5 py-1 text-gray-700 hover:bg-gray-200"
                >
                  {t(`plan_preset__${key}`)}
                </button>
              ))}
            </div>
          </fieldset>

          <div className="grid grid-cols-2 gap-3">
            <label className="text-sm font-medium text-gray-700">
              {t("starting")}
              <DateInput value={start} onChange={setStart} className="mt-1" />
            </label>
            <label className="text-sm font-medium text-gray-700">
              {t("number_of_sessions")}
              <input
                type="number"
                min={1}
                max={100}
                value={sessions}
                onChange={(e) => setSessions(Number(e.target.value))}
                className={fieldClass}
              />
            </label>
          </div>

          {preview && (
            <div className="rounded-lg border border-gray-200 p-3 text-sm">
              <p className="mb-2 flex items-center gap-2 font-medium text-emerald-700">
                <CalendarCheck className="size-4" aria-hidden />
                {t("plan_will_book", { count: preview.booked.length })}
              </p>
              <div className="flex flex-wrap gap-1">
                {preview.booked.map((d) => (
                  <span
                    key={d}
                    className="rounded bg-emerald-50 px-1.5 py-0.5 text-xs text-emerald-800"
                  >
                    {dayjs(d).format("ddd D MMM")}
                  </span>
                ))}
              </div>
              {!!preview.skipped.length && (
                <>
                  <p className="mb-1 mt-3 flex items-center gap-2 font-medium text-amber-700">
                    <CalendarX className="size-4" aria-hidden />
                    {t("plan_skipped", { count: preview.skipped.length })}
                  </p>
                  <ul className="max-h-32 overflow-auto text-xs text-gray-600">
                    {preview.skipped.map((s) => (
                      <li key={s.date}>
                        {dayjs(s.date).format("ddd D MMM")} -{" "}
                        {t(`skip_reason__${s.reason}`)}
                      </li>
                    ))}
                  </ul>
                </>
              )}
              {preview.booked.length < sessions && (
                <p className="mt-2 text-xs text-rose-700">
                  {t("plan_short", { count: sessions - preview.booked.length })}
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
              {t("book_n_sessions", { count: preview.booked.length })}
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

/** Repeats this week's shifts for the next N weeks (days already planned are kept). */
export function RepeatWeekDialog({
  open,
  date,
  onClose,
  onDone,
}: {
  open: boolean;
  date: string;
  onClose: () => void;
  onDone: () => void;
}) {
  const { t } = useTranslation();
  const { fmt } = useDateFormat();
  const [weeks, setWeeks] = useState(4);
  const monday = weekStart(date);
  const repeat = useMutation({
    mutationFn: () =>
      callApi(BoardRoutes.repeatWeek, {
        body: { week_start: monday, weeks },
        silent: true,
      }),
    onSuccess: (r) => {
      Notification.Success({ msg: t("shifts_copied", { count: r.copied }) });
      onDone();
    },
    onError: (e: Error) => apiError(e, t("board_save_failed")),
  });
  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="sm:max-w-sm">
        <DialogHeader>
          <DialogTitle>{t("repeat_week")}</DialogTitle>
          <DialogDescription>
            {t("repeat_week_help", {
              date: fmt(monday),
            })}
          </DialogDescription>
        </DialogHeader>
        <label className="text-sm font-medium text-gray-700">
          {t("for_next_weeks")}
          <input
            type="number"
            min={1}
            max={26}
            value={weeks}
            onChange={(e) => setWeeks(Number(e.target.value))}
            className={fieldClass}
          />
        </label>
        <DialogFooter className="gap-2">
          <Button type="button" variant="outline" onClick={onClose}>
            {t("cancel")}
          </Button>
          <Button
            type="button"
            variant="primary"
            disabled={weeks < 1 || repeat.isPending}
            onClick={() => repeat.mutate()}
          >
            {t("repeat")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
