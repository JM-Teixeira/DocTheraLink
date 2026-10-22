import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

import { chipClass } from "@/components/Common/FilterChips";
import { MissInput } from "@/components/TimeBoard/types";

/**
 * Marking a session missed: was it justified (ill, warned in advance) or
 * not? Unjustified misses add up and flag the patient to the coordinator.
 */
export default function MissDialog({
  patientName,
  initial,
  pending,
  onClose,
  onConfirm,
}: {
  /** null = closed */
  patientName: string | null;
  initial?: MissInput;
  pending?: boolean;
  onClose: () => void;
  onConfirm: (v: MissInput) => void;
}) {
  const { t } = useTranslation();
  const [justified, setJustified] = useState(false);
  const [reason, setReason] = useState("");
  // Reset when it opens for a patient (not on every render of the caller)
  useEffect(() => {
    if (patientName !== null) {
      setJustified(initial?.justified ?? false);
      setReason(initial?.reason ?? "");
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [patientName]);

  return (
    <Dialog open={patientName !== null} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>
            {t("mark_missed_title", { name: patientName })}
          </DialogTitle>
          <DialogDescription>{t("mark_missed_help")}</DialogDescription>
        </DialogHeader>
        <form
          className="flex flex-col gap-3"
          onSubmit={(e) => {
            e.preventDefault();
            onConfirm({ justified, reason: reason.trim() });
          }}
        >
          <div
            role="radiogroup"
            aria-label={t("miss_kind")}
            className="flex flex-wrap gap-2"
          >
            {[false, true].map((j) => (
              <button
                key={String(j)}
                type="button"
                role="radio"
                aria-checked={justified === j}
                className={chipClass(justified === j)}
                onClick={() => setJustified(j)}
              >
                {j ? t("miss_justified") : t("miss_unjustified")}
              </button>
            ))}
          </div>
          <label className="text-sm font-medium text-gray-700">
            {t("reason")}
            <input
              value={reason}
              maxLength={200}
              onChange={(e) => setReason(e.target.value)}
              placeholder={
                justified
                  ? t("miss_reason_justified_placeholder")
                  : t("miss_reason_placeholder")
              }
              className="mt-1 h-10 w-full rounded-lg border border-gray-300 px-3 text-sm shadow-sm focus:border-primary-500 focus:ring-primary-500"
            />
          </label>
          <DialogFooter className="gap-2">
            <Button type="button" variant="outline" onClick={onClose}>
              {t("cancel")}
            </Button>
            <Button type="submit" variant="primary" disabled={pending}>
              {t("mark_missed")}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

/** Status text of a session, saying whether a miss was justified. */
export function statusLabel(
  t: (k: string) => string,
  ap: { status: string; miss_justified?: boolean | null },
) {
  if (ap.status !== "missed") return t(`appointment_status__${ap.status}`);
  return ap.miss_justified
    ? t("missed_justified_short")
    : t("missed_unjustified_short");
}
