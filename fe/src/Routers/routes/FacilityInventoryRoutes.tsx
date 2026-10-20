import { Redirect } from "raviger";

import { AddInventoryForm } from "@/components/Facility/AddInventoryForm";
import InventoryList from "@/components/Facility/InventoryList";
import InventoryLog from "@/components/Facility/InventoryLog";
import MinQuantityList from "@/components/Facility/MinQuantityList";
import { SetInventoryForm } from "@/components/Facility/SetInventoryForm";

import { AppRoutes } from "@/Routers/AppRouter";

const FacilityInventoryRoutes: AppRoutes = {
  "/facility/:facilityId/inventory": ({ facilityId }) => (
    <InventoryList facilityId={facilityId} />
  ),
  "/facility/:facilityId/inventory/add": ({ facilityId }) => (
    <AddInventoryForm facilityId={facilityId} />
  ),
  "/facility/:facilityId/inventory/min_quantity/set": ({ facilityId }) => (
    <SetInventoryForm facilityId={facilityId} />
  ),
  "/facility/:facilityId/inventory/min_quantity/list": ({ facilityId }) => (
    <MinQuantityList facilityId={facilityId} />
  ),
  "/facility/:facilityId/inventory/min_quantity": ({ facilityId }) => (
    <Redirect to={`/facility/${facilityId}/inventory/min_quantity/list`} />
  ),
  "/facility/:facilityId/inventory/:inventoryId": ({
    facilityId,
    inventoryId,
  }) => <InventoryLog facilityId={facilityId} inventoryId={inventoryId} />,
};

export default FacilityInventoryRoutes;
