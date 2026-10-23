import { useQuery } from "@tanstack/react-query";
import { LogIn, LogOut, Plane, RotateCcw } from "lucide-react";
import { useTranslation } from "react-i18next";

import { cn } from "@/lib/utils";

import StaffRoutes from "@/components/Staff/api";
import { workSteps } from "@/components/Staff/staffUtils";

import useDateFormat from "@/hooks/useDateFormat";

import query from "@/Utils/request/query";

const STEP_STYLE = {
  joined: {
    icon: LogIn,
    tone: "bg-emerald-50 text-emerald-800 ring-emerald-200",
  },
  back: {
    icon: RotateCcw,
    tone: "bg-emerald-50 text-emerald-800 ring-emerald-200",
  },
  leave: { icon: Plane, tone: "bg-amber-50 text-amber-900 ring-amber-200" },
  left: { icon: LogOut, tone: "bg-rose-50 text-rose-800 ring-rose-200" },
} as const;

/**
 * A person's work story in one line: joined, leaves (from / until) and when
 * they left, so their situation is clear at a glance.
 */
export default function WorkTimeline({ username }: { username: string }) {
  const { t } = useTranslation();
  const { fmt } = useDateFormat();
  const { data } = useQuery({
    queryKey: ["employment-history", username],
    queryFn: query(StaffRoutes.employmentHistory, {
      pathParams: { username },
      silent: true,
    }),
  });
  if (!data) return null;
  return (
    <ol
      aria-label={t("work_history")}
      className="mt-3 flex flex-wrap items-center gap-2 text-xs"
    >
      {workSteps(data).map((s, i) => {
        const { icon: Icon, tone } = STEP_STYLE[s.kind];
        return (
          <li
            key={i}
            title={s.note || undefined}
            className={cn(
              "inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 font-medium ring-1 ring-inset",
              tone,
            )}
          >
            <Icon className="size-3.5" aria-hidden />
            {s.until
              ? t(`work_step__${s.kind}_until`, {
                  from: fmt(s.from),
                  to: fmt(s.until),
                })
              : t(`work_step__${s.kind}`, { date: fmt(s.from) })}
          </li>
        );
      })}
    </ol>
  );
}
