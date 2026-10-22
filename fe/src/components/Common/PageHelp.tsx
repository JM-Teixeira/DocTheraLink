import { CircleHelp } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

/**
 * Splits a help text into blocks: lines starting with "- " become a list,
 * other lines paragraphs. Help texts live in the locales as help__<topic>.
 */
export function helpBlocks(text: string) {
  const blocks: (
    | { kind: "p"; text: string }
    | { kind: "ul"; items: string[] }
  )[] = [];
  for (const raw of text.split("\n")) {
    const line = raw.trim();
    if (!line) continue;
    if (line.startsWith("- ")) {
      const last = blocks[blocks.length - 1];
      if (last?.kind === "ul") last.items.push(line.slice(2));
      else blocks.push({ kind: "ul", items: [line.slice(2)] });
    } else blocks.push({ kind: "p", text: line });
  }
  return blocks;
}

/**
 * The "?" next to a page title: what the page is for and what it can do, in
 * a pop-up, so the page itself stays short.
 */
export default function PageHelp({
  topic,
  title,
}: {
  topic: string;
  title: string;
}) {
  const { t } = useTranslation();
  const [open, setOpen] = useState(false);
  return (
    <>
      <Button
        variant="ghost"
        size="icon"
        className="size-8 rounded-full text-gray-500 hover:text-primary-700"
        aria-label={t("page_help_open", { page: title })}
        onClick={() => setOpen(true)}
      >
        <CircleHelp className="size-5" aria-hidden />
      </Button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>{title}</DialogTitle>
            <DialogDescription>{t("page_help_subtitle")}</DialogDescription>
          </DialogHeader>
          <div className="flex flex-col gap-3 text-sm text-gray-700">
            {helpBlocks(t(`help__${topic}`)).map((b, i) =>
              b.kind === "p" ? (
                <p key={i}>{b.text}</p>
              ) : (
                <ul key={i} className="list-disc space-y-1 pl-5">
                  {b.items.map((it) => (
                    <li key={it}>{it}</li>
                  ))}
                </ul>
              ),
            )}
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}
