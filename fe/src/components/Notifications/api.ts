import { NotificationPage } from "@/components/Notifications/types";

import { Type } from "@/Utils/request/api";

/** Routes of be/internal/modules/notification (current user only). */
const NotificationRoutes = {
  list: {
    path: "/api/v1/notifications/",
    method: "GET",
    TRes: Type<NotificationPage>(),
  },
  unreadCount: {
    path: "/api/v1/notifications/unread-count/",
    method: "GET",
    TRes: Type<{ unread: number }>(),
  },
  markRead: {
    path: "/api/v1/notifications/{id}/read/",
    method: "POST",
    TBody: Type<Record<string, never>>(),
    TRes: Type<void>(),
  },
  markAllRead: {
    path: "/api/v1/notifications/read-all/",
    method: "POST",
    TBody: Type<Record<string, never>>(),
    TRes: Type<void>(),
  },
} as const;

export default NotificationRoutes;
