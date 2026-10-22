import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import PatientTimeline from "@/components/Patient/PatientTimeline";
import { PatientTimeline as Data } from "@/components/Patient/api/routes";

// Keys are shown as text (no i18n instance in unit tests).
const state = vi.hoisted(() => ({
  result: { data: undefined, isLoading: false, isError: false } as {
    data?: Data;
    isLoading: boolean;
    isError: boolean;
  },
}));
vi.mock("@/components/Patient/api", () => ({
  usePatientTimeline: () => state.result,
}));
vi.mock("@/hooks/useDateFormat", () => ({
  default: () => ({
    fmt: (d: string) => d,
    fmtTime: (d: string) => d.slice(0, 10),
  }),
}));

const session = (i: number) => ({
  at: `2099-01-0${i}T09:00:00Z`,
  kind: "session" as const,
  status: "booked",
  item_id: `s${i}`,
  date: `2099-01-0${i}`,
  hour: 9,
  who: "Rui",
});

const data: Data = {
  upcoming: [1, 2, 3, 4].map(session),
  history: [
    {
      at: "2026-10-12T09:00:00Z",
      kind: "session",
      status: "missed",
      date: "2026-10-12",
      hour: 9,
      justified: false,
      text: "doente",
    },
    {
      at: "2026-10-10T10:00:00Z",
      kind: "note",
      status: "progress",
      referral_id: "r1",
      text: "Melhor",
      pain: 4,
      who: "Sara",
    },
    {
      at: "2026-09-02T10:00:00Z",
      kind: "referral",
      status: "created",
      referral_id: "r1",
      text: "Lombalgia",
    },
    { at: "2026-09-01T10:00:00Z", kind: "registered" },
  ],
};

describe("PatientTimeline", () => {
  it("shows the next sessions, then the history by month", async () => {
    state.result = { data, isLoading: false, isError: false };
    render(<PatientTimeline patientId="p1" />);

    const region = screen.getByRole("region", { name: "patient_timeline" });
    expect(within(region).getAllByText(/2099-01-0\d 09:00/)).toHaveLength(3); // first 3
    await userEvent.click(screen.getByRole("button", { name: "show_all" }));
    expect(within(region).getAllByText(/2099-01-0\d 09:00/)).toHaveLength(4);

    expect(screen.getByText("timeline__session__missed")).toBeInTheDocument();
    expect(screen.getByText("(miss_unjustified)")).toBeInTheDocument();
    expect(screen.getByText("pain_n")).toBeInTheDocument();
    expect(
      screen.getByRole("link", { name: "timeline__note__progress" }),
    ).toHaveAttribute("href", "/referrals/r1");
    expect(
      screen.getAllByRole("heading", { level: 4 }).length,
    ).toBeGreaterThanOrEqual(3); // next + 2 months
  });

  it("filters by kind", async () => {
    state.result = { data, isLoading: false, isError: false };
    render(<PatientTimeline patientId="p1" />);
    await userEvent.click(
      screen.getByRole("button", { name: "timeline_filter__note (1)" }),
    );
    expect(screen.getByText("timeline__note__progress")).toBeInTheDocument();
    expect(
      screen.queryByText("timeline__referral__created"),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /timeline_filter__file/ }),
    ).not.toBeInTheDocument();
  });

  it("pages long histories and reports errors", async () => {
    const many = Array.from({ length: 35 }, (_, i) => ({
      at: `2026-08-${String((i % 28) + 1).padStart(2, "0")}T10:00:00Z`,
      kind: "file" as const,
      item_id: `f${i}`,
      text: `file-${i}.pdf`,
    }));
    state.result = {
      data: { upcoming: [], history: many },
      isLoading: false,
      isError: false,
    };
    const { unmount } = render(<PatientTimeline patientId="p1" />);
    expect(screen.getAllByText(/^file-\d+\.pdf$/)).toHaveLength(30);
    await userEvent.click(screen.getByRole("button", { name: "show_more" }));
    expect(screen.getAllByText(/^file-\d+\.pdf$/)).toHaveLength(35);
    unmount();

    state.result = { data: undefined, isLoading: false, isError: true };
    render(<PatientTimeline patientId="p1" />);
    expect(screen.getByText("timeline_failed")).toBeInTheDocument();
  });
});
