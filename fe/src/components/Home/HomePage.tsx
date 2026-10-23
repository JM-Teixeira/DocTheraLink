import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import dayjs from "dayjs";
import {
  Bell,
  CalendarCheck,
  CalendarClock,
  CalendarDays,
  Check,
  ClipboardList,
  Inbox,
  ListOrdered,
  Package,
  PackageX,
  Plane,
  Stethoscope,
  TriangleAlert,
  UserPlus,
  UserX,
  Users,
} from "lucide-react";
import { navigate } from "raviger";
import { useState } from "react";
import { useTranslation } from "react-i18next";

import { cn } from "@/lib/utils";

import { Button } from "@/components/ui/button";

import Page from "@/components/Common/Page";
import {
  arrivals,
  clinicLoad,
  greetingKey,
  myDay,
  nextSession,
} from "@/components/Home/homeUtils";
import InventoryRoutes from "@/components/Inventory/api";
import NotificationRoutes from "@/components/Notifications/api";
import {
  notificationLink,
  notificationText,
} from "@/components/Notifications/notificationUtils";
import ReferralStatusBadge from "@/components/Referral/ReferralStatusBadge";
import ReferralRoutes from "@/components/Referral/api";
import {
  canCreateReferral,
  sumCounts,
} from "@/components/Referral/referralUtils";
import MissDialog, { statusLabel } from "@/components/TimeBoard/MissDialog";
import PatientQuickView from "@/components/TimeBoard/PatientQuickView";
import BoardRoutes from "@/components/TimeBoard/api";
import {
  CARE_LEVEL_STYLES,
  DIVISION_COLORS,
  coveringFor,
  formatHour,
  today,
  utilisation,
} from "@/components/TimeBoard/boardUtils";
import {
  Absence,
  AppointmentStatus,
  Board,
} from "@/components/TimeBoard/types";

import useAuthUser from "@/hooks/useAuthUser";

import * as Notification from "@/Utils/Notifications";
import query, { callApi } from "@/Utils/request/query";
import { isAdmin, isCoordinator, isDoctor, isTherapist } from "@/Utils/roles";

/**
 * Home page (logo / "Home"): a greeting and what matters today for the
 * signed-in user's role — a physio's own day, the clinic's load and what
 * needs attention for coordinators, a doctor's referrals, today's arrivals
 * for the front desk — plus the latest notifications.
 */
