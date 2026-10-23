import { describe, expect, it } from "vitest";

import { newPasswordProblem } from "@/components/Auth/ForcePasswordChange";

describe("newPasswordProblem", () => {
  it("asks for 8 characters, a different password and a matching confirmation", () => {
    expect(newPasswordProblem("Velha#2026", "curta", "curta")).toBe(
      "password_min_8",
    );
    expect(newPasswordProblem("Velha#2026", "Velha#2026", "Velha#2026")).toBe(
      "password_must_differ",
    );
    expect(newPasswordProblem("Velha#2026", "Nova#20260", "Nova#2026")).toBe(
      "passwords_do_not_match",
    );
    expect(newPasswordProblem("Velha#2026", "Nova#20260", "Nova#20260")).toBe(
      "",
    );
  });
});
