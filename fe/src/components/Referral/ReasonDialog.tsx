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

import { ReferralAction } from "@/components/Referral/types";

interface Props {
  /** The action awaiting a reason, or null when closed. */
  action: ReferralAction | null;
  pending: boolean;
  onClose: () => void;
  onConfirm: (reason: string) => void;
}

/**
 * Asks for the free-text reason that decline/cancel need, or the outcome
 * summary for complete (sent to the doctor).
 */
export default function ReasonDialog({
  action,
  pending,
  onClose,
  onConfirm,
}: Props) {
  const { t } = useTranslation();
  const [reason, setReason] = useState("");

  // Start empty every time the dialog opens
  useEffect(() => {
    if (action) setReason("");
  }, [action]);

  return (
    <Dialog open={action !== null} onOpenChange={(open) => !open && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{action && t(`referral_action__${action}`)}</DialogTitle>
          <DialogDescription>
            {action && t(`referral_action_hint__${action}`)}
          </DialogDescription>
        </DialogHeader>
        <form
          onSubmit={(e) => {
            e.preventDefault();
            if (reason.trim()) onConfirm(reason.trim());
          }}
        >
          <label htmlFor="referral-reason" className="sr-only">
            {t("reason")}
          </label>
          <textarea
            id="referral-reason"
            autoFocus
            rows={5}
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            className="w-full rounded-lg border border-gray-300 p-3 text-sm shadow-sm focus:border-primary-500 focus:ring-primary-500"
            placeholder={
              action === "complete"
                ? t("outcome_summary_placeholder")
                : t("reason_placeholder")
            }
          />
          <DialogFooter className="mt-4 gap-2">
            <Button type="button" variant="outline" onClick={onClose}>
              {t("cancel")}
            </Button>
            <Button
              type="submit"
              variant={action === "complete" ? "primary" : "destructive"}
              disabled={!reason.trim() || pending}
            >
              {action && t(`referral_action__${action}`)}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
