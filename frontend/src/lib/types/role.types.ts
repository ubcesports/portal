import type { PermissionKey } from "./permission.types";

export type RoleUser = {
  user_id: string;
  name: string;
  profile_picture: string | null;
};

/* A role as listed by GET /admin/roles. */
export type Role = {
  id: string;
  name: string;
  description: string | null;
  color: string | null;
  created_by: RoleUser | null;
  created_at: string;
  updated_at: string;
};

/* A role as returned by GET, POST and PATCH /admin/roles/{id}. */
export type RoleDetails = Role & {
  permission_keys: PermissionKey[];
  users_with_role: RoleUser[];
};

/* A role held by a user (GET /users/{id}/roles, and roles on GET /profile). */
export type UserRole = {
  id: string;
  name: string;
  color: string | null;
  permission_keys: PermissionKey[];
};

export type CreateRoleRequest = {
  name: string;
  description?: string;
  color?: string;
  permission_keys: PermissionKey[];
};

/*
  Every field is optional and an absent field is left untouched. When present,
  permission_keys replaces the role's whole permission set.
*/
export type UpdateRoleRequest = Partial<CreateRoleRequest>;

export type AssignRoleRequest = {
  user_id: string;
};

export const ROLE_NAME_MAX_LENGTH = 50;
