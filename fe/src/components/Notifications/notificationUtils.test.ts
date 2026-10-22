import { describe, expect, it } from "vitest";

import {
  badgeLabel,
  notificationLink,
  notificationText,
} from "@/components/Notifications/notificationUtils";
import { AppNotification } from "@/components/Notifications/types";

const base: AppNotification = {
  id: "n1",
  event: "referral_accepted",
  referral_id: "r1",
  actor_name: "Sarah Johnson",
  patient_name: "Maria Silva",
  excerpt: "",
  read_at: null,
  created_date: "2026-10-21T09:00:00Z",
};

describe("notification helpers", () => {
  it("builds the translated line from the event and snapshot", () => {
    expect(notificationText(base)).toEqual({
      key: "notification__referral_accepted",
      params: { actor: "Sarah Johnson", patient: "Maria Silva" },
    });
    expect(
      notificationText({ ...base, actor_name: "", patient_name: "" }).params,
    ).toEqual({
      actor: "-",
      patient: "-",
    });
  });

  it("links to the referral when there is one", () => {
    expect(notificationLink(base)).toBe("/referrals/r1");
    expect(notificationLink({ ...base, referral_id: null })).toBeNull();
  });

  it("formats the badge", () => {
    expect(badgeLabel(0)).toBeNull();
    expect(badgeLabel(undefined)).toBeNull();
    expect(badgeLabel(3)).toBe("3");
    expect(badgeLabel(12)).toBe("9+");
  });
});
