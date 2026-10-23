/**
 * Pure helpers for rendering notifications (tested). Display text is built
 * here from the event + snapshot so it is translated in the reader's language.
 */
import { AppNotification } from "@/components/Notifications/types";

/** Translation key and parameters for a notification's main line. */
export function notificationText(n: AppNotification) {
  return {
    key: `notification__${n.event}`,
    params: {
      actor: n.actor_name || "-",
      patient: n.patient_name || "-",
    },
  };
}

/** Where clicking the notification goes (null when there is nowhere to go). */
export function notificationLink(n: AppNotification) {
  return n.referral_id ? `/referrals/${n.referral_id}` : null;
}

/** Badge text: hides zero, caps at 9+. */
export function badgeLabel(count?: number) {
  if (!count) return null;
  return count > 9 ? "9+" : String(count);
}
