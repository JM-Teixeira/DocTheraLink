import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
import { describe, expect, it, vi } from "vitest";

import ClinicalListEditor from "@/components/Patient/ClinicalListEditor";

type Row = { substance: string; severity: string };
const empty: Row = { substance: "", severity: "" };

function Editor({
  error,
  onRows,
}: {
  error?: string;
  onRows: (r: Row[]) => void;
}) {
  const [rows, setRows] = useState<Row[]>([]);
  return (
    <ClinicalListEditor
      id="allergy_list"
      title="Allergies"
      rows={rows}
      empty={empty}
      addLabel="Add"
      error={error}
      columns={[
        { cell: "substance", label: "Substance", wide: true },
        {
          cell: "severity",
          label: "Severity",
          options: [{ value: "severe", label: "Severe" }],
        },
      ]}
      onChange={(r) => {
        setRows(r);
        onRows(r);
      }}
    />
  );
}

describe("ClinicalListEditor", () => {
  it("adds, fills and removes rows", async () => {
    const onRows = vi.fn();
    render(<Editor onRows={onRows} />);
    await userEvent.click(screen.getByRole("button", { name: "Add" }));
    await userEvent.type(screen.getByLabelText("Substance 1"), "Pó");
    await userEvent.selectOptions(
      screen.getByLabelText("Severity 1"),
      "severe",
    );
    expect(onRows).toHaveBeenLastCalledWith([
      { substance: "Pó", severity: "severe" },
    ]);

    await userEvent.click(screen.getByRole("button", { name: "Add" }));
    expect(screen.getByLabelText("Substance 2")).toHaveValue("");
    await userEvent.click(screen.getByRole("button", { name: "remove 1" }));
    expect(onRows).toHaveBeenLastCalledWith([empty]);
  });

  it("shows a translated error", () => {
    render(<Editor onRows={vi.fn()} error="error__invalid_year" />);
    expect(screen.getByText("error__invalid_year")).toBeInTheDocument();
  });
});
