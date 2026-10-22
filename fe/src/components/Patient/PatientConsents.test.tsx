import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import PatientConsents from "@/components/Patient/PatientConsents";
import { ConsentView } from "@/components/Patient/api/routes";

// Texts are translation keys (no i18n instance in unit tests).
const m = vi.hoisted(() => ({
  view: undefined as ConsentView | undefined,
  mutate: vi.fn(),
}));
vi.mock("@/components/Patient/api", () => ({
  usePatientConsents: () => ({ data: m.view, isLoading: false }),
  useRecordConsent: () => ({ mutate: m.mutate, isPending: false }),
}));
vi.mock("@/hooks/useDateFormat", () => ({
  default: () => ({
    fmt: (d: string) => d.slice(0, 10),
    fmtTime: (d: string) => d,
  }),
}));

const sms = {
  id: "c1",
  patient_id: "p1",
  kind: "contact_sms" as const,
  granted: false,
  method: "verbal" as const,
  note: "pediu",
  recorded_by: { id: 1, first_name: "Rita", last_name: "Lopes" },
  created_date: "2026-10-20T10:00:00Z",
};

beforeEach(() => vi.clearAllMocks());

describe("PatientConsents", () => {
  it("warns when health-data consent is missing and records a change", async () => {
    m.view = { current: { contact_sms: sms }, history: [sms] };
    render(<PatientConsents patientId="p1" canRecord />);
    expect(screen.getByRole("alert")).toHaveTextContent(
      "consent_data_processing_missing",
    );
    expect(screen.getByText("consent_status__withdrawn")).toBeInTheDocument();
    expect(screen.getAllByText("consent_status__none")).toHaveLength(4);
    expect(
      screen.getByText("2026-10-20 · consent_method__verbal · Rita Lopes"),
    ).toBeInTheDocument();

    await userEvent.click(
      screen.getByRole("button", { name: /show_consent_history/ }),
    );
    expect(screen.getByText(/pediu/)).toBeInTheDocument();

    await userEvent.click(
      screen.getByRole("button", {
        name: "record_consent: consent__data_processing",
      }),
    );
    const dialog = screen.getByRole("dialog");
    await userEvent.selectOptions(
      within(dialog).getByLabelText("consent_method"),
      "digital",
    );
    await userEvent.click(within(dialog).getByRole("button", { name: "save" }));
    expect(m.mutate.mock.calls[0][0]).toEqual({
      kind: "data_processing",
      granted: true, // nothing recorded yet: defaults to giving it
      method: "digital",
      note: "",
    });
  });

  it("read-only users only see the state", () => {
    m.view = {
      current: {
        data_processing: { ...sms, kind: "data_processing", granted: true },
      },
      history: [],
    };
    render(<PatientConsents patientId="p1" canRecord={false} />);
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /record_consent/ }),
    ).not.toBeInTheDocument();
  });
});
