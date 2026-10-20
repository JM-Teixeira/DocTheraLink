import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import ReferralCard from "@/components/Referral/ReferralCard";
import { Referral } from "@/components/Referral/types";

// react-i18next without an initialised instance returns the key, so
// assertions check translation keys rather than English text.

const base: Referral = {
  id: "r-42",
  patient_id: "p1",
  doctor_id: 1,
  therapist_id: null,
  diagnosis: "Rotator cuff tear",
  reason: "Restore shoulder mobility",
  priority: "routine",
  status: "pending",
  patient: { id: "p1", name: "Maria Silva", date_of_birth: "1970-01-01" },
  doctor: {
    id: 1,
    username: "doc",
    first_name: "John",
    last_name: "Smith",
    user_type: "Doctor",
  },
  created_date: new Date().toISOString(),
  modified_date: new Date().toISOString(),
};

describe("ReferralCard", () => {
  it("shows patient, diagnosis, status and links to the detail page", () => {
    render(<ReferralCard referral={base} />);

    expect(screen.getByText("Maria Silva")).toBeInTheDocument();
    expect(screen.getByText("MS")).toBeInTheDocument(); // avatar initials
    expect(screen.getByText("Rotator cuff tear")).toBeInTheDocument();
    expect(screen.getByText("referral_status__pending")).toBeInTheDocument();
    expect(screen.getByText("referral_unassigned")).toBeInTheDocument();
    expect(screen.getByRole("link")).toHaveAttribute("href", "/referrals/r-42");
  });

  it("flags urgent referrals and names the assigned therapist", () => {
    render(
      <ReferralCard
        referral={{
          ...base,
          priority: "urgent",
          therapist_id: 10,
          therapist: {
            id: 10,
            username: "sj",
            first_name: "Sarah",
            last_name: "Johnson",
            user_type: "PhysioTherapist",
          },
        }}
      />,
    );

    expect(screen.getByText("referral_priority__urgent")).toBeInTheDocument();
    expect(screen.getByText("Sarah Johnson")).toBeInTheDocument();
    expect(screen.queryByText("referral_unassigned")).not.toBeInTheDocument();
  });
});
