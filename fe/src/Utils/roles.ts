/**
 * Role helpers shared by all modules. Mirror be/internal/auth/middleware.go and
 * be/internal/models/user.go — the backend is the source of truth; these only
 * decide what the UI shows.
 */

/** The fields of the signed-in user the helpers need. */
export interface RoleUser {
  id: number;
  user_type: string;
}

const ADMIN_ROLES = ["StateAdmin", "DistrictAdmin", "LocalBodyAdmin"];
const READ_ONLY_ROLES = [
  "StaffReadOnly",
  "DistrictReadOnlyAdmin",
  "StateReadOnlyAdmin",
];

export const isAdmin = (u: RoleUser) => ADMIN_ROLES.includes(u.user_type);
export const isDoctor = (u: RoleUser) => u.user_type === "Doctor";
export const isTherapist = (u: RoleUser) => u.user_type === "PhysioTherapist";
export const isReceptionist = (u: RoleUser) => u.user_type === "Receptionist";
export const isReadOnly = (u: RoleUser) =>
  READ_ONLY_ROLES.includes(u.user_type);

/** Anyone except read-only roles may register and edit patients. */
export const canWritePatients = (u: RoleUser) => !isReadOnly(u);

/**
 * Referrals carry clinical content: only doctors, physiotherapists and admins
 * see them (receptionists work with appointments instead).
 */
export const canSeeReferrals = (u: RoleUser) =>
  isAdmin(u) || isDoctor(u) || isTherapist(u);
