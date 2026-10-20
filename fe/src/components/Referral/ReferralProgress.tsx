import { Check, X } from "lucide-react";
import { useTranslation } from "react-i18next";

import { cn } from "@/lib/utils";

import { STATUS_STEPS, stepIndex } from "@/components/Referral/referralUtils";
import { Referral } from "@/components/Referral/types";

/**
 * Horizontal tracker: Pending → Accepted → In progress → Completed.
 * Declined/cancelled referrals show a single closed state instead.
 */
export default function ReferralProgress({ referral }: { referral: Referral }) {
  const { t } = useTranslation();
  const current = stepIndex(referral.status);

  if (current === -1) {
    return (
      <div className="flex items-center gap-3 rounded-lg bg-gray-50 px-4 py-3 text-sm text-gray-700">
        <span className="flex size-7 items-center justify-center rounded-full bg-gray-300 text-white">
          <X className="size-4" aria-hidden />
        </span>
        {t(`referral_status__${referral.status}`)}
      </div>
    );
  }

  return (
    <ol className="flex items-center" aria-label={t("referral_progress")}>
      {STATUS_STEPS.map((step, i) => {
        const done = i < current || referral.status === "completed";
        const active = i === current && referral.status !== "completed";
        return (
          <li
            key={step}
            className={cn(
              "flex items-center",
              i < STATUS_STEPS.length - 1 && "flex-1",
            )}
            aria-current={active ? "step" : undefined}
          >
            <div className="flex flex-col items-center gap-1.5">
              <span
                className={cn(
                  "flex size-8 items-center justify-center rounded-full text-xs font-semibold transition",
                  done && "bg-primary-600 text-white",
                  active && "bg-white text-primary-700 ring-2 ring-primary-600",
                  !done && !active && "bg-gray-100 text-gray-400",
                )}
              >
                {done ? <Check className="size-4" aria-hidden /> : i + 1}
              </span>
              <span
                className={cn(
                  "whitespace-nowrap text-xs",
                  done || active
                    ? "font-medium text-gray-900"
                    : "text-gray-400",
                )}
              >
                {t(`referral_status__${step}`)}
              </span>
            </div>
            {i < STATUS_STEPS.length - 1 && (
              <span
                className={cn(
                  "mx-2 mb-5 h-0.5 flex-1 rounded",
                  i < current || referral.status === "completed"
                    ? "bg-primary-600"
                    : "bg-gray-200",
                )}
                aria-hidden
              />
            )}
          </li>
        );
      })}
    </ol>
  );
}
