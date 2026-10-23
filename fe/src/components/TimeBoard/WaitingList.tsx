import { useMutation, useQuery } from "@tanstack/react-query";
import dayjs from "dayjs";
import { CalendarRange, Plus, Search, TriangleAlert, X } from "lucide-react";
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

import { chipClass } from "@/components/Common/FilterChips";
import PatientIds from "@/components/Patient/PatientIds";
import { apiError } from "@/components/TimeBoard/BoardDialogs";
import BoardRoutes from "@/components/TimeBoard/api";
import { CARE_LEVEL_STYLES } from "@/components/TimeBoard/boardUtils";
import { WaitingEntry, WaitingPeriod } from "@/components/TimeBoard/types";

import useDebouncedValue from "@/hooks/useDebouncedValue";

import query, { callApi } from "@/Utils/request/query";

type Physio = { id: number; first_name: string; last_name: string };
const WEEK_ORDER = [1, 2, 3, 4, 5, 6, 0];
const PERIODS: WaitingPeriod[] = ["any", "morning", "afternoon"];

/** Days a patient has been waiting. */
export const daysWaiting = (created: string, now = dayjs()) =>
  Math.max(now.diff(dayjs(created), "day"), 0);

/**
 * Patients waiting for a slot: urgent first, then whoever waits longest,
 * with their preferred days and time of day. "Book plan" opens the weekly
 * plan for the patient; once booked the entry leaves the list.
 */
export default function WaitingListDialog({
  open,
  physios,
  onClose,
  onPlan,
}: {
  open: boolean;
  physios: Physio[];
  onClose: () => void;
  onPlan: (e: WaitingEntry) => void;
}) {
  const { t } = useTranslation();
  const [adding, setAdding] = useState(false);
  const { data: list, refetch } = useQuery({
    queryKey: ["waiting-list"],
    queryFn: query(BoardRoutes.waiting),
    enabled: open,
  });
  const remove = useMutation({
    mutationFn: (id: string) =>
      callApi(BoardRoutes.updateWaiting, {
        pathParams: { id },
        body: { status: "removed" },
        silent: true,
      }),
    onSuccess: () => refetch(),
    onError: (e: Error) => apiError(e, t("board_save_failed")),
  });
  const physioName = (id?: number | null) => {
    const p = physios.find((x) => x.id === id);
    return p ? `${p.first_name} ${p.last_name}` : "";
  };

  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-3xl">
        <DialogHeader>
          <DialogTitle>{t("waiting_list")}</DialogTitle>
          <DialogDescription>{t("waiting_list_help")}</DialogDescription>
        </DialogHeader>

        {adding ? (
          <AddWaiting
            physios={physios}
            onCancel={() => setAdding(false)}
            onAdded={() => {
              setAdding(false);
              refetch();
            }}
          />
        ) : (
          <Button
            variant="outline"
            className="self-start"
            onClick={() => setAdding(true)}
          >
            <Plus className="mr-1 size-4" aria-hidden />
            {t("add_to_waiting_list")}
          </Button>
        )}

        {!list?.length ? (
          <p className="text-sm text-gray-500">{t("waiting_list_empty")}</p>
        ) : (
          <ol className="flex flex-col divide-y divide-gray-100 rounded-lg border border-gray-200">
            {list.map((e) => (
              <li
                key={e.id}
                className={cn(
                  "flex flex-wrap items-center gap-3 px-3 py-2.5 text-sm",
                  // a free hour that suits the patient: stands out
                  e.slot && "bg-emerald-50 ring-1 ring-inset ring-emerald-300",
                )}
              >
                <div className="min-w-48 flex-1">
                  <p className="flex flex-wrap items-center gap-2 font-medium text-gray-900">
                    {e.urgent && (
                      <span className="flex items-center gap-1 rounded-full bg-rose-50 px-2 text-xs font-semibold text-rose-700">
                        <TriangleAlert className="size-3" aria-hidden />
                        {t("referral_priority__urgent")}
                      </span>
                    )}
                    <span
                      className={cn(
                        "rounded px-1 text-[11px] font-semibold",
                        CARE_LEVEL_STYLES[e.patient?.care_level ?? 1],
                      )}
                    >
                      {e.patient?.care_level ?? 1}
                    </span>
                    {e.patient?.name}
                  </p>
                  <p className="text-xs text-gray-500">
                    {[
                      t("waiting_n_days", {
                        count: daysWaiting(e.created_date),
                      }),
                      e.weekdays.length
                        ? e.weekdays
                            .slice()
                            .sort((a, b) => ((a + 6) % 7) - ((b + 6) % 7))
                            .map((d) => dayjs().day(d).format("ddd"))
                            .join(", ")
                        : t("any_day"),
                      t(`period__${e.period}`),
                      physioName(e.preferred_therapist_id),
                    ]
                      .filter(Boolean)
                      .join(" · ")}
                  </p>
                  {e.note && <p className="text-xs text-gray-700">{e.note}</p>}
                  {e.slot && (
                    <p className="mt-1 text-xs font-semibold text-emerald-800">
                      {t("slot_available", {
                        when: `${dayjs(e.slot.date).format("ddd D/M")} ${String(e.slot.hour).padStart(2, "0")}:00`,
                        physio: e.slot.therapist,
                      })}
                    </p>
                  )}
                </div>
                <Button size="sm" variant="primary" onClick={() => onPlan(e)}>
                  <CalendarRange className="mr-1 size-4" aria-hidden />
                  {t("book_plan")}
                </Button>
                <Button
                  size="icon"
                  variant="ghost"
                  aria-label={t("remove_from_waiting_list", {
                    name: e.patient?.name,
                  })}
                  onClick={() =>
                    window.confirm(t("confirm_remove_waiting")) &&
                    remove.mutate(e.id)
                  }
                >
                  <X className="size-4 text-gray-500" aria-hidden />
                </Button>
              </li>
            ))}
          </ol>
        )}
      </DialogContent>
    </Dialog>
  );
}

