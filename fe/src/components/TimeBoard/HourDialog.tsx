import { useMutation, useQuery } from "@tanstack/react-query";
import { CalendarRange, Check, Plus, Search, UserX, X } from "lucide-react";
import { useEffect, useState } from "react";
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
import BoardRoutes from "@/components/TimeBoard/api";
import {
  CARE_LEVEL_STYLES,
  DIVISION_COLORS,
  formatHour,
  hourIsFull,
} from "@/components/TimeBoard/boardUtils";
import {
  AppointmentStatus,
  Board,
  BoardTherapist,
  HourLoad,
} from "@/components/TimeBoard/types";

import useDebouncedValue from "@/hooks/useDebouncedValue";

import * as Notification from "@/Utils/Notifications";
import query, { callApi } from "@/Utils/request/query";

const fieldClass =
  "h-10 w-full rounded-lg border border-gray-300 px-3 text-sm shadow-sm focus:border-primary-500 focus:ring-primary-500";

export interface PatientRef {
  id: string;
  name: string;
  care_level?: number;
}

/**
 * One physiotherapist-hour: its patients (click to see the patient), their
 * status, and — for coordinators — assigning a patient to this hour only or
 * starting a weekly plan from it.
 */
export default function HourDialog({
  board,
  therapist,
  load,
  onClose,
  onChanged,
  onOpenPatient,
  onPlan,
}: {
  board: Board;
  therapist: BoardTherapist | null;
  load: HourLoad | null;
  onClose: () => void;
  onChanged: () => void;
  onOpenPatient: (patientId: string) => void;
  onPlan: (patient: PatientRef) => void;
}) {
  const { t } = useTranslation();
  const [search, setSearch] = useState("");
  const debounced = useDebouncedValue(search);
  useEffect(() => setSearch(""), [load?.hour, therapist?.id]);
  const canBook = board.permissions.book_patients;
  const division = board.divisions.find((d) => d.id === load?.division_id);
  const full = load ? hourIsFull(load) : false;

  const { data: patients } = useQuery({
    queryKey: ["board-patient-search", debounced],
    queryFn: query(BoardRoutes.searchPatients, {
      queryParams: { search: debounced, limit: 6 },
    }),
    enabled: !!load && canBook && !full && debounced.trim().length >= 2,
  });

  const fail = (e: Error) => apiError(e, t("board_save_failed"));
  const book = useMutation({
    mutationFn: (patientId: string) =>
      callApi(BoardRoutes.book, {
        body: {
          date: board.date,
          hour: load!.hour,
          therapist_id: therapist!.id,
          patient_id: patientId,
        },
        silent: true,
      }),
    onSuccess: () => {
      setSearch("");
      onChanged();
    },
    onError: fail,
  });
  const status = useMutation({
    mutationFn: (v: { id: string; status: AppointmentStatus }) =>
      callApi(BoardRoutes.setStatus, {
        pathParams: { id: v.id },
        body: { status: v.status },
        silent: true,
      }),
    onSuccess: (res) => {
      // e.g. "Gloves: not enough in stock" when the session used consumables
      for (const w of res.warnings ?? []) Notification.Warn({ msg: w });
      onChanged();
    },
    onError: fail,
  });
  const unbook = useMutation({
    mutationFn: (id: string) =>
      callApi(BoardRoutes.unbook, { pathParams: { id }, silent: true }),
    onSuccess: onChanged,
    onError: fail,
  });

  return (
    <Dialog open={!!load} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>
            {therapist?.first_name} {therapist?.last_name} ·{" "}
            {load && formatHour(load.hour)}
          </DialogTitle>
          <DialogDescription className="flex flex-wrap items-center gap-2">
            {division && (
              <span
                className={cn(
                  "size-2.5 rounded-full",
                  DIVISION_COLORS[division.color].dot,
                )}
              />
            )}
            {division?.name} ·{" "}
            {t("n_of_capacity_patients", {
              count: load?.appointments.length ?? 0,
              capacity: load?.capacity ?? 0,
            })}{" "}
            ·{" "}
            {t("credits_used", {
              count: load?.credits ?? 0,
              limit: load?.credit_limit ?? 0,
            })}
          </DialogDescription>
        </DialogHeader>

        {load?.appointments.length ? (
          <ul className="divide-y divide-gray-100 rounded-lg border border-gray-200">
            {load.appointments.map((ap) => (
              <li
                key={ap.id}
                className="flex items-center gap-2 px-3 py-2 text-sm"
              >
                <button
                  type="button"
                  onClick={() => onOpenPatient(ap.patient_id)}
                  className="flex min-w-0 flex-1 items-center gap-2 text-left"
                >
                  <span className="truncate font-medium text-primary-700 underline-offset-2 hover:underline">
                    {ap.patient?.name}
                  </span>
                  <span
                    className={cn(
                      "rounded px-1.5 text-xs font-semibold",
                      CARE_LEVEL_STYLES[ap.credits],
                    )}
                    title={t("care_level_n", { level: ap.credits })}
                  >
                    {ap.credits}
                  </span>
                  {ap.series_id && (
                    <CalendarRange
                      className="size-3.5 text-gray-400"
                      aria-label={t("part_of_plan")}
                    />
                  )}
                </button>
                <span
                  className={cn(
                    "rounded-full px-2 py-0.5 text-xs",
                    ap.status === "attended" &&
                      "bg-emerald-50 text-emerald-700",
                    ap.status === "missed" && "bg-rose-50 text-rose-700",
                    ap.status === "booked" && "bg-gray-100 text-gray-700",
                  )}
                >
                  {t(`appointment_status__${ap.status}`)}
                </span>
                <Button
                  size="icon"
                  variant="ghost"
                  aria-label={t("mark_attended")}
                  onClick={() =>
                    status.mutate({ id: ap.id, status: "attended" })
                  }
                >
                  <Check className="size-4 text-emerald-600" aria-hidden />
                </Button>
                <Button
                  size="icon"
                  variant="ghost"
                  aria-label={t("mark_missed")}
                  onClick={() => status.mutate({ id: ap.id, status: "missed" })}
                >
                  <UserX className="size-4 text-amber-600" aria-hidden />
                </Button>
                {canBook && (
                  <Button
                    size="icon"
                    variant="ghost"
                    aria-label={t("remove_from_hour")}
                    onClick={() => unbook.mutate(ap.id)}
                  >
                    <X className="size-4 text-rose-600" aria-hidden />
                  </Button>
                )}
              </li>
            ))}
          </ul>
        ) : (
          <p className="text-sm text-gray-500">{t("no_patients_this_hour")}</p>
        )}

        {canBook &&
          (full ? (
            <p className="rounded-lg bg-rose-50 p-3 text-sm text-rose-700">
              {t("hour_is_full")}
            </p>
          ) : (
            <div>
              <div className="relative">
                <Search
                  className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-gray-400"
                  aria-hidden
                />
                <input
                  aria-label={t("assign_patient")}
                  placeholder={t("assign_patient_placeholder")}
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                  className={cn(fieldClass, "pl-9")}
                />
              </div>
              {!!patients?.results.length && (
                <ul className="mt-2 divide-y divide-gray-100 rounded-lg border border-gray-200">
                  {patients.results.map((p) => {
                    const level = p.care_level ?? 1;
                    const fits = load
                      ? load.credits + level <= load.credit_limit
                      : true;
                    return (
                      <li
                        key={p.id}
                        className="flex items-center gap-2 px-3 py-2 text-sm"
                      >
                        <span className="min-w-0 flex-1 truncate font-medium text-gray-900">
                          {p.name}
                        </span>
                        <span
                          className={cn(
                            "rounded px-1.5 text-xs font-semibold",
                            CARE_LEVEL_STYLES[level],
                          )}
                        >
                          {level}
                        </span>
                        <Button
                          size="sm"
                          variant="outline"
                          disabled={book.isPending || !fits}
                          title={fits ? undefined : t("not_enough_credits")}
                          onClick={() => book.mutate(p.id!)}
                        >
                          <Plus className="mr-1 size-3.5" aria-hidden />
                          {t("this_hour")}
                        </Button>
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() =>
                            onPlan({
                              id: p.id!,
                              name: p.name ?? "",
                              care_level: level,
                            })
                          }
                        >
                          <CalendarRange
                            className="mr-1 size-3.5"
                            aria-hidden
                          />
                          {t("weekly_plan")}
                        </Button>
                      </li>
                    );
                  })}
                </ul>
              )}
            </div>
          ))}
      </DialogContent>
    </Dialog>
  );
}
