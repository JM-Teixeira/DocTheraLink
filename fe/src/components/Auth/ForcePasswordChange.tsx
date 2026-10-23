import { useMutation } from "@tanstack/react-query";
import { KeyRound } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";

import { Button } from "@/components/ui/button";

import useAuthUser, { useAuthContext } from "@/hooks/useAuthUser";

import routes from "@/Utils/request/api";
import { callApi } from "@/Utils/request/query";
import { HTTPError } from "@/Utils/request/types";

const fieldClass =
  "mt-1 h-10 w-full rounded-lg border border-gray-300 px-3 text-sm shadow-sm focus:border-primary-500 focus:ring-primary-500";

/** What is wrong with the new password ("" when it is fine). */
export function newPasswordProblem(
  current: string,
  next: string,
  confirm: string,
) {
  if (next.length < 8) return "password_min_8";
  if (next === current) return "password_must_differ";
  if (next !== confirm) return "passwords_do_not_match";
  return "";
}

/**
 * First sign-in of an account someone else created: the person chooses their
 * own password before anything else (the server refuses every other call
 * until then).
 */
export default function ForcePasswordChange() {
  const { t } = useTranslation();
  const user = useAuthUser();
  const { refetchUser, signOut } = useAuthContext();
  const [current, setCurrent] = useState("");
  const [next, setNext] = useState("");
  const [confirm, setConfirm] = useState("");
  const [error, setError] = useState("");
  const problem = newPasswordProblem(current, next, confirm);

  const save = useMutation({
    mutationFn: () =>
      callApi(routes.updatePassword, {
        body: {
          username: user.username,
          old_password: current,
          new_password: next,
        },
        silent: true,
      }),
    onSuccess: () => refetchUser(),
    onError: (e: Error) =>
      setError(
        e instanceof HTTPError && e.cause?.message === "Invalid old password"
          ? t("current_password_wrong")
          : t("password_change_failed"),
      ),
  });

  return (
    <main className="flex min-h-screen items-center justify-center bg-gray-100 p-4">
      <form
        className="flex w-full max-w-md flex-col gap-4 rounded-xl border border-gray-200 bg-white p-6 shadow-sm"
        onSubmit={(e) => {
          e.preventDefault();
          setError("");
          if (!problem && current) save.mutate();
        }}
      >
        <div className="flex items-center gap-3">
          <span className="flex size-10 items-center justify-center rounded-full bg-primary-100 text-primary-700">
            <KeyRound className="size-5" aria-hidden />
          </span>
          <div>
            <h1 className="text-lg font-semibold text-gray-900">
              {t("choose_your_password")}
            </h1>
            <p className="text-sm text-gray-600">
              {t("choose_your_password_help", { username: user.username })}
            </p>
          </div>
        </div>
        <label className="text-sm font-medium text-gray-700">
          {t("current_password")}
          <input
            type="password"
            autoComplete="current-password"
            value={current}
            onChange={(e) => setCurrent(e.target.value)}
            className={fieldClass}
          />
        </label>
        <label className="text-sm font-medium text-gray-700">
          {t("new_password")}
          <input
            type="password"
            autoComplete="new-password"
            value={next}
            onChange={(e) => setNext(e.target.value)}
            className={fieldClass}
          />
        </label>
        <label className="text-sm font-medium text-gray-700">
          {t("confirm_password")}
          <input
            type="password"
            autoComplete="new-password"
            value={confirm}
            onChange={(e) => setConfirm(e.target.value)}
            className={fieldClass}
          />
        </label>
        {next && problem && (
          <p className="text-xs text-gray-600">{t(problem)}</p>
        )}
        {error && (
          <p role="alert" className="text-sm text-rose-700">
            {error}
          </p>
        )}
        <div className="flex justify-between gap-2">
          <Button type="button" variant="ghost" onClick={signOut}>
            {t("sign_out")}
          </Button>
          <Button
            type="submit"
            variant="primary"
            disabled={!!problem || !current || save.isPending}
          >
            {t("save_new_password")}
          </Button>
        </div>
      </form>
    </main>
  );
}
