import { describe, expect, it } from "vitest";

import {
  MAX_FILE_BYTES,
  checkFile,
  formatBytes,
  uploadErrorKey,
} from "@/components/Referral/attachments";

describe("attachment helpers", () => {
  it("formats sizes", () => {
    expect(formatBytes(532)).toBe("532 B");
    expect(formatBytes(12_700)).toBe("12.4 KB");
    expect(formatBytes(3.1 * 1024 * 1024)).toBe("3.1 MB");
  });

  it("checks files before upload", () => {
    expect(checkFile({ size: 10, type: "application/pdf" })).toBeNull();
    expect(checkFile({ size: 10, type: "" })).toBeNull(); // unknown type: server decides
    expect(checkFile({ size: 0, type: "application/pdf" })).toBe(
      "attachment_error__empty",
    );
    expect(checkFile({ size: MAX_FILE_BYTES + 1, type: "image/png" })).toBe(
      "attachment_error__too_large",
    );
    expect(checkFile({ size: 10, type: "application/x-msdownload" })).toBe(
      "attachment_error__type",
    );
  });

  it("maps server statuses to messages", () => {
    expect(uploadErrorKey(413)).toBe("attachment_error__too_large");
    expect(uploadErrorKey(415)).toBe("attachment_error__type");
    expect(uploadErrorKey(403)).toBe("attachment_error__forbidden");
    expect(uploadErrorKey(500)).toBe("attachment_error__generic");
  });
});
