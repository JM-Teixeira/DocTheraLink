import {
  CalendarCheck,
  CalendarClock,
  CalendarX,
  FileText,
  MessageSquare,
  Stethoscope,
  UserPlus,
} from "lucide-react";
import { Link } from "raviger";
import { useState } from "react";
import { useTranslation } from "react-i18next";

import { cn } from "@/lib/utils";

import { Button } from "@/components/ui/button";

import { usePatientTimeline } from "@/components/Patient/api";
import { TimelineEntry } from "@/components/Patient/api/routes";
import {
  TIMELINE_FILTERS,
  TimelineFilter,
  countKinds,
  filterEntries,
  groupByMonth,
  titleKey,
} from "@/components/Patient/timelineUtils";
import { today } from "@/components/TimeBoard/boardUtils";

import useDateFormat from "@/hooks/useDateFormat";

const PAGE = 30;

/** Everything that happened to a patient, newest first, with the next sessions on top. */
export default function PatientTimeline({ patientId }: { patientId: string }) {
  const { t, i18n } = useTranslation();
  const { data, isLoading, isError } = usePatientTimeline(patientId);
  const [filter, setFilter] = useState<TimelineFilter>("all");
  const [shown, setShown] = useState(PAGE);
  const [allUpcoming, setAllUpcoming] = useState(false);

  const monthName = (m: string) =>
    new Intl.DateTimeFormat(i18n.language, {
      month: "long",
      year: "numeric",
    }).format(new Date(Number(m.slice(0, 4)), Number(m.slice(5, 7)) - 1, 1));

  if (isLoading) return <Box>{t("loading")}</Box>;
  if (isError || !data) return <Box>{t("timeline_failed")}</Box>;

  const counts = countKinds(data.history);
  const entries = filterEntries(data.history, filter);
  const upcoming = allUpcoming ? data.upcoming : data.upcoming.slice(0, 3);

  return (
    <section aria-labelledby="timeline-title">
      <h3
        id="timeline-title"
        className="mb-3 text-base font-semibold text-gray-900"
      >
        {t("patient_timeline")}
      </h3>

      {data.upcoming.length > 0 && (
        <div className="mb-4 rounded-xl border border-primary-200 bg-primary-50/50 p-4">
          <h4 className="mb-2 text-sm font-semibold text-primary-900">
            {t("upcoming_sessions", { count: data.upcoming.length })}
          </h4>
          <ul className="space-y-1 text-sm">
            {upcoming.map((e) => (
              <li key={e.item_id} className="flex items-center gap-2">
                <CalendarClock
                  className="size-4 text-primary-600"
                  aria-hidden
                />
                <SessionWhen e={e} />
                {e.who && <span className="text-gray-500">· {e.who}</span>}
              </li>
            ))}
          </ul>
          {data.upcoming.length > 3 && (
            <Button
              variant="link"
              size="sm"
              className="mt-1 h-auto p-0"
              onClick={() => setAllUpcoming((v) => !v)}
            >
              {allUpcoming ? t("show_less") : t("show_all")}
            </Button>
          )}
        </div>
      )}

      <div
        className="mb-3 flex flex-wrap gap-2"
        role="group"
        aria-label={t("filter")}
      >
        {TIMELINE_FILTERS.filter((f) => f === "all" || counts[f] > 0).map(
          (f) => (
            <button
              key={f}
              type="button"
              aria-pressed={filter === f}
              onClick={() => {
                setFilter(f);
                setShown(PAGE);
              }}
              className={cn(
                "rounded-full px-3 py-1 text-xs font-medium ring-1 ring-inset transition",
                filter === f
                  ? "bg-primary-600 text-white ring-primary-600"
                  : "bg-white text-gray-700 ring-gray-300 hover:bg-gray-50",
              )}
            >
              {t(`timeline_filter__${f}`)} ({counts[f]})
            </button>
          ),
        )}
      </div>

      <div className="rounded-xl border border-gray-200 bg-white p-5 shadow-sm">
        {groupByMonth(entries.slice(0, shown)).map((g) => (
          <div key={g.month} className="mb-4 last:mb-0">
            <h4 className="mb-2 text-xs font-semibold uppercase tracking-wide text-gray-500">
              {monthName(g.month)}
            </h4>
            <ol className="relative ml-3 border-l border-gray-200">
              {g.entries.map((e, i) => (
                <Item key={`${e.kind}-${e.item_id ?? e.at}-${i}`} e={e} />
              ))}
            </ol>
          </div>
        ))}
        {entries.length > shown && (
          <Button
            variant="outline"
            size="sm"
            onClick={() => setShown((n) => n + PAGE)}
          >
            {t("show_more")}
          </Button>
        )}
      </div>
    </section>
  );
}

const ICONS = {
  registered: UserPlus,
  referral: Stethoscope,
  note: MessageSquare,
  file: FileText,
  session: CalendarCheck,
};

function Item({ e }: { e: TimelineEntry }) {
  const { t } = useTranslation();
  const { fmtTime } = useDateFormat();
  const missed = e.kind === "session" && e.status === "missed";
  const Icon = missed ? CalendarX : ICONS[e.kind];
  const tone = missed
    ? "bg-rose-100 text-rose-700"
    : e.kind === "session"
      ? "bg-emerald-100 text-emerald-700"
      : e.kind === "referral"
        ? "bg-primary-100 text-primary-700"
        : "bg-gray-100 text-gray-600";
  const title = t(titleKey(e, today()));

  return (
    <li className="mb-3 ml-5 last:mb-0">
      <span
        className={cn(
          "absolute -left-3 flex size-6 items-center justify-center rounded-full ring-4 ring-white",
          tone,
        )}
      >
        <Icon className="size-3.5" aria-hidden />
      </span>
      <p className="text-sm font-medium text-gray-900">
        {e.referral_id ? (
          <Link
            href={`/referrals/${e.referral_id}`}
            className="hover:underline"
          >
            {title}
          </Link>
        ) : (
          title
        )}
        {missed && e.justified !== undefined && (
          <span className="ml-1 font-normal text-gray-500">
            ({e.justified ? t("miss_justified") : t("miss_unjustified")})
          </span>
        )}
        {e.pain !== undefined && (
          <span className="ml-2 rounded bg-amber-100 px-1.5 text-xs text-amber-800">
            {t("pain_n", { n: e.pain })}
          </span>
        )}
      </p>
      <p className="text-xs text-gray-500">
        {e.kind === "session" ? <SessionWhen e={e} /> : fmtTime(e.at)}
        {e.who && ` · ${e.who}`}
      </p>
      {e.text && (
        <p className="mt-0.5 whitespace-pre-wrap text-sm text-gray-700">
          {e.text}
        </p>
      )}
    </li>
  );
}

function SessionWhen({ e }: { e: TimelineEntry }) {
  const { fmt } = useDateFormat();
  return (
    <span>
      {fmt(e.date)} {String(e.hour ?? 0).padStart(2, "0")}:00
    </span>
  );
}

function Box({ children }: { children: React.ReactNode }) {
  return (
    <p className="rounded-xl border-2 border-dashed border-gray-200 bg-white p-6 text-center text-sm text-gray-500">
      {children}
    </p>
  );
}
