/**
 * Referral attachments: types, pure helpers (tested) and the two calls that
 * cannot go through callApi because they send/receive files.
 */
import careConfig from "@careConfig";

import { getAuthorizationHeader } from "@/Utils/request/utils";

export interface Attachment {
  id: string;
  referral_id: string;
  uploader_id: number;
  file_name: string;
  content_type: string;
  size: number;
  uploader?: { id: number; first_name: string; last_name: string };
  created_date: string;
}

/** Types the backend accepts (it re-checks the real content). */
export const ACCEPTED_FILE_TYPES =
  "application/pdf,image/jpeg,image/png,image/webp,text/plain";

/** Matches the backend default MAX_UPLOAD_SIZE (10 MB). */
export const MAX_FILE_BYTES = 10 * 1024 * 1024;

/** Human-readable size: 532 B, 12.4 KB, 3.1 MB. */
export function formatBytes(n: number) {
  if (n < 1024) return `${n} B`;
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`;
  return `${(n / (1024 * 1024)).toFixed(1)} MB`;
}

/** Translation key of the client-side check, or null if the file may be sent. */
export function checkFile(file: { size: number; type: string }) {
  if (file.size === 0) return "attachment_error__empty";
  if (file.size > MAX_FILE_BYTES) return "attachment_error__too_large";
  if (file.type && !ACCEPTED_FILE_TYPES.split(",").includes(file.type))
    return "attachment_error__type";
  return null;
}

/** Translation key for a failed upload's HTTP status. */
export function uploadErrorKey(status: number) {
  if (status === 413) return "attachment_error__too_large";
  if (status === 415) return "attachment_error__type";
  if (status === 403) return "attachment_error__forbidden";
  return "attachment_error__generic";
}

function authHeaders() {
  const h = new Headers();
  const auth = getAuthorizationHeader();
  if (auth) h.set("Authorization", auth);
  return h;
}

/** Uploads one file (multipart). Throws an Error whose message is a translation key. */
export async function uploadAttachment(referralId: string, file: File) {
  const body = new FormData();
  body.append("file", file);
  const res = await fetch(
    `${careConfig.apiUrl}/api/v1/referrals/${referralId}/attachments/`,
    { method: "POST", headers: authHeaders(), body },
  );
  if (!res.ok) throw new Error(uploadErrorKey(res.status));
  return (await res.json()) as Attachment;
}

/**
 * Downloads with the auth header and saves through a temporary link (a plain
 * <a href> cannot send the bearer token).
 */
export async function downloadAttachment(a: Attachment) {
  const res = await fetch(
    `${careConfig.apiUrl}/api/v1/attachments/${a.id}/download/`,
    { headers: authHeaders() },
  );
  if (!res.ok) throw new Error("attachment_error__generic");
  const url = URL.createObjectURL(await res.blob());
  const link = document.createElement("a");
  link.href = url;
  link.download = a.file_name;
  link.click();
  setTimeout(() => URL.revokeObjectURL(url), 1000);
}
