import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { useTranslation } from "react-i18next";

import { cn } from "@/lib/utils";

import { Button } from "@/components/ui/button";

import Page from "@/components/Common/Page";
import { STATUS_STYLES } from "@/components/Settings/SuggestionsCard";
import SuggestionRoutes, {
  SUGGESTION_STATUSES,
  Suggestion,
  SuggestionStatus,
} from "@/components/Settings/api";
import { roleKey } from "@/components/Staff/staffUtils";

import useDateFormat from "@/hooks/useDateFormat";

import query, { callApi } from "@/Utils/request/query";

/**
 * Admins read every suggestion in one place (new ones first), set where it
 * stands and answer; the author sees the status and answer in Settings.
 */
export default function SuggestionsAdmin() {
  const { t } = useTranslation();
  const [status, setStatus] = useState<SuggestionStatus | "">("new");
  const { data, isLoading } = useQuery({
    queryKey: ["suggestions", "all", status],
    queryFn: query(SuggestionRoutes.all, {
      queryParams: status ? { status } : {},
    }),
  });

  return (
    <Page title={t("suggestions")} hideBack help="suggestions">
      <div
        role="group"
        aria-label={t("status")}
        className="mb-4 flex flex-wrap gap-2"
      >
        {(["", ...SUGGESTION_STATUSES] as const).map((s) => (
          <button
            key={s || "all"}
            type="button"
            aria-pressed={status === s}
            onClick={() => setStatus(s)}
            className={cn(
              "rounded-full px-3 py-1 text-sm ring-1 ring-inset",
              status === s
                ? "bg-primary-600 text-white ring-primary-600"
                : "bg-white text-gray-700 ring-gray-300",
            )}
          >
            {s ? t(`suggestion_status__${s}`) : t("all")}
          </button>
        ))}
      </div>
      {isLoading ? (
        <p className="text-sm text-gray-500">{t("loading")}</p>
      ) : !data?.length ? (
        <p className="rounded-xl border-2 border-dashed border-gray-200 bg-white p-6 text-center text-sm text-gray-500">
          {t("no_suggestions")}
        </p>
      ) : (
        <ul className="flex max-w-3xl flex-col gap-3">
          {data.map((s) => (
            <SuggestionItem key={s.id} s={s} />
          ))}
        </ul>
      )}
    </Page>
  );
}

function SuggestionItem({ s }: { s: Suggestion }) {
  const { t } = useTranslation();
  const { fmtTime } = useDateFormat();
  const queryClient = useQueryClient();
  const [note, setNote] = useState(s.admin_note);
  const review = useMutation({
    mutationFn: (body: { status?: SuggestionStatus; admin_note?: string }) =>
      callApi(SuggestionRoutes.review, { pathParams: { id: s.id }, body }),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["suggestions"] }),
  });
  const who = s.author
    ? `${s.author.first_name} ${s.author.last_name} (${t(roleKey(s.author.user_type), { defaultValue: s.author.user_type })})`
    : "";

  return (
    <li className="rounded-xl border border-gray-200 bg-white p-4 shadow-sm">
      <div className="mb-2 flex flex-wrap items-center gap-2 text-xs text-gray-500">
        <span
          className={cn(
            "rounded-full px-2 py-0.5 font-medium",
            STATUS_STYLES[s.status],
          )}
        >
          {t(`suggestion_status__${s.status}`)}
        </span>
        <span>{t(`suggestion_kind__${s.kind}`)}</span>
        <span>· {fmtTime(s.created_date)}</span>
        {who && <span>· {who}</span>}
      </div>
      <p className="whitespace-pre-wrap text-sm text-gray-900">{s.text}</p>
      <div className="mt-3 flex flex-wrap items-end gap-2">
        <label className="text-xs font-medium text-gray-700">
          {t("status")}
          <select
            value={s.status}
            disabled={review.isPending}
            onChange={(e) =>
              review.mutate({ status: e.target.value as SuggestionStatus })
            }
            className="mt-1 block h-9 rounded-md border border-gray-300 px-2 text-sm"
          >
            {SUGGESTION_STATUSES.map((st) => (
              <option key={st} value={st}>
                {t(`suggestion_status__${st}`)}
              </option>
            ))}
          </select>
        </label>
        <label className="min-w-60 flex-1 text-xs font-medium text-gray-700">
          {t("suggestion_answer")}
          <input
            value={note}
            maxLength={1000}
            onChange={(e) => setNote(e.target.value)}
            className="mt-1 block h-9 w-full rounded-md border border-gray-300 px-2 text-sm"
          />
        </label>
        <Button
          variant="outline"
          size="sm"
          disabled={note === s.admin_note || review.isPending}
          onClick={() => review.mutate({ admin_note: note })}
        >
          {t("save")}
        </Button>
      </div>
    </li>
  );
}
