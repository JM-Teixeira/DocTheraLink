import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { useTranslation } from "react-i18next";

import { cn } from "@/lib/utils";

import ReferralRoutes from "@/components/Referral/api";
import { catalogLang } from "@/components/Referral/referralUtils";
import { IcdCode } from "@/components/Referral/types";

import useDebouncedValue from "@/hooks/useDebouncedValue";

import query from "@/Utils/request/query";

/**
 * ICD-10 code field with suggestions: type a code ("M54") or words of the
 * diagnosis ("low back", "lombalgia"); pick with mouse or arrows + Enter.
 * Free text stays allowed for codes outside the curated list.
 */
export default function IcdPicker({
  id,
  value,
  onChange,
  onPick,
  className,
}: {
  id: string;
  value: string;
  onChange: (text: string) => void;
  onPick: (code: IcdCode) => void;
  className?: string;
}) {
  const { t, i18n } = useTranslation();
  const lang = catalogLang(i18n.language);
  const [open, setOpen] = useState(false);
  const [active, setActive] = useState(0);
  const search = useDebouncedValue(value.trim(), 200);

  const { data: options = [] } = useQuery({
    queryKey: ["icd10", lang, search],
    queryFn: query(ReferralRoutes.icd10, {
      queryParams: { search, lang, limit: 8 },
    }),
    enabled: open && search.length >= 2,
    staleTime: Infinity, // reference data
  });
  const shown = open && search.length >= 2 && options.length > 0;
  const listId = `${id}-options`;

  const pick = (c: IcdCode) => {
    onPick(c);
    setOpen(false);
  };

  return (
    <div className="relative">
      <input
        id={id}
        role="combobox"
        aria-expanded={shown}
        aria-controls={listId}
        aria-autocomplete="list"
        aria-activedescendant={shown ? `${listId}-${active}` : undefined}
        autoComplete="off"
        maxLength={32}
        value={value}
        placeholder={t("icd_code_placeholder")}
        className={className}
        onChange={(e) => {
          onChange(e.target.value);
          setActive(0);
          setOpen(true);
        }}
        onFocus={() => setOpen(true)}
        onBlur={() => setOpen(false)}
        onKeyDown={(e) => {
          if (!shown) return;
          if (e.key === "ArrowDown" || e.key === "ArrowUp") {
            e.preventDefault();
            const step = e.key === "ArrowDown" ? 1 : -1;
            setActive((a) => (a + step + options.length) % options.length);
          } else if (e.key === "Enter") {
            e.preventDefault(); // pick instead of submitting the form
            pick(options[Math.min(active, options.length - 1)]);
          } else if (e.key === "Escape") {
            setOpen(false);
          }
        }}
      />
      {shown && (
        <ul
          id={listId}
          role="listbox"
          className="absolute right-0 z-20 mt-1 max-h-72 w-80 overflow-auto rounded-lg border border-gray-200 bg-white py-1 shadow-lg sm:w-96"
        >
          {options.map((c, i) => (
            <li
              key={c.code}
              id={`${listId}-${i}`}
              role="option"
              aria-selected={i === active}
              // mousedown fires before the input's blur closes the list
              onMouseDown={(e) => {
                e.preventDefault();
                pick(c);
              }}
              onMouseEnter={() => setActive(i)}
              className={cn(
                "flex cursor-pointer items-baseline gap-3 px-3 py-2 text-sm",
                i === active ? "bg-primary-50" : "hover:bg-gray-50",
              )}
            >
              <span className="w-12 shrink-0 font-mono text-xs font-semibold text-primary-700">
                {c.code}
              </span>
              <span className="text-gray-800">{c.description}</span>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
