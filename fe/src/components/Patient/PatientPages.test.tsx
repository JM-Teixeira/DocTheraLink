import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import DuplicatePatientsDialog from "@/components/Patient/DuplicatePatientsDialog";
import PatientDetail from "@/components/Patient/PatientDetail";
import PatientIds from "@/components/Patient/PatientIds";
import PatientList from "@/components/Patient/PatientList";
import PatientMetrics from "@/components/Patient/PatientMetrics";
import { PatientModel } from "@/components/Patient/models";

// Data hooks, routing and page chrome are replaced; texts are translation keys.
const m = vi.hoisted(() => ({
  navigate: vi.fn(),
  download: vi.fn(async () => undefined),
  user: { id: 1, user_type: "Doctor" },
  patient: undefined as PatientModel | undefined,
  list: { data: undefined as unknown, isLoading: false, isError: false },
  lastListParams: {} as Record<string, unknown>,
}));
vi.mock("@/components/Patient/api", () => ({
  usePatient: () => ({
    data: m.patient,
    isLoading: false,
    isError: !m.patient,
  }),
  usePatientTimeline: () => ({
    data: { upcoming: [], history: [] },
    isLoading: false,
    isError: false,
  }),
  usePatientList: (p: Record<string, unknown>) => {
    m.lastListParams = p;
    return m.list;
  },
  usePatientStats: () => ({
    data: {
      total: 2,
      care_levels: { "1": 1, "3": 1 },
      genders: { "2": 2 },
      age_bands: { "65+": 1 },
      average_age: 61.5,
      monthly: [
        { month: "2026-09", count: 1 },
        { month: "2026-10", count: 3 },
      ],
    },
  }),
}));
vi.mock("@/components/Patient/exportPatients", () => ({
  downloadPatientsCsv: m.download,
}));
vi.mock("raviger", async (orig) => ({
  ...(await orig<object>()),
  navigate: m.navigate,
}));
vi.mock("@/components/Common/Page", () => ({
  default: ({
    children,
    options,
  }: {
    children: ReactNode;
    options?: ReactNode;
  }) => (
    <div>
      {options}
      {children}
    </div>
  ),
}));
vi.mock("@/hooks/useAuthUser", () => ({
  default: () => m.user,
  useAuthContext: () => ({ user: { date_format: "ymd" } }),
}));
vi.mock("@/Utils/request/query", () => ({
  default: () => async () => ({ results: [], count: 0 }),
  callApi: vi.fn(),
}));

const wrap = (ui: ReactNode) =>
  render(
    <QueryClientProvider client={new QueryClient()}>{ui}</QueryClientProvider>,
  );

beforeEach(() => {
  vi.clearAllMocks();
  m.user = { id: 1, user_type: "Doctor" };
});

describe("PatientDetail", () => {
  it("shows contact, emergency contact, insurance and the clinical lists", async () => {
    m.patient = {
      id: "p1",
      name: "Ana Lopes",
      address: "Rua A 1",
      postal_code: "4000-123",
      city: "Porto",
      emergency_contact_name: "Rui",
      emergency_contact_relation: "spouse",
      emergency_phone_number: "912345678",
      insurer: "ADSE",
      insurance_number: "123",
      allergy_list: [{ substance: "Penicilina", severity: "severe" }],
      condition_list: [{ name: "Asma", since: 2010 }],
      medication_list: [{ name: "Ibuprofeno", dose: "400 mg" }],
    };
    wrap(<PatientDetail id="p1" />);
    expect(screen.getByText("Ana Lopes")).toBeInTheDocument();
    expect(screen.getByText("allergic_to")).toBeInTheDocument();
    expect(screen.getByText("Rua A 1 4000-123 Porto")).toBeInTheDocument();
    expect(
      screen.getByText("Rui (relation__spouse) 912345678"),
    ).toBeInTheDocument();
    expect(screen.getByText("ADSE · 123")).toBeInTheDocument();
    expect(
      screen.getByText("Penicilina · severity__severe"),
    ).toBeInTheDocument();
    expect(screen.getByText("Asma (since_n)")).toBeInTheDocument();
    expect(screen.getByText("Ibuprofeno · 400 mg")).toBeInTheDocument();
    expect(
      await screen.findByText("no_referrals_for_patient"),
    ).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "edit" }));
    expect(m.navigate).toHaveBeenCalledWith("/patients/p1/edit");
    await userEvent.click(
      screen.getByRole("button", { name: "refer_to_physiotherapy" }),
    );
    expect(m.navigate).toHaveBeenCalledWith("/referrals/new?patient=p1");
  });

  it("says when the patient does not exist", () => {
    m.patient = undefined;
    wrap(<PatientDetail id="nope" />);
    expect(screen.getByText("patient_not_found")).toBeInTheDocument();
  });
});

