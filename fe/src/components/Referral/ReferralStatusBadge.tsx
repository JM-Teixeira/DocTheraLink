import { useTranslation } from "react-i18next";

import { cn } from "@/lib/utils";

import { STATUS_STYLES } from "@/components/Referral/referralUtils";
import { ReferralStatus } from "@/components/Referral/types";

/** Pill showing a referral status with a coloured dot. */
export default function ReferralStatusBadge({
  status,
  className,
}: {
  status: ReferralStatus;
  className?: string;
}) {
  const { t } = useTranslation();
  const style = STATUS_STYLES[status];
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-medium ring-1 ring-inset",
        style.badge,
        className,
      )}
    >
      <span className={cn("size-1.5 rounded-full", style.dot)} aria-hidden />
      {t(`referral_status__${status}`)}
    </span>
  );
}
