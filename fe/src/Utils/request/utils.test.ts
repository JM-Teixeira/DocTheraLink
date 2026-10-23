import { describe, expect, it } from "vitest";

import { makeUrl } from "@/Utils/request/utils";

describe("makeUrl", () => {
  it("fills path params and skips undefined query values", () => {
    expect(
      makeUrl(
        "/api/v1/users/{username}/",
        { a: 1, b: undefined },
        { username: "ana" },
      ),
    ).toBe("/api/v1/users/ana/?a=1");
  });

  it("repeats the key for array values", () => {
    expect(
      makeUrl("/api/v1/patient/", { care_level: [1, 3], search: "x" }),
    ).toBe("/api/v1/patient/?care_level=1&care_level=3&search=x");
  });
});
