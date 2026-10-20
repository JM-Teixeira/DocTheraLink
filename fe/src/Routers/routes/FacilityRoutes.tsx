import { FacilityConfigure } from "@/components/Facility/FacilityConfigure";
import { FacilityCreate } from "@/components/Facility/FacilityCreate";
import { FacilityHome } from "@/components/Facility/FacilityHome";
import { FacilityList } from "@/components/Facility/FacilityList";
import FacilityUsers from "@/components/Facility/FacilityUsers";

import { AppRoutes } from "@/Routers/AppRouter";
import FacilityInventoryRoutes from "@/Routers/routes/FacilityInventoryRoutes";

const FacilityRoutes: AppRoutes = {
  "/facility": () => <FacilityList />,
  "/facility/create": () => <FacilityCreate />,
  "/facility/:facilityId/update": ({ facilityId }) => (
    <FacilityCreate facilityId={facilityId} />
  ),
  "/facility/:facilityId/configure": ({ facilityId }) => (
    <FacilityConfigure facilityId={facilityId} />
  ),
  "/facility/:facilityId": ({ facilityId }) => (
    <FacilityHome facilityId={facilityId} />
  ),
  "/facility/:facilityId/users": ({ facilityId }) => (
    <FacilityUsers facilityId={facilityId} />
  ),
  ...FacilityInventoryRoutes,
};

export default FacilityRoutes;
