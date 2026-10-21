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

import { useTherapists } from "@/components/Referral/hooks";

export interface TherapistChoice {
  therapistId?: number;
  reason: string;
}

interface Props {
  /** "reassign" requires a therapist; "reopen" may send it back to the inbox. */
  mode: "reassign" | "reopen" | null;
  /** The currently assigned therapist (excluded from the hand-over list). */
  currentTherapistId?: number | null;
  pending: boolean;
  onClose: () => void;
  onConfirm: (choice: TherapistChoice) => void;
}

const fieldClass =
  "mt-1 w-full rounded-lg border border-gray-300 p-2.5 text-sm shadow-sm focus:border-primary-500 focus:ring-primary-500";

/** Picks a physiotherapist (and a reason) for hand-over or re-open. */
export default function TherapistDialog({
  mode,
  currentTherapistId,
  pending,
  onClose,
  onConfirm,
}: Props) {
  const { t } = useTranslation();
  const { data: therapists } = useTherapists();
  const [therapist, setTherapist] = useState("");
  const [reason, setReason] = useState("");

  // Start empty every time the dialog opens
  useEffect(() => {
    if (mode) {
      setTherapist("");
      setReason("");
    }
  }, [mode]);

  const options =
    therapists?.results.filter(
      (u) => mode === "reopen" || u.id !== currentTherapistId,
    ) ?? [];
  const valid = reason.trim() && (mode === "reopen" || therapist);

  return (
    <Dialog open={mode !== null} onOpenChange={(open) => !open && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{mode && t(`referral_action__${mode}`)}</DialogTitle>
          <DialogDescription>
            {mode && t(`referral_action_hint__${mode}`)}
          </DialogDescription>
        </DialogHeader>
        <form
          className="flex flex-col gap-4"
          onSubmit={(e) => {
            e.preventDefault();
            if (!valid) return;
            onConfirm({
              therapistId: therapist ? Number(therapist) : undefined,
              reason: reason.trim(),
            });
          }}
        >
          <div>
            <label
              htmlFor="therapist-select"
              className="text-sm font-medium text-gray-700"
            >
              {t("physiotherapist")}
            </label>
            <select
              id="therapist-select"
              value={therapist}
              onChange={(e) => setTherapist(e.target.value)}
              className={fieldClass}
              required={mode === "reassign"}
            >
              <option value="">
                {mode === "reopen"
                  ? t("any_available_physiotherapist")
                  : t("choose_physiotherapist")}
              </option>
              {options.map((u) => (
                <option key={u.id} value={u.id}>
                  {`${u.first_name} ${u.last_name}`}
                </option>
              ))}
            </select>
          </div>
          <div>
            <label
              htmlFor="therapist-reason"
              className="text-sm font-medium text-gray-700"
            >
              {t("reason")}
            </label>
            <textarea
              id="therapist-reason"
              rows={3}
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              placeholder={t("reason_placeholder")}
              className={fieldClass}
            />
          </div>
          <DialogFooter className="gap-2">
            <Button type="button" variant="outline" onClick={onClose}>
              {t("cancel")}
            </Button>
            <Button
              type="submit"
              variant="primary"
              disabled={!valid || pending}
            >
              {mode && t(`referral_action__${mode}`)}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