function AddWaiting({
  physios,
  onCancel,
  onAdded,
}: {
  physios: Physio[];
  onCancel: () => void;
  onAdded: () => void;
}) {
  const { t } = useTranslation();
  const [search, setSearch] = useState("");
  const debounced = useDebouncedValue(search);
  const [patient, setPatient] = useState<{ id: string; name: string } | null>(
    null,
  );
  const [urgent, setUrgent] = useState(false);
  const [days, setDays] = useState<number[]>([]);
  const [period, setPeriod] = useState<WaitingPeriod>("any");
  const [physio, setPhysio] = useState(0);
  const [note, setNote] = useState("");
  useEffect(() => setPatient(null), [debounced]);

  const { data: found } = useQuery({
    queryKey: ["waiting-patient-search", debounced],
    queryFn: query(BoardRoutes.searchPatients, {
      queryParams: { search: debounced, limit: 6 },
    }),
    enabled: debounced.trim().length >= 2 && !patient,
  });
  const add = useMutation({
    mutationFn: () =>
      callApi(BoardRoutes.addWaiting, {
        body: {
          patient_id: patient!.id,
          urgent,
          weekdays: days,
          period,
          ...(physio ? { preferred_therapist_id: physio } : {}),
          note,
        },
        silent: true,
      }),
    onSuccess: onAdded,
    onError: (e: Error) => apiError(e, t("board_save_failed")),
  });
  const toggle = (d: number) =>
    setDays((ds) => (ds.includes(d) ? ds.filter((x) => x !== d) : [...ds, d]));

  return (
    <form
      className="flex flex-col gap-3 rounded-lg border border-primary-200 bg-primary-50/40 p-3"
      onSubmit={(e) => {
        e.preventDefault();
        if (patient) add.mutate();
      }}
    >
      {patient ? (
        <p className="flex items-center justify-between rounded-lg bg-white px-3 py-2 text-sm ring-1 ring-primary-200">
          <span className="font-medium">{patient.name}</span>
          <Button
            type="button"
            size="sm"
            variant="ghost"
            onClick={() => setPatient(null)}
          >
            {t("change")}
          </Button>
        </p>
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
              className="h-10 w-full rounded-lg border border-gray-300 pl-9 pr-3 text-sm"
            />
          </div>
          {!!found?.results.length && (
            <ul className="mt-1 divide-y divide-gray-100 rounded-lg border border-gray-200 bg-white">
              {found.results.map((p) => (
                <li key={p.id}>
                  <button
                    type="button"
                    className="w-full px-3 py-2 text-left text-sm hover:bg-gray-50"
                    onClick={() =>
                      setPatient({ id: p.id!, name: p.name ?? "" })
                    }
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
      <div
        className="flex flex-wrap items-center gap-1.5"
        role="group"
        aria-label={t("preferred_days")}
      >
        <span className="w-28 text-xs font-medium uppercase tracking-wide text-gray-500">
          {t("preferred_days")}
        </span>
        {WEEK_ORDER.map((d) => (
          <button
            key={d}
            type="button"
            aria-pressed={days.includes(d)}
            className={chipClass(days.includes(d))}
            onClick={() => toggle(d)}
          >
            <span className="capitalize">{dayjs().day(d).format("ddd")}</span>
          </button>
        ))}
      </div>
      <div
        className="flex flex-wrap items-center gap-1.5"
        role="group"
        aria-label={t("preferred_time")}
      >
        <span className="w-28 text-xs font-medium uppercase tracking-wide text-gray-500">
          {t("preferred_time")}
        </span>
        {PERIODS.map((p) => (
          <button
            key={p}
            type="button"
            aria-pressed={period === p}
            className={chipClass(period === p)}
            onClick={() => setPeriod(p)}
          >
            {t(`period__${p}`)}
          </button>
        ))}
      </div>
      <div className="grid gap-3 sm:grid-cols-2">
        <label className="text-sm font-medium text-gray-700">
          {t("preferred_physio")}
          <select
            value={physio}
            onChange={(e) => setPhysio(Number(e.target.value))}
            className="mt-1 h-10 w-full rounded-lg border border-gray-300 px-3 text-sm"
          >
            <option value={0}>{t("anyone")}</option>
            {physios.map((p) => (
              <option key={p.id} value={p.id}>
                {p.first_name} {p.last_name}
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
            className="mt-1 h-10 w-full rounded-lg border border-gray-300 px-3 text-sm"
          />
        </label>
      </div>
      <label className="flex items-center gap-2 text-sm text-gray-700">
        <input
          type="checkbox"
          checked={urgent}
          onChange={(e) => setUrgent(e.target.checked)}
          className="rounded border-gray-300"
        />
        {t("urgent_waiting")}
      </label>
      <div className="flex gap-2">
        <Button
          type="submit"
          variant="primary"
          disabled={!patient || add.isPending}
        >
          {t("add_to_waiting_list")}
        </Button>
        <Button type="button" variant="outline" onClick={onCancel}>
          {t("cancel")}
        </Button>
      </div>
    </form>
  );
}
