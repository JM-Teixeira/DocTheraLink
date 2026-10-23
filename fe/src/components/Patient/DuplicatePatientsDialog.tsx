import { TriangleAlert, UserRound } from "lucide-react";
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

import { PossibleDuplicate } from "@/components/Patient/patientFormUtils";

import useDateFormat from "@/hooks/useDateFormat";

/**
 * Shown when the backend answers 409 duplicate_patient. Same NIF / SNS number
 * (blocking) only offers to open the existing record; name-based matches can
 * be confirmed as a different person.
 */
export default function DuplicatePatientsDialog({
  open,
  blocking,
  matches,
  pending,
  onOpenPatient,
  onConfirmDifferent,
  onClose,
}: {
  open: boolean;
  blocking: boolean;
  matches: PossibleDuplicate[];
  pending: boolean;
  onOpenPatient: (id: string) => void;
  onConfirmDifferent: () => void;
  onClose: () => void;
}) {
  const { t } = useTranslation();
  const { fmt } = useDateFormat();
  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <TriangleAlert className="size-5 text-amber-500" aria-hidden />
            {blocking
              ? t("duplicate_patient_blocking_title")
              : t("duplicate_patient_title")}
          </DialogTitle>
          <DialogDescription>
            {blocking
              ? t("duplicate_patient_blocking_help")
              : t("duplicate_patient_help")}
          </DialogDescription>
        </DialogHeader>

        <ul className="divide-y divide-gray-100 rounded-lg border border-gray-200">
          {matches.map((m) => (
            <li key={m.id} className="flex items-center gap-3 p-3">
              <UserRound
                className="size-5 shrink-0 text-gray-400"
                aria-hidden
              />
              <div className="min-w-0 flex-1">
                <p className="truncate font-medium text-gray-900">{m.name}</p>
                <p className="text-xs text-gray-500">
                  {m.date_of_birth ? fmt(m.date_of_birth) : t("age_unknown")}
                  {m.phone_number && ` · ${m.phone_number}`}
                </p>
                <p
                  className={
                    m.blocking
                      ? "mt-0.5 text-xs font-medium text-rose-700"
                      : "mt-0.5 text-xs text-amber-700"
                  }
                >
                  {t(`duplicate_reason__${m.reason}`)}
                </p>
              </div>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => onOpenPatient(m.id)}
              >
                {t("open")}
              </Button>
            </li>
          ))}
          {matches.length === 0 && (
            <li className="p-3 text-sm text-gray-600">
              {t("duplicate_identifier_exists")}
            </li>
          )}
        </ul>

        <DialogFooter className="gap-2">
          <Button type="button" variant="outline" onClick={onClose}>
            {t("back_to_form")}
          </Button>
          {!blocking && (
            <Button
              type="button"
              variant="primary"
              disabled={pending}
              onClick={onConfirmDifferent}
            >
              {t("register_as_different_person")}
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
