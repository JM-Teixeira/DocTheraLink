import { resolveApiUrl } from "@careConfig";
import { describe, expect, it } from "vitest";

describe("resolveApiUrl", () => {
  it("uses the page's own address for same-origin (works from any computer)", () => {
    expect(resolveApiUrl("same-origin", "http://jm-teixeira.local:8088")).toBe(
      "http://jm-teixeira.local:8088",
    );
    expect(resolveApiUrl(undefined, "http://localhost:8088")).toBe(
      "http://localhost:8088",
    );
  });

  it("keeps an explicit address", () => {
    expect(resolveApiUrl("https://api.klyniva.com", "http://x")).toBe(
      "https://api.klyniva.com",
    );
  });
});
