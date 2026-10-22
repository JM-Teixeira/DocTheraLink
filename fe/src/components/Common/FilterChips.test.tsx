import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { FilterChips } from "@/components/Common/FilterChips";

const options = [
  { value: 1, label: "One" },
  { value: 2, label: "Two" },
];

describe("FilterChips", () => {
  it("marks the chosen chips and 'all' when nothing is chosen", () => {
    const { rerender } = render(
      <FilterChips
        label="Level"
        allLabel="All"
        options={options}
        value={[]}
        onChange={vi.fn()}
      />,
    );
    expect(screen.getByRole("button", { name: "All" })).toHaveAttribute(
      "aria-pressed",
      "true",
    );
    rerender(
      <FilterChips
        label="Level"
        allLabel="All"
        options={options}
        value={[2]}
        onChange={vi.fn()}
      />,
    );
    expect(screen.getByRole("button", { name: "All" })).toHaveAttribute(
      "aria-pressed",
      "false",
    );
    expect(screen.getByRole("button", { name: "Two" })).toHaveAttribute(
      "aria-pressed",
      "true",
    );
  });

  it("toggles chips and clears with 'all'", async () => {
    const onChange = vi.fn();
    render(
      <FilterChips
        label="Level"
        allLabel="All"
        options={options}
        value={[1]}
        onChange={onChange}
      />,
    );
    await userEvent.click(screen.getByRole("button", { name: "Two" }));
    expect(onChange).toHaveBeenLastCalledWith([1, 2]);
    await userEvent.click(screen.getByRole("button", { name: "One" }));
    expect(onChange).toHaveBeenLastCalledWith([]);
    await userEvent.click(screen.getByRole("button", { name: "All" }));
    expect(onChange).toHaveBeenLastCalledWith([]);
  });

  it("is a labelled group", () => {
    render(
      <FilterChips
        label="Level"
        options={options}
        value={[]}
        onChange={vi.fn()}
      />,
    );
    expect(screen.getByRole("group", { name: "Level" })).toBeInTheDocument();
  });
});