describe("PatientList", () => {
  const page = (n: number, count = n) => ({
    data: {
      count,
      results: Array.from({ length: n }, (_, i) => ({
        id: `p${i}`,
        name: `Paciente ${i}`,
        care_level: 2,
        gender: 2,
        nif: i === 0 ? "123456789" : null,
        created_date: "2026-10-01T10:00:00Z",
      })),
    },
    isLoading: false,
    isError: false,
  });

  it("lists patients with metrics, filters and paging", async () => {
    m.list = page(20, 45);
    wrap(<PatientList />);
    expect(screen.getByText("Paciente 0")).toBeInTheDocument();
    expect(
      screen.getByRole("region", { name: "patient_metrics" }),
    ).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "next" }));
    expect(m.lastListParams).toMatchObject({ offset: 20 });
    await userEvent.click(screen.getByText("Paciente 1"));
    expect(m.navigate).toHaveBeenCalledWith("/patients/p1");

    await userEvent.click(screen.getByRole("button", { name: "new_patient" }));
    expect(m.navigate).toHaveBeenCalledWith("/patients/new");
  });

  it("explains an empty list and load errors", () => {
    m.list = page(0);
    const { unmount } = wrap(<PatientList />);
    expect(screen.getByText("no_patients_yet")).toBeInTheDocument();
    unmount();
    m.list = { data: undefined, isLoading: false, isError: true };
    wrap(<PatientList />);
    expect(screen.getByText("patients_load_error")).toBeInTheDocument();
  });
});

describe("PatientList for a coordinator", () => {
  it("exports the filtered list and shows the monthly trend", async () => {
    m.user = { id: 2, user_type: "Coordinator" };
    m.list = {
      data: { count: 1, results: [{ id: "p1", name: "Ana" }] },
      isLoading: false,
      isError: false,
    };
    wrap(<PatientList />);
    expect(
      screen.getByRole("list", { name: "registrations_per_month" }),
    ).toBeInTheDocument();
    await userEvent.type(screen.getByLabelText("diagnosis"), "lombalgia");
    await userEvent.click(screen.getByRole("button", { name: "export_csv" }));
    expect(m.download).toHaveBeenCalled();
  });

  it("hides the diagnosis filter and export from the front desk", () => {
    m.user = { id: 3, user_type: "Receptionist" };
    m.list = {
      data: { count: 0, results: [] },
      isLoading: false,
      isError: false,
    };
    wrap(<PatientList />);
    expect(screen.queryByLabelText("diagnosis")).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "export_csv" }),
    ).not.toBeInTheDocument();
    expect(screen.getByLabelText("physiotherapist")).toBeInTheDocument();
  });
});

describe("small pieces", () => {
  it("PatientIds shows what is known", () => {
    const { container } = render(<PatientIds />);
    expect(container).toBeEmptyDOMElement();
    render(<PatientIds nif="123456789" sns="987654321" />);
    expect(
      screen.getByText("nif_short 123 456 789 · sns_short 987 654 321"),
    ).toBeInTheDocument();
  });

  it("PatientMetrics waits for numbers", () => {
    const { container } = render(<PatientMetrics />);
    expect(container.querySelector(".animate-pulse")).not.toBeNull();
  });

  it("DuplicatePatientsDialog lets the user open a match or confirm another person", async () => {
    const onOpen = vi.fn();
    const onConfirm = vi.fn();
    render(
      <DuplicatePatientsDialog
        open
        blocking={false}
        pending={false}
        matches={[
          {
            id: "p1",
            name: "Ana Lopes",
            reason: "name_dob",
            blocking: false,
            date_of_birth: "1980-01-02",
          },
        ]}
        onOpenPatient={onOpen}
        onConfirmDifferent={onConfirm}
        onClose={vi.fn()}
      />,
    );
    const dialog = screen.getByRole("dialog");
    expect(
      within(dialog).getByText("duplicate_patient_title"),
    ).toBeInTheDocument();
    await userEvent.click(within(dialog).getByRole("button", { name: "open" }));
    expect(onOpen).toHaveBeenCalledWith("p1");
    await userEvent.click(
      within(dialog).getByRole("button", {
        name: /register_as_different_person/,
      }),
    );
    expect(onConfirm).toHaveBeenCalled();
  });
});
