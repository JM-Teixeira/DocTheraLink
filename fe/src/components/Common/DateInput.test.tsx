import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
import { describe, expect, it, vi } from "vitest";

import DateInput from "@/components/Common/DateInput";
import { UserModel } from "@/components/Users/models";

import { AuthUserContext } from "@/hooks/useAuthUser";

// Signed in as someone with the given date preference
function renderWith(
  pref: UserModel["date_format"],
  props: Partial<React.ComponentProps<typeof DateInput>> = {},
) {
  const onChange = vi.fn();
  function Field() {
    const [value, setValue] = useState(props.value ?? "");
    return (
      <DateInput
        aria-label="date"
        {...props}
        value={value}
        onChange={(v) => {
          onChange(v);
          setValue(v);
        }}
      />
    );
  }
  render(
    <AuthUserContext.Provider
      value={{
        user: { date_format: pref } as UserModel,
        signIn: vi.fn(),
        signOut: vi.fn(),
        refetchUser: vi.fn(),
      }}
    >
      <Field />
    </AuthUserContext.Provider>,
  );
  return { onChange, input: screen.getByLabelText("date") };
}

describe("DateInput", () => {
  it("shows the value in the person's format", () => {
    expect(renderWith("dmy", { value: "2026-10-21" }).input).toHaveValue(
      "21/10/2026",
    );
  });

  it("follows another preference", () => {
    expect(renderWith("ymd", { value: "2026-10-21" }).input).toHaveValue(
      "2026-10-21",
    );
  });

  it("turns what is typed into an ISO date", async () => {
    const { input, onChange } = renderWith("mdy");
    await userEvent.type(input, "10/21/2026");
    await userEvent.tab();
    expect(onChange).toHaveBeenCalledWith("2026-10-21");
    expect(input).toHaveValue("10/21/2026");
  });

  it("refuses an impossible date and says so", async () => {
    const { input, onChange } = renderWith("dmy");
    await userEvent.type(input, "31/02/2026{Enter}");
    expect(onChange).not.toHaveBeenCalled();
    expect(input).toHaveAttribute("aria-invalid", "true");
    expect(screen.getByRole("alert")).toHaveTextContent("invalid_date_format");
  });

  it("respects min and max", async () => {
    const { input, onChange } = renderWith("dmy", {
      min: "2026-01-01",
      max: "2026-12-31",
    });
    await userEvent.type(input, "01/01/2027{Enter}");
    expect(onChange).not.toHaveBeenCalled();
  });

  it("clears to an empty value", async () => {
    const { input, onChange } = renderWith("dmy", { value: "2026-10-21" });
    await userEvent.clear(input);
    await userEvent.tab();
    expect(onChange).toHaveBeenCalledWith("");
  });
});
