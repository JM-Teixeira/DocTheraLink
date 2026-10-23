import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import dayjs from "dayjs";
import {
  CircleHelp,
  CornerDownRight,
  Info,
  Reply,
  Send,
  X,
} from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";

import { cn } from "@/lib/utils";

import { Button } from "@/components/ui/button";

import { chipClass } from "@/components/Common/FilterChips";
import ReferralRoutes from "@/components/Referral/api";
import {
  CurrentUser,
  canPostNote,
  fullName,
  initials,
} from "@/components/Referral/referralUtils";
import {
  ThreadFilter,
  filterThread,
  painSeries,
  sparkPoints,
  waitingFor,
} from "@/components/Referral/threadUtils";
import { Referral, ReferralNote } from "@/components/Referral/types";

import * as Notification from "@/Utils/Notifications";
import mutate from "@/Utils/request/mutate";
import query from "@/Utils/request/query";

type PostKind = "progress" | "question";
const FILTERS: ThreadFilter[] = ["all", "progress", "questions"];

/**
 * The shared doctor ↔ therapist conversation on a referral:
 *   - Progress: the physio's update after sessions, optionally with the
 *     pain score (0-10), drawn as the patient's evolution;
 *   - Question: either side asks; it stays "waiting" (badge on the referral
 *     card) until the other side answers it with "Reply";
 *   - status changes appear as system entries, so the thread is also the
 *     referral's history.
 */
