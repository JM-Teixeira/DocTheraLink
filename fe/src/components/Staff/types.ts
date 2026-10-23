/** Staff API types (be/internal/modules/staff). */

export type EmploymentStatus = "active" | "on_leave" | "left";

export const EMPLOYMENT_STATUSES: EmploymentStatus[] = [
  "active",
  "on_leave",
  "left",
];

export interface Skill {
  id: string;
  name: string;
  description?: string | null;
}

/** A skill linked to a user (id is the link, not the skill). */
export interface UserSkill {
  id: string;
  skill_object: Skill;
}

/** One row of the users list. */
export interface StaffRow {
  id: number;
  username: string;
  first_name: string;
  last_name: string;
  email: string;
  user_type: string;
  last_login?: string | null;
  employment_status?: EmploymentStatus;
  employment_since?: string | null;
  clinic_name?: string;
}

export interface EmploymentRequest {
  status: EmploymentStatus;
  since?: string;
  note?: string;
}

export interface Clinic {
  id: string;
  name: string;
}
