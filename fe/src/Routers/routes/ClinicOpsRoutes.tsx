import InventoryPage from "@/components/Inventory/InventoryPage";
import TimeBoardPage from "@/components/TimeBoard/TimeBoardPage";

import { AppRoutes } from "@/Routers/AppRouter";

/** Running the clinic day: who works where (time board) and stock. */
const ClinicOpsRoutes: AppRoutes = {
  "/time-board": () => <TimeBoardPage />,
  "/inventory": () => <InventoryPage />,
};

export default ClinicOpsRoutes;
