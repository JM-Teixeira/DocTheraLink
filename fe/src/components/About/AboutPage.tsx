import careConfig from "@careConfig";
import { useQuery } from "@tanstack/react-query";
import { Link } from "raviger";
import { useTranslation } from "react-i18next";

import Page from "@/components/Common/Page";

import useDateFormat from "@/hooks/useDateFormat";

import { APP_VERSION } from "@/Utils/version";

/**
 * About Klyniva: the release (app and server), when this copy was built,
 * how version numbers work, and the "check for update" of this browser.
 */
export default function AboutPage() {
  const { t } = useTranslation();
  const { fmtTime } = useDateFormat();
  const { data: server } = useQuery({
    queryKey: ["server-version"],
    queryFn: async () =>
      (await fetch("/api/v1/version/")).json() as Promise<{ version: string }>,
  });
  const { data: build } = useQuery({
    queryKey: ["build-meta"],
    queryFn: async () =>
      (
        await fetch("/build-meta.json", { cache: "no-store" })
      ).json() as Promise<{
        version: string;
        built_at?: string;
      }>,
  });
  const mismatch = !!server && server.version !== APP_VERSION;

  return (
    <Page title={t("about_klyniva")}>
      <div className="flex max-w-3xl flex-col gap-5">
        <section className="flex items-center gap-5 rounded-xl border border-gray-200 bg-white p-6 shadow-sm">
          <img
            src={careConfig.mainLogo?.light}
            alt=""
            className="h-14 w-auto"
          />
          <div>
            <h2 className="text-xl font-semibold text-gray-900">
              {`Klyniva ${APP_VERSION}`}
            </h2>
            <p className="text-sm text-gray-600">{t("about_tagline")}</p>
          </div>
        </section>

        <section className="rounded-xl border border-gray-200 bg-white p-6 shadow-sm">
          <h2 className="mb-3 text-base font-semibold text-gray-900">
            {t("about_versions")}
          </h2>
          <dl className="grid grid-cols-[auto_1fr] gap-x-6 gap-y-2 text-sm">
            <dt className="text-gray-600">{t("app_version")}</dt>
            <dd className="font-mono text-gray-900">{APP_VERSION}</dd>
            <dt className="text-gray-600">{t("server_version")}</dt>
            <dd className="font-mono text-gray-900">
              {server?.version ?? "-"}
            </dd>
            <dt className="text-gray-600">{t("built_at")}</dt>
            <dd className="text-gray-900">
              {build?.built_at ? fmtTime(build.built_at) : "-"}
            </dd>
          </dl>
          {mismatch && (
            <p className="mt-3 rounded-lg bg-amber-50 px-3 py-2 text-sm text-amber-900">
              {t("about_version_mismatch", { server: server?.version })}
            </p>
          )}
          <p className="mt-4 text-xs text-gray-500">{t("about_semver")}</p>
        </section>

        <section className="rounded-xl border border-gray-200 bg-white p-6 shadow-sm">
          <h2 className="text-base font-semibold text-gray-900">
            {t("software_update")}
          </h2>
          <p className="mt-1 text-sm text-gray-600">
            {t("software_update_moved")}{" "}
            <Link href="/settings" className="text-primary-700 underline">
              {t("settings")}
            </Link>
          </p>
        </section>
      </div>
    </Page>
  );
}

/** One line for the profile page: the version and a link to the About page. */
export function AboutSummary() {
  const { t } = useTranslation();
  return (
    <div className="flex flex-wrap items-center gap-4 rounded-lg bg-white px-4 py-3 shadow-sm ring-1 ring-gray-200">
      <span className="font-mono text-sm text-gray-900">
        {`Klyniva ${APP_VERSION}`}
      </span>
      <Link
        href="/about"
        className="text-sm font-medium text-primary-700 hover:underline"
      >
        {t("about_open")}
      </Link>
    </div>
  );
}
