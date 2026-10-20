import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import dayjs from "dayjs";
import { Info, Send } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";

import { cn } from "@/lib/utils";

import { Button } from "@/components/ui/button";

import ReferralRoutes from "@/components/Referral/api";
import {
  CurrentUser,
  canPostNote,
  fullName,
  initials,
} from "@/components/Referral/referralUtils";
import {
  NoteRequest,
  Referral,
  ReferralNote,
} from "@/components/Referral/types";

import * as Notification from "@/Utils/Notifications";
import mutate from "@/Utils/request/mutate";
import query from "@/Utils/request/query";

type PostKind = NonNullable<NoteRequest["kind"]>;
const KINDS: PostKind[] = ["progress", "question", "reply"];

/**
 * The shared doctor ↔ therapist conversation on a referral. Status changes
 * appear inline as system entries so the thread doubles as an audit trail.
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
      queryClient.invalidateQueries({
        queryKey: ["referral-notes", referral.id],
      });
    },
    onError: () => Notification.Error({ msg: t("note_post_failed") }),
  });

  return (
    <section className="flex flex-col rounded-xl border border-gray-200 bg-white shadow-sm">
      <header className="border-b border-gray-100 px-5 py-4">
        <h2 className="text-base font-semibold text-gray-900">
          {t("referral_thread")}
        </h2>
        <p className="text-sm text-gray-500">
          {t("referral_thread_description")}
        </p>
      </header>

      <ol className="flex max-h-[32rem] flex-col gap-4 overflow-y-auto px-5 py-4">
        {isLoading && <li className="text-sm text-gray-500">{t("loading")}</li>}
        {notes?.length === 0 && (
          <li className="py-6 text-center text-sm text-gray-500">
            {t("no_notes_yet")}
          </li>
        )}
        {notes?.map((n) => (
          <NoteItem key={n.id} note={n} mine={n.author_id === user.id} />
        ))}
      </ol>

      {canPostNote(referral, user) && (
        <form
          className="border-t border-gray-100 p-4"
          onSubmit={(e) => {
            e.preventDefault();
            if (body.trim()) post({ kind, body: body.trim() });
          }}
        >
          <div
            className="mb-2 flex gap-1.5"
            role="radiogroup"
            aria-label={t("note_type")}
          >
            {KINDS.map((k) => (
              <button
                type="button"
                key={k}
                role="radio"
                aria-checked={kind === k}
                onClick={() => setKind(k)}
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
          </div>
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
                if (
                  e.key === "Enter" &&
                  (e.ctrlKey || e.metaKey) &&
                  body.trim()
                ) {
                  post({ kind, body: body.trim() });
                }
              }}
              placeholder={t("note_placeholder")}
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

function NoteItem({ note, mine }: { note: ReferralNote; mine: boolean }) {
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
        <p className="mb-1 text-xs text-gray-500">
          <span className="font-medium text-gray-700">{name}</span> ·{" "}
          {t(`note_kind__${note.kind}`)} · {time}
        </p>
        <div
          className={cn(
            "whitespace-pre-wrap rounded-2xl px-4 py-2.5 text-left text-sm",
            mine
              ? "rounded-tr-sm bg-primary-600 text-white"
              : "rounded-tl-sm bg-gray-100 text-gray-900",
          )}
        >
          {note.body}
        </div>
      </div>
    </li>
  );
}
