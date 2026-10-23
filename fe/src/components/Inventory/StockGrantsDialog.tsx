import { useMutation } from "@tanstack/react-query";
import { Plus, Trash2 } from "lucide-react";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

import { FilterChips } from "@/components/Common/FilterChips";
import InventoryRoutes from "@/components/Inventory/api";
import { GrantRule, StockGrant } from "@/components/Inventory/types";

import * as Notification from "@/Utils/Notifications";
import { callApi } from "@/Utils/request/query";

/**
 * What a person may change in the stock: rules of floors × categories
 * ("Higiene on floors 2 and 3"). No floor chosen = every floor; no category
 * = every category; no rule = may only look. Coordinators and admins.
 */
export default function StockGrantsDialog({
  person,
  grants,
  floors,
  categories,
  onClose,
  onSaved,
}: {
  person: { id: number; name: string } | null;
  grants: StockGrant[];
  floors: number[];
  categories: string[];
  onClose: () => void;
  onSaved: () => void;
}) {
  const { t } = useTranslation();
  const [rules, setRules] = useState<GrantRule[]>([]);
  useEffect(() => {
    if (person)
      setRules(
        grants.map((g) => ({ floors: g.floors, categories: g.categories })),
      );
    // only when the dialog opens for someone
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [person?.id]);

  const save = useMutation({
    mutationFn: () =>
      callApi(InventoryRoutes.setGrants, {
        pathParams: { userId: person!.id },
        body: { rules },
        silent: true,
      }),
    onSuccess: () => {
      Notification.Success({ msg: t("stock_access_saved") });
      onSaved();
    },
    onError: () => Notification.Error({ msg: t("stock_access_save_failed") }),
  });
  const update = (i: number, patch: Partial<GrantRule>) =>
    setRules((rs) => rs.map((r, j) => (j === i ? { ...r, ...patch } : r)));
  // Categories of the rules that are not (or no longer) in the item list
  const allCategories = [
    ...new Set([...categories, ...rules.flatMap((r) => r.categories)]),
  ].sort();

  return (
    <Dialog open={!!person} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>
            {t("stock_access_for", { name: person?.name })}
          </DialogTitle>
          <DialogDescription>{t("stock_access_help")}</DialogDescription>
        </DialogHeader>

        {!rules.length ? (
          <p className="rounded-lg bg-gray-50 p-3 text-sm text-gray-600">
            {t("stock_access_none")}
          </p>
        ) : (
          <ol className="flex flex-col gap-3">
            {rules.map((r, i) => (
              <li key={i} className="rounded-lg border border-gray-200 p-3">
                <div className="mb-2 flex items-center justify-between">
                  <span className="text-sm font-medium text-gray-900">
                    {t("stock_access_rule_n", { n: i + 1 })}
                  </span>
                  <Button
                    size="icon"
                    variant="ghost"
                    aria-label={t("remove_rule")}
                    onClick={() =>
                      setRules((rs) => rs.filter((_, j) => j !== i))
                    }
                  >
                    <Trash2 className="size-4 text-rose-600" aria-hidden />
                  </Button>
                </div>
                <div className="flex flex-col gap-2">
                  <FilterChips
                    label={t("floor")}
                    allLabel={t("all_floors")}
                    options={floors.map((n) => ({
                      value: n,
                      label: t("floor_n", { n }),
                    }))}
                    value={r.floors}
                    onChange={(v) => update(i, { floors: v })}
                  />
                  <FilterChips
                    label={t("category")}
                    allLabel={t("all_categories")}
                    options={allCategories.map((c) => ({ value: c, label: c }))}
                    value={r.categories}
                    onChange={(v) => update(i, { categories: v })}
                  />
                </div>
              </li>
            ))}
          </ol>
        )}
        <Button
          variant="outline"
          className="self-start"
          onClick={() =>
            setRules((rs) => [...rs, { floors: [], categories: [] }])
          }
        >
          <Plus className="mr-1 size-4" aria-hidden />
          {t("add_rule")}
        </Button>

        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={onClose}>
            {t("cancel")}
          </Button>
          <Button
            variant="primary"
            disabled={save.isPending}
            onClick={() => save.mutate()}
          >
            {t("save")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
