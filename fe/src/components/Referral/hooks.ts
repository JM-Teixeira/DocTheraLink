import { useQuery } from "@tanstack/react-query";

import ReferralRoutes from "@/components/Referral/api";

import query from "@/Utils/request/query";

/**
 * Physiotherapists a referral can be addressed or handed over to. The backend
 * rejects therapists of another clinic, so this list is a convenience only.
 */
export function useTherapists() {
  return useQuery({
    queryKey: ["physiotherapists"],
    queryFn: query(ReferralRoutes.listTherapists, {
      queryParams: { user_type: "PhysioTherapist", limit: 100 },
    }),
    staleTime: 5 * 60_000,
  });
}
