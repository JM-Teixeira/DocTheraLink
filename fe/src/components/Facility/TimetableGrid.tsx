import { useState } from "react";
import { useTranslation } from "react-i18next";

import Card from "@/CAREUI/display/Card";
import CareIcon from "@/CAREUI/icons/CareIcon";

import ButtonV2 from "@/components/Common/ButtonV2";
import Loading from "@/components/Common/Loading";
import Page from "@/components/Common/Page";

import routes from "@/Utils/request/api";
import useTanStackQueryInstead from "@/Utils/request/useQuery";

interface TimetableGridProps {
  facilityId: string;
}

const DAYS_OF_WEEK = [
  "Sunday",
  "Monday",
  "Tuesday",
  "Wednesday",
  "Thursday",
  "Friday",
  "Saturday",
];

const TIME_SLOTS = [
  "08:00",
  "09:00",
  "10:00",
  "11:00",
  "12:00",
  "13:00",
  "14:00",
  "15:00",
  "16:00",
  "17:00",
  "18:00",
  "19:00",
  "20:00",
];

export default function TimetableGrid({ facilityId }: TimetableGridProps) {
  const { t } = useTranslation();
  const [selectedDay, setSelectedDay] = useState(new Date().getDay());

  const { data: sections, loading: sectionsLoading } =
    useTanStackQueryInstead(routes.timetable.listSections, {
      pathParams: { id: facilityId },
    });

  const { data: slots, loading: slotsLoading } = useTanStackQueryInstead(
    routes.timetable.getTimetable,
    {
      pathParams: { id: facilityId },
      query: { day_of_week: selectedDay },
    },
  );

  const isLoading = sectionsLoading || slotsLoading;

  const getSlotsForSectionAndTime = (sectionId: string, time: string) => {
    if (!slots) return [];

    return slots.filter((slot: any) => {
      const slotStart = slot.start_time;
      const slotEnd = slot.end_time;
      return (
        slot.section_id === sectionId && time >= slotStart && time < slotEnd
      );
    });
  };

  if (isLoading) {
    return <Loading />;
  }

  return (
    <Page title={t("timetable")} hideBack>
      <Card className="mt-4">
        {/* Day selector */}
        <div className="mb-6 flex gap-2 overflow-x-auto">
          {DAYS_OF_WEEK.map((day, index) => (
            <ButtonV2
              key={day}
              variant={selectedDay === index ? "primary" : "secondary"}
              onClick={() => setSelectedDay(index)}
              className="whitespace-nowrap"
            >
              {t(day.toLowerCase())}
            </ButtonV2>
          ))}
        </div>

        {/* Timetable grid */}
        <div className="overflow-x-auto">
          <table className="min-w-full border-collapse border border-gray-300">
            <thead>
              <tr className="bg-gray-100">
                <th className="border border-gray-300 px-4 py-2 text-left font-semibold">
                  {t("time")}
                </th>
                {sections?.map((section: any) => (
                  <th
                    key={section.id}
                    className="border border-gray-300 px-4 py-2 text-left font-semibold"
                  >
                    <div className="flex items-center justify-between">
                      <span>{section.name}</span>
                      <span className="text-xs text-gray-500">
                        ({t("max")}: {section.max_capacity})
                      </span>
                    </div>
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {TIME_SLOTS.map((time) => (
                <tr key={time} className="hover:bg-gray-50">
                  <td className="border border-gray-300 px-4 py-2 font-medium">
                    {time}
                  </td>
                  {sections?.map((section: any) => {
                    const sectionSlots = getSlotsForSectionAndTime(
                      section.id,
                      time,
                    );
                    const isOverCapacity =
                      sectionSlots.length > section.max_capacity;

                    return (
                      <td
                        key={section.id}
                        className={`border border-gray-300 px-2 py-2 ${isOverCapacity ? "bg-red-50" : ""
                          }`}
                      >
                        <div className="flex flex-wrap gap-1">
                          {sectionSlots.map((slot: any) => (
                            <div
                              key={slot.id}
                              className="flex items-center gap-1 rounded bg-blue-100 px-2 py-1 text-xs"
                            >
                              <CareIcon
                                icon="l-user"
                                className="text-blue-600"
                              />
                              <span className="font-medium">
                                {slot.user?.first_name} {slot.user?.last_name}
                              </span>
                            </div>
                          ))}
                          {isOverCapacity && (
                            <div className="flex items-center gap-1 text-xs text-red-600">
                              <CareIcon icon="l-exclamation-triangle" />
                              <span>{t("over_capacity")}</span>
                            </div>
                          )}
                        </div>
                      </td>
                    );
                  })}
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {/* Legend */}
        <div className="mt-4 flex items-center gap-4 text-sm text-gray-600">
          <div className="flex items-center gap-2">
            <div className="h-4 w-4 rounded bg-blue-100"></div>
            <span>{t("assigned_staff")}</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="h-4 w-4 rounded bg-red-50"></div>
            <span>{t("over_capacity")}</span>
          </div>
        </div>
      </Card>
    </Page>
  );
}
