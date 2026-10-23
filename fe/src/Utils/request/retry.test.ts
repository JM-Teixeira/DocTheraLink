import { describe, expect, it } from "vitest";

import { isTransient, retryDelay, retryTransient } from "@/Utils/request/retry";
import { HTTPError, NetworkError } from "@/Utils/request/types";

const http = (status: number) =>
  new HTTPError({ message: "x", status, silent: true });

describe("retry on network errors", () => {
  it("retries only failures the server most likely never handled", () => {
    expect(isTransient(new NetworkError())).toBe(true);
    expect(isTransient(http(503))).toBe(true);
    expect(isTransient(http(500))).toBe(false);
    expect(isTransient(http(409))).toBe(false);
    expect(isTransient(new Error("boom"))).toBe(false);
  });

  it("stops after two more tries, waiting longer each time", () => {
    expect(retryTransient(0, new NetworkError())).toBe(true);
    expect(retryTransient(1, new NetworkError())).toBe(true);
    expect(retryTransient(2, new NetworkError())).toBe(false);
    expect([0, 1, 2, 5].map(retryDelay)).toEqual([1000, 2000, 4000, 10000]);
  });
});
