import { useMutation, useQuery, useQueryClient, type QueryClient } from "@tanstack/react-query";
import {
  assignRoleToUser,
  createRole,
  deleteRole,
  fetchRole,
  fetchRoles,
  fetchUserRoles,
  unassignRoleFromUser,
  updateRole,
} from "./admin.roles.api";
import type { CreateRoleRequest, RoleDetails, UpdateRoleRequest } from "@/lib/types/role.types";

const ROLES_KEY = ["admin", "roles"] as const;
const roleKey = (roleId: string) => [...ROLES_KEY, roleId] as const;
const userRolesKey = (userId: string) => ["admin", "user", userId, "roles"] as const;

// Refreshes every user's role list, e.g. after a role is renamed, recolored or deleted.
const invalidateAllUserRoles = (queryClient: QueryClient) =>
  queryClient.invalidateQueries({
    predicate: (query) =>
      query.queryKey[0] === "admin" &&
      query.queryKey[1] === "user" &&
      query.queryKey[3] === "roles",
  });

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

export function useUserRoles(userId: string, options?: { enabled?: boolean }) {
  return useQuery({
    queryKey: userRolesKey(userId),
    queryFn: ({ signal }) => fetchUserRoles(userId, signal),
    enabled: options?.enabled !== false && !!userId,
    retry: false,
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
        invalidateAllUserRoles(queryClient),
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
        invalidateAllUserRoles(queryClient),
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
    onSuccess: async (_data, { roleId, userId }) => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: roleKey(roleId) }),
        queryClient.invalidateQueries({ queryKey: userRolesKey(userId) }),
        queryClient.invalidateQueries({ queryKey: ["auth"] }),
      ]);
    },
  });
}

export function useUnassignRole() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ roleId, userId }: RoleAssignment) => unassignRoleFromUser(roleId, userId),
    onSuccess: async (_data, { roleId, userId }) => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: roleKey(roleId) }),
        queryClient.invalidateQueries({ queryKey: userRolesKey(userId) }),
        queryClient.invalidateQueries({ queryKey: ["auth"] }),
      ]);
    },
  });
}
