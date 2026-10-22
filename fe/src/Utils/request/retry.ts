/**
 * When to try a save again. Only failures where the server most likely never
 * did the work are retried: no connection at all, or a gateway / restart
 * answer (502, 503, 504). A 4xx is the user's to fix and a 500 may have
 * half-happened, so neither is repeated.
 */
import { HTTPError, NetworkError } from "@/Utils/request/types";

export const MAX_RETRIES = 2;

export function isTransient(error: unknown) {
  if (error instanceof NetworkError) return true;
  return error instanceof HTTPError && [502, 503, 504].includes(error.status);
}

/** TanStack `retry` option: up to MAX_RETRIES more tries for transient failures. */
export const retryTransient = (failureCount: number, error: Error) =>
  failureCount < MAX_RETRIES && isTransient(error);

/** TanStack `retryDelay`: 1 s, 2 s, 4 s ... capped at 10 s. */
export const retryDelay = (attempt: number) =>
  Math.min(1000 * 2 ** attempt, 10000);
