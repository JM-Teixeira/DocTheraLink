import CreateReferral from "@/components/Referral/CreateReferral";
import ReferralDetail from "@/components/Referral/ReferralDetail";
import ReferralLetter from "@/components/Referral/ReferralLetter";
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
  "/referrals/:id/letter": ({ id }) => <ReferralLetter id={id} />,
};

export default ReferralRoutes;
