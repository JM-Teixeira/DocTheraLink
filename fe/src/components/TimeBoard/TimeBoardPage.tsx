import { useQuery, useQueryClient } from "@tanstack/react-query";
import dayjs from "dayjs";
import {
  CalendarRange,
  ChevronLeft,
  ChevronRight,
  Copy,
  LayoutGrid,
  Plus,
  Repeat,
  Users,
} from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";

import { cn } from "@/lib/utils";

import { Button } from "@/components/ui/button";

import Loading from "@/components/Common/Loading";
import Page from "@/components/Common/Page";
import {
  CopyDayDialog,
  DivisionsDialog,
  ShiftDialog,
} from "@/components/TimeBoard/BoardDialogs";
import HourDialog from "@/components/TimeBoard/HourDialog";
import PatientQuickView from "@/components/TimeBoard/PatientQuickView";
import {
  PlanPreset,
  RepeatWeekDialog,
  SeriesDialog,
} from "@/components/TimeBoard/PlanDialogs";
import {
  MonthView,
  WeekView,
  YearView,
} from "@/components/TimeBoard/RangeViews";
import BoardRoutes from "@/components/TimeBoard/api";
import {
  DIVISION_COLORS,
  addDays,
  boardHours,
  formatHour,
  hourIsFull,
  hourOf,
  staffingByDivision,
  today,
  utilisation,
  weekStart,
} from "@/components/TimeBoard/boardUtils";
import {
  Board,
  BoardTherapist,
  HourLoad,
  Shift,
} from "@/components/TimeBoard/types";

import query from "@/Utils/request/query";

type Mode = "day" | "week" | "month" | "year";
type DayView = "physio" | "division";

/** Previous / next period for each mode. */
function step(mode: Mode, date: string, dir: 1 | -1) {
  const d = dayjs(date);
  switch (mode) {
    case "day":
      return addDays(date, dir);
    case "week":
      return addDays(date, 7 * dir);
    case "month":
      return d.add(dir, "month").format("YYYY-MM-DD");
    case "year":
      return d.add(dir, "year").format("YYYY-MM-DD");
  }
}

function periodLabel(mode: Mode, date: string) {
  const d = dayjs(date);
  switch (mode) {
    case "day":
      return d.format("dddd, D MMMM YYYY");
    case "week": {
      const start = dayjs(weekStart(date));
      return `${start.format("D MMM")} – ${start.add(6, "day").format("D MMM YYYY")}`;
    }
    case "month":
      return d.format("MMMM YYYY");
    case "year":
      return d.format("YYYY");
  }
}

/**
 * Time board: which physiotherapist works in which division and when, and
 * how many patients (and care credits) each one has per hour. Coordinators
 * and admins plan shifts and assign patients; week, month and year views
 * show the load and the sessions held.
 */
