import { useTranslation } from "react-i18next";

const STEPS = ["refer", "accept", "treat", "close"] as const;

/**
 * The referral workflow in four steps, shown inside the page help ("?")
 * of the referrals page.
 */
export default function ReferralHelp() {
  const { t } = useTranslation();
  return (
    <div>
      <p className="mb-3 text-sm text-gray-700">{t("referral_help_intro")}</p>
      <ol className="grid gap-3 sm:grid-cols-2">
        {STEPS.map((s, i) => (
          <li
            key={s}
            className="rounded-lg bg-white p-3 text-sm shadow-sm ring-1 ring-gray-100"
          >
            <p className="mb-1 flex items-center gap-2 font-semibold text-gray-900">
              <span className="flex size-6 items-center justify-center rounded-full bg-primary-600 text-xs text-white">
                {i + 1}
              </span>
              {t(`referral_help_step__${s}`)}
            </p>
            <p className="text-gray-600">
              {t(`referral_help_step__${s}_text`)}
            </p>
          </li>
        ))}
      </ol>
    </div>
  );
}
