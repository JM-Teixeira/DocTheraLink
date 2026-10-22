import { useMutation, useQuery } from "@tanstack/react-query";
import { navigate } from "raviger";
import { ReactNode, useState } from "react";
import { useTranslation } from "react-i18next";

import { Button } from "@/components/ui/button";

import Page from "@/components/Common/Page";
import StaffRoutes from "@/components/Staff/api";
import { roleKey } from "@/components/Staff/staffUtils";

import useAuthUser from "@/hooks/useAuthUser";
import useDebouncedValue from "@/hooks/useDebouncedValue";

import * as Notification from "@/Utils/Notifications";
import routes from "@/Utils/request/api";
import query, { callApi } from "@/Utils/request/query";
import { HTTPError } from "@/Utils/request/types";
import { isAdmin } from "@/Utils/roles";

/** Roles a coordinator may give (be/internal/api/handlers/access.go). */
const CLINIC_ROLES = [
  "Doctor",
  "PhysioTherapist",
  "Receptionist",
  "Nurse",
  "Coordinator",
  "Staff",
  "StaffReadOnly",
] as const;
/** Organisation roles, admins only; these have no clinic. */
const ADMIN_ROLES = ["StateAdmin", "DistrictAdmin"] as const;

const fieldClass =
  "mt-1 h-10 w-full rounded-lg border border-gray-300 px-3 text-sm shadow-sm focus:border-primary-500 focus:ring-primary-500";

/**
 * Adds a person to a clinic. The username is not typed: the server builds it
 * from the clinic code and the name (cm.joao.teixeira) and the preview shows
 * it while typing.
 */
