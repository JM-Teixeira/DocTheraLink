import { Building2, MapPin, Users } from "lucide-react";
import { Link } from "raviger";
import { useTranslation } from "react-i18next";

import { Button } from "@/components/ui/button";

import { Avatar } from "@/components/Common/Avatar";
import { FacilityModel } from "@/components/Facility/models";

import { FACILITY_TYPES } from "@/common/constants";

/** One clinic in the clinic list. */
export const FacilityCard = (props: {
  facility: FacilityModel;
  /** Kept for callers; no role-specific actions remain on the card. */
  userType?: string;
}) => {
  const { facility } = props;
  const { t } = useTranslation();
  const type =
    typeof facility.facility_type === "number"
      ? FACILITY_TYPES.find((ft) => ft.id === facility.facility_type)?.text
      : facility.facility_type;

  return (
    <div className="flex flex-col gap-4 rounded-xl border border-gray-200 bg-white p-5 shadow-sm transition hover:border-primary-300 sm:flex-row sm:items-center">
      <Link href={`/facility/${facility.id}`} className="size-20 shrink-0">
        <Avatar
          name={facility.name || ""}
          imageUrl={facility.read_cover_image_url}
        />
      </Link>
      <div className="min-w-0 flex-1">
        <Link
          href={`/facility/${facility.id}`}
          className="text-lg font-semibold text-gray-900 hover:text-primary-700"
        >
          {facility.name}
        </Link>
        <div className="mt-1 flex flex-wrap gap-x-4 gap-y-1 text-sm text-gray-500">
          {type && (
            <span className="inline-flex items-center gap-1">
              <Building2 className="size-4" aria-hidden />
              {type}
            </span>
          )}
          {(facility.district_object?.name || facility.address) && (
            <span className="inline-flex items-center gap-1">
              <MapPin className="size-4" aria-hidden />
              {facility.district_object?.name ?? facility.address}
            </span>
          )}
        </div>
      </div>
      <div className="flex gap-2">
        <Button variant="outline" asChild>
          <Link href={`/users?clinic=${facility.id}`}>
            <Users className="mr-1.5 size-4" aria-hidden />
            {t("staff")}
          </Link>
        </Button>
        <Button variant="primary" asChild>
          <Link href={`/facility/${facility.id}`}>{t("open_clinic")}</Link>
        </Button>
      </div>
    </div>
  );
};
