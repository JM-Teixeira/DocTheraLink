import { useState } from "react";
import { useTranslation } from "react-i18next";

import CareIcon from "@/CAREUI/icons/CareIcon";

import ButtonV2 from "@/components/Common/ButtonV2";

import * as Notification from "@/Utils/Notifications";
import routes from "@/Utils/request/api";
import request from "@/Utils/request/request";
import useTanStackQueryInstead from "@/Utils/request/useQuery";

export default function ClockInOutButton() {
  const { t } = useTranslation();
  const [isLoading, setIsLoading] = useState(false);

  const { data: statusData, refetch } = useTanStackQueryInstead(
    routes.timeLog.getStatus,
    {},
  );

  const isClockedIn = statusData?.clocked_in || false;
  const currentLog = statusData?.time_log;

  const handleClockIn = async () => {
    setIsLoading(true);
    const { res } = await request(routes.timeLog.clockIn, {
      body: { notes: "" },
    });

    if (res?.ok) {
      Notification.Success({ msg: t("clocked_in_successfully") });
      refetch();
    } else {
      Notification.Error({ msg: t("failed_to_clock_in") });
    }
    setIsLoading(false);
  };

  const handleClockOut = async () => {
    setIsLoading(true);
    const { res } = await request(routes.timeLog.clockOut, {
      body: { notes: "" },
    });

    if (res?.ok) {
      Notification.Success({ msg: t("clocked_out_successfully") });
      refetch();
    } else {
      Notification.Error({ msg: t("failed_to_clock_out") });
    }
    setIsLoading(false);
  };

  const formatTime = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
  };

  return (
    <div className="flex items-center gap-3">
      {isClockedIn && currentLog && (
        <div className="flex items-center gap-2 text-sm text-gray-700">
          <CareIcon icon="l-clock" className="text-lg text-green-600" />
          <span>
            {t("clocked_in_at")} {formatTime(currentLog.clock_in)}
          </span>
        </div>
      )}

      <ButtonV2
        onClick={isClockedIn ? handleClockOut : handleClockIn}
        disabled={isLoading}
        variant={isClockedIn ? "danger" : "primary"}
        className="min-w-[120px]"
      >
        <CareIcon
          icon={isClockedIn ? "l-sign-out-alt" : "l-sign-in-alt"}
          className="mr-2"
        />
        {isLoading
          ? t("loading")
          : isClockedIn
            ? t("clock_out")
            : t("clock_in")}
      </ButtonV2>
    </div>
  );
}
