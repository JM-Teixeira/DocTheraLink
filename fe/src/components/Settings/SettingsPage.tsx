import { useMutation } from "@tanstack/react-query";
import { Check } from "lucide-react";
import { ReactNode } from "react";
import { useTranslation } from "react-i18next";

import { cn } from "@/lib/utils";

import Page from "@/components/Common/Page";
import SuggestionsCard from "@/components/Settings/SuggestionsCard";
import StaffRoutes from "@/components/Staff/api";
import UserSoftwareUpdate from "@/components/Users/UserSoftwareUpdate";

import useAuthUser, { useAuthContext } from "@/hooks/useAuthUser";

import * as Notification from "@/Utils/Notifications";
import { DATE_FORMATS, dateFormatOf, formatDate } from "@/Utils/dates";
import { callApi } from "@/Utils/request/query";
import {
  LANGUAGES,
  START_PAGES,
  TEXT_SIZES,
  THEMES,
  textSizeOf,
  themeOf,
} from "@/Utils/settings";

type Prefs = {
  date_format?: string;
  theme?: string;
  text_size?: string;
  start_page?: string;
  language?: string;
};

/**
 * The person's own settings: appearance, language and dates, start page,
 * software update and suggestions. Saved on the account, so they follow the
 * person to any computer.
 */
export default function SettingsPage() {
  const { t, i18n } = useTranslation();
  const user = useAuthUser();
  const { refetchUser } = useAuthContext();
  const save = useMutation({
    mutationFn: (body: Prefs) =>
      callApi(StaffRoutes.setPreferences, {
        pathParams: { username: user.username },
        body: body as never,
        silent: true,
      }),
    onSuccess: async () => {
      Notification.Success({ msg: t("preferences_saved") });
      await refetchUser(); // the whole app reads the settings from the user
    },
    onError: () => Notification.Error({ msg: t("preferences_save_failed") }),
  });
  const today = new Date();

  return (
    <Page title={t("settings")} hideBack help="settings">
      <div className="flex max-w-3xl flex-col gap-6">
        <Card title={t("appearance")}>
          <Choice
            label={t("theme")}
            value={themeOf(user.theme)}
            options={THEMES.map((v) => ({ value: v, label: t(`theme__${v}`) }))}
            onChange={(theme) => save.mutate({ theme })}
            pending={save.isPending}
          />
          <Choice
            label={t("text_size")}
            value={textSizeOf(user.text_size)}
            options={TEXT_SIZES.map((v) => ({
              value: v,
              label: t(`text_size__${v}`),
            }))}
            onChange={(text_size) => save.mutate({ text_size })}
            pending={save.isPending}
          />
        </Card>

        <Card title={t("language_and_dates")}>
          <Choice
            label={t("language")}
            value={
              user.language || (i18n.language.startsWith("en") ? "en" : "pt")
            }
            options={LANGUAGES.map((v) => ({
              value: v,
              label: t(`language__${v}`),
            }))}
            onChange={(language) => {
              i18n.changeLanguage(language);
              save.mutate({ language });
            }}
            pending={save.isPending}
          />
          <Choice
            label={t("date_format")}
            hint={t("date_format_help")}
            value={dateFormatOf(user.date_format)}
            options={DATE_FORMATS.map((f) => ({
              value: f,
              label: formatDate(today, f),
              detail: t(`date_format__${f}`),
            }))}
            onChange={(date_format) => save.mutate({ date_format })}
            pending={save.isPending}
          />
        </Card>

        <Card title={t("start_page")}>
          <Choice
            label={t("start_page")}
            hint={t("start_page_help")}
            value={user.start_page || "home"}
            options={START_PAGES.map((v) => ({
              value: v,
              label: t(`start_page__${v}`),
            }))}
            onChange={(start_page) => save.mutate({ start_page })}
            pending={save.isPending}
          />
        </Card>

        <Card title={t("software_update")}>
          <p className="mb-3 text-sm text-gray-600">
            {t("software_update_explained")}
          </p>
          <UserSoftwareUpdate />
        </Card>

        <SuggestionsCard />
      </div>
    </Page>
  );
}

function Card({ title, children }: { title: string; children: ReactNode }) {
  return (
    <section
      aria-label={title}
      className="flex flex-col gap-5 rounded-xl border border-gray-200 bg-white p-5 shadow-sm"
    >
      <h2 className="text-base font-semibold text-gray-900">{title}</h2>
      {children}
    </section>
  );
}

/** A labelled group of option cards; clicking one saves it. */
function Choice({
  label,
  hint,
  value,
  options,
  onChange,
  pending,
}: {
  label: string;
  hint?: string;
  value: string;
  options: { value: string; label: string; detail?: string }[];
  onChange: (v: string) => void;
  pending: boolean;
}) {
  return (
    <div>
      <p className="text-sm font-medium text-gray-700">{label}</p>
      {hint && <p className="text-xs text-gray-500">{hint}</p>}
      <div
        role="radiogroup"
        aria-label={label}
        className="mt-2 grid gap-2 sm:grid-cols-3"
      >
        {options.map((o) => (
          <button
            key={o.value}
            type="button"
            role="radio"
            aria-checked={value === o.value}
            disabled={pending}
            onClick={() => o.value !== value && onChange(o.value)}
            className={cn(
              "flex items-center justify-between rounded-lg border p-3 text-left text-sm transition",
              value === o.value
                ? "border-primary-500 bg-primary-50 ring-1 ring-primary-500"
                : "border-gray-200 hover:bg-gray-50",
            )}
          >
            <span>
              <span className="block font-medium text-gray-900">{o.label}</span>
              {o.detail && (
                <span className="block text-xs text-gray-500">{o.detail}</span>
              )}
            </span>
            {value === o.value && (
              <Check className="size-4 text-primary-600" aria-hidden />
            )}
          </button>
        ))}
      </div>
    </div>
  );
}