export default function TimeBoardPage() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [date, setDate] = useState(today());
  const [mode, setMode] = useState<Mode>("day");
  const [view, setView] = useState<DayView>("physio");
  const [hour, setHour] = useState<{
    therapistId: number;
    hour: number;
  } | null>(null);
  const [shift, setShift] = useState<Shift | null>(null);
  const [preset, setPreset] = useState<{
    therapistId: number;
    hour: number;
  } | null>(null);
  const [plan, setPlan] = useState<PlanPreset | null>(null);
  const [patientId, setPatientId] = useState<string | null>(null);
  const [divisionsOpen, setDivisionsOpen] = useState(false);
  const [copyOpen, setCopyOpen] = useState(false);
  const [repeatOpen, setRepeatOpen] = useState(false);

  const { data: board, isLoading } = useQuery({
    queryKey: ["board", date],
    queryFn: query(BoardRoutes.board, { queryParams: { date } }),
    refetchInterval: 60_000, // others change it too
  });

  const refresh = async () => {
    await queryClient.invalidateQueries({ queryKey: ["board"] });
    await queryClient.invalidateQueries({ queryKey: ["board-range"] });
    await queryClient.invalidateQueries({ queryKey: ["board-year"] });
  };
  const openDay = (d: string) => {
    setDate(d);
    setMode("day");
  };

  if (isLoading || !board) return <Loading />;
  const manage = board.permissions.manage_shifts;
  const nowHour = date === today() ? dayjs().hour() : -1;
  const current = hour
    ? board.therapists.find((x) => x.id === hour.therapistId)
    : undefined;
  const currentLoad = current && hour ? hourOf(current, hour.hour) : undefined;

  return (
    <Page
      title={t("time_board")}
      options={
        manage && (
          <div className="flex flex-wrap gap-2">
            <Button variant="outline" onClick={() => setDivisionsOpen(true)}>
              <LayoutGrid className="mr-2 size-4" aria-hidden />
              {t("divisions")}
            </Button>
            <Button variant="outline" onClick={() => setCopyOpen(true)}>
              <Copy className="mr-2 size-4" aria-hidden />
              {t("copy_day")}
            </Button>
            <Button variant="outline" onClick={() => setRepeatOpen(true)}>
              <Repeat className="mr-2 size-4" aria-hidden />
              {t("repeat_week")}
            </Button>
            <Button
              variant="outline"
              disabled={!board.therapists.length}
              onClick={() =>
                setPlan({
                  patient: null,
                  therapistId: board.therapists[0].id,
                  hour: 10,
                  startDate: date,
                })
              }
            >
              <CalendarRange className="mr-2 size-4" aria-hidden />
              {t("weekly_plan")}
            </Button>
            <Button
              variant="primary"
              disabled={!board.therapists.length || !board.divisions.length}
              onClick={() =>
                setPreset({ therapistId: board.therapists[0].id, hour: 9 })
              }
            >
              <Plus className="mr-2 size-4" aria-hidden />
              {t("new_shift")}
            </Button>
          </div>
        )
      }
    >
      {/* Period navigation and mode */}
      <div className="mb-4 flex flex-wrap items-center gap-3">
        <div className="flex items-center gap-1">
          <Button
            size="icon"
            variant="outline"
            aria-label={t("previous")}
            onClick={() => setDate(step(mode, date, -1))}
          >
            <ChevronLeft className="size-4" aria-hidden />
          </Button>
          <input
            type="date"
            aria-label={t("date")}
            value={date}
            onChange={(e) => e.target.value && setDate(e.target.value)}
            className="h-9 rounded-lg border border-gray-300 px-2 text-sm"
          />
          <Button
            size="icon"
            variant="outline"
            aria-label={t("next")}
            onClick={() => setDate(step(mode, date, 1))}
          >
            <ChevronRight className="size-4" aria-hidden />
          </Button>
          <Button variant="ghost" size="sm" onClick={() => setDate(today())}>
            {t("today")}
          </Button>
        </div>
        <span className="text-sm font-medium capitalize text-gray-700">
          {periodLabel(mode, date)}
        </span>
        <div className="ml-auto flex flex-wrap gap-2">
          <div
            className="flex rounded-lg bg-gray-100 p-0.5"
            role="tablist"
            aria-label={t("period")}
          >
            {(["day", "week", "month", "year"] as const).map((m) => (
              <button
                key={m}
                role="tab"
                aria-selected={mode === m}
                onClick={() => setMode(m)}
                className={cn(
                  "rounded-md px-3 py-1.5 text-sm",
                  mode === m ? "bg-white shadow-sm" : "text-gray-600",
                )}
              >
                {t(`period__${m}`)}
              </button>
            ))}
          </div>
          {mode === "day" && (
            <div
              className="flex rounded-lg bg-gray-100 p-0.5"
              role="tablist"
              aria-label={t("group_by")}
            >
              {(["physio", "division"] as const).map((v) => (
                <button
                  key={v}
                  role="tab"
                  aria-selected={view === v}
                  onClick={() => setView(v)}
                  className={cn(
                    "flex items-center gap-1.5 rounded-md px-3 py-1.5 text-sm",
                    view === v ? "bg-white shadow-sm" : "text-gray-600",
                  )}
                >
                  {v === "physio" ? (
                    <Users className="size-4" aria-hidden />
                  ) : (
                    <LayoutGrid className="size-4" aria-hidden />
                  )}
                  {t(`board_view__${v}`)}
                </button>
              ))}
            </div>
          )}
        </div>
      </div>

      {mode !== "year" && (
        <div className="mb-3 flex flex-wrap gap-3 text-xs text-gray-600">
          {board.divisions.map((d) => (
            <span key={d.id} className="flex items-center gap-1.5">
              <span
                className={cn(
                  "size-2.5 rounded-full",
                  DIVISION_COLORS[d.color].dot,
                )}
              />
              {d.name}{" "}
              <span className="text-gray-400">
                (
                {t("division_limits", {
                  count: d.capacity,
                  credits: d.credit_limit,
                })}
                )
              </span>
            </span>
          ))}
          <span className="flex items-center gap-1.5">
            <span className="size-2.5 rounded-full ring-2 ring-rose-400" />{" "}
            {t("full")}
          </span>
        </div>
      )}

      {mode === "day" &&
        (board.therapists.length === 0 ? (
          <p className="rounded-xl border border-dashed border-gray-300 bg-white p-8 text-center text-sm text-gray-600">
            {t("no_physios_in_clinic")}
          </p>
        ) : view === "physio" ? (
          <PhysioGrid
            board={board}
            nowHour={nowHour}
            onHour={(th, load) =>
              setHour({ therapistId: th.id, hour: load.hour })
            }
            onShift={manage ? setShift : undefined}
            onEmpty={
              manage
                ? (therapistId, h) => setPreset({ therapistId, hour: h })
                : undefined
            }
          />
        ) : (
          <DivisionGrid
            board={board}
            nowHour={nowHour}
            onHour={(th, load) =>
              setHour({ therapistId: th.id, hour: load.hour })
            }
          />
        ))}
      {mode === "week" && <WeekView date={date} onOpenDay={openDay} />}
      {mode === "month" && <MonthView date={date} onOpenDay={openDay} />}
      {mode === "year" && (
        <YearView
          year={dayjs(date).year()}
          onOpenMonth={(d) => {
            setDate(d);
            setMode("month");
          }}
        />
      )}

      <HourDialog
        board={board}
        therapist={current ?? null}
        load={currentLoad ?? null}
        onClose={() => setHour(null)}
        onChanged={refresh}
        onOpenPatient={setPatientId}
        onPlan={(p) => {
          if (!hour) return;
          setPlan({
            patient: p,
            therapistId: hour.therapistId,
            hour: hour.hour,
            startDate: date,
          });
          setHour(null);
        }}
      />
      <PatientQuickView
        patientId={patientId}
        onClose={() => setPatientId(null)}
      />
      <SeriesDialog
        preset={plan}
        therapists={board.therapists}
        onClose={() => setPlan(null)}
        onDone={() => {
          setPlan(null);
          refresh();
        }}
      />
      <ShiftDialog
        board={board}
        shift={shift}
        preset={preset}
        onClose={() => {
          setShift(null);
          setPreset(null);
        }}
        onChanged={refresh}
      />
      <DivisionsDialog
        open={divisionsOpen}
        divisions={board.divisions}
        onClose={() => setDivisionsOpen(false)}
        onChanged={refresh}
      />
      <CopyDayDialog
        open={copyOpen}
        date={date}
        onClose={() => setCopyOpen(false)}
        onCopied={(to) => {
          setCopyOpen(false);
          setDate(to);
          refresh();
        }}
      />
      <RepeatWeekDialog
        open={repeatOpen}
        date={date}
        onClose={() => setRepeatOpen(false)}
        onDone={() => {
          setRepeatOpen(false);
          refresh();
        }}
      />
    </Page>
  );
}