export default function HomePage() {
  const { t } = useTranslation();
  const user = useAuthUser();
  const [patientId, setPatientId] = useState<string | null>(null);
  const physio = isTherapist(user);
  const manager = isCoordinator(user) || isAdmin(user);
  const doctor = isDoctor(user);
  const date = today();

  const { data: board } = useQuery({
    queryKey: ["board", date],
    queryFn: query(BoardRoutes.board, { queryParams: { date }, silent: true }),
    retry: false,
    refetchInterval: 60_000,
  });
  const { data: counts } = useQuery({
    queryKey: ["referral-counts", "home"],
    queryFn: query(ReferralRoutes.counts, { silent: true }),
    enabled: physio || doctor || manager,
    retry: false,
  });
  const { data: stock } = useQuery({
    queryKey: ["stock-summary"],
    queryFn: query(InventoryRoutes.summary, { silent: true }),
    retry: false,
  });

  const now = dayjs();
  const clinic = user.home_facility_object?.name;
  const load = board ? clinicLoad(board) : undefined;
  const mine = board && physio ? myDay(board, user.id) : [];
  const next = nextSession(mine, now.hour());
  const minePatients = mine.reduce((n, h) => n + h.appointments.length, 0);

  return (
    <Page title={t("home")} hideTitleOnPage help="home">
      {/* Greeting */}
      <section className="mb-6 overflow-hidden rounded-2xl bg-gradient-to-r from-primary-700 to-primary-500 p-6 text-white shadow-sm">
        <p className="text-sm capitalize text-white/80">
          {now.format("dddd, D MMMM YYYY")}
        </p>
        <h1 className="mt-1 text-2xl font-semibold">
          {t(greetingKey(now.hour()))}, {user.first_name}
        </h1>
        {clinic && <p className="mt-1 text-sm text-white/80">{clinic}</p>}
        <div className="mt-4 flex flex-wrap gap-2">
          {canCreateReferral(user) && (
            <QuickAction
              icon={Stethoscope}
              label={t("new_referral")}
              to="/referrals/new"
            />
          )}
          <QuickAction
            icon={UserPlus}
            label={t("register_patient")}
            to="/patients/new"
          />
          <QuickAction
            icon={CalendarDays}
            label={t("time_board")}
            to="/time-board"
          />
          {(physio || doctor) && (
            <QuickAction
              icon={Inbox}
              label={physio ? t("referral_inbox") : t("referrals")}
              to="/referrals"
            />
          )}
          <QuickAction icon={Package} label={t("inventory")} to="/inventory" />
        </div>
      </section>

      {/* Key numbers for the role */}
      <div className="mb-6 grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
        {physio && (
          <>
            <Stat
              icon={CalendarClock}
              label={t("my_sessions_today")}
              value={minePatients}
            />
            <Stat
              icon={CalendarClock}
              label={t("next_session")}
              value={next ? formatHour(next.hour) : "-"}
              hint={
                next
                  ? next.appointments.map((a) => a.patient?.name).join(", ")
                  : t("no_more_today")
              }
            />
            <Stat
              icon={Inbox}
              label={t("referrals_waiting")}
              value={counts?.pending ?? 0}
              to="/referrals"
            />
            <Stat
              icon={ClipboardList}
              label={t("in_treatment")}
              value={counts?.in_progress ?? 0}
              to="/referrals"
            />
          </>
        )}
        {doctor && (
          <>
            <Stat
              icon={Inbox}
              label={t("referral_status__pending")}
              value={counts?.pending ?? 0}
              to="/referrals"
            />
            <Stat
              icon={ClipboardList}
              label={t("in_treatment")}
              value={(counts?.in_progress ?? 0) + (counts?.accepted ?? 0)}
              to="/referrals"
            />
            <Stat
              icon={Check}
              label={t("referral_status__completed")}
              value={counts?.completed ?? 0}
              to="/referrals"
            />
            <Stat
              icon={Stethoscope}
              label={t("all_referrals")}
              value={counts ? sumCounts(counts) : 0}
              to="/referrals"
            />
          </>
        )}
        {!physio && !doctor && (
          <>
            <Stat
              icon={Users}
              label={t("physios_working_today")}
              value={load?.physios ?? 0}
              to="/time-board"
            />
            <Stat
              icon={CalendarClock}
              label={t("sessions_today")}
              value={load?.patients ?? 0}
              hint={
                load ? t("n_attended", { count: load.attended }) : undefined
              }
              to="/time-board"
            />
            <Stat
              icon={CalendarDays}
              label={t("clinic_load")}
              value={`${load?.percent ?? 0}%`}
              hint={
                load
                  ? t("credits_used", {
                      count: load.credits,
                      limit: load.capacity,
                    })
                  : undefined
              }
            />
            <Stat
              icon={PackageX}
              label={t("stock_attention")}
              value={
                (stock?.low ?? 0) +
                (stock?.out ?? 0) +
                (stock?.open_expired ?? 0)
              }
              tone={
                (stock?.low ?? 0) + (stock?.out ?? 0) > 0 ? "warn" : undefined
              }
              to="/inventory"
            />
          </>
        )}
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        <div className="flex flex-col gap-6 lg:col-span-2">
          {physio &&
            board &&
            coveringFor(board.absences, user.id).map((a) => (
              <CoverCard key={a.id} absence={a} onPatient={setPatientId} />
            ))}
          {physio && board && (
            <MyDay board={board} userId={user.id} onPatient={setPatientId} />
          )}
          {manager && board && <ClinicToday board={board} />}
          {manager && <NeedsAttention pending={counts?.pending ?? 0} />}
          {doctor && <MyReferrals />}
          {!physio && !doctor && board && (
            <Arrivals board={board} onPatient={setPatientId} />
          )}
        </div>
        <LatestNotifications />
      </div>

      <PatientQuickView
        patientId={patientId}
        onClose={() => setPatientId(null)}
      />
    </Page>
  );
}

function QuickAction({
  icon: Icon,
  label,
  to,
}: {
  icon: typeof Bell;
  label: string;
  to: string;
}) {
  return (
    <button
      type="button"
      onClick={() => navigate(to)}
      className="flex items-center gap-2 rounded-full bg-white/15 px-3 py-1.5 text-sm font-medium text-white ring-1 ring-inset ring-white/30 transition hover:bg-white/25"
    >
      <Icon className="size-4" aria-hidden />
      {label}
    </button>
  );
}

