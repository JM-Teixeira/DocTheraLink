import { useQuery } from "@tanstack/react-query";
import { Eye } from "lucide-react";
import { useState } from "react";
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

import StaffRoutes from "@/components/Staff/api";
import { roleKey } from "@/components/Staff/staffUtils";

import query from "@/Utils/request/query";
import { getViewAs, setViewAs } from "@/Utils/viewAs";

/** Roles an admin may view as (same list as the server, auth/viewas.go). */
export const VIEW_AS_ROLES = [
  "Coordinator",
  "PhysioTherapist",
  "Doctor",
  "Receptionist",
  "Nurse",
  "Staff",
  "StaffReadOnly",
] as const;

const fieldClass =
  "mt-1 h-10 w-full rounded-lg border border-gray-300 px-3 text-sm shadow-sm focus:border-primary-500 focus:ring-primary-500";

/** Restart the app so every page and cached answer uses the new view. */
const restart = () => window.location.assign("/");

/** Admins choose a role and clinic to look at the app as (read-only). */
export function ViewAsDialog({
  open,
  onClose,
}: {
  open: boolean;
  onClose: () => void;
}) {
  const { t } = useTranslation();
  const [role, setRole] = useState<string>("PhysioTherapist");
  const [clinic, setClinic] = useState("");
  const { data: clinics } = useQuery({
    queryKey: ["staff-clinics"],
    queryFn: query(StaffRoutes.clinics, { queryParams: { limit: 100 } }),
    enabled: open,
  });
  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t("view_as")}</DialogTitle>
          <DialogDescription>{t("view_as_help")}</DialogDescription>
        </DialogHeader>
        <form
          className="flex flex-col gap-4"
          onSubmit={(e) => {
            e.preventDefault();
            const name =
              clinics?.results.find((c) => c.id === clinic)?.name ?? "";
            setViewAs({ role, clinic, clinicName: name });
            restart();
          }}
        >
          <label className="text-sm font-medium text-gray-700">
            {t("role")}
            <select
              value={role}
              onChange={(e) => setRole(e.target.value)}
              className={fieldClass}
            >
              {VIEW_AS_ROLES.map((r) => (
                <option key={r} value={r}>
                  {t(roleKey(r), { defaultValue: r })}
                </option>
              ))}
            </select>
          </label>
          <label className="text-sm font-medium text-gray-700">
            {t("clinic")}
            <select
              value={clinic}
              onChange={(e) => setClinic(e.target.value)}
              className={fieldClass}
            >
              <option value="">{t("choose_clinic")}</option>
              {clinics?.results.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name}
                </option>
              ))}
            </select>
          </label>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={onClose}>
              {t("cancel")}
            </Button>
            <Button type="submit" variant="primary" disabled={!clinic}>
              {t("view_as_start")}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

/** Shown on every page while viewing as a role, with the way back. */
export function ViewAsBanner() {
  const { t } = useTranslation();
  const v = getViewAs();
  if (!v) return null;
  return (
    <div
      role="status"
      className="flex flex-wrap items-center justify-center gap-3 bg-amber-100 px-4 py-2 text-sm text-amber-900"
    >
      <Eye className="size-4" aria-hidden />
      <span>
        {t("viewing_as", {
          role: t(roleKey(v.role), { defaultValue: v.role }),
          clinic: v.clinicName,
        })}
      </span>
      <Button
        size="sm"
        variant="outline"
        className="h-7 bg-white"
        onClick={() => {
          setViewAs(null);
          restart();
        }}
      >
        {t("view_as_stop")}
      </Button>
    </div>
  );
}
