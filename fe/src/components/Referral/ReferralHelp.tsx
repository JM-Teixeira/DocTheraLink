import { ChevronDown, CircleHelp } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";

import { cn } from "@/lib/utils";

const STORAGE_KEY = "klyniva.referral-help.open";
const STEPS = ["refer", "accept", "treat", "close"] as const;

function initiallyOpen() {
  try {
    return localStorage.getItem(STORAGE_KEY) !== "false";
  } catch {
    return true;
  }
}

/**
 * Explains the referral workflow in four steps. Open by default; the
 * choice to hide it is remembered on this browser.
 */
export default function ReferralHelp() {
  const { t } = useTranslation();
  const [open, setOpen] = useState(initiallyOpen);
  const toggle = () => {
    setOpen((o) => {
      try {
        localStorage.setItem(STORAGE_KEY, String(!o));
      } catch {
        // storage unavailable: just toggle for this visit
      }
      return !o;
    });
  };

  return (
    <section className="mb-4 rounded-xl border border-primary-100 bg-primary-50/50">
      <button
        type="button"
        aria-expanded={open}
        onClick={toggle}
        className="flex w-full items-center gap-2 px-4 py-3 text-left text-sm font-medium text-primary-800"
      >
        <CircleHelp className="size-4" aria-hidden />
        {t("referral_help_title")}
        <ChevronDown
          className={cn("ml-auto size-4 transition", open && "rotate-180")}
          aria-hidden
        />
      </button>
      {open && (
        <div className="px-4 pb-4">
          <p className="mb-3 text-sm text-gray-700">
            {t("referral_help_intro")}
          </p>
          <ol className="grid gap-3 sm:grid-cols-4">
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
          {/* What the list itself offers (the only help on this page) */}
          <p className="mt-3 text-sm text-gray-700">
            {t("referral_help_page")}
          </p>
        </div>
      )}
    </section>
  );
}
