import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import dayjs from "dayjs";
import relativeTime from "dayjs/plugin/relativeTime";
import {
  ArrowLeftRight,
  BellOff,
  CheckCheck,
  CircleCheck,
  CircleX,
  MessageSquare,
  RotateCcw,
  Stethoscope,
  ThumbsUp,
} from "lucide-react";
import { navigate } from "raviger";
import { useState } from "react";
import { useTranslation } from "react-i18next";

import { cn } from "@/lib/utils";

import CareIcon from "@/CAREUI/icons/CareIcon";
import SlideOver from "@/CAREUI/misc/interactive/SlideOver";

import { Button } from "@/components/ui/button";

import {
  ShrinkedSidebarItem,
  SidebarItem,
} from "@/components/Common/Sidebar/SidebarItem";
import NotificationRoutes from "@/components/Notifications/api";
import {
  badgeLabel,
  notificationLink,
  notificationText,
} from "@/components/Notifications/notificationUtils";
import {
  AppNotification,
  NotificationEvent,
} from "@/components/Notifications/types";

import { callApi } from "@/Utils/request/query";
import query from "@/Utils/request/query";

dayjs.extend(relativeTime);

/** How often the bell checks for new notifications. */
const POLL_MS = 30_000;

const EVENT_ICON: Record<NotificationEvent, typeof Stethoscope> = {
  referral_created: Stethoscope,
  referral_accepted: ThumbsUp,
  referral_declined: CircleX,
  referral_completed: CircleCheck,
  referral_cancelled: CircleX,
  referral_reassigned: ArrowLeftRight,
  referral_reopened: RotateCcw,
  referral_message: MessageSquare,
};

interface Props {
  shrinked: boolean;
  handleOverflow: (value: boolean) => void;
  onClickCB?: () => void;
}

/** Sidebar bell with unread badge, opening a panel with the user's notifications. */
export default function NotificationsPanel({
  shrinked,
  handleOverflow,
  onClickCB,
}: Props) {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [open, setOpen] = useState(false);
  const [unreadOnly, setUnreadOnly] = useState(false);

  const { data: unread } = useQuery({
    queryKey: ["notifications", "unread"],
    queryFn: query(NotificationRoutes.unreadCount, { silent: true }),
    refetchInterval: POLL_MS,
  });

  const { data: page, isLoading } = useQuery({
    queryKey: ["notifications", "list", unreadOnly],
    queryFn: query(NotificationRoutes.list, {
      queryParams: { unread: unreadOnly || undefined, limit: 50 },
    }),
    enabled: open,
  });

  const refresh = () =>
    queryClient.invalidateQueries({ queryKey: ["notifications"] });

  const markRead = useMutation({
    mutationFn: (id: string) =>
      callApi(NotificationRoutes.markRead, { pathParams: { id }, body: {} }),
    onSuccess: refresh,
  });
  const markAll = useMutation({
    mutationFn: () => callApi(NotificationRoutes.markAllRead, { body: {} }),
    onSuccess: refresh,
  });

  const openNotification = (n: AppNotification) => {
    if (!n.read_at) markRead.mutate(n.id);
    const link = notificationLink(n);
    if (link) {
      setOpen(false);
      onClickCB?.();
      navigate(link);
    }
  };

  const Item = shrinked ? ShrinkedSidebarItem : SidebarItem;
  const unreadCount = unread?.unread ?? 0;

  return (
    <>
      <Item
        text={t("notifications")}
        id="notification-slide-btn"
        do={() => setOpen(!open)}
        icon={<CareIcon icon="d-bell" className="h-5" />}
        badgeCount={unreadCount}
        handleOverflow={handleOverflow}
      />
      <SlideOver
        open={open}
        setOpen={setOpen}
        slideFrom="right"
        title={
          <span className="flex items-center gap-2">
            {t("notifications")}
            {badgeLabel(unreadCount) && (
              <span className="rounded-full bg-primary-100 px-2 py-0.5 text-xs font-semibold text-primary-700">
                {t("n_unread", { count: unreadCount })}
              </span>
            )}
          </span>
        }
        dialogClass="md:w-[420px]"
        onCloseClick={onClickCB}
      >
        <div className="mb-3 flex items-center justify-between gap-2">
          <div role="tablist" className="flex gap-1 rounded-lg bg-gray-100 p-1">
            {[false, true].map((u) => (
              <button
                key={String(u)}
                role="tab"
                aria-selected={unreadOnly === u}
                onClick={() => setUnreadOnly(u)}
                className={cn(
                  "rounded-md px-3 py-1 text-xs font-medium transition",
                  unreadOnly === u
                    ? "bg-white text-gray-900 shadow-sm"
                    : "text-gray-600",
                )}
              >
                {u ? t("unread") : t("all")}
              </button>
            ))}
          </div>
          <Button
            variant="ghost"
            size="sm"
            disabled={!unreadCount || markAll.isPending}
            onClick={() => markAll.mutate()}
          >
            <CheckCheck className="mr-1 size-4" aria-hidden />
            {t("mark_all_as_read")}
          </Button>
        </div>

        {isLoading ? (
          <p className="py-8 text-center text-sm text-gray-500">
            {t("loading")}
          </p>
        ) : page?.results.length ? (
          <ul className="flex flex-col gap-1">
            {page.results.map((n) => (
              <NotificationRow
                key={n.id}
                n={n}
                onOpen={() => openNotification(n)}
              />
            ))}
          </ul>
        ) : (
          <div className="flex flex-col items-center py-12 text-center">
            <BellOff className="mb-3 size-8 text-gray-300" aria-hidden />
            <p className="text-sm text-gray-500">
              {unreadOnly
                ? t("no_unread_notifications")
                : t("no_notifications")}
            </p>
          </div>
        )}
      </SlideOver>
    </>
  );
}

function NotificationRow({
  n,
  onOpen,
}: {
  n: AppNotification;
  onOpen: () => void;
}) {
  const { t } = useTranslation();
  const { key, params } = notificationText(n);
  const Icon = EVENT_ICON[n.event] ?? Stethoscope;
  const unread = !n.read_at;
  return (
    <li>
      <button
        onClick={onOpen}
        className={cn(
          "flex w-full gap-3 rounded-lg p-3 text-left transition hover:bg-gray-50",
          unread && "bg-primary-50/60",
        )}
      >
        <span
          className={cn(
            "flex size-9 shrink-0 items-center justify-center rounded-full",
            unread
              ? "bg-primary-100 text-primary-700"
              : "bg-gray-100 text-gray-500",
          )}
        >
          <Icon className="size-4" aria-hidden />
        </span>
        <span className="min-w-0 flex-1">
          <span
            className={cn(
              "block text-sm",
              unread ? "font-medium text-gray-900" : "text-gray-700",
            )}
          >
            {t(key, params)}
          </span>
          {n.excerpt && (
            <span className="mt-0.5 block truncate text-xs text-gray-500">
              “{n.excerpt}”
            </span>
          )}
          <time
            dateTime={n.created_date}
            className="mt-1 block text-xs text-gray-400"
          >
            {dayjs(n.created_date).fromNow()}
          </time>
        </span>
        {unread && (
          <span
            className="mt-1.5 size-2 shrink-0 rounded-full bg-primary-500"
            aria-label={t("unread")}
          />
        )}
      </button>
    </li>
  );
}
