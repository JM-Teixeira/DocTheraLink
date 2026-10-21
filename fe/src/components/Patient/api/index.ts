import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { callApi } from "@/Utils/request/query";
import query from "@/Utils/request/query";
import { QueryParams } from "@/Utils/request/types";

import { PatientModel } from "../models";
import PatientRoutes from "./routes";

// Query Keys
export const patientKeys = {
  all: ["patients"] as const,
  lists: () => [...patientKeys.all, "list"] as const,
  list: (filters: QueryParams) => [...patientKeys.lists(), filters] as const,
  details: () => [...patientKeys.all, "detail"] as const,
  detail: (id: string) => [...patientKeys.details(), id] as const,
  search: (filters: QueryParams) =>
    [...patientKeys.all, "search", filters] as const,
  stats: (filters: QueryParams) =>
    [...patientKeys.all, "stats", filters] as const,
};

export function useCreatePatient() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (body: Partial<PatientModel>) =>
      // silent: the form shows field errors and the duplicate dialog itself
      callApi(PatientRoutes.createPatient, { body, silent: true }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: patientKeys.lists() });
    },
    retry: (failureCount, error) => {
      if (error && "status" in error && typeof error.status === "number") {
        const status = error.status;
        if (status >= 400 && status < 500) {
          return false;
        }
      }
      return failureCount < 2;
    },
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
  });
}

export function useSearchPatients(
  searchParams: QueryParams,
  options?: { enabled?: boolean },
) {
  return useQuery({
    queryKey: patientKeys.search(searchParams),
    queryFn: query(PatientRoutes.searchPatient, {
      queryParams: searchParams,
    }),
    enabled: options?.enabled ?? true,
    staleTime: 30000,
    retry: (failureCount, error) => {
      if (error && "status" in error && typeof error.status === "number") {
        const status = error.status;
        if (status >= 400 && status < 500) {
          return false;
        }
      }
      return failureCount < 2;
    },
  });
}

export function usePatientList(
  filters: QueryParams = {},
  options?: { enabled?: boolean },
) {
  return useQuery({
    queryKey: patientKeys.list(filters),
    queryFn: query(PatientRoutes.listPatients, {
      queryParams: filters,
    }),
    enabled: options?.enabled ?? true,
    staleTime: 60000,
    retry: 2,
  });
}

/** Metrics for the patients matching the list filters. */
export function usePatientStats(filters: QueryParams = {}) {
  return useQuery({
    queryKey: patientKeys.stats(filters),
    queryFn: query(PatientRoutes.patientStats, { queryParams: filters }),
    staleTime: 60000,
  });
}

export function usePatient(patientId: string, options?: { enabled?: boolean }) {
  return useQuery({
    queryKey: patientKeys.detail(patientId),
    queryFn: query(PatientRoutes.getPatient, {
      pathParams: { id: patientId },
    }),
    enabled: options?.enabled ?? !!patientId,
    staleTime: 60000,
    retry: 2,
  });
}

export function useUpdatePatient() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<PatientModel> }) =>
      callApi(PatientRoutes.updatePatient, { pathParams: { id }, body: data }),
    onSuccess: (data, variables) => {
      queryClient.invalidateQueries({
        queryKey: patientKeys.detail(variables.id),
      });
      queryClient.invalidateQueries({ queryKey: patientKeys.lists() });
    },
    retry: (failureCount, error) => {
      if (error && "status" in error && typeof error.status === "number") {
        const status = error.status;
        if (status >= 400 && status < 500) {
          return false;
        }
      }
      return failureCount < 2;
    },
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
  });
}

export function usePatchPatient() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<PatientModel> }) =>
      callApi(PatientRoutes.patchPatient, {
        pathParams: { id },
        body: data,
        silent: true,
      }),
    onSuccess: (data, variables) => {
      queryClient.invalidateQueries({
        queryKey: patientKeys.detail(variables.id),
      });
      queryClient.invalidateQueries({ queryKey: patientKeys.lists() });
    },
    retry: (failureCount, error) => {
      if (error && "status" in error && typeof error.status === "number") {
        const status = error.status;
        if (status >= 400 && status < 500) {
          return false;
        }
      }
      return failureCount < 2;
    },
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
  });
}

export function getDuplicateCheckParams(
  patientData: Partial<PatientModel>,
): QueryParams {
  const params: QueryParams = {};

  if (patientData.phone_number) {
    params.phone_number = patientData.phone_number;
  }

  if (patientData.name) {
    params.name = patientData.name;
  }

  if (patientData.date_of_birth) {
    params.date_of_birth = patientData.date_of_birth;
  }

  if (!patientData.date_of_birth && patientData.year_of_birth) {
    params.year_of_birth = patientData.year_of_birth;
  }

  return params;
}

export { PatientRoutes };