export default function ReferralNotes({
  referral,
  user,
}: {
  referral: Referral;
  user: CurrentUser;
}) {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [body, setBody] = useState("");
  const [kind, setKind] = useState<PostKind>("progress");
  const [pain, setPain] = useState("");
  const [replyTo, setReplyTo] = useState<ReferralNote | null>(null);
  const [filter, setFilter] = useState<ThreadFilter>("all");

  const { data: notes, isLoading } = useQuery({
    queryKey: ["referral-notes", referral.id],
    queryFn: query(ReferralRoutes.listNotes, {
      pathParams: { id: referral.id },
    }),
  });

  const { mutate: post, isPending } = useMutation({
    mutationFn: mutate(ReferralRoutes.addNote, {
      pathParams: { id: referral.id },
    }),
    onSuccess: () => {
      setBody("");
      setPain("");
      setReplyTo(null);
      queryClient.invalidateQueries({
        queryKey: ["referral-notes", referral.id],
      });
      queryClient.invalidateQueries({ queryKey: ["referrals"] });
    },
    onError: () => Notification.Error({ msg: t("note_post_failed") }),
  });
  const send = () => {
    const text = body.trim();
    if (!text) return;
    if (replyTo) post({ kind: "reply", body: text, reply_to: replyTo.id });
    else
      post({
        kind,
        body: text,
        ...(kind === "progress" && pain !== "" ? { pain: Number(pain) } : {}),
      });
  };

  const all = notes ?? [];
  const byId = new Map(all.map((n) => [n.id, n]));
  const shown = filterThread(all, filter);
  const series = painSeries(all);
  const waiting = waitingFor(all, user.id);
  const canPost = canPostNote(referral, user);

  return (
    <section className="flex flex-col rounded-xl border border-gray-200 bg-white shadow-sm">
      <header className="flex flex-col gap-3 border-b border-gray-100 px-5 py-4">
        <div>
          <h2 className="text-base font-semibold text-gray-900">
            {t("referral_thread")}
          </h2>
          <p className="text-sm text-gray-500">
            {t("referral_thread_explained")}
          </p>
        </div>
        {waiting.length > 0 && (
          <p className="flex items-center gap-2 rounded-lg bg-amber-50 px-3 py-2 text-sm text-amber-900">
            <CircleHelp className="size-4 shrink-0" aria-hidden />
            {t("questions_waiting_for_you", { count: waiting.length })}
          </p>
        )}
        {series.length >= 2 && <PainTrend series={series} />}
        <div
          role="group"
          aria-label={t("show_messages")}
          className="flex flex-wrap gap-1.5"
        >
          {FILTERS.map((f) => (
            <button
              key={f}
              type="button"
              aria-pressed={filter === f}
              className={chipClass(filter === f)}
              onClick={() => setFilter(f)}
            >
              {t(`thread_filter__${f}`)}
            </button>
          ))}
        </div>
      </header>

      <ol className="flex max-h-[32rem] flex-col gap-4 overflow-y-auto px-5 py-4">
        {isLoading && <li className="text-sm text-gray-500">{t("loading")}</li>}
        {!isLoading && shown.length === 0 && (
          <li className="py-6 text-center text-sm text-gray-500">
            {t("no_notes_yet")}
          </li>
        )}
        {shown.map((n) => (
          <NoteItem
            key={n.id}
            note={n}
            mine={n.author_id === user.id}
            question={n.reply_to_id ? byId.get(n.reply_to_id) : undefined}
            onReply={
              canPost && n.kind === "question" && n.author_id !== user.id
                ? () => setReplyTo(n)
                : undefined
            }
          />
        ))}
      </ol>

      {canPost && (
        <form
          className="border-t border-gray-100 p-4"
          onSubmit={(e) => {
            e.preventDefault();
            send();
          }}
        >
          {replyTo ? (
            <div className="mb-2 flex items-start gap-2 rounded-lg bg-primary-50 px-3 py-2 text-sm text-primary-900">
              <Reply className="mt-0.5 size-4 shrink-0" aria-hidden />
              <span className="flex-1">
                <span className="block text-xs font-medium">
                  {t("replying_to", { name: fullName(replyTo.author) })}
                </span>
                <span className="line-clamp-2">{replyTo.body}</span>
              </span>
              <Button
                type="button"
                size="icon"
                variant="ghost"
                aria-label={t("cancel")}
                onClick={() => setReplyTo(null)}
              >
                <X className="size-4" aria-hidden />
              </Button>
            </div>
          ) : (
            <div
              className="mb-2 flex flex-wrap items-center gap-1.5"
              role="radiogroup"
              aria-label={t("note_type")}
            >
              {(["progress", "question"] as const).map((k) => (
                <button
                  type="button"
                  key={k}
                  role="radio"
                  aria-checked={kind === k}
                  onClick={() => setKind(k)}
                  title={t(`note_kind_help__${k}`)}
                  className={cn(
                    "rounded-full px-3 py-1 text-xs font-medium transition",
                    kind === k
                      ? "bg-primary-100 text-primary-800 ring-1 ring-primary-300"
                      : "text-gray-600 hover:bg-gray-100",
                  )}
                >
                  {t(`note_kind__${k}`)}
                </button>
              ))}
              {kind === "progress" && (
                <label className="ml-auto flex items-center gap-1.5 text-xs text-gray-600">
                  {t("pain_0_10")}
                  <select
                    value={pain}
                    onChange={(e) => setPain(e.target.value)}
                    className="h-8 rounded-md border border-gray-300 px-2 text-xs"
                  >
                    <option value="">-</option>
                    {Array.from({ length: 11 }, (_, i) => (
                      <option key={i} value={i}>
                        {i}
                      </option>
                    ))}
                  </select>
                </label>
              )}
            </div>
          )}
          <div className="flex items-end gap-2">
            <label htmlFor="note-body" className="sr-only">
              {t("message")}
            </label>
            <textarea
              id="note-body"
              rows={2}
              value={body}
              onChange={(e) => setBody(e.target.value)}
              onKeyDown={(e) => {
                // Ctrl/Cmd+Enter sends
                if (e.key === "Enter" && (e.ctrlKey || e.metaKey)) send();
              }}
              placeholder={
                replyTo
                  ? t("reply_placeholder")
                  : t(`note_placeholder__${kind}`)
              }
              className="min-h-[2.75rem] flex-1 resize-y rounded-lg border border-gray-300 p-2.5 text-sm shadow-sm focus:border-primary-500 focus:ring-primary-500"
            />
            <Button
              type="submit"
              variant="primary"
              size="icon"
              disabled={!body.trim() || isPending}
              aria-label={t("send")}
            >
              <Send className="size-4" aria-hidden />
            </Button>
          </div>
        </form>
      )}
    </section>
  );
}

