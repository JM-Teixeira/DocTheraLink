import { CircleHelp } from "lucide-react";
import { useTranslation } from "react-i18next";

/**
 * How automatic per-session use works, step by step with an example
 * (gloves and swim caps in the gym and in hydrotherapy).
 */
export default function PerSessionHelp({ open = false }: { open?: boolean }) {
  const { t } = useTranslation();
  const steps = [1, 2, 3, 4, 5, 6] as const;
  return (
    <details
      open={open}
      className="group rounded-lg border border-primary-100 bg-primary-50/50 p-3 text-sm text-gray-700"
    >
      <summary className="flex cursor-pointer list-none items-center gap-2 font-medium text-primary-800">
        <CircleHelp className="size-4" aria-hidden />
        {t("per_session_explain_title")}
      </summary>
      <ol className="mt-3 list-decimal space-y-1.5 pl-5">
        {steps.map((n) => (
          <li key={n}>{t(`per_session_explain_${n}`)}</li>
        ))}
      </ol>
      <div className="mt-3 overflow-x-auto">
        <table className="min-w-full text-xs">
          <caption className="mb-1 text-left font-medium text-gray-800">
            {t("per_session_example_title")}
          </caption>
          <thead className="text-left text-gray-500">
            <tr>
              <th className="py-1 pr-3">{t("item_name")}</th>
              <th className="py-1 pr-3">{t("per_session_default")}</th>
              <th className="py-1 pr-3">{t("per_session_example_gym")}</th>
              <th className="py-1">{t("per_session_example_pool")}</th>
            </tr>
          </thead>
          <tbody className="text-gray-700">
            <tr>
              <td className="py-1 pr-3">{t("per_session_example_gloves")}</td>
              <td className="py-1 pr-3">1</td>
              <td className="py-1 pr-3">
                {t("per_session_example_default_used")}
              </td>
              <td className="py-1">2</td>
            </tr>
            <tr>
              <td className="py-1 pr-3">{t("per_session_example_caps")}</td>
              <td className="py-1 pr-3">0</td>
              <td className="py-1 pr-3">0</td>
              <td className="py-1">1</td>
            </tr>
          </tbody>
        </table>
        <p className="mt-1 text-xs text-gray-600">
          {t("per_session_example_result")}
        </p>
      </div>
    </details>
  );
}
