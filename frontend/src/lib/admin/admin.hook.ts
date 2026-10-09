import { keepPreviousData, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  fetchUser,
  fetchUserMemberships,
  fetchUsers,
  updateUser,
  fetchAuditLogs,
  fetchEligibleMembershipsForUser,
  addOfflineMembership,
  fetchAdminMembershipTierOptions,
  fetchMembershipInvitations,
  createMembershipInvitation,
  deleteMembershipInvitation,
} from "./admin.api";
import type {
  AddOfflineMembershipRequest,
  AdminUserFilters,
  AdminPagination,
  AppliedSearch,
  UpdateUserRequest,
  CreateMembershipInvitationRequest,
} from "@/lib/types/admin.types";
import type { User } from "@/lib/types/user.types";

export function useUsers(
  appliedSearch: AppliedSearch,
  filters: AdminUserFilters,
  pagination: AdminPagination,
  options?: { enabled?: boolean },
) {
  return useQuery({
    queryKey: ["admin", "users", { appliedSearch, filters, pagination }],
    queryFn: ({ signal }) => fetchUsers(appliedSearch, filters, pagination, signal),
    placeholderData: keepPreviousData,
    enabled: options?.enabled ?? true,
  });
}

export function useAdminMembershipTierOptions(options?: { enabled?: boolean }) {
  return useQuery({
    queryKey: ["admin", "membership-tier-options"],
    queryFn: ({ signal }) => fetchAdminMembershipTierOptions(signal),
    enabled: options?.enabled ?? true,
  });
}

export function useMembershipInvitations(options?: { enabled?: boolean }) {
  return useQuery({
    queryKey: ["admin", "membership-invitations"],
    queryFn: ({ signal }) => fetchMembershipInvitations(signal),
    enabled: options?.enabled ?? true,
  });
}

export function useCreateMembershipInvitation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (body: CreateMembershipInvitationRequest) => createMembershipInvitation(body),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["admin", "membership-invitations"],
      });
    },
  });
}

export function useDeleteMembershipInvitation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => deleteMembershipInvitation(id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["admin", "membership-invitations"],
      });
    },
  });
}

export function useUser(userId: string, options?: { enabled?: boolean }) {
  return useQuery({
    queryKey: ["admin", "user", userId],
    queryFn: ({ signal }) => fetchUser(userId, signal),
    enabled: (options?.enabled ?? true) && Boolean(userId),
    // A 404 for an unknown user is a final answer, not a transient failure.
    retry: false,
  });
}

export function useUserMemberships(userId: string, options?: { enabled?: boolean }) {
  return useQuery({
    queryKey: ["admin", "user", userId, "memberships"],
    queryFn: ({ signal }) => fetchUserMemberships(userId, signal),
    enabled: (options?.enabled ?? true) && Boolean(userId),
    retry: false,
  });
}

export function useAdminEligibleMemberships(userId: string, options?: { enabled?: boolean }) {
  return useQuery({
    queryKey: ["admin", "user", userId, "eligible-memberships"],
    queryFn: ({ signal }) => fetchEligibleMembershipsForUser(userId, signal),
    enabled: (options?.enabled ?? true) && Boolean(userId),
    retry: false,
  });
}

export function useAddOfflineMembership(userId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (body: AddOfflineMembershipRequest) => addOfflineMembership(userId, body),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: ["admin", "user", userId, "memberships"],
        }),
        queryClient.invalidateQueries({
          queryKey: ["admin", "user", userId, "eligible-memberships"],
        }),
        queryClient.invalidateQueries({ queryKey: ["admin", "users"] }),
      ]);
    },
  });
}

export function useUpdateUser(userId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (body: UpdateUserRequest) => updateUser(userId, body),
    onSuccess: (user: User, body: UpdateUserRequest) => {
      // The response carries the full updated profile, so the detail view can
      // refresh without a second round trip.
      queryClient.setQueryData(["admin", "user", userId], user);
      queryClient.invalidateQueries({ queryKey: ["admin", "users"] });

      if (body.cancel_membership_id) {
        queryClient.invalidateQueries({ queryKey: ["admin", "user", userId, "memberships"] });
      }
    },
  });
}

export function useAuditLogs(
  actorName?: string,
  pagination?: AdminPagination,
  options?: { enabled?: boolean },
) {
  return useQuery({
    queryKey: ["admin", "audit-logs", { actorName, pagination }],
    queryFn: ({ signal }) => fetchAuditLogs(actorName, pagination, signal),
    placeholderData: keepPreviousData,
    enabled: options?.enabled ?? true,
  });
}
