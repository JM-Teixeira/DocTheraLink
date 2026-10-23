import { Redirect } from "raviger";

import { FacilityConfigure } from "@/components/Facility/FacilityConfigure";
import { FacilityCreate } from "@/components/Facility/FacilityCreate";
import { FacilityHome } from "@/components/Facility/FacilityHome";
import { FacilityList } from "@/components/Facility/FacilityList";

import { AppRoutes } from "@/Routers/AppRouter";

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
  // The clinic's people are the people list filtered by clinic
  "/facility/:facilityId/users": ({ facilityId }) => (
    <Redirect to={`/users?clinic=${facilityId}`} replace />
  ),
};

export default FacilityRoutes;
