import { useState } from "react";
import { useTranslation } from "react-i18next";

import Card from "@/CAREUI/display/Card";
import CareIcon from "@/CAREUI/icons/CareIcon";

import Loading from "@/components/Common/Loading";
import Page from "@/components/Common/Page";
import Pagination from "@/components/Common/Pagination";

import routes from "@/Utils/request/api";
import useTanStackQueryInstead from "@/Utils/request/useQuery";
import { formatDateTime } from "@/Utils/utils";

interface TimeLogHistoryProps {
  userId?: number;
}

export default function TimeLogHistory({ userId }: TimeLogHistoryProps) {
  const { t } = useTranslation();
  const [currentPage, setCurrentPage] = useState(1);
  const limit = 20;

  const { data, loading } = useTanStackQueryInstead(
    userId ? routes.timeLog.getUserLogs : routes.timeLog.getMyLogs,
    {
      pathParams: userId ? { id: String(userId) } : undefined,
      query: {
        limit,
        offset: (currentPage - 1) * limit,
      },
    },
  );

  const timeLogs = data?.results || [];
  const totalCount = data?.count || 0;

  const formatDurationMinutes = (minutes: number) => {
    const hours = Math.floor(minutes / 60);
    const mins = minutes % 60;
    return `${hours}h ${mins}m`;
  };

  if (loading) {
    return <Loading />;
  }

  return (
    <Page title={t("time_log_history")} hideBack>
      <Card className="mt-4">
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                  {t("date")}
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                  {t("clock_in")}
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                  {t("clock_out")}
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                  {t("duration")}
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                  {t("notes")}
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 bg-white">
              {timeLogs.length === 0 ? (
                <tr>
                  <td colSpan={5} className="px-6 py-4 text-center text-gray-500">
                    {t("no_time_logs_found")}
                  </td>
                </tr>
              ) : (
                timeLogs.map((log: any) => (
                  <tr key={log.id} className="hover:bg-gray-50">
                    <td className="whitespace-nowrap px-6 py-4 text-sm text-gray-900">
                      {formatDateTime(log.date)}
                    </td>
                    <td className="whitespace-nowrap px-6 py-4 text-sm text-gray-900">
                      <div className="flex items-center gap-2">
                        <CareIcon
                          icon="l-sign-in-alt"
                          className="text-green-600"
                        />
                        {new Date(log.clock_in).toLocaleTimeString([], {
                          hour: "2-digit",
                          minute: "2-digit",
                        })}
                      </div>
                    </td>
                    <td className="whitespace-nowrap px-6 py-4 text-sm text-gray-900">
                      {log.clock_out ? (
                        <div className="flex items-center gap-2">
                          <CareIcon
                            icon="l-sign-out-alt"
                            className="text-red-600"
                          />
                          {new Date(log.clock_out).toLocaleTimeString([], {
                            hour: "2-digit",
                            minute: "2-digit",
                          })}
                        </div>
                      ) : (
                        <span className="text-yellow-600">{t("active")}</span>
                      )}
                    </td>
                    <td className="whitespace-nowrap px-6 py-4 text-sm text-gray-900">
                      {log.duration > 0 ? (
                        formatDurationMinutes(log.duration)
                      ) : (
                        <span className="text-gray-400">-</span>
                      )}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-500">
                      {log.notes || "-"}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {totalCount > limit && (
          <div className="mt-4 flex justify-center">
            <Pagination
              data={{ totalCount }}
              cPage={currentPage}
              defaultPerPage={limit}
              onChange={(page) => setCurrentPage(page)}
            />
          </div>
        )}
      </Card>
    </Page>
  );
}
