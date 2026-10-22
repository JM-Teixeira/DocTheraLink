/**
 * "View as": an admin looks at the app as a clinic role (read-only on the
 * server, be/internal/auth/viewas.go). The choice lives in this browser tab
 * only (sessionStorage) and is sent with every API call.
 */
export interface ViewAsChoice {
  role: string;
  clinic: string;
  clinicName: string;
}

const KEY = "klyniva:view-as";

export function getViewAs(): ViewAsChoice | null {
  try {
    const v = JSON.parse(sessionStorage.getItem(KEY) ?? "null");
    return v && typeof v.role === "string" && typeof v.clinic === "string"
      ? v
      : null;
  } catch {
    return null;
  }
}

export function setViewAs(v: ViewAsChoice | null) {
  try {
    if (v) sessionStorage.setItem(KEY, JSON.stringify(v));
    else sessionStorage.removeItem(KEY);
  } catch {
    // storage blocked: view-as simply stays off
  }
}

/** Headers for the API while viewing as someone ({} otherwise). */
export function viewAsHeaders(v = getViewAs()): Record<string, string> {
  return v ? { "X-View-As-Role": v.role, "X-View-As-Clinic": v.clinic } : {};
}
