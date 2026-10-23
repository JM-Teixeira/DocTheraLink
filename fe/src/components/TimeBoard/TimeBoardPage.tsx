import { useQuery, useQueryClient } from "@tanstack/react-query";
import dayjs from "dayjs";
import {
  CalendarRange,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  Copy,
  LayoutGrid,
  ListOrdered,
  Plane,
  Plus,
  Repeat,
  Settings2,
  Users,
} from "lucide-react";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";

import { cn } from "@/lib/utils";

import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

import DateInput from "@/components/Common/DateInput";
import Loading from "@/components/Common/Loading";
import Page from "@/components/Common/Page";
import StaffRoutes from "@/components/Staff/api";
import { AbsencesDialog } from "@/components/TimeBoard/AbsenceDialogs";
import {
  CopyDayDialog,
  DivisionsDialog,
  ShiftDialog,
} from "@/components/TimeBoard/BoardDialogs";
import BoardFilters, {
  loadFilter,
  saveFilter,
} from "@/components/TimeBoard/BoardFilters";
import HourDialog from "@/components/TimeBoard/HourDialog";
import PatientQuickView from "@/components/TimeBoard/PatientQuickView";
import PhysioDayDialog from "@/components/TimeBoard/PhysioDayDialog";
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
import WaitingListDialog from "@/components/TimeBoard/WaitingList";
import BoardRoutes from "@/components/TimeBoard/api";
import {
  BoardFilter,
  DIVISION_COLORS,
  absenceOf,
  addDays,
  boardHours,
  filterBoard,
  floorsOf,
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

import useAuthUser from "@/hooks/useAuthUser";

import query, { callApi } from "@/Utils/request/query";
import { isAdmin } from "@/Utils/roles";
import { getWorkingClinic, setWorkingClinic } from "@/Utils/workingClinic";

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
      return `${start.format("D MMM")} - ${start.add(6, "day").format("D MMM YYYY")}`;
    }
    case "month":
      return d.format("MMMM YYYY");
    case "year":
      return d.format("YYYY");
  }
}

/**
 * Time board: who works where and when, and how loaded each physio is.
 * Filters (floor, division, physio) keep it readable; clicking a physio
 * opens their day with the patients' names; coordinators plan shifts,
 * absences (with handover) and assign patients.
 */
