import CreateReferral from "@/components/Referral/CreateReferral";
import ReferralDetail from "@/components/Referral/ReferralDetail";
import ReferralList from "@/components/Referral/ReferralList";

import { AppRoutes } from "@/Routers/AppRouter";

/** Doctor → physiotherapist referral workflow. */
const ReferralRoutes: AppRoutes = {
  "/referrals": () => <ReferralList />,
  "/referrals/new": () => (
    <CreateReferral
      patientId={
        new URLSearchParams(window.location.search).get("patient") ?? undefined
      }
    />
  ),
  "/referrals/:id": ({ id }) => <ReferralDetail id={id} />,
};

export default ReferralRoutes;
