import { useTranslation } from "react-i18next";

import { formatId } from "@/components/Patient/patientFormUtils";

/** "NIF 123 456 789 · SNS 987 654 321" under a patient's name in searches. */
export default function PatientIds({
  nif,
  sns,
}: {
  nif?: string | null;
  sns?: string | null;
}) {
  const { t } = useTranslation();
  if (!nif && !sns) return null;
  return (
    <span className="block text-xs text-gray-500">
      {[
        nif && `${t("nif_short")} ${formatId(nif)}`,
        sns && `${t("sns_short")} ${formatId(sns)}`,
      ]
        .filter(Boolean)
        .join(" · ")}
    </span>
  );
}