function Stat({
  icon: Icon,
  label,
  value,
  hint,
  tone,
  to,
}: {
  icon: typeof Bell;
  label: string;
  value: number | string;
  hint?: string;
  tone?: "warn";
  to?: string;
}) {
  const body = (
    <>
      <span
        className={cn(
          "flex size-10 shrink-0 items-center justify-center rounded-xl",
          tone === "warn"
            ? "bg-amber-50 text-amber-600"
            : "bg-primary-50 text-primary-600",
        )}
      >
        <Icon className="size-5" aria-hidden />
      </span>
      <span className="min-w-0">
        <span className="block text-xs font-medium text-gray-500">{label}</span>
        <span className="block text-2xl font-semibold text-gray-900">
          {value}
        </span>
        {hint && (
          <span className="block truncate text-xs text-gray-500">{hint}</span>
        )}
      </span>
    </>
  );
  const cls =
    "flex items-center gap-3 rounded-xl border border-gray-200 bg-white p-4 text-left shadow-sm";
  return to ? (
    <button
      type="button"
      onClick={() => navigate(to)}
      className={cn(cls, "transition hover:border-primary-300")}
    >
      {body}
    </button>
  ) : (
    <div className={cls}>{body}</div>
  );
}

function Card({
  title,
  action,
  children,
}: {
  title: string;
  action?: React.ReactNode;
  children: React.ReactNode;
}) {
  return (
    <section className="rounded-xl border border-gray-200 bg-white shadow-sm">
      <header className="flex items-center justify-between border-b border-gray-100 px-5 py-3">
        <h2 className="text-base font-semibold text-gray-900">{title}</h2>
        {action}
      </header>
      <div className="p-5">{children}</div>
    </section>
  );
}

