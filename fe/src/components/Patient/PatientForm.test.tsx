import { act, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ReactNode } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import PatientForm from "@/components/Patient/PatientForm";
import { PatientModel } from "@/components/Patient/models";
import {
  EMPTY_PATIENT_FORM,
  draftKey,
} from "@/components/Patient/patientFormUtils";

import { HTTPError, NetworkError } from "@/Utils/request/types";

// The API hooks, routing and page chrome are replaced; the form logic is real.
// Texts are translation keys (no i18n instance in unit tests).
const m = vi.hoisted(() => ({
  create: vi.fn(),
  patch: vi.fn(),
  navigate: vi.fn(),
  existing: undefined as PatientModel | undefined,
}));
vi.mock("@/components/Patient/api", () => ({
  useCreatePatient: () => ({
    mutate: m.create,
    isPending: false,
    failureCount: 0,
  }),
  usePatchPatient: () => ({
    mutate: m.patch,
    isPending: false,
    failureCount: 0,
  }),
  usePatient: (_id: string, o: { enabled: boolean }) => ({
    data: o.enabled ? m.existing : undefined,
    isLoading: false,
  }),
}));
vi.mock("raviger", () => ({ navigate: m.navigate }));
vi.mock("@/components/Common/Page", () => ({
  default: ({ children }: { children: ReactNode }) => <div>{children}</div>,
}));
vi.mock("@/hooks/useAuthUser", () => ({
  default: () => ({ id: 7, username: "ana" }),
  useAuthContext: () => ({ user: { date_format: "dmy" } }),
}));
vi.mock("@/Utils/Notifications", () => ({ Success: vi.fn(), Error: vi.fn() }));

const NEW = draftKey(7);
const saveDraft = (key: string, form: object) =>
  localStorage.setItem(
    key,
    JSON.stringify({
      saved_at: new Date().toISOString(),
      form: { ...EMPTY_PATIENT_FORM, ...form },
    }),
  );

beforeEach(() => {
  localStorage.clear();
  m.existing = undefined;
  vi.clearAllMocks();
});
afterEach(() => localStorage.clear());

describe("PatientForm", () => {
  it("validates before sending and sends the lists", async () => {
    render(<PatientForm />);
    await userEvent.click(
      screen.getByRole("button", { name: "register_patient" }),
    );
    expect(screen.getByText("error__name_min_2")).toBeInTheDocument();
    expect(m.create).not.toHaveBeenCalled();

    await userEvent.type(screen.getByLabelText(/full_name/), "Ana Lopes");
    await userEvent.click(screen.getByRole("button", { name: "add_allergy" }));
    await userEvent.type(
      screen.getByLabelText("allergy_substance 1"),
      "Penicilina",
    );
    await userEvent.click(
      screen.getByRole("button", { name: "add_condition" }),
    );
    await userEvent.type(screen.getByLabelText("condition_name 1"), "Asma");
    await userEvent.type(screen.getByLabelText("since_year 1"), "1890");
    await userEvent.click(
      screen.getByRole("button", { name: "register_patient" }),
    );
    expect(screen.getByText("error__invalid_year")).toBeInTheDocument();

    await userEvent.clear(screen.getByLabelText("since_year 1"));
    await userEvent.type(screen.getByLabelText("since_year 1"), "2010");
    await userEvent.click(
      screen.getByRole("button", { name: "register_patient" }),
    );
    const body = m.create.mock.calls[0][0];
    expect(body).toMatchObject({
      name: "Ana Lopes",
      allergy_list: [{ substance: "Penicilina" }],
      condition_list: [{ name: "Asma", since: 2010 }],
    });

    // Saved: the draft goes and the patient page opens
    act(() => m.create.mock.calls[0][1].onSuccess({ id: "p9" }));
    expect(localStorage.getItem(NEW)).toBeNull();
    expect(m.navigate).toHaveBeenCalledWith("/patients/p9");
  });

  it("saves a draft while typing, restores it and can discard it", async () => {
    const { unmount } = render(<PatientForm />);
    await userEvent.type(screen.getByLabelText(/full_name/), "Rui");
    await waitFor(() => expect(localStorage.getItem(NEW)).toContain("Rui"), {
      timeout: 2000,
    });
    unmount();

    render(<PatientForm />);
    expect(screen.getByLabelText(/full_name/)).toHaveValue("Rui");
    expect(screen.getByRole("status")).toHaveTextContent("draft_restored");
    await userEvent.click(
      screen.getByRole("button", { name: "discard_draft" }),
    );
    expect(screen.getByLabelText(/full_name/)).toHaveValue("");
    expect(localStorage.getItem(NEW)).toBeNull();
  });

  it("offers to try again when the server cannot be reached", async () => {
    saveDraft(NEW, { name: "Ana Lopes" });
    render(<PatientForm />);
    await userEvent.click(
      screen.getByRole("button", { name: "register_patient" }),
    );
    act(() => m.create.mock.calls[0][1].onError(new NetworkError(true)));

    expect(screen.getByRole("alert")).toHaveTextContent("save_unreachable");
    expect(localStorage.getItem(NEW)).toContain("Ana Lopes"); // nothing lost
    await userEvent.click(screen.getByRole("button", { name: "try_again" }));
    expect(m.create).toHaveBeenCalledTimes(2);
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("shows server field errors and duplicates", async () => {
    saveDraft(NEW, { name: "Ana Lopes" });
    render(<PatientForm />);
    await userEvent.click(
      screen.getByRole("button", { name: "register_patient" }),
    );
    act(() =>
      m.create.mock.calls[0][1].onError(
        new HTTPError({
          message: "x",
          status: 400,
          silent: true,
          cause: { details: { nif: "already used" } },
        }),
      ),
    );
    expect(screen.getByText("already used")).toBeInTheDocument();
  });

  it("edits a patient, preferring an unsaved draft of it", async () => {
    m.existing = {
      id: "p1",
      name: "Ana",
      city: "Porto",
      allergy_list: [{ substance: "Látex" }],
    };
    render(<PatientForm patientId="p1" />);
    expect(screen.getByLabelText("city")).toHaveValue("Porto");
    expect(screen.getByLabelText("allergy_substance 1")).toHaveValue("Látex");
    expect(screen.queryByRole("status")).not.toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "save_changes" }));
    expect(m.patch.mock.calls[0][0]).toMatchObject({
      id: "p1",
      data: { city: "Porto" },
    });
  });

  it("restores a draft of an edit and cancelling throws it away", async () => {
    m.existing = { id: "p1", name: "Ana", city: "Porto" };
    saveDraft(draftKey(7, "p1"), { name: "Ana", city: "Braga" });
    render(<PatientForm patientId="p1" />);
    expect(screen.getByLabelText("city")).toHaveValue("Braga");
    await userEvent.click(screen.getByRole("button", { name: "cancel" }));
    expect(localStorage.getItem(draftKey(7, "p1"))).toBeNull();
    expect(m.navigate).toHaveBeenCalledWith("/patients/p1");
  });
});
