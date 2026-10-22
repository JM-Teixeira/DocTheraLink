import { useMutation, useQuery } from "@tanstack/react-query";
import { Check } from "lucide-react";
import { useTranslation } from "react-i18next";

import { cn } from "@/lib/utils";

import { userChildProps } from "@/components/Common/UserColumns";
import StaffRoutes from "@/components/Staff/api";

import { useAuthContext } from "@/hooks/useAuthUser";

import * as Notification from "@/Utils/Notifications";
import {
  DATE_FORMATS,
  DateFormat,
  dateFormatOf,
  formatDate,
} from "@/Utils/dates";
import { callApi } from "@/Utils/request/query";
import { APP_VERSION } from "@/Utils/version";

/** The person's own display preferences (today: how dates are shown). */
export default function PreferencesTab({
  userData,
  username,
  refetchUserData,
}: userChildProps) {
  const { t } = useTranslation();
  const { refetchUser } = useAuthContext();
  const current = dateFormatOf(userData.date_format);
  const today = new Date();

  const { data: server } = useQuery({
    queryKey: ["server-version"],
    queryFn: async () =>
      (await fetch("/api/v1/version/")).json() as Promise<{ version: string }>,
  });

  const save = useMutation({
    mutationFn: (date_format: DateFormat) =>
      callApi(StaffRoutes.setPreferences, {
        pathParams: { username },
        body: { date_format },
        silent: true,
      }),
    onSuccess: async () => {
      Notification.Success({ msg: t("preferences_saved") });
      // The whole app reads the format from the signed-in user
      await refetchUser();
      refetchUserData?.();
    },
    onError: () => Notification.Error({ msg: t("preferences_save_failed") }),
  });

  return (
    <section className="mt-6 max-w-3xl rounded-xl border border-gray-200 bg-white p-5 shadow-sm">
      <h2 className="text-base font-semibold text-gray-900">
        {t("date_format")}
      </h2>
      <p className="mt-1 text-sm text-gray-500">{t("date_format_help")}</p>
      <div
        role="radiogroup"
        aria-label={t("date_format")}
        className="mt-4 grid gap-3 sm:grid-cols-2"
      >
        {DATE_FORMATS.map((f) => (
          <button
            key={f}
            type="button"
            role="radio"
            aria-checked={current === f}
            disabled={save.isPending}
            onClick={() => f !== current && save.mutate(f)}
            className={cn(
              "flex items-center justify-between rounded-lg border p-3 text-left transition",
              current === f
                ? "border-primary-500 bg-primary-50 ring-1 ring-primary-500"
                : "border-gray-200 hover:bg-gray-50",
            )}
          >
            <span>
              <span className="block text-sm font-medium text-gray-900">
                {formatDate(today, f)}
              </span>
              <span className="block text-xs text-gray-500">
                {t(`date_format__${f}`)}
              </span>
            </span>
            {current === f && (
              <Check className="size-4 text-primary-600" aria-hidden />
            )}
          </button>
        ))}
      </div>

      <h2 className="mt-8 text-base font-semibold text-gray-900">
        {t("about_klyniva")}
      </h2>
      <dl className="mt-2 grid grid-cols-2 gap-x-6 gap-y-1 text-sm sm:w-fit">
        <dt className="text-gray-600">{t("app_version")}</dt>
        <dd className="font-mono text-gray-900">{APP_VERSION}</dd>
        <dt className="text-gray-600">{t("server_version")}</dt>
        <dd className="font-mono text-gray-900">{server?.version ?? "-"}</dd>
      </dl>
    </section>
  );
}
