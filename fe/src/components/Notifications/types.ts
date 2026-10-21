/** Mirrors be/internal/modules/notification/model.go. */

export const NOTIFICATION_EVENTS = [
  "referral_created",
  "referral_accepted",
  "referral_declined",
  "referral_completed",
  "referral_cancelled",
  "referral_reassigned",
  "referral_reopened",
  "referral_message",
] as const;

export type NotificationEvent = (typeof NOTIFICATION_EVENTS)[number];

export interface AppNotification {
  id: string;
  event: NotificationEvent;
  referral_id?: string | null;
  actor_id?: number | null;
  actor_name: string;
  patient_name: string;
  excerpt: string;
  read_at?: string | null;
  created_date: string;
}

export interface NotificationPage {
  count: number;
  unread: number;
  results: AppNotification[];
}