/** Pain over the treatment, from the progress notes (0-10). */
function PainTrend({ series }: { series: { date: string; pain: number }[] }) {
  const { t } = useTranslation();
  const first = series[0].pain;
  const last = series[series.length - 1].pain;
  const W = 160;
  const H = 32;
  return (
    <figure className="flex items-center gap-3 rounded-lg bg-gray-50 px-3 py-2">
      <svg
        width={W}
        height={H + 4}
        viewBox={`0 -2 ${W} ${H + 4}`}
        role="img"
        aria-label={t("pain_trend_label", { first, last })}
      >
        <polyline
          points={sparkPoints(
            series.map((s) => s.pain),
            W,
            H,
          )}
          fill="none"
          stroke="currentColor"
          strokeWidth={2}
          className="text-primary-600"
        />
      </svg>
      <figcaption className="text-xs text-gray-700">
        <span className="block font-medium">{t("pain_trend")}</span>
        {t("pain_from_to", { first, last })}
      </figcaption>
    </figure>
  );
}

function NoteItem({
  note,
  mine,
  question,
  onReply,
}: {
  note: ReferralNote;
  mine: boolean;
  /** For replies: the question they answer. */
  question?: ReferralNote;
  onReply?: () => void;
}) {
  const { t } = useTranslation();
  const time = (
    <time
      dateTime={note.created_date}
      title={dayjs(note.created_date).format("LLL")}
    >
      {dayjs(note.created_date).format("D MMM, HH:mm")}
    </time>
  );

  if (note.kind === "system") {
    return (
      <li className="flex items-center justify-center gap-2 text-xs text-gray-500">
        <Info className="size-3.5" aria-hidden />
        <span>
          {note.body} · {fullName(note.author)} · {time}
        </span>
      </li>
    );
  }

  const name = fullName(note.author);
  return (
    <li className={cn("flex gap-2.5", mine && "flex-row-reverse")}>
      <span
        className="flex size-8 shrink-0 items-center justify-center rounded-full bg-gray-100 text-xs font-semibold text-gray-600"
        aria-hidden
      >
        {initials(name)}
      </span>
      <div className={cn("max-w-[80%]", mine && "text-right")}>
        <p
          className="mb-1 flex flex-wrap items-center gap-x-1.5 text-xs text-gray-500"
          style={{ justifyContent: mine ? "flex-end" : undefined }}
        >
          <span className="font-medium text-gray-700">{name}</span>·{" "}
          {t(`note_kind__${note.kind}`)} · {time}
          {note.kind === "question" && (
            <span
              className={cn(
                "rounded-full px-1.5 font-medium",
                note.answered
                  ? "bg-emerald-50 text-emerald-700"
                  : "bg-amber-50 text-amber-800",
              )}
            >
              {note.answered ? t("question_answered") : t("question_waiting")}
            </span>
          )}
          {typeof note.pain === "number" && (
            <span className="rounded-full bg-rose-50 px-1.5 font-medium text-rose-700">
              {t("pain_n", { n: note.pain })}
            </span>
          )}
        </p>
        {question && (
          <p className="mb-1 flex items-start gap-1 text-left text-xs text-gray-500">
            <CornerDownRight className="mt-0.5 size-3 shrink-0" aria-hidden />
            <span className="line-clamp-1">
              {t("in_reply_to", { text: question.body })}
            </span>
          </p>
        )}
        <div
          className={cn(
            "whitespace-pre-wrap rounded-2xl px-4 py-2.5 text-left text-sm",
            mine
              ? "rounded-tr-sm bg-primary-600 text-white"
              : "rounded-tl-sm bg-gray-100 text-gray-900",
            note.kind === "question" &&
              !mine &&
              !note.answered &&
              "ring-2 ring-amber-300",
          )}
        >
          {note.body}
        </div>
        {onReply && (
          <Button
            type="button"
            size="sm"
            variant="ghost"
            className="mt-1 h-7 px-2 text-xs"
            onClick={onReply}
          >
            <Reply className="mr-1 size-3.5" aria-hidden />
            {t("reply_action")}
          </Button>
        )}
      </div>
    </li>
  );
}
