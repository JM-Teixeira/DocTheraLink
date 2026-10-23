import { ReactNode } from "react";

import { cn } from "@/lib/utils";

/** Classes of a filter chip, on or off. */
export const chipClass = (on: boolean) =>
  cn(
    "flex items-center gap-1.5 rounded-full px-3 py-1 text-sm ring-1 ring-inset transition",
    on
      ? "bg-primary-600 text-white ring-primary-600"
      : "bg-white text-gray-700 ring-gray-300 hover:bg-gray-50",
  );

export interface ChipOption<T> {
  value: T;
  label: ReactNode;
}

/**
 * A labelled row of chips. Single choice by default; with `multiple` each
 * chip toggles and an empty selection means "all".
 */
export function FilterChips<T extends string | number>({
  label,
  options,
  value,
  onChange,
  allLabel,
}: {
  label: string;
  options: ChipOption<T>[];
  value: T[];
  onChange: (v: T[]) => void;
  /** Adds an "all" chip that clears the selection. */
  allLabel?: string;
}) {
  const toggle = (v: T) =>
    onChange(value.includes(v) ? value.filter((x) => x !== v) : [...value, v]);
  return (
    <div
      className="flex flex-wrap items-center gap-2"
      role="group"
      aria-label={label}
    >
      <span className="w-24 shrink-0 text-xs font-medium uppercase tracking-wide text-gray-500">
        {label}
      </span>
      {allLabel && (
        <button
          type="button"
          aria-pressed={!value.length}
          className={chipClass(!value.length)}
          onClick={() => onChange([])}
        >
          {allLabel}
        </button>
      )}
      {options.map((o) => (
        <button
          key={o.value}
          type="button"
          aria-pressed={value.includes(o.value)}
          className={chipClass(value.includes(o.value))}
          onClick={() => toggle(o.value)}
        >
          {o.label}
        </button>
      ))}
    </div>
  );
}

/** The white card that holds a page's filters. */
export function FilterPanel({ children }: { children: ReactNode }) {
  return (
    <div className="mb-5 flex flex-col gap-3 rounded-xl border border-gray-200 bg-white p-4 shadow-sm">
      {children}
    </div>
  );
}
