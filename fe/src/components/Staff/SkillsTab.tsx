import { useMutation, useQuery } from "@tanstack/react-query";
import { Plus, X } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";

import { Button } from "@/components/ui/button";

import { userChildProps } from "@/components/Common/UserColumns";
import StaffRoutes from "@/components/Staff/api";

import useAuthUser from "@/hooks/useAuthUser";

import * as Notification from "@/Utils/Notifications";
import query, { callApi } from "@/Utils/request/query";
import { managesStaff } from "@/Utils/roles";

/** The skills of a person: chips, plus "add" for them and their managers. */
export default function SkillsTab({ userData, username }: userChildProps) {
  const { t } = useTranslation();
  const me = useAuthUser();
  const canEdit = me.username === username || managesStaff(me, userData);
  const [pick, setPick] = useState("");

  const { data: linked, refetch } = useQuery({
    queryKey: ["user-skills", username],
    queryFn: query(StaffRoutes.userSkills, { pathParams: { username } }),
  });
  const { data: catalogue } = useQuery({
    queryKey: ["skills"],
    queryFn: query(StaffRoutes.skills),
    enabled: canEdit,
  });
  const fail = () => Notification.Error({ msg: t("skill_save_failed") });
  const add = useMutation({
    mutationFn: (skill: string) =>
      callApi(StaffRoutes.addSkill, {
        pathParams: { username },
        body: { skill },
        silent: true,
      }),
    onSuccess: () => {
      setPick("");
      refetch();
    },
    onError: fail,
  });
  const remove = useMutation({
    mutationFn: (id: string) =>
      callApi(StaffRoutes.removeSkill, {
        pathParams: { username, id },
        silent: true,
      }),
    onSuccess: () => refetch(),
    onError: fail,
  });

  const have = new Set(linked?.results.map((s) => s.skill_object.id));
  const available = catalogue?.results.filter((s) => !have.has(s.id)) ?? [];

  return (
    <section className="mt-6 max-w-3xl rounded-xl border border-gray-200 bg-white p-5 shadow-sm">
      <h2 className="text-base font-semibold text-gray-900">{t("skills")}</h2>
      <p className="mt-1 text-sm text-gray-500">{t("skills_help")}</p>

      {linked?.results.length ? (
        <ul className="mt-4 flex flex-wrap gap-2">
          {linked.results.map((s) => (
            <li
              key={s.id}
              className="flex items-center gap-1 rounded-full bg-primary-50 py-1 pl-3 pr-1 text-sm text-primary-800 ring-1 ring-primary-200"
            >
              <span title={s.skill_object.description ?? undefined}>
                {s.skill_object.name}
              </span>
              {canEdit && (
                <button
                  type="button"
                  aria-label={t("remove_skill", { name: s.skill_object.name })}
                  onClick={() => remove.mutate(s.id)}
                  className="rounded-full p-0.5 hover:bg-primary-100"
                >
                  <X className="size-3.5" aria-hidden />
                </button>
              )}
            </li>
          ))}
        </ul>
      ) : (
        <p className="mt-4 text-sm text-gray-500">{t("no_skills")}</p>
      )}

      {canEdit && (
        <form
          className="mt-5 flex flex-wrap gap-2"
          onSubmit={(e) => {
            e.preventDefault();
            if (pick) add.mutate(pick);
          }}
        >
          <select
            aria-label={t("add_skill")}
            value={pick}
            onChange={(e) => setPick(e.target.value)}
            className="h-10 min-w-64 rounded-lg border border-gray-300 px-3 text-sm shadow-sm focus:border-primary-500 focus:ring-primary-500"
          >
            <option value="">{t("choose_skill")}</option>
            {available.map((s) => (
              <option key={s.id} value={s.id}>
                {s.name}
              </option>
            ))}
          </select>
          <Button
            type="submit"
            variant="primary"
            disabled={!pick || add.isPending}
          >
            <Plus className="mr-1 size-4" aria-hidden />
            {t("add_skill")}
          </Button>
        </form>
      )}
    </section>
  );
}
