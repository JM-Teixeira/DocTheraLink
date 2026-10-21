import { Link, usePath } from "raviger";
import { useState } from "react";
import { useTranslation } from "react-i18next";

import CareIcon from "@/CAREUI/icons/CareIcon";

import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { Button } from "@/components/ui/button";

import useAppHistory from "@/hooks/useAppHistory";

import { classNames } from "@/Utils/utils";

/** Path segments shown with a translated name. */
const SEGMENT_KEYS: Record<string, string> = {
  facility: "clinic",
  patients: "patients",
  patient: "patients",
  referrals: "referrals",
  "time-board": "time_board",
  inventory: "inventory",
  users: "users",
  user: "users",
  profile: "profile",
  skills: "skills",
  facilities: "USERMANAGEMENT_TAB__FACILITIES",
  new: "new",
  edit: "edit",
  letter: "letter",
  notifications: "notifications",
};

/** Longer paths collapse their middle into "•••". */
const COLLAPSE_AFTER = 4;

const looksLikeId = (s: string) =>
  /^[0-9a-f]{8}-[0-9a-f]{4}-/i.test(s) || /^\d+$/.test(s);

const capitalize = (s: string) =>
  s
    .replace(/[_-]/g, " ")
    .split(" ")
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");

interface BreadcrumbsProps {
  replacements?: {
    [key: string]: { name?: string; uri?: string; style?: string };
  };
  className?: string;
  hideBack?: boolean;
  backUrl?: string;
  onBackClick?: () => boolean | void;
}

/**
 * Back button and the path as links: Home › Utentes › Maria Silva › Editar.
 * Ids are replaced by names when the page gives them (replacements).
 */
export default function Breadcrumbs({
  replacements = {},
  className = "",
  hideBack = false,
  backUrl,
  onBackClick,
}: BreadcrumbsProps) {
  const { t } = useTranslation();
  const { goBack } = useAppHistory();
  const path = usePath() ?? "";
  const [expanded, setExpanded] = useState(false);

  const segments = path.split("/").filter(Boolean);
  const crumbs = segments.map((field, i) => ({
    name:
      replacements[field]?.name ||
      (SEGMENT_KEYS[field]
        ? t(SEGMENT_KEYS[field])
        : looksLikeId(field)
          ? t("details")
          : capitalize(field)),
    uri: replacements[field]?.uri || "/" + segments.slice(0, i + 1).join("/"),
    style: replacements[field]?.style || "",
  }));
  const collapse = !expanded && crumbs.length > COLLAPSE_AFTER;
  const shown = collapse ? [crumbs[0], null, ...crumbs.slice(-2)] : crumbs;

  return (
    <Breadcrumb>
      <nav className={classNames("w-full", className)} aria-label="Breadcrumb">
        <BreadcrumbList>
          <ol className="flex flex-wrap items-center">
            {!hideBack && (
              <li className="mr-3 flex items-center">
                <Button
                  variant="link"
                  className="rounded bg-gray-200/50 px-1 text-sm font-normal text-gray-800 transition hover:bg-gray-200/75 hover:no-underline"
                  size="xs"
                  onClick={() => {
                    if (onBackClick && onBackClick() === false) return;
                    goBack(backUrl);
                  }}
                >
                  <CareIcon icon="l-arrow-left" className="h-5 text-gray-700" />
                  <span className="pr-2">{t("back")}</span>
                </Button>
              </li>
            )}
            <BreadcrumbItem>
              <Button
                asChild
                variant="link"
                className="p-1 font-normal text-gray-800 hover:text-gray-700"
              >
                <Link href="/">{t("home")}</Link>
              </Button>
            </BreadcrumbItem>
            {shown.map((crumb, i) => (
              <li
                key={crumb ? crumb.uri : "collapsed"}
                className={classNames(
                  "flex items-center text-sm font-normal",
                  crumb?.style ?? "",
                )}
              >
                <BreadcrumbSeparator className="text-gray-400" />
                {crumb === null ? (
                  <Button
                    variant="link"
                    className="h-auto p-1 font-light text-gray-500 hover:text-gray-700"
                    onClick={() => setExpanded(true)}
                    aria-label={t("show_full_path")}
                  >
                    •••
                  </Button>
                ) : i === shown.length - 1 ? (
                  <span className="p-1 text-gray-600">{crumb.name}</span>
                ) : (
                  <Button
                    asChild
                    variant="link"
                    className="p-1 font-normal text-gray-800 hover:text-gray-700"
                  >
                    <Link href={crumb.uri}>{crumb.name}</Link>
                  </Button>
                )}
              </li>
            ))}
          </ol>
        </BreadcrumbList>
      </nav>
    </Breadcrumb>
  );
}