export default function AddPerson() {
  const { t } = useTranslation();
  const user = useAuthUser();
  const admin = isAdmin(user);
  const [form, setForm] = useState({
    first_name: "",
    last_name: "",
    email: "",
    phone_number: "",
    user_type: "PhysioTherapist",
    home_facility: "",
    password: "",
    confirm: "",
  });
  const [error, setError] = useState("");
  const set =
    (k: keyof typeof form) =>
    (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
      setForm((f) => ({ ...f, [k]: e.target.value }));

  const orgRole = (ADMIN_ROLES as readonly string[]).includes(form.user_type);
  const roles = admin ? [...CLINIC_ROLES, ...ADMIN_ROLES] : CLINIC_ROLES;

  const { data: clinics } = useQuery({
    queryKey: ["staff-clinics"],
    queryFn: query(StaffRoutes.clinics, { queryParams: { limit: 100 } }),
    enabled: admin,
  });

  // Username preview (the server decides again when saving)
  const names = useDebouncedValue(
    `${form.first_name.trim()}|${form.last_name.trim()}|${orgRole ? "" : form.home_facility}`,
  );
  const [first, last, clinic] = names.split("|");
  const preview = useQuery({
    queryKey: ["username-suggestion", names],
    queryFn: query(StaffRoutes.usernameSuggestion, {
      queryParams: {
        first_name: first,
        last_name: last,
        home_facility: admin ? clinic || undefined : undefined,
      },
      silent: true,
    }),
    enabled: !!first && !!last && (!admin || orgRole || !!clinic),
    retry: false,
  });

  const problem =
    form.first_name.trim().length < 1 || form.last_name.trim().length < 1
      ? "add_person_names_required"
      : !/^[^@\s]+@[^@\s]+\.[^@\s]+$/.test(form.email.trim())
        ? "error__invalid_email"
        : admin && !orgRole && !form.home_facility
          ? "choose_clinic"
          : form.password.length < 8
            ? "password_min_8"
            : form.password !== form.confirm
              ? "passwords_do_not_match"
              : "";

  const save = useMutation({
    mutationFn: () =>
      callApi(routes.addUser, {
        body: {
          first_name: form.first_name.trim(),
          last_name: form.last_name.trim(),
          email: form.email.trim(),
          phone_number: form.phone_number.trim() || undefined,
          user_type: form.user_type,
          home_facility: admin && !orgRole ? form.home_facility : undefined,
          password: form.password,
        },
        silent: true,
      }),
    onSuccess: (u) => {
      Notification.Success({
        msg: t("person_added", { username: u.username }),
      });
      navigate(`/users/${u.username}`);
    },
    onError: (e: Error) =>
      setError(
        e instanceof HTTPError && typeof e.cause?.message === "string"
          ? e.cause.message
          : t("person_add_failed"),
      ),
  });

  return (
    <Page title={t("add_person")} backUrl="/users" help="add_person">
      <form
        noValidate
        className="mx-auto flex max-w-2xl flex-col gap-5 rounded-xl border border-gray-200 bg-white p-5 shadow-sm"
        onSubmit={(e) => {
          e.preventDefault();
          setError("");
          if (!problem) save.mutate();
        }}
      >
        <div className="grid gap-4 sm:grid-cols-2">
          <Field label={`${t("first_name")} *`}>
            <input
              value={form.first_name}
              onChange={set("first_name")}
              className={fieldClass}
            />
          </Field>
          <Field label={`${t("last_name")} *`}>
            <input
              value={form.last_name}
              onChange={set("last_name")}
              className={fieldClass}
            />
          </Field>
          <Field label={`${t("role")} *`}>
            <select
              value={form.user_type}
              onChange={set("user_type")}
              className={fieldClass}
            >
              {roles.map((r) => (
                <option key={r} value={r}>
                  {t(roleKey(r), { defaultValue: r })}
                </option>
              ))}
            </select>
          </Field>
          {admin && !orgRole && (
            <Field label={`${t("clinic")} *`}>
              <select
                value={form.home_facility}
                onChange={set("home_facility")}
                className={fieldClass}
              >
                <option value="">{t("choose_clinic")}</option>
                {clinics?.results.map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.name}
                  </option>
                ))}
              </select>
            </Field>
          )}
          <Field label={`${t("email")} *`}>
            <input
              type="email"
              value={form.email}
              onChange={set("email")}
              className={fieldClass}
            />
          </Field>
          <Field label={t("phone")}>
            <input
              type="tel"
              value={form.phone_number}
              onChange={set("phone_number")}
              className={fieldClass}
            />
          </Field>
          <Field label={`${t("password")} *`}>
            <input
              type="password"
              autoComplete="new-password"
              value={form.password}
              onChange={set("password")}
              className={fieldClass}
            />
          </Field>
          <Field label={`${t("confirm_password")} *`}>
            <input
              type="password"
              autoComplete="new-password"
              value={form.confirm}
              onChange={set("confirm")}
              className={fieldClass}
            />
          </Field>
        </div>

        <div className="rounded-lg bg-gray-50 p-3 text-sm" aria-live="polite">
          <span className="text-gray-600">{t("username_will_be")}: </span>
          <span className="font-mono font-semibold text-gray-900">
            {preview.data?.username ??
              (preview.error instanceof HTTPError &&
              typeof preview.error.cause?.message === "string"
                ? preview.error.cause.message
                : "-")}
          </span>
          <p className="mt-1 text-xs text-gray-500">
            {t("username_standard_help")}
          </p>
        </div>

        {error && (
          <p role="alert" className="text-sm text-rose-700">
            {error}
          </p>
        )}
        <div className="flex justify-end gap-2">
          <Button
            type="button"
            variant="outline"
            onClick={() => navigate("/users")}
          >
            {t("cancel")}
          </Button>
          <Button
            type="submit"
            variant="primary"
            disabled={!!problem || save.isPending}
          >
            {t("add_person")}
          </Button>
        </div>
        {problem && (
          <p className="-mt-3 text-right text-xs text-gray-500">{t(problem)}</p>
        )}
      </form>
    </Page>
  );
}

function Field({ label, children }: { label: string; children: ReactNode }) {
  return (
    <label className="text-sm font-medium text-gray-700">
      {label}
      {children}
    </label>
  );
}
