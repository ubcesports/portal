import { useMutation, useQueries, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  assignRoleToUser,
  createRole,
  deleteRole,
  fetchRole,
  fetchRoles,
  unassignRoleFromUser,
  updateRole,
} from "./admin.roles.api";
import type { CreateRoleRequest, RoleDetails, UpdateRoleRequest } from "@/lib/types/role.types";

const ROLES_KEY = ["admin", "roles"] as const;
const roleKey = (roleId: string) => [...ROLES_KEY, roleId] as const;

export function useRoles(options?: { enabled?: boolean }) {
  return useQuery({
    queryKey: ROLES_KEY,
    queryFn: ({ signal }) => fetchRoles(signal),
    enabled: options?.enabled ?? true,
  });
}

export function useRole(roleId: string, options?: { enabled?: boolean }) {
  return useQuery({
    queryKey: roleKey(roleId),
    queryFn: ({ signal }) => fetchRole(roleId, signal),
    enabled: options?.enabled !== false && !!roleId,
    // A 404 for an unknown role is a final answer, not a transient failure.
    retry: false,
  });
}

/*
  Loads every role's details, which include the users holding it. There is no
  endpoint for one user's roles, so this is how the user page finds them.
*/
export function useAllRoleDetails(roleIds: string[], options?: { enabled?: boolean }) {
  return useQueries({
    queries: roleIds.map((roleId) => ({
      queryKey: roleKey(roleId),
      queryFn: ({ signal }: { signal: AbortSignal }) => fetchRole(roleId, signal),
      enabled: options?.enabled ?? true,
      retry: false,
    })),
  });
}

export function useCreateRole() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (body: CreateRoleRequest) => createRole(body),
    onSuccess: async (role: RoleDetails) => {
      queryClient.setQueryData(roleKey(role.id), role);
      await queryClient.invalidateQueries({ queryKey: ROLES_KEY, exact: true });
    },
  });
}

export function useUpdateRole(roleId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (body: UpdateRoleRequest) => updateRole(roleId, body),
    onSuccess: async (role: RoleDetails) => {
      queryClient.setQueryData(roleKey(roleId), role);
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ROLES_KEY, exact: true }),
        // The current user's permissions may have changed.
        queryClient.invalidateQueries({ queryKey: ["auth"] }),
      ]);
    },
  });
}

export function useDeleteRole() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (roleId: string) => deleteRole(roleId),
    onSuccess: async (_data, roleId) => {
      await Promise.all([
        // Mark the deleted role stale without refetching it. Removing or refetching
        // it while its page is still mounted would request it again and show a 404 toast.
        queryClient.invalidateQueries({ queryKey: roleKey(roleId), refetchType: "none" }),
        queryClient.invalidateQueries({ queryKey: ROLES_KEY, exact: true }),
        queryClient.invalidateQueries({ queryKey: ["auth"] }),
      ]);
    },
  });
}

export type RoleAssignment = {
  roleId: string;
  userId: string;
};

export function useAssignRole() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ roleId, userId }: RoleAssignment) =>
      assignRoleToUser(roleId, { user_id: userId }),
    onSuccess: async (_data, { roleId }) => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: roleKey(roleId) }),
        queryClient.invalidateQueries({ queryKey: ["auth"] }),
      ]);
    },
  });
}

export function useUnassignRole() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ roleId, userId }: RoleAssignment) => unassignRoleFromUser(roleId, userId),
    onSuccess: async (_data, { roleId }) => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: roleKey(roleId) }),
        queryClient.invalidateQueries({ queryKey: ["auth"] }),
      ]);
    },
  });
}