export default function TimeBoardPage() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  // Admins belong to no clinic: they choose which clinic's board to see
  const admin = isAdmin(useAuthUser());
  const [workingClinic, setClinicState] = useState(getWorkingClinic);
  const { data: clinics } = useQuery({
    queryKey: ["staff-clinics"],
    queryFn: query(StaffRoutes.clinics, { queryParams: { limit: 100 } }),
    enabled: admin,
  });
  const chooseClinic = (id: string) => {
    setWorkingClinic(id);
    setClinicState(id);
    queryClient.invalidateQueries(); // every answer was for the old clinic
  };
  useEffect(() => {
    if (admin && !workingClinic && clinics?.results[0])
      chooseClinic(clinics.results[0].id);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [admin, workingClinic, clinics]);
  const [date, setDate] = useState(today());
  // Bumped on every "Today" click: the views pulse today's cell
  const [todayFlash, setTodayFlash] = useState(0);
  const onToday = date === today();
  const [mode, setMode] = useState<Mode>("day");
  const [view, setView] = useState<DayView>("physio");
  const [filter, setFilterState] = useState<BoardFilter>(loadFilter);
  const [hour, setHour] = useState<{
    therapistId: number;
    hour: number;
  } | null>(null);
  const [physio, setPhysio] = useState<number | null>(null);
  const [shift, setShift] = useState<Shift | null>(null);
  const [preset, setPreset] = useState<{
    therapistId: number;
    hour: number;
  } | null>(null);
  const [plan, setPlan] = useState<PlanPreset | null>(null);
  const [patientId, setPatientId] = useState<string | null>(null);
  const [dialog, setDialog] = useState<
    "divisions" | "copy" | "repeat" | "absences" | "waiting" | null
  >(null);
  // The waiting-list entry a plan is being booked for
  const [waitingFor, setWaitingFor] = useState<string | null>(null);

  const setFilter = (f: BoardFilter) => {
    setFilterState(f);
    saveFilter(f);
  };

  const { data: board, isLoading } = useQuery({
    queryKey: ["board", date],
    queryFn: query(BoardRoutes.board, { queryParams: { date } }),
    refetchInterval: 60_000, // others change it too
  });

  const { data: waiting } = useQuery({
    queryKey: ["waiting-list"],
    queryFn: query(BoardRoutes.waiting, { silent: true }),
    retry: false,
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
  const shown = filterBoard(board, filter);
  const find = (id: number | null | undefined) =>
    board.therapists.find((x) => x.id === id);
  const current = hour ? find(hour.therapistId) : undefined;
  const currentLoad = current && hour ? hourOf(current, hour.hour) : undefined;

  return (
    <Page
      help="time_board"
      title={t("time_board")}
      options={
        manage && (
          // Daily actions stay visible; occasional ones live in "Organizar"
          <div className="flex flex-wrap justify-end gap-2">
            <Button variant="outline" onClick={() => setDialog("waiting")}>
              <ListOrdered className="mr-2 size-4" aria-hidden />
              {t("waiting_list")}
              {!!waiting?.length && (
                <span className="ml-1.5 rounded-full bg-amber-100 px-1.5 text-xs font-semibold text-amber-800">
                  {waiting.length}
                </span>
              )}
              {/* how many of them have a free slot that suits them now */}
              {!!waiting?.filter((w) => w.slot).length && (
                <span
                  title={t("n_with_free_slot", {
                    count: waiting.filter((w) => w.slot).length,
                  })}
                  className="ml-1 rounded-full bg-emerald-600 px-1.5 text-xs font-semibold text-white"
                >
                  {waiting.filter((w) => w.slot).length}
                </span>
              )}
            </Button>
            <DropdownMenu modal={false}>
              <DropdownMenuTrigger asChild>
                <Button variant="outline">
                  <Settings2 className="mr-2 size-4" aria-hidden />
                  {t("organize")}
                  <ChevronDown className="ml-1.5 size-4" aria-hidden />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-56">
                <DropdownMenuLabel>{t("team")}</DropdownMenuLabel>
                <DropdownMenuItem onSelect={() => setDialog("absences")}>
                  <Plane className="mr-2 size-4" aria-hidden />
                  {t("absences")}
                </DropdownMenuItem>
                <DropdownMenuItem onSelect={() => setDialog("divisions")}>
                  <LayoutGrid className="mr-2 size-4" aria-hidden />
                  {t("divisions")}
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuLabel>{t("repeat_shifts")}</DropdownMenuLabel>
                <DropdownMenuItem onSelect={() => setDialog("copy")}>
                  <Copy className="mr-2 size-4" aria-hidden />
                  {t("copy_day")}
                </DropdownMenuItem>
                <DropdownMenuItem onSelect={() => setDialog("repeat")}>
                  <Repeat className="mr-2 size-4" aria-hidden />
                  {t("repeat_week")}
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
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
      <div className="mb-3 flex flex-wrap items-center gap-3">
        {admin && (
          <select
            aria-label={t("clinic")}
            value={workingClinic}
            onChange={(e) => chooseClinic(e.target.value)}
            className="h-9 rounded-lg border border-gray-300 px-3 text-sm shadow-sm focus:border-primary-500 focus:ring-primary-500"
          >
            {clinics?.results.map((c) => (
              <option key={c.id} value={c.id}>
                {c.name}
              </option>
            ))}
          </select>
        )}
        <div className="flex items-center gap-1">
          <Button
            size="icon"
            variant="outline"
            aria-label={t("previous")}
            onClick={() => setDate(step(mode, date, -1))}
          >
            <ChevronLeft className="size-4" aria-hidden />
          </Button>
          <DateInput
            aria-label={t("date")}
            value={date}
            onChange={(d) => d && setDate(d)}
            className="w-40"
          />
          <Button
            size="icon"
            variant="outline"
            aria-label={t("next")}
            onClick={() => setDate(step(mode, date, 1))}
          >
            <ChevronRight className="size-4" aria-hidden />
          </Button>
          <Button
            variant={onToday && mode === "day" ? "outline" : "ghost"}
            size="sm"
            aria-pressed={onToday}
            className={cn(onToday && "font-semibold text-primary-700")}
            onClick={() => {
              setDate(today());
              setTodayFlash((n) => n + 1);
            }}
          >
            {t("today")}
          </Button>
        </div>
        <span
          className={cn(
            "flex items-center gap-2 text-sm capitalize",
            mode === "day" && onToday
              ? "font-bold text-primary-800"
              : "font-medium text-gray-700",
          )}
        >
          {periodLabel(mode, date)}
          {mode === "day" && onToday && (
            <span className="rounded-full bg-primary-100 px-2 py-0.5 text-xs font-semibold normal-case text-primary-800">
              {t("today")}
            </span>
          )}
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

      {(mode === "day" || mode === "week") && (
        <BoardFilters
          divisions={board.divisions}
          physios={board.therapists}
          value={filter}
          onChange={setFilter}
        />
      )}

      {mode === "day" &&
        (shown.therapists.length === 0 ? (
          <p className="rounded-xl border border-dashed border-gray-300 bg-white p-8 text-center text-sm text-gray-600">
            {board.therapists.length
              ? t("nothing_matches_filters")
              : t("no_physios_in_clinic")}
          </p>
        ) : view === "physio" ? (
          <PhysioGrid
            board={shown}
            nowHour={nowHour}
            onHour={(th, load) =>
              setHour({ therapistId: th.id, hour: load.hour })
            }
            onPhysio={(th) => setPhysio(th.id)}
            onEmpty={
              manage
                ? (therapistId, h) => setPreset({ therapistId, hour: h })
                : undefined
            }
          />
        ) : (
          <DivisionGrid
            board={shown}
            nowHour={nowHour}
            onHour={(th, load) =>
              setHour({ therapistId: th.id, hour: load.hour })
            }
          />
        ))}
      {mode === "week" && (
        <WeekView
          date={date}
          onOpenDay={openDay}
          physios={filter.physios}
          flash={todayFlash}
        />
      )}
      {mode === "month" && (
        <MonthView date={date} onOpenDay={openDay} flash={todayFlash} />
      )}
      {mode === "year" && (
        <YearView
          year={dayjs(date).year()}
          flash={todayFlash}
          onOpenMonth={(d) => {
            setDate(d);
            setMode("month");
          }}
        />
      )}

      <PhysioDayDialog
        board={board}
        therapist={find(physio) ?? null}
        onClose={() => setPhysio(null)}
        onChanged={refresh}
        onPatient={setPatientId}
        onShift={
          manage
            ? (s) => {
                setPhysio(null);
                setShift(s);
              }
            : undefined
        }
        onHour={(load) => {
          if (physio) setHour({ therapistId: physio, hour: load.hour });
          setPhysio(null);
        }}
      />
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
        onClose={() => {
          setPlan(null);
          setWaitingFor(null);
        }}
        onDone={async (r) => {
          setPlan(null);
          // Booked from the waiting list: the patient leaves it
          if (waitingFor && r.series_id) {
            await callApi(BoardRoutes.updateWaiting, {
              pathParams: { id: waitingFor },
              body: { status: "scheduled", series_id: r.series_id },
              silent: true,
            }).catch(() => undefined);
            setWaitingFor(null);
            queryClient.invalidateQueries({ queryKey: ["waiting-list"] });
          }
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
        open={dialog === "divisions"}
        divisions={board.divisions}
        onClose={() => setDialog(null)}
        onChanged={refresh}
      />
      <WaitingListDialog
        open={dialog === "waiting"}
        physios={board.therapists}
        onClose={() => setDialog(null)}
        onPlan={(e) => {
          setDialog(null);
          setWaitingFor(e.id);
          setPlan({
            patient: e.patient
              ? {
                  id: e.patient.id,
                  name: e.patient.name,
                  care_level: e.patient.care_level,
                }
              : null,
            // the free slot when there is one, otherwise the preferences
            therapistId:
              e.slot?.therapist_id ??
              e.preferred_therapist_id ??
              board.therapists[0]?.id ??
              0,
            hour: e.slot?.hour ?? (e.period === "afternoon" ? 15 : 9),
            startDate: e.slot?.date ?? date,
            weekdays: e.weekdays,
          });
        }}
      />
      <AbsencesDialog
        open={dialog === "absences"}
        physios={board.therapists}
        onClose={() => setDialog(null)}
        onChanged={refresh}
      />
      <CopyDayDialog
        open={dialog === "copy"}
        date={date}
        onClose={() => setDialog(null)}
        onCopied={(to) => {
          setDialog(null);
          setDate(to);
          refresh();
        }}
      />
      <RepeatWeekDialog
        open={dialog === "repeat"}
        date={date}
        onClose={() => setDialog(null)}
        onDone={() => {
          setDialog(null);
          refresh();
        }}
      />
    </Page>
  );
}

const nameCell =
  "sticky left-0 z-10 w-52 min-w-52 border-r border-gray-200 bg-white px-3 py-2 text-left align-middle";

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
        <th className={cn(nameCell, "font-medium")}>{first}</th>
        {boardHours(board).map((h) => (
          <th
            key={h}
            className={cn(
              "min-w-14 px-0.5 py-2 font-medium",
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

/**
 * Rows = physios. Each hour shows only how many patients there are, with
 * the division as colour and the care credits as a thin bar; the details
 * (division, names) are in the tooltip and in the physio's day.
 */
function PhysioGrid({
  board,
  nowHour,
  onHour,
  onPhysio,
  onEmpty,
}: {
  board: Board;
  nowHour: number;
  onHour: (t: BoardTherapist, load: HourLoad) => void;
  onPhysio: (t: BoardTherapist) => void;
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
          {board.therapists.map((th) => {
            const absence = absenceOf(board.absences, th.id);
            const pct = utilisation(th);
            return (
              <tr
                key={th.id}
                data-physio-row
                className={cn(absence && "bg-amber-50/40")}
              >
                <th className={nameCell}>
                  <button
                    type="button"
                    onClick={() => onPhysio(th)}
                    className="group w-full text-left"
                  >
                    <span className="block truncate font-medium text-gray-900 group-hover:text-primary-700 group-hover:underline">
                      {th.first_name} {th.last_name}
                    </span>
                    {absence ? (
                      <span className="flex items-center gap-1 text-xs text-amber-800">
                        <Plane className="size-3" aria-hidden />
                        {t(`absence_reason__${absence.reason}`)}
                        {absence.cover && ` · ${absence.cover.first_name}`}
                      </span>
                    ) : (
                      <span className="flex items-center gap-2 text-xs text-gray-500">
                        <span className="h-1.5 w-16 rounded-full bg-gray-100">
                          <span
                            className={cn(
                              "block h-1.5 rounded-full",
                              pct >= 90 ? "bg-rose-500" : "bg-primary-500",
                            )}
                            style={{ width: `${pct}%` }}
                          />
                        </span>
                        {t("n_patients", { count: th.patients })}
                      </span>
                    )}
                  </button>
                </th>
                {boardHours(board).map((h) => {
                  const load = hourOf(th, h);
                  if (!load)
                    return (
                      <td
                        key={h}
                        className={cn(
                          "p-0.5",
                          h === nowHour && "bg-primary-50/40",
                        )}
                      >
                        {onEmpty && !absence ? (
                          <button
                            type="button"
                            aria-label={t("add_shift_at", {
                              hour: formatHour(h),
                            })}
                            onClick={() => onEmpty(th.id, h)}
                            className="group flex h-11 w-full items-center justify-center rounded-md hover:bg-gray-50"
                          >
                            <Plus
                              className="size-3.5 text-gray-300 opacity-0 group-hover:opacity-100"
                              aria-hidden
                            />
                          </button>
                        ) : (
                          <div className="h-11" />
                        )}
                      </td>
                    );
                  const d = division(load.division_id);
                  const full = hourIsFull(load);
                  const n = load.appointments.length;
                  const names = load.appointments
                    .map((a) => a.patient?.name)
                    .join(", ");
                  return (
                    <td
                      key={h}
                      className={cn(
                        "p-0.5",
                        h === nowHour && "bg-primary-50/40",
                      )}
                    >
                      <button
                        type="button"
                        onClick={() => onHour(th, load)}
                        title={`${d?.name ?? ""} · ${load.credits}/${load.credit_limit} ${t("credits_short")}${names ? `\n${names}` : ""}`}
                        className={cn(
                          "relative flex h-11 w-full flex-col items-center justify-center overflow-hidden rounded-md border transition hover:brightness-95",
                          d
                            ? DIVISION_COLORS[d.color].cell
                            : "border-gray-200 bg-gray-50",
                          full && "ring-2 ring-rose-400",
                        )}
                      >
                        <span
                          className={cn(
                            "text-sm font-semibold",
                            n === 0
                              ? "text-gray-300"
                              : full
                                ? "text-rose-700"
                                : "text-gray-900",
                          )}
                        >
                          {n || "·"}
                        </span>
                        <span className="absolute inset-x-1 bottom-1 h-1 rounded-full bg-white/70">
                          <span
                            className={cn(
                              "block h-1 rounded-full",
                              d ? DIVISION_COLORS[d.color].dot : "bg-gray-400",
                            )}
                            style={{
                              width: `${Math.min(100, (load.credits / Math.max(load.credit_limit, 1)) * 100)}%`,
                            }}
                          />
                        </span>
                      </button>
                    </td>
                  );
                })}
              </tr>
            );
          })}
        </tbody>
      </table>
      <Legend board={board} />
    </div>
  );
}

function Legend({ board }: { board: Board }) {
  const { t } = useTranslation();
  return (
    <div className="flex flex-wrap items-center gap-x-4 gap-y-1 border-t border-gray-100 px-3 py-2 text-xs text-gray-500">
      {board.divisions.map((d) => (
        <span key={d.id} className="flex items-center gap-1.5">
          <span
            className={cn(
              "size-2.5 rounded-full",
              DIVISION_COLORS[d.color].dot,
            )}
          />
          {d.name}
        </span>
      ))}
      <span>{t("board_legend")}</span>
    </div>
  );
}

/** Rows = divisions grouped by floor; cells = who works there that hour. */
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
  const hours = boardHours(board);
  return (
    <div className="overflow-x-auto rounded-xl border border-gray-200 bg-white shadow-sm">
      <table className="w-full border-collapse text-sm">
        <HourHeader board={board} nowHour={nowHour} first={t("division")} />
        {floorsOf(board.divisions).map((floor) => (
          <tbody key={floor} className="divide-y divide-gray-100">
            <tr className="bg-gray-50">
              <th
                colSpan={hours.length + 1}
                className="px-3 py-1.5 text-left text-xs font-semibold uppercase tracking-wide text-gray-500"
              >
                {t("floor_n", { n: floor })}
              </th>
            </tr>
            {board.divisions
              .filter((d) => (d.floor ?? 1) === floor)
              .map((d) => (
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
                  </th>
                  {hours.map((h) => (
                    <td
                      key={h}
                      className={cn(
                        "p-0.5 align-top",
                        h === nowHour && "bg-primary-50/40",
                      )}
                    >
                      <div className="flex min-h-11 flex-col gap-0.5">
                        {(staffing.get(d.id)?.get(h) ?? []).map(
                          ({ therapist, load }) => (
                            <button
                              key={therapist.id}
                              type="button"
                              onClick={() => onHour(therapist, load)}
                              title={load.appointments
                                .map((a) => a.patient?.name)
                                .join(", ")}
                              className={cn(
                                "flex items-center justify-between gap-1 rounded px-1 py-0.5 text-[11px]",
                                DIVISION_COLORS[d.color].chip,
                                hourIsFull(load) && "ring-2 ring-rose-400",
                              )}
                            >
                              <span className="truncate">
                                {therapist.first_name}
                              </span>
                              <span className="font-semibold">
                                {load.appointments.length}
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
        ))}
      </table>
    </div>
  );
}
