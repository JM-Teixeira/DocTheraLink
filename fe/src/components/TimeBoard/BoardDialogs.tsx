import { useMutation } from "@tanstack/react-query";
import { Plus, Trash2 } from "lucide-react";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";

import { cn } from "@/lib/utils";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

import DateInput from "@/components/Common/DateInput";
import BoardRoutes from "@/components/TimeBoard/api";
import {
  COLOR_NAMES,
  DIVISION_COLORS,
  addDays,
  formatHour,
} from "@/components/TimeBoard/boardUtils";
import {
  Board,
  Division,
  DivisionColor,
  Shift,
} from "@/components/TimeBoard/types";

import * as Notification from "@/Utils/Notifications";
import { callApi } from "@/Utils/request/query";
import { HTTPError } from "@/Utils/request/types";

const fieldClass =
  "mt-1 h-10 w-full rounded-lg border border-gray-300 px-3 text-sm shadow-sm focus:border-primary-500 focus:ring-primary-500";

/** The API's explanation ("this hour is full…"), or a generic message. */
export function apiError(e: Error, fallback: string) {
  const msg = e instanceof HTTPError ? e.cause?.message : undefined;
  Notification.Error({ msg: typeof msg === "string" && msg ? msg : fallback });
}

/** Creates or edits a shift (who, where, from–to). */
export function ShiftDialog({
  board,
  shift,
  preset,
  onClose,
  onChanged,
}: {
  board: Board;
  shift: Shift | null;
  /** New shift for this physio starting at this hour (cell click). */
  preset: { therapistId: number; hour: number } | null;
  onClose: () => void;
  onChanged: () => void;
}) {
  const { t } = useTranslation();
  const open = !!shift || !!preset;
  const [form, setForm] = useState({
    therapist: "",
    division: "",
    start: 9,
    end: 13,
    notes: "",
  });
  useEffect(() => {
    if (shift)
      setForm({
        therapist: String(shift.therapist_id),
        division: shift.division_id,
        start: shift.start_hour,
        end: shift.end_hour,
        notes: shift.notes,
      });
    else if (preset)
      setForm({
        therapist: String(preset.therapistId),
        division: board.divisions[0]?.id ?? "",
        start: preset.hour,
        end: Math.min(preset.hour + 4, 24),
        notes: "",
      });
  }, [shift, preset, board.divisions]);
  const hours = Array.from({ length: 25 }, (_, i) => i);
  const fail = (e: Error) => apiError(e, t("board_save_failed"));

  const save = useMutation({
    mutationFn: () => {
      const body = {
        date: board.date,
        therapist_id: Number(form.therapist),
        division_id: form.division,
        start_hour: form.start,
        end_hour: form.end,
        notes: form.notes,
      };
      return shift
        ? callApi(BoardRoutes.updateShift, {
            pathParams: { id: shift.id },
            body,
            silent: true,
          })
        : callApi(BoardRoutes.createShift, { body, silent: true });
    },
    onSuccess: () => {
      onChanged();
      onClose();
    },
    onError: fail,
  });
  const remove = useMutation({
    mutationFn: () =>
      callApi(BoardRoutes.deleteShift, {
        pathParams: { id: shift!.id },
        silent: true,
      }),
    onSuccess: () => {
      onChanged();
      onClose();
    },
    onError: fail,
  });
  const valid = form.therapist && form.division && form.end > form.start;

  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{shift ? t("edit_shift") : t("new_shift")}</DialogTitle>
          <DialogDescription>{board.date}</DialogDescription>
        </DialogHeader>
        <div className="grid gap-3">
          <label className="text-sm font-medium text-gray-700">
            {t("physiotherapist")}
            <select
              value={form.therapist}
              onChange={(e) =>
                setForm((f) => ({ ...f, therapist: e.target.value }))
              }
              className={fieldClass}
            >
              <option value="">-</option>
              {board.therapists.map((th) => (
                <option key={th.id} value={th.id}>
                  {th.first_name} {th.last_name}
                </option>
              ))}
            </select>
          </label>
          <label className="text-sm font-medium text-gray-700">
            {t("division")}
            <select
              value={form.division}
              onChange={(e) =>
                setForm((f) => ({ ...f, division: e.target.value }))
              }
              className={fieldClass}
            >
              {board.divisions.map((d) => (
                <option key={d.id} value={d.id}>
                  {d.name}
                </option>
              ))}
            </select>
          </label>
          <div className="grid grid-cols-2 gap-3">
            <label className="text-sm font-medium text-gray-700">
              {t("from")}
              <select
                value={form.start}
                onChange={(e) =>
                  setForm((f) => ({ ...f, start: Number(e.target.value) }))
                }
                className={fieldClass}
              >
                {hours.slice(0, 24).map((h) => (
                  <option key={h} value={h}>
                    {formatHour(h)}
                  </option>
                ))}
              </select>
            </label>
            <label className="text-sm font-medium text-gray-700">
              {t("until")}
              <select
                value={form.end}
                onChange={(e) =>
                  setForm((f) => ({ ...f, end: Number(e.target.value) }))
                }
                className={fieldClass}
              >
                {hours.slice(1).map((h) => (
                  <option key={h} value={h}>
                    {formatHour(h)}
                  </option>
                ))}
              </select>
            </label>
          </div>
          <label className="text-sm font-medium text-gray-700">
            {t("note_optional")}
            <input
              value={form.notes}
              maxLength={200}
              onChange={(e) =>
                setForm((f) => ({ ...f, notes: e.target.value }))
              }
              className={fieldClass}
            />
          </label>
        </div>
        <DialogFooter className="mt-2 gap-2">
          {shift && (
            <Button
              type="button"
              variant="ghost"
              className="mr-auto text-rose-600"
              onClick={() => remove.mutate()}
              disabled={remove.isPending}
            >
              <Trash2 className="mr-1 size-4" aria-hidden />
              {t("delete")}
            </Button>
          )}
          <Button type="button" variant="outline" onClick={onClose}>
            {t("cancel")}
          </Button>
          <Button
            type="button"
            variant="primary"
            disabled={!valid || save.isPending}
            onClick={() => save.mutate()}
          >
            {t("save")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

type DivisionDraft = {
  name: string;
  color: DivisionColor;
  capacity: number;
  credit_limit: number;
};
const EMPTY_DIVISION: DivisionDraft = {
  name: "",
  color: "teal",
  capacity: 3,
  credit_limit: 8,
};

/**
 * Add, rename, recolour or remove divisions and set their limits: patients
 * per physio per hour, and care credits per physio per hour (patients count
 * their care level 1–5).
 */
export function DivisionsDialog({
  open,
  divisions,
  onClose,
  onChanged,
}: {
  open: boolean;
  divisions: Division[];
  onClose: () => void;
  onChanged: () => void;
}) {
  const { t } = useTranslation();
  const [draft, setDraft] = useState<DivisionDraft>(EMPTY_DIVISION);
  const fail = (e: Error) => apiError(e, t("board_save_failed"));
  const save = useMutation({
    mutationFn: (v: DivisionDraft & { id?: string; sort_order: number }) =>
      v.id
        ? callApi(BoardRoutes.updateDivision, {
            pathParams: { id: v.id },
            body: v,
            silent: true,
          })
        : callApi(BoardRoutes.createDivision, { body: v, silent: true }),
    onSuccess: () => {
      setDraft(EMPTY_DIVISION);
      onChanged();
    },
    onError: fail,
  });
  const remove = useMutation({
    mutationFn: (id: string) =>
      callApi(BoardRoutes.removeDivision, { pathParams: { id }, silent: true }),
    onSuccess: onChanged,
    onError: fail,
  });

  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>{t("divisions")}</DialogTitle>
          <DialogDescription>{t("divisions_help")}</DialogDescription>
        </DialogHeader>
        <ul className="divide-y divide-gray-100 rounded-lg border border-gray-200">
          {divisions.map((d, i) => (
            <DivisionRow
              key={d.id}
              division={d}
              onSave={(v) => save.mutate({ id: d.id, sort_order: i, ...v })}
              onRemove={() => {
                if (
                  window.confirm(t("confirm_remove_division", { name: d.name }))
                )
                  remove.mutate(d.id);
              }}
            />
          ))}
        </ul>
        <form
          className="flex flex-wrap items-end gap-2"
          onSubmit={(e) => {
            e.preventDefault();
            if (draft.name.trim().length >= 2)
              save.mutate({
                ...draft,
                name: draft.name.trim(),
                sort_order: divisions.length,
              });
          }}
        >
          <label className="min-w-40 flex-1 text-sm font-medium text-gray-700">
            {t("new_division")}
            <input
              value={draft.name}
              onChange={(e) =>
                setDraft((d) => ({ ...d, name: e.target.value }))
              }
              className={fieldClass}
            />
          </label>
          <ColorPicker
            value={draft.color}
            onChange={(color) => setDraft((d) => ({ ...d, color }))}
          />
          <LimitInputs
            value={draft}
            onChange={(v) => setDraft((d) => ({ ...d, ...v }))}
            labelled
          />
          <Button
            type="submit"
            variant="primary"
            disabled={draft.name.trim().length < 2 || save.isPending}
          >
            <Plus className="mr-1 size-4" aria-hidden />
            {t("add")}
          </Button>
        </form>
      </DialogContent>
    </Dialog>
  );
}

function LimitInputs({
  value,
  onChange,
  labelled,
}: {
  value: { capacity: number; credit_limit: number };
  onChange: (v: { capacity?: number; credit_limit?: number }) => void;
  labelled?: boolean;
}) {
  const { t } = useTranslation();
  return (
    <>
      <label className="w-24 text-sm font-medium text-gray-700">
        {labelled ? (
          t("patients_per_hour")
        ) : (
          <span className="sr-only">{t("patients_per_hour")}</span>
        )}
        <input
          type="number"
          min={1}
          max={20}
          value={value.capacity}
          onChange={(e) => onChange({ capacity: Number(e.target.value) })}
          className={cn(fieldClass, !labelled && "mt-0")}
        />
      </label>
      <label className="w-24 text-sm font-medium text-gray-700">
        {labelled ? (
          t("credits_per_hour")
        ) : (
          <span className="sr-only">{t("credits_per_hour")}</span>
        )}
        <input
          type="number"
          min={1}
          max={40}
          value={value.credit_limit}
          onChange={(e) => onChange({ credit_limit: Number(e.target.value) })}
          className={cn(fieldClass, !labelled && "mt-0")}
        />
      </label>
    </>
  );
}

function DivisionRow({
  division,
  onSave,
  onRemove,
}: {
  division: Division;
  onSave: (v: DivisionDraft) => void;
  onRemove: () => void;
}) {
  const { t } = useTranslation();
  const [v, setV] = useState<DivisionDraft>({
    name: division.name,
    color: division.color,
    capacity: division.capacity,
    credit_limit: division.credit_limit,
  });
  const dirty =
    v.name !== division.name ||
    v.color !== division.color ||
    v.capacity !== division.capacity ||
    v.credit_limit !== division.credit_limit;
  return (
    <li className="flex flex-wrap items-center gap-2 px-3 py-2">
      <input
        aria-label={t("division")}
        value={v.name}
        onChange={(e) => setV((x) => ({ ...x, name: e.target.value }))}
        className={cn(fieldClass, "mt-0 min-w-40 flex-1")}
      />
      <ColorPicker
        value={v.color}
        onChange={(color) => setV((x) => ({ ...x, color }))}
      />
      <LimitInputs
        value={v}
        onChange={(lim) => setV((x) => ({ ...x, ...lim }))}
      />
      <Button
        size="sm"
        variant="outline"
        disabled={!dirty || v.name.trim().length < 2}
        onClick={() => onSave({ ...v, name: v.name.trim() })}
      >
        {t("save")}
      </Button>
      <Button
        size="icon"
        variant="ghost"
        aria-label={t("delete")}
        onClick={onRemove}
      >
        <Trash2 className="size-4 text-rose-600" aria-hidden />
      </Button>
    </li>
  );
}

function ColorPicker({
  value,
  onChange,
}: {
  value: DivisionColor;
  onChange: (c: DivisionColor) => void;
}) {
  const { t } = useTranslation();
  return (
    <div className="flex gap-1" role="radiogroup" aria-label={t("color")}>
      {COLOR_NAMES.map((c) => (
        <button
          key={c}
          type="button"
          role="radio"
          aria-checked={value === c}
          aria-label={c}
          onClick={() => onChange(c)}
          className={cn(
            "size-6 rounded-full ring-offset-2",
            DIVISION_COLORS[c].dot,
            value === c && "ring-2 ring-gray-900",
          )}
        />
      ))}
    </div>
  );
}

/** Copies this day's shifts onto another day (default: a week later). */
export function CopyDayDialog({
  open,
  date,
  onClose,
  onCopied,
}: {
  open: boolean;
  date: string;
  onClose: () => void;
  onCopied: (to: string) => void;
}) {
  const { t } = useTranslation();
  const [to, setTo] = useState(addDays(date, 7));
  useEffect(() => setTo(addDays(date, 7)), [date, open]);
  const copy = useMutation({
    mutationFn: () =>
      callApi(BoardRoutes.copyDay, {
        body: { from_date: date, to_date: to },
        silent: true,
      }),
    onSuccess: (r) => {
      Notification.Success({ msg: t("shifts_copied", { count: r.copied }) });
      onCopied(to);
    },
    onError: (e: Error) => apiError(e, t("board_save_failed")),
  });
  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="sm:max-w-sm">
        <DialogHeader>
          <DialogTitle>{t("copy_day")}</DialogTitle>
          <DialogDescription>{t("copy_day_help", { date })}</DialogDescription>
        </DialogHeader>
        <label className="text-sm font-medium text-gray-700">
          {t("copy_to")}
          <DateInput value={to} onChange={setTo} className="mt-1" />
        </label>
        <DialogFooter className="gap-2">
          <Button type="button" variant="outline" onClick={onClose}>
            {t("cancel")}
          </Button>
          <Button
            type="button"
            variant="primary"
            disabled={!to || to === date || copy.isPending}
            onClick={() => copy.mutate()}
          >
            {t("copy")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
