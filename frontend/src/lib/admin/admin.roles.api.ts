import apiClient from "../client";
import type {
  AssignRoleRequest,
  CreateRoleRequest,
  Role,
  RoleDetails,
  UpdateRoleRequest,
} from "@/lib/types/role.types";

export async function fetchRoles(signal?: AbortSignal): Promise<Role[]> {
  const response = await apiClient.get<Role[]>("/admin/roles", { signal });

  return response.data ?? [];
}

export async function fetchRole(roleId: string, signal?: AbortSignal): Promise<RoleDetails> {
  const response = await apiClient.get<RoleDetails>(`/admin/roles/${roleId}`, { signal });

  return withDefaults(response.data);
}

export async function createRole(body: CreateRoleRequest): Promise<RoleDetails> {
  const response = await apiClient.post<RoleDetails>("/admin/roles", body);

  return withDefaults(response.data);
}

export async function updateRole(roleId: string, body: UpdateRoleRequest): Promise<RoleDetails> {
  const response = await apiClient.patch<RoleDetails>(`/admin/roles/${roleId}`, body);

  return withDefaults(response.data);
}

export async function deleteRole(roleId: string): Promise<void> {
  await apiClient.delete(`/admin/roles/${roleId}`);
}

export async function assignRoleToUser(roleId: string, body: AssignRoleRequest): Promise<void> {
  await apiClient.post(`/admin/roles/${roleId}/users`, body);
}

export async function unassignRoleFromUser(roleId: string, userId: string): Promise<void> {
  await apiClient.delete(`/admin/roles/${roleId}/users/${userId}`);
}

// Go encodes empty slices as null, so normalize them to arrays.
function withDefaults(role: RoleDetails): RoleDetails {
  return {
    ...role,
    permission_keys: role.permission_keys ?? [],
    users_with_role: role.users_with_role ?? [],
  };
}
