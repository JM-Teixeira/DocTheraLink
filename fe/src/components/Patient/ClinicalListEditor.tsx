import { Plus, X } from "lucide-react";
import { useTranslation } from "react-i18next";

import { Button } from "@/components/ui/button";

/** One column of a list: a text cell, a year or a choice. */
export interface ListColumn {
  cell: string;
  label: string;
  placeholder?: string;
  options?: { value: string; label: string }[];
  year?: boolean;
  wide?: boolean;
}

const cellClass =
  "h-9 w-full rounded-md border border-gray-300 px-2 text-sm focus:border-primary-500 focus:ring-primary-500";

/**
 * Editable rows (allergies, medications, conditions). Rows are plain objects
 * of strings; empty rows are dropped when the form is saved.
 */
export default function ClinicalListEditor<Row extends Record<string, string>>({
  id,
  title,
  rows,
  empty,
  columns,
  addLabel,
  error,
  onChange,
}: {
  id: string;
  title: string;
  rows: Row[];
  empty: Row;
  columns: ListColumn[];
  addLabel: string;
  error?: string;
  onChange: (rows: Row[]) => void;
}) {
  const { t } = useTranslation();
  const setCell = (i: number, cell: string, v: string) =>
    onChange(rows.map((r, j) => (j === i ? { ...r, [cell]: v } : r)));

  return (
    <fieldset
      className="sm:col-span-2"
      aria-describedby={error ? `${id}-error` : undefined}
    >
      <legend className="text-sm font-medium text-gray-700">{title}</legend>
      {rows.length > 0 && (
        <ul className="mt-2 space-y-2" id={id}>
          {rows.map((r, i) => (
            <li key={i} className="flex flex-wrap items-center gap-2">
              {columns.map((c) => {
                const label = `${c.label} ${i + 1}`;
                return (
                  <span
                    key={c.cell}
                    className={
                      c.wide
                        ? "min-w-40 flex-[2]"
                        : c.year
                          ? "w-24"
                          : "min-w-28 flex-1"
                    }
                  >
                    {c.options ? (
                      <select
                        aria-label={label}
                        value={r[c.cell]}
                        onChange={(e) => setCell(i, c.cell, e.target.value)}
                        className={cellClass}
                      >
                        <option value="">{c.label}</option>
                        {c.options.map((o) => (
                          <option key={o.value} value={o.value}>
                            {o.label}
                          </option>
                        ))}
                      </select>
                    ) : (
                      <input
                        aria-label={label}
                        value={r[c.cell]}
                        inputMode={c.year ? "numeric" : undefined}
                        maxLength={c.year ? 4 : 120}
                        placeholder={c.placeholder ?? c.label}
                        onChange={(e) => setCell(i, c.cell, e.target.value)}
                        className={cellClass}
                      />
                    )}
                  </span>
                );
              })}
              <Button
                type="button"
                variant="ghost"
                size="icon"
                aria-label={`${t("remove")} ${i + 1}`}
                onClick={() => onChange(rows.filter((_, j) => j !== i))}
              >
                <X className="size-4" aria-hidden />
              </Button>
            </li>
          ))}
        </ul>
      )}
      <Button
        type="button"
        variant="outline"
        size="sm"
        className="mt-2"
        disabled={rows.length >= 30}
        onClick={() => onChange([...rows, { ...empty }])}
      >
        <Plus className="mr-1 size-4" aria-hidden />
        {addLabel}
      </Button>
      {error && (
        <p id={`${id}-error`} className="mt-1 text-xs text-rose-600">
          {error.startsWith("error__") ? t(error) : error}
        </p>
      )}
    </fieldset>
  );
}
