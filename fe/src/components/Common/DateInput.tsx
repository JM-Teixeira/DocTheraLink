import { CalendarDays } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";

import { cn } from "@/lib/utils";

import useDateFormat from "@/hooks/useDateFormat";

/**
 * A date field in the person's own format (profile → preferences). Typing
 * works ("21/10/2026"); the calendar button opens the browser's picker.
 * value/onChange use ISO "YYYY-MM-DD" ("" = empty).
 */
export default function DateInput({
  value,
  onChange,
  min,
  max,
  className,
  id,
  required,
  "aria-label": ariaLabel,
}: {
  value: string;
  onChange: (iso: string) => void;
  min?: string;
  max?: string;
  className?: string;
  id?: string;
  required?: boolean;
  "aria-label"?: string;
}) {
  const { t } = useTranslation();
  const { fmt, parse, pattern } = useDateFormat();
  const picker = useRef<HTMLInputElement>(null);
  const [text, setText] = useState(fmt(value));
  const [invalid, setInvalid] = useState(false);

  // Follow outside changes (and a change of preference)
  useEffect(() => {
    setText(value ? fmt(value) : "");
    setInvalid(false);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [value, pattern]);

  const commit = (raw: string) => {
    if (!raw.trim()) {
      setInvalid(false);
      if (value) onChange("");
      return;
    }
    const iso = parse(raw);
    const outOfRange = !!iso && ((min && iso < min) || (max && iso > max));
    if (!iso || outOfRange) {
      setInvalid(true);
      return;
    }
    setInvalid(false);
    setText(fmt(iso));
    if (iso !== value) onChange(iso);
  };

  // A typed numeric date reads like the pattern: DD/MM/AAAA in Portuguese
  const placeholder = pattern
    .replace("YYYY", t("date_year_letters"))
    .replace("MM", t("date_month_letters"))
    .replace("DD", t("date_day_letters"));

  return (
    <div className={cn("relative", className)}>
      <input
        id={id}
        type="text"
        inputMode="numeric"
        aria-label={ariaLabel}
        aria-invalid={invalid || undefined}
        required={required}
        value={text}
        placeholder={placeholder}
        onChange={(e) => {
          setText(e.target.value);
          setInvalid(false);
        }}
        onBlur={(e) => commit(e.target.value)}
        onKeyDown={(e) => e.key === "Enter" && commit(text)}
        className={cn(
          "h-10 w-full rounded-lg border px-3 pr-10 text-sm shadow-sm focus:ring-primary-500",
          invalid
            ? "border-rose-400 focus:border-rose-500"
            : "border-gray-300 focus:border-primary-500",
        )}
      />
      <button
        type="button"
        aria-label={t("open_calendar")}
        onClick={() => {
          const el = picker.current;
          if (!el) return;
          try {
            el.showPicker();
          } catch {
            el.focus();
          }
        }}
        className="absolute inset-y-0 right-0 flex w-10 items-center justify-center text-gray-500 hover:text-gray-800"
      >
        <CalendarDays className="size-4" aria-hidden />
      </button>
      {/* The browser's calendar, kept out of sight; it only feeds the value */}
      <input
        ref={picker}
        type="date"
        tabIndex={-1}
        aria-hidden
        value={value}
        min={min}
        max={max}
        onChange={(e) => onChange(e.target.value)}
        className="pointer-events-none absolute bottom-0 right-0 h-0 w-0 opacity-0"
      />
      {invalid && (
        <p className="mt-1 text-xs text-rose-600" role="alert">
          {t("invalid_date_format", { pattern: placeholder })}
        </p>
      )}
    </div>
  );
}
