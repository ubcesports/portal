import type { GroupType, RoleType, User } from "./user.types";
import type { PaymentMethod } from "./membership.types";

export type SearchMode = "full_name" | "email" | "student_id";

export type IsStudent = "yes" | "no";

export type AppliedSearch = {
  mode: SearchMode;
  value: string;
} | null;

export type AdminUserFilters = {
  role?: RoleType;
  groups?: GroupType[];
  membershipTierIds?: string[];
  isStudent?: boolean;
};

export type AdminActiveMembership = {
  tier_title: string;
};

export type AdminUser = User & {
  active_memberships: AdminActiveMembership[];
};

export type AdminMembershipTierOption = {
  id: string;
  title: string;
  program_name: string;
};

export type AuditLogActor = {
  actor_user_id: string;
  actor_full_name: string;
  actor_avatar_url: string;
};

export type AuditLogOutcome = "success" | "failed" | "denied";

export type AuditLogResponse = {
  logs: AuditLogEntry[];
  total: number;
};

export type AuditLogEntry = {
  actor: AuditLogActor;
  occured_at: string;
  action: string;
  description: string | null;
  outcome: AuditLogOutcome;
  request_id: string;
  target_user: AuditLogActor | null;
};

export type UsersResponse = {
  users: AdminUser[];
  total: number;
};

export type UserResponse = {
  user: User;
};

/*
  Every field is optional. An absent field is left untouched, so only send the
  ones the admin actually changed.
*/
export type UpdateUserRequest = {
  full_name?: string;
  student_id?: string;
  is_student?: boolean;
  groups_add?: GroupType[];
  groups_remove?: GroupType[];
  role?: RoleType;
  cancel_membership_id?: string;
};

export type OfflinePaymentMethod = Exclude<PaymentMethod, "stripe">;

export type AddOfflineMembershipRequest = {
  tier_id: string;
  amount_paid_cents: number;
  payment_method: OfflinePaymentMethod;
};

export type MembershipInvitation = {
  id: string;
  email: string;
  tier_id: string;
  tier_title: string;
  program_name: string;
  amount_paid_cents: number;
  payment_method: OfflinePaymentMethod;
  done: boolean;
  created_by_user_id: string;
  created_by_name: string;
  invitation_sent_at: string | null;
  purchased_at: string;
  created_at: string;
  updated_at: string;
};

export type CreateMembershipInvitationRequest = {
  email: string;
  tier_id: string;
  amount_paid_cents: number;
  payment_method: OfflinePaymentMethod;
};

export type AdminPagination = {
  limit: number;
  offset: number;
};

export const PAGE_SIZE_OPTIONS = [10, 25, 50, 100] as const;

export const DEFAULT_PAGE_SIZE = 25;

export const GROUP_OPTIONS: { value: GroupType; label: string }[] = [
  { value: "member", label: "Member" },
  { value: "competitive_team", label: "Competitive Team" },
  { value: "executive", label: "Executive" },
  { value: "director", label: "Director" },
  { value: "board", label: "Board" },
];

export const ROLE_OPTIONS: { value: RoleType; label: string }[] = [
  { value: "member", label: "Member" },
  { value: "admin", label: "Admin" },
];

export const SEARCH_MODE_OPTIONS: { value: SearchMode; label: string }[] = [
  { value: "full_name", label: "Full name" },
  { value: "email", label: "Email" },
  { value: "student_id", label: "Student ID" },
];

export const IS_STUDENT_OPTIONS: { value: IsStudent; label: string }[] = [
  { value: "yes", label: "Yes" },
  { value: "no", label: "No" },
];

export type IsStudentFilter = "all" | IsStudent;
