import { describe, expect, it } from "vitest";

import { exportFileName } from "@/components/Patient/exportPatients";

describe("exportFileName", () => {
  it("uses the local date", () => {
    expect(exportFileName(new Date(2026, 0, 5))).toBe("utentes-2026-01-05.csv");
  });
});
