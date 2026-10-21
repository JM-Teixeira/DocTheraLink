import { useMutation, useQuery } from "@tanstack/react-query";
import dayjs from "dayjs";
import { Building2 } from "lucide-react";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";

import { cn } from "@/lib/utils";

import { Button } from "@/components/ui/button";

import { chipClass } from "@/components/Common/FilterChips";
import { userChildProps } from "@/components/Common/UserColumns";
import StaffRoutes from "@/components/Staff/api";
import { STATUS_STYLES, statusOf } from "@/components/Staff/staffUtils";
import {
  EMPLOYMENT_STATUSES,
  EmploymentStatus,
} from "@/components/Staff/types";

import useAuthUser from "@/hooks/useAuthUser";

import * as Notification from "@/Utils/Notifications";
import query, { callApi } from "@/Utils/request/query";
import { isAdmin, managesStaff } from "@/Utils/roles";

const fieldClass =
  "mt-1 h-10 w-full rounded-lg border border-gray-300 px-3 text-sm shadow-sm focus:border-primary-500 focus:ring-primary-500";

/**
 * Where a person works and whether they still do. People who left keep
 * their record (and history) but can no longer sign in.
 */
export default function EmploymentTab({
  userData,
  username,
  refetchUserData,
}: userChildProps) {
  const { t } = useTranslation();
  const me = useAuthUser();
  const admin = isAdmin(me);
  // Nobody changes their own status
  const canEdit = me.username !== username && managesStaff(me, userData);
  const current = statusOf(userData.employment_status);

  const [status, setStatus] = useState<EmploymentStatus>(current);
  const [since, setSince] = useState(dayjs().format("YYYY-MM-DD"));
  const [note, setNote] = useState(userData.employment_note ?? "");
  const [clinic, setClinic] = useState(userData.home_facility ?? "");
  useEffect(() => {
    setStatus(current);
    setNote(userData.employment_note ?? "");
    setClinic(userData.home_facility ?? "");
  }, [userData, current]);

  const { data: clinics } = useQuery({
    queryKey: ["staff-clinics"],
    queryFn: query(StaffRoutes.clinics, { queryParams: { limit: 100 } }),
    enabled: admin,
  });
  const saved = (msg: string) => () => {
    Notification.Success({ msg });
    refetchUserData?.();
  };
  const saveStatus = useMutation({
    mutationFn: () =>
      callApi(StaffRoutes.setEmployment, {
        pathParams: { username },
        body: { status, since, note },
        silent: true,
      }),
    onSuccess: saved(t("status_saved")),
    onError: () => Notification.Error({ msg: t("status_save_failed") }),
  });
  const saveClinic = useMutation({
    mutationFn: () =>
      callApi(StaffRoutes.setClinic, {
        pathParams: { username },
        body: { facility: clinic },
        silent: true,
      }),
    onSuccess: saved(t("clinic_saved")),
    onError: () => Notification.Error({ msg: t("clinic_save_failed") }),
  });

  return (
    <div className="mt-6 grid max-w-4xl gap-5 md:grid-cols-2">
      <section className="rounded-xl border border-gray-200 bg-white p-5 shadow-sm">
        <h2 className="flex items-center gap-2 text-base font-semibold text-gray-900">
          <Building2 className="size-4 text-gray-500" aria-hidden />
          {t("clinic")}
        </h2>
        <p className="mt-2 text-sm text-gray-700">
          {userData.home_facility_object?.name ?? t("no_clinic")}
        </p>
        {admin && (
          <div className="mt-4 flex gap-2">
            <select
              aria-label={t("clinic")}
              value={clinic}
              onChange={(e) => setClinic(e.target.value)}
              className={fieldClass}
            >
              <option value="">—</option>
              {clinics?.results.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name}
                </option>
              ))}
            </select>
            <Button
              variant="primary"
              className="mt-1"
              disabled={
                !clinic ||
                clinic === userData.home_facility ||
                saveClinic.isPending
              }
              onClick={() => saveClinic.mutate()}
            >
              {t("save")}
            </Button>
          </div>
        )}
      </section>

      <section className="rounded-xl border border-gray-200 bg-white p-5 shadow-sm">
        <h2 className="text-base font-semibold text-gray-900">
          {t("employment_status")}
        </h2>
        <p className="mt-2 flex flex-wrap items-center gap-2 text-sm">
          <span
            className={cn(
              "rounded-full px-2 py-0.5 text-xs font-medium",
              STATUS_STYLES[current],
            )}
          >
            {t(`employment__${current}`)}
          </span>
          {userData.employment_since && (
            <span className="text-gray-500">
              {t("since_date", {
                date: dayjs(userData.employment_since).format("D MMM YYYY"),
              })}
            </span>
          )}
        </p>
        {userData.employment_note && !canEdit && (
          <p className="mt-2 text-sm text-gray-600">
            {userData.employment_note}
          </p>
        )}

        {canEdit && (
          <form
            className="mt-4 flex flex-col gap-3"
            onSubmit={(e) => {
              e.preventDefault();
              if (
                status === "left" &&
                current !== "left" &&
                !window.confirm(t("confirm_left"))
              )
                return;
              saveStatus.mutate();
            }}
          >
            <div
              className="flex flex-wrap gap-2"
              role="group"
              aria-label={t("employment_status")}
            >
              {EMPLOYMENT_STATUSES.map((s) => (
                <button
                  key={s}
                  type="button"
                  aria-pressed={status === s}
                  className={chipClass(status === s)}
                  onClick={() => setStatus(s)}
                >
                  {t(`employment__${s}`)}
                </button>
              ))}
            </div>
            <label className="text-sm font-medium text-gray-700">
              {t("since")}
              <input
                type="date"
                value={since}
                onChange={(e) => setSince(e.target.value)}
                className={fieldClass}
              />
            </label>
            <label className="text-sm font-medium text-gray-700">
              {t("note")}
              <input
                value={note}
                maxLength={300}
                onChange={(e) => setNote(e.target.value)}
                placeholder={t("employment_note_placeholder")}
                className={fieldClass}
              />
            </label>
            {status === "left" && (
              <p className="rounded-lg bg-amber-50 p-2 text-xs text-amber-900">
                {t("left_help")}
              </p>
            )}
            <Button
              type="submit"
              variant="primary"
              className="self-start"
              disabled={saveStatus.isPending}
            >
              {t("save")}
            </Button>
          </form>
        )}
      </section>
    </div>
  );
}
