import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { useTranslation } from "react-i18next";

import { cn } from "@/lib/utils";

import { Button } from "@/components/ui/button";

import SuggestionRoutes, {
  SUGGESTION_KINDS,
  SuggestionKind,
  SuggestionStatus,
} from "@/components/Settings/api";

import useDateFormat from "@/hooks/useDateFormat";

import * as Notification from "@/Utils/Notifications";
import query, { callApi } from "@/Utils/request/query";

export const STATUS_STYLES: Record<SuggestionStatus, string> = {
  new: "bg-gray-100 text-gray-700",
  read: "bg-sky-100 text-sky-800",
  planned: "bg-amber-100 text-amber-800",
  done: "bg-emerald-100 text-emerald-800",
  declined: "bg-rose-100 text-rose-800",
};

/** Send an idea or a problem; see what happened to the ones already sent. */
export default function SuggestionsCard() {
  const { t } = useTranslation();
  const { fmt } = useDateFormat();
  const queryClient = useQueryClient();
  const [kind, setKind] = useState<SuggestionKind>("idea");
  const [text, setText] = useState("");
  const { data: mine } = useQuery({
    queryKey: ["suggestions", "mine"],
    queryFn: query(SuggestionRoutes.mine),
  });
  const send = useMutation({
    mutationFn: () =>
      callApi(SuggestionRoutes.send, {
        body: { kind, text, page: "" },
      }),
    onSuccess: () => {
      setText("");
      Notification.Success({ msg: t("suggestion_sent") });
      queryClient.invalidateQueries({ queryKey: ["suggestions"] });
    },
  });

  return (
    <section
      aria-label={t("suggestions")}
      className="rounded-xl border border-gray-200 bg-white p-5 shadow-sm"
    >
      <h2 className="text-base font-semibold text-gray-900">
        {t("suggestions")}
      </h2>
      <p className="mb-3 text-sm text-gray-600">{t("suggestions_help")}</p>
      <form
        className="flex flex-col gap-3"
        onSubmit={(e) => {
          e.preventDefault();
          if (text.trim().length >= 5) send.mutate();
        }}
      >
        <div
          role="radiogroup"
          aria-label={t("suggestion_kind")}
          className="flex flex-wrap gap-2"
        >
          {SUGGESTION_KINDS.map((k) => (
            <button
              key={k}
              type="button"
              role="radio"
              aria-checked={kind === k}
              onClick={() => setKind(k)}
              className={cn(
                "rounded-full px-3 py-1 text-sm ring-1 ring-inset",
                kind === k
                  ? "bg-primary-600 text-white ring-primary-600"
                  : "bg-white text-gray-700 ring-gray-300",
              )}
            >
              {t(`suggestion_kind__${k}`)}
            </button>
          ))}
        </div>
        <label className="text-sm font-medium text-gray-700">
          {t("suggestion_text")}
          <textarea
            value={text}
            maxLength={2000}
            rows={4}
            onChange={(e) => setText(e.target.value)}
            placeholder={t("suggestion_placeholder")}
            className="mt-1 w-full rounded-lg border border-gray-300 p-2.5 text-sm shadow-sm focus:border-primary-500 focus:ring-primary-500"
          />
        </label>
        <div className="flex justify-end">
          <Button
            type="submit"
            variant="primary"
            disabled={text.trim().length < 5 || send.isPending}
          >
            {t("send_suggestion")}
          </Button>
        </div>
      </form>

      {!!mine?.length && (
        <div className="mt-4">
          <h3 className="mb-2 text-sm font-semibold text-gray-800">
            {t("my_suggestions")}
          </h3>
          <ul className="space-y-2">
            {mine.map((s) => (
              <li
                key={s.id}
                className="rounded-lg border border-gray-100 p-3 text-sm"
              >
                <div className="mb-1 flex flex-wrap items-center gap-2 text-xs text-gray-500">
                  <span>{fmt(s.created_date)}</span>
                  <span>· {t(`suggestion_kind__${s.kind}`)}</span>
                  <span
                    className={cn(
                      "rounded-full px-2 py-0.5 font-medium",
                      STATUS_STYLES[s.status],
                    )}
                  >
                    {t(`suggestion_status__${s.status}`)}
                  </span>
                </div>
                <p className="whitespace-pre-wrap text-gray-800">{s.text}</p>
                {s.admin_note && (
                  <p className="mt-2 rounded bg-gray-50 p-2 text-xs text-gray-700">
                    <span className="font-medium">
                      {t("suggestion_answer")}:{" "}
                    </span>
                    {s.admin_note}
                  </p>
                )}
              </li>
            ))}
          </ul>
        </div>
      )}
    </section>
  );
}
