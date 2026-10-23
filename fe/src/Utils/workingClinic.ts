/**
 * The clinic an admin works on (admins belong to no clinic): the time board
 * and the home page show that clinic. Kept on this browser; sent as the
 * X-Clinic header, which the server only honours for admins.
 */
const KEY = "klyniva:admin-clinic";

export function getWorkingClinic(): string {
  try {
    return localStorage.getItem(KEY) ?? "";
  } catch {
    return "";
  }
}

export function setWorkingClinic(id: string) {
  try {
    if (id) localStorage.setItem(KEY, id);
    else localStorage.removeItem(KEY);
  } catch {
    // storage blocked: the server's default clinic is used
  }
}

export const workingClinicHeaders = (id = getWorkingClinic()) =>
  id ? { "X-Clinic": id } : ({} as Record<string, string>);