const nameCell =
  "sticky left-0 z-10 w-56 min-w-56 border-r border-gray-200 bg-white p-3 text-left align-top";

function HourHeader({
  board,
  nowHour,
  first,
}: {
  board: Board;
  nowHour: number;
  first: string;
}) {
  return (
    <thead>
      <tr className="text-xs text-gray-500">
        <th className={cn(nameCell, "py-2 font-medium")}>{first}</th>
        {boardHours(board).map((h) => (
          <th
            key={h}
            className={cn(
              "min-w-20 px-1 py-2 font-medium",
              h === nowHour && "text-primary-700",
            )}
          >
            {formatHour(h)}
          </th>
        ))}
      </tr>
    </thead>
  );
}

/** Rows = physiotherapists; cells = division, patients and credits per hour. */
function PhysioGrid({
  board,
  nowHour,
  onHour,
  onShift,
  onEmpty,
}: {
  board: Board;
  nowHour: number;
  onHour: (t: BoardTherapist, load: HourLoad) => void;
  onShift?: (s: Shift) => void;
  onEmpty?: (therapistId: number, hour: number) => void;
}) {
  const { t } = useTranslation();
  const division = (id: string) => board.divisions.find((d) => d.id === id);
  return (
    <div className="overflow-x-auto rounded-xl border border-gray-200 bg-white shadow-sm">
      <table className="w-full border-collapse text-sm">
        <HourHeader
          board={board}
          nowHour={nowHour}
          first={t("physiotherapist")}
        />
        <tbody className="divide-y divide-gray-100">
          {board.therapists.map((th) => (
            <tr key={th.id}>
              <th className={nameCell}>
                <p className="font-medium text-gray-900">
                  {th.first_name} {th.last_name}
                </p>
                <p className="text-xs text-gray-500">
                  {t("n_patients", { count: th.patients })} ·{" "}
                  {t("credits_used", {
                    count: th.credits,
                    limit: th.credit_capacity,
                  })}
                </p>
                <div className="mt-1 h-1.5 rounded-full bg-gray-100">
                  <div
                    className="h-1.5 rounded-full bg-primary-500"
                    style={{ width: `${utilisation(th)}%` }}
                  />
                </div>
                <div className="mt-2 flex flex-wrap gap-1">
                  {th.shifts.map((s) => {
                    const d = division(s.division_id);
                    return (
                      <button
                        key={s.id}
                        type="button"
                        disabled={!onShift}
                        onClick={() => onShift?.(s)}
                        className={cn(
                          "rounded px-1.5 py-0.5 text-xs",
                          d ? DIVISION_COLORS[d.color].chip : "bg-gray-100",
                          onShift && "hover:opacity-80",
                        )}
                      >
                        {s.start_hour}–{s.end_hour} {d?.name}
                      </button>
                    );
                  })}
                  {!th.shifts.length && (
                    <span className="text-xs text-gray-400">
                      {t("off_today")}
                    </span>
                  )}
                </div>
              </th>
              {boardHours(board).map((h) => {
                const load = hourOf(th, h);
                if (!load)
                  return (
                    <td
                      key={h}
                      className={cn("p-1", h === nowHour && "bg-primary-50/40")}
                    >
                      {onEmpty ? (
                        <button
                          type="button"
                          aria-label={t("add_shift_at", {
                            hour: formatHour(h),
                          })}
                          onClick={() => onEmpty(th.id, h)}
                          className="group flex h-16 w-full items-center justify-center rounded-md border border-dashed border-transparent hover:border-gray-300"
                        >
                          <Plus
                            className="size-4 text-gray-300 opacity-0 group-hover:opacity-100"
                            aria-hidden
                          />
                        </button>
                      ) : (
                        <div className="h-16" />
                      )}
                    </td>
                  );
                const d = division(load.division_id);
                const full = hourIsFull(load);
                return (
                  <td
                    key={h}
                    className={cn("p-1", h === nowHour && "bg-primary-50/40")}
                  >
                    <button
                      type="button"
                      onClick={() => onHour(th, load)}
                      title={load.appointments
                        .map((a) => a.patient?.name)
                        .join(", ")}
                      className={cn(
                        "flex h-16 w-full flex-col items-center justify-center rounded-md border text-xs transition hover:brightness-95",
                        d
                          ? DIVISION_COLORS[d.color].cell
                          : "bg-gray-50 border-gray-200",
                        full && "ring-2 ring-rose-400",
                      )}
                    >
                      <span className="max-w-full truncate px-1 text-[11px] text-gray-600">
                        {d?.name}
                      </span>
                      <span
                        className={cn(
                          "font-semibold",
                          full ? "text-rose-700" : "text-gray-900",
                        )}
                      >
                        {t("n_patients_short", {
                          count: load.appointments.length,
                        })}
                      </span>
                      <span className="text-[11px] text-gray-500">
                        {load.credits}/{load.credit_limit} {t("credits_short")}
                      </span>
                    </button>
                  </td>
                );
              })}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

/** Rows = divisions; cells = the physiotherapists working there that hour. */
function DivisionGrid({
  board,
  nowHour,
  onHour,
}: {
  board: Board;
  nowHour: number;
  onHour: (t: BoardTherapist, load: HourLoad) => void;
}) {
  const { t } = useTranslation();
  const staffing = staffingByDivision(board);
  return (
    <div className="overflow-x-auto rounded-xl border border-gray-200 bg-white shadow-sm">
      <table className="w-full border-collapse text-sm">
        <HourHeader board={board} nowHour={nowHour} first={t("division")} />
        <tbody className="divide-y divide-gray-100">
          {board.divisions.map((d) => (
            <tr key={d.id}>
              <th className={nameCell}>
                <p className="flex items-center gap-2 font-medium text-gray-900">
                  <span
                    className={cn(
                      "size-2.5 rounded-full",
                      DIVISION_COLORS[d.color].dot,
                    )}
                  />
                  {d.name}
                </p>
                <p className="text-xs text-gray-500">
                  {t("division_limits", {
                    count: d.capacity,
                    credits: d.credit_limit,
                  })}
                </p>
              </th>
              {boardHours(board).map((h) => (
                <td
                  key={h}
                  className={cn(
                    "p-1 align-top",
                    h === nowHour && "bg-primary-50/40",
                  )}
                >
                  <div className="flex min-h-14 flex-col gap-1">
                    {(staffing.get(d.id)?.get(h) ?? []).map(
                      ({ therapist, load }) => (
                        <button
                          key={therapist.id}
                          type="button"
                          onClick={() => onHour(therapist, load)}
                          className={cn(
                            "flex items-center justify-between gap-1 rounded px-1.5 py-0.5 text-xs",
                            DIVISION_COLORS[d.color].chip,
                            hourIsFull(load) && "ring-2 ring-rose-400",
                          )}
                        >
                          <span className="truncate">
                            {therapist.first_name}
                          </span>
                          <span className="font-semibold">
                            {load.credits}/{load.credit_limit}
                          </span>
                        </button>
                      ),
                    )}
                  </div>
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