/** A physiotherapist's own hours today, with their patients. */
/** A physio covering a colleague: whose patients, until when, and the handover. */
function CoverCard({
  absence,
  onPatient,
}: {
  absence: Absence;
  onPatient: (id: string) => void;
}) {
  const { t } = useTranslation();
  const who = `${absence.therapist?.first_name ?? ""} ${absence.therapist?.last_name ?? ""}`;
  return (
    <section className="rounded-xl border border-amber-200 bg-amber-50 p-5 shadow-sm">
      <h2 className="flex items-center gap-2 text-base font-semibold text-amber-900">
        <Plane className="size-4" aria-hidden />
        {t("covering_for", {
          name: who,
          date: dayjs(absence.to).format("D MMM"),
        })}
      </h2>
      {absence.handover && (
        <p className="mt-2 whitespace-pre-wrap text-sm text-gray-800">
          {absence.handover}
        </p>
      )}
      {!!absence.notes.length && (
        <ul className="mt-3 flex flex-col gap-2 text-sm">
          {absence.notes.map((n) => (
            <li key={n.id} className="rounded-lg bg-white p-2 shadow-sm">
              <button
                type="button"
                onClick={() => onPatient(n.patient_id)}
                className="font-medium text-primary-700 hover:underline"
              >
                {n.patient?.name}
              </button>
              <p className="text-gray-700">{n.note}</p>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}

function MyDay({
  board,
  userId,
  onPatient,
}: {
  board: Board;
  userId: number;
  onPatient: (id: string) => void;
}) {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const day = myDay(board, userId);
  const [missing, setMissing] = useState<{
    id: string;
    name: string;
    justified?: boolean | null;
    reason?: string;
  } | null>(null);
  const nowHour = dayjs().hour();
  const status = useMutation({
    mutationFn: (v: {
      id: string;
      status: AppointmentStatus;
      justified?: boolean;
      reason?: string;
    }) =>
      callApi(BoardRoutes.setStatus, {
        pathParams: { id: v.id },
        body: { status: v.status, justified: v.justified, reason: v.reason },
        silent: true,
      }),
    onSuccess: (res) => {
      for (const w of res.warnings ?? []) Notification.Warn({ msg: w });
      setMissing(null);
      queryClient.invalidateQueries({ queryKey: ["board"] });
    },
    onError: () => Notification.Error({ msg: t("board_save_failed") }),
  });

  return (
    <Card
      title={t("my_day")}
      action={
        <Button
          size="sm"
          variant="ghost"
          onClick={() => navigate("/time-board")}
        >
          {t("time_board")}
        </Button>
      }
    >
      {!day.length ? (
        <p className="text-sm text-gray-500">{t("not_working_today")}</p>
      ) : (
        <ol className="relative flex flex-col gap-3 border-l-2 border-gray-100 pl-4">
          {day.map((h) => (
            <li key={h.hour} className="relative">
              <span
                className={cn(
                  "absolute -left-[23px] top-1 size-3 rounded-full ring-4 ring-white",
                  h.hour < nowHour
                    ? "bg-gray-300"
                    : h.hour === nowHour
                      ? "bg-primary-600"
                      : "bg-primary-200",
                )}
              />
              <div className="flex flex-wrap items-baseline gap-2">
                <span
                  className={cn(
                    "font-mono text-sm font-semibold",
                    h.hour === nowHour ? "text-primary-700" : "text-gray-700",
                  )}
                >
                  {formatHour(h.hour)}
                </span>
                {h.division && (
                  <span
                    className={cn(
                      "rounded px-1.5 text-xs",
                      DIVISION_COLORS[h.division.color].chip,
                    )}
                  >
                    {h.division.name}
                  </span>
                )}
              </div>
              {h.appointments.length ? (
                <ul className="mt-1 divide-y divide-gray-100 rounded-lg border border-gray-100">
                  {h.appointments.map((a) => (
                    <li
                      key={a.id}
                      className="flex items-center gap-2 px-3 py-1.5 text-sm"
                    >
                      <button
                        type="button"
                        onClick={() => onPatient(a.patient_id)}
                        className="min-w-0 flex-1 truncate text-left font-medium text-primary-700 hover:underline"
                      >
                        {a.patient?.name}
                      </button>
                      <span
                        className={cn(
                          "rounded px-1.5 text-xs font-semibold",
                          CARE_LEVEL_STYLES[a.credits],
                        )}
                      >
                        {a.credits}
                      </span>
                      <span className="w-20 text-right text-xs text-gray-500">
                        {statusLabel(t, a)}
                      </span>
                      <Button
                        size="icon"
                        variant="ghost"
                        aria-label={t("mark_attended")}
                        onClick={() =>
                          status.mutate({ id: a.id, status: "attended" })
                        }
                      >
                        <Check
                          className="size-4 text-emerald-600"
                          aria-hidden
                        />
                      </Button>
                      <Button
                        size="icon"
                        variant="ghost"
                        aria-label={t("mark_missed")}
                        onClick={() =>
                          setMissing({
                            id: a.id,
                            name: a.patient?.name ?? "",
                            justified: a.miss_justified,
                            reason: a.miss_reason,
                          })
                        }
                      >
                        <UserX className="size-4 text-amber-600" aria-hidden />
                      </Button>
                    </li>
                  ))}
                </ul>
              ) : (
                <p className="mt-0.5 text-xs text-gray-400">{t("free")}</p>
              )}
            </li>
          ))}
        </ol>
      )}
      <MissDialog
        patientName={missing?.name ?? null}
        initial={
          missing
            ? { justified: !!missing.justified, reason: missing.reason ?? "" }
            : undefined
        }
        pending={status.isPending}
        onClose={() => setMissing(null)}
        onConfirm={(v) =>
          missing && status.mutate({ id: missing.id, status: "missed", ...v })
        }
      />
    </Card>
  );
}

/** Coordinators: each physio's load today. */
function ClinicToday({ board }: { board: Board }) {
  const { t } = useTranslation();
  const working = board.therapists.filter((x) => x.shifts.length > 0);
  return (
    <Card
      title={t("clinic_today")}
      action={
        <Button
          size="sm"
          variant="ghost"
          onClick={() => navigate("/time-board")}
        >
          {t("time_board")}
        </Button>
      }
    >
      {!working.length ? (
        <p className="text-sm text-gray-500">{t("nobody_working_today")}</p>
      ) : (
        <ul className="flex flex-col gap-3">
          {working.map((th) => {
            const pct = utilisation(th);
            return (
              <li
                key={th.id}
                className="grid grid-cols-[10rem_1fr_auto] items-center gap-3 text-sm"
              >
                <span className="truncate font-medium text-gray-900">
                  {th.first_name} {th.last_name}
                </span>
                <div className="h-2 rounded-full bg-gray-100">
                  <div
                    className={cn(
                      "h-2 rounded-full",
                      pct >= 90
                        ? "bg-rose-500"
                        : pct >= 60
                          ? "bg-amber-500"
                          : "bg-primary-500",
                    )}
                    style={{ width: `${pct}%` }}
                  />
                </div>
                <span className="w-36 text-right text-xs text-gray-500">
                  {t("n_patients", { count: th.patients })} · {pct}%
                </span>
              </li>
            );
          })}
        </ul>
      )}
    </Card>
  );
}

/** Coordinators: stock to reorder, containers open too long, referrals waiting. */
function NeedsAttention({ pending }: { pending: number }) {
  const { t } = useTranslation();
  const { data: reorder } = useQuery({
    queryKey: ["stock-reorder"],
    queryFn: query(InventoryRoutes.reorder, { silent: true }),
    retry: false,
  });
  const { data: items } = useQuery({
    queryKey: ["stock-items", "", "all"],
    queryFn: query(InventoryRoutes.list, { silent: true }),
    retry: false,
  });
  const { data: misses } = useQuery({
    queryKey: ["missed-flags"],
    queryFn: query(BoardRoutes.missed, { silent: true }),
    retry: false,
  });
  const expired = items?.filter((i) => i.open_expired) ?? [];
  const flagged = misses?.patients ?? [];
  const { data: waiting } = useQuery({
    queryKey: ["waiting-list"],
    queryFn: query(BoardRoutes.waiting, { silent: true }),
    retry: false,
  });
  return (
    <Card title={t("needs_attention")}>
      <ul className="flex flex-col gap-2 text-sm">
        {!!waiting?.filter((w) => w.slot).length && (
          <AttentionRow
            icon={CalendarCheck}
            to="/time-board"
            text={t("n_with_free_slot", {
              count: waiting.filter((w) => w.slot).length,
            })}
          />
        )}
        {!!waiting?.length && (
          <AttentionRow
            icon={ListOrdered}
            to="/time-board"
            text={t("n_waiting_for_slot", { count: waiting.length })}
            tone={waiting.some((w) => w.urgent) ? "danger" : "warn"}
          />
        )}
        {flagged.map((m) => (
          <AttentionRow
            key={m.patient_id}
            icon={UserX}
            to={`/patients/${m.patient_id}`}
            text={t("misses_flag_line", {
              name: m.patient?.name ?? "",
              count: m.unjustified,
              days: misses?.window_days ?? 90,
            })}
            tone="danger"
          />
        ))}
        {pending > 0 && (
          <AttentionRow
            icon={Inbox}
            to="/referrals"
            text={t("referrals_waiting_n", { count: pending })}
          />
        )}
        {expired.map((i) => (
          <AttentionRow
            key={i.id}
            icon={TriangleAlert}
            to="/inventory"
            text={t("open_too_long", { name: i.name })}
            tone="danger"
          />
        ))}
        {reorder?.slice(0, 6).map((i) => (
          <AttentionRow
            key={i.id}
            icon={PackageX}
            to="/inventory"
            text={t("reorder_line", {
              name: i.name,
              count: i.reorder_quantity,
              unit: i.unit,
            })}
            tone={i.status === "out" ? "danger" : "warn"}
          />
        ))}
        {!pending &&
          !expired.length &&
          !reorder?.length &&
          !flagged.length &&
          !waiting?.length && (
            <li className="text-gray-500">{t("all_in_order")}</li>
          )}
      </ul>
    </Card>
  );
}

function AttentionRow({
  icon: Icon,
  text,
  to,
  tone,
}: {
  icon: typeof Bell;
  text: string;
  to: string;
  tone?: "warn" | "danger";
}) {
  return (
    <li>
      <button
        type="button"
        onClick={() => navigate(to)}
        className="flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-left hover:bg-gray-50"
      >
        <Icon
          className={cn(
            "size-4 shrink-0",
            tone === "danger"
              ? "text-rose-600"
              : tone === "warn"
                ? "text-amber-600"
                : "text-primary-600",
          )}
          aria-hidden
        />
        <span className="text-gray-800">{text}</span>
      </button>
    </li>
  );
}

/** Doctors: their latest referrals. */
function MyReferrals() {
  const { t } = useTranslation();
  const { data } = useQuery({
    queryKey: ["referrals", "home"],
    queryFn: query(ReferralRoutes.list, {
      queryParams: { limit: 6 },
      silent: true,
    }),
    retry: false,
  });
  return (
    <Card
      title={t("my_referrals")}
      action={
        <Button
          size="sm"
          variant="ghost"
          onClick={() => navigate("/referrals")}
        >
          {t("see_all")}
        </Button>
      }
    >
      {!data?.results.length ? (
        <p className="text-sm text-gray-500">{t("no_referrals_yet")}</p>
      ) : (
        <ul className="divide-y divide-gray-100">
          {data.results.map((r) => (
            <li key={r.id}>
              <button
                type="button"
                onClick={() => navigate(`/referrals/${r.id}`)}
                className="flex w-full items-center gap-3 py-2 text-left text-sm hover:bg-gray-50"
              >
                <span className="min-w-0 flex-1">
                  <span className="block truncate font-medium text-gray-900">
                    {r.patient?.name}
                  </span>
                  <span className="block truncate text-xs text-gray-500">
                    {r.diagnosis}
                  </span>
                </span>
                <ReferralStatusBadge status={r.status} />
              </button>
            </li>
          ))}
        </ul>
      )}
    </Card>
  );
}

/** Front desk: who arrives today, when and with whom. */
function Arrivals({
  board,
  onPatient,
}: {
  board: Board;
  onPatient: (id: string) => void;
}) {
  const { t } = useTranslation();
  const list = arrivals(board);
  const nowHour = dayjs().hour();
  return (
    <Card
      title={t("arrivals_today")}
      action={
        <span className="text-xs text-gray-500">
          {t("n_patients", { count: list.length })}
        </span>
      }
    >
      {!list.length ? (
        <p className="text-sm text-gray-500">{t("no_arrivals_today")}</p>
      ) : (
        <ul className="max-h-[28rem] divide-y divide-gray-100 overflow-auto">
          {list.map((a, i) => (
            <li
              key={`${a.patientId}-${a.hour}-${i}`}
              className={cn(
                "flex items-center gap-3 py-2 text-sm",
                a.hour < nowHour && "opacity-60",
              )}
            >
              <span className="w-12 font-mono font-semibold text-gray-700">
                {formatHour(a.hour)}
              </span>
              <button
                type="button"
                onClick={() => onPatient(a.patientId)}
                className="min-w-0 flex-1 truncate text-left font-medium text-primary-700 hover:underline"
              >
                {a.patient}
              </button>
              <span
                className={cn(
                  "rounded px-1.5 text-xs font-semibold",
                  CARE_LEVEL_STYLES[a.careLevel],
                )}
              >
                {a.careLevel}
              </span>
              <span className="hidden w-36 truncate text-xs text-gray-600 sm:block">
                {a.physio}
              </span>
              {a.division && (
                <span
                  className={cn(
                    "hidden rounded px-1.5 text-xs md:block",
                    DIVISION_COLORS[a.division.color].chip,
                  )}
                >
                  {a.division.name}
                </span>
              )}
              <span className="w-20 text-right text-xs text-gray-500">
                {t(`appointment_status__${a.status}`)}
              </span>
            </li>
          ))}
        </ul>
      )}
    </Card>
  );
}

/** Everyone: the latest notifications. */
function LatestNotifications() {
  const { t } = useTranslation();
  const { data } = useQuery({
    queryKey: ["notifications", "home"],
    queryFn: query(NotificationRoutes.list, {
      queryParams: { limit: 6 },
      silent: true,
    }),
    retry: false,
  });
  return (
    <Card title={t("notifications")}>
      {!data?.results.length ? (
        <p className="text-sm text-gray-500">{t("no_notifications")}</p>
      ) : (
        <ul className="flex flex-col gap-1">
          {data.results.map((n) => {
            const { key, params } = notificationText(n);
            const link = notificationLink(n);
            return (
              <li key={n.id}>
                <button
                  type="button"
                  disabled={!link}
                  onClick={() => link && navigate(link)}
                  className="flex w-full gap-2 rounded-lg px-2 py-2 text-left text-sm hover:bg-gray-50"
                >
                  <span
                    className={cn(
                      "mt-1.5 size-2 shrink-0 rounded-full",
                      n.read_at ? "bg-transparent" : "bg-primary-500",
                    )}
                  />
                  <span className="min-w-0">
                    <span className="block text-gray-800">
                      {t(key, params)}
                    </span>
                    <span className="block text-xs text-gray-500">
                      {dayjs(n.created_date).fromNow()}
                    </span>
                  </span>
                </button>
              </li>
            );
          })}
        </ul>
      )}
    </Card>
  );
}
