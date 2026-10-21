import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import dayjs from "dayjs";
import {
  Download,
  FileImage,
  FileText,
  Loader2,
  Paperclip,
  Trash2,
  Upload,
} from "lucide-react";
import { useRef, useState } from "react";
import { useTranslation } from "react-i18next";

import { cn } from "@/lib/utils";

import { Button } from "@/components/ui/button";

import ReferralRoutes from "@/components/Referral/api";
import {
  ACCEPTED_FILE_TYPES,
  Attachment,
  checkFile,
  downloadAttachment,
  formatBytes,
  uploadAttachment,
} from "@/components/Referral/attachments";
import {
  CurrentUser,
  canPostNote,
  isAdmin,
} from "@/components/Referral/referralUtils";
import { Referral } from "@/components/Referral/types";

import * as Notification from "@/Utils/Notifications";
import query, { callApi } from "@/Utils/request/query";

/** Files attached to a referral: list, download, upload (drag & drop) and delete. */
export default function ReferralAttachments({
  referral,
  user,
}: {
  referral: Referral;
  user: CurrentUser;
}) {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const inputRef = useRef<HTMLInputElement>(null);
  const [dragging, setDragging] = useState(false);
  const canUpload = canPostNote(referral, user); // participants only
  const key = ["referral-attachments", referral.id];

  const { data: files, isLoading } = useQuery({
    queryKey: key,
    queryFn: query(ReferralRoutes.listAttachments, {
      pathParams: { id: referral.id },
    }),
  });

  const upload = useMutation({
    mutationFn: (file: File) => uploadAttachment(referral.id, file),
    onSuccess: (a) => {
      Notification.Success({
        msg: t("attachment_uploaded", { name: a.file_name }),
      });
      queryClient.invalidateQueries({ queryKey: key });
    },
    onError: (e: Error) => Notification.Error({ msg: t(e.message) }),
  });

  const remove = useMutation({
    mutationFn: (id: string) =>
      callApi(ReferralRoutes.deleteAttachment, { pathParams: { id } }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: key }),
    onError: () => Notification.Error({ msg: t("attachment_error__generic") }),
  });

  const send = (list: FileList | null) => {
    for (const file of Array.from(list ?? [])) {
      const problem = checkFile(file);
      if (problem) Notification.Error({ msg: `${file.name}: ${t(problem)}` });
      else upload.mutate(file);
    }
  };

  const download = (a: Attachment) =>
    downloadAttachment(a).catch(() =>
      Notification.Error({ msg: t("attachment_error__generic") }),
    );

  return (
    <section className="rounded-xl border border-gray-200 bg-white p-5 shadow-sm">
      <h3 className="mb-4 flex items-center gap-2 text-base font-semibold text-gray-900">
        <Paperclip className="size-4 text-gray-500" aria-hidden />
        {t("attachments")}
      </h3>

      {isLoading ? (
        <p className="text-sm text-gray-500">{t("loading")}</p>
      ) : files?.length ? (
        <ul className="mb-4 divide-y divide-gray-100 rounded-lg border border-gray-100">
          {files.map((a) => {
            const Icon = a.content_type.startsWith("image/")
              ? FileImage
              : FileText;
            const canDelete = a.uploader_id === user.id || isAdmin(user);
            return (
              <li key={a.id} className="flex items-center gap-3 p-3">
                <Icon
                  className="size-5 shrink-0 text-primary-600"
                  aria-hidden
                />
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-medium text-gray-900">
                    {a.file_name}
                  </p>
                  <p className="text-xs text-gray-500">
                    {formatBytes(a.size)}
                    {a.uploader &&
                      ` · ${a.uploader.first_name} ${a.uploader.last_name}`}
                    {` · ${dayjs(a.created_date).format("D MMM YYYY")}`}
                  </p>
                </div>
                <Button
                  variant="ghost"
                  size="icon"
                  aria-label={t("download_file", { name: a.file_name })}
                  onClick={() => download(a)}
                >
                  <Download className="size-4" aria-hidden />
                </Button>
                {canDelete && (
                  <Button
                    variant="ghost"
                    size="icon"
                    aria-label={t("delete_file", { name: a.file_name })}
                    disabled={remove.isPending}
                    onClick={() => {
                      if (
                        window.confirm(
                          t("confirm_delete_file", { name: a.file_name }),
                        )
                      )
                        remove.mutate(a.id);
                    }}
                  >
                    <Trash2 className="size-4 text-rose-600" aria-hidden />
                  </Button>
                )}
              </li>
            );
          })}
        </ul>
      ) : (
        <p className="mb-4 text-sm text-gray-500">{t("no_attachments")}</p>
      )}

      {canUpload && (
        <div
          onDragOver={(e) => {
            e.preventDefault();
            setDragging(true);
          }}
          onDragLeave={() => setDragging(false)}
          onDrop={(e) => {
            e.preventDefault();
            setDragging(false);
            send(e.dataTransfer.files);
          }}
          className={cn(
            "flex flex-col items-center gap-2 rounded-lg border-2 border-dashed p-5 text-center transition",
            dragging ? "border-primary-500 bg-primary-50" : "border-gray-200",
          )}
        >
          {upload.isPending ? (
            <Loader2
              className="size-5 animate-spin text-primary-600"
              aria-hidden
            />
          ) : (
            <Upload className="size-5 text-gray-400" aria-hidden />
          )}
          <p className="text-sm text-gray-600">{t("drop_files_here")}</p>
          <input
            ref={inputRef}
            type="file"
            multiple
            accept={ACCEPTED_FILE_TYPES}
            className="hidden"
            onChange={(e) => {
              send(e.target.files);
              e.target.value = "";
            }}
          />
          <Button
            type="button"
            variant="outline"
            size="sm"
            disabled={upload.isPending}
            onClick={() => inputRef.current?.click()}
          >
            {t("choose_files")}
          </Button>
          <p className="text-xs text-gray-400">{t("attachment_limits")}</p>
        </div>
      )}
    </section>
  );
}
